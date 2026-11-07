package aiclient

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
)

func newSigningServer(t *testing.T, calls *int32, wantURI string) *httptest.Server {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(calls, 1)

		body, _ := io.ReadAll(r.Body)
		timestamp, _ := strconv.ParseInt(r.Header.Get("X-Praxis-Timestamp"), 10, 64)
		want := sign(testSecret, timestamp, r.Method, r.URL.RequestURI(), r.Header.Get("X-Praxis-Actor"), body)
		if got := r.Header.Get("X-Praxis-Signature"); got != want {
			t.Errorf("signature mismatch: got %s want %s", got, want)
		}
		if r.URL.RequestURI() != wantURI {
			t.Errorf("unexpected path %q, want %q", r.URL.RequestURI(), wantURI)
		}
		w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestAssessorPacketIsReviewerOrAdminOnly(t *testing.T) {
	var calls int32
	srv := newSigningServer(t, &calls, "/internal/v1/scenarios/CS-01-A/assessor-packet?version=1.1.0")

	client, err := New(srv.URL, testSecret)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	_, err = client.AssessorPacket(ctx, CandidateActor("abc"), "", "CS-01-A", "1.1.0")
	if !errors.Is(err, ErrActorNotAllowed) {
		t.Fatalf("candidate must be refused, got %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 0 {
		t.Fatalf("a refused call must not reach the service, got %d calls", got)
	}

	if _, err := client.AssessorPacket(ctx, ReviewerActor("abc"), "", "CS-01-A", "1.1.0"); err != nil {
		t.Fatalf("reviewer should be allowed: %v", err)
	}
	if _, err := client.AssessorPacket(ctx, AdminActor("abc"), "", "CS-01-A", "1.1.0"); err != nil {
		t.Fatalf("admin should be allowed: %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("expected 2 calls, got %d", got)
	}
}

func TestCompareAllowsReviewerAndSystemOnly(t *testing.T) {
	var calls int32
	srv := newSigningServer(t, &calls, "/internal/v1/diagnostics/compare")

	client, err := New(srv.URL, testSecret)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	payload := map[string]string{"request_id": "r1"}

	if _, err := client.Compare(ctx, CandidateActor("abc"), "", payload); !errors.Is(err, ErrActorNotAllowed) {
		t.Fatalf("candidate must be refused, got %v", err)
	}
	if _, err := client.Compare(ctx, AdminActor("abc"), "", payload); !errors.Is(err, ErrActorNotAllowed) {
		t.Fatalf("admin must be refused, got %v", err)
	}
	if _, err := client.Compare(ctx, ReviewerActor("abc"), "", payload); err != nil {
		t.Fatalf("reviewer should be allowed: %v", err)
	}
	if _, err := client.Compare(ctx, SystemActor, "", payload); err != nil {
		t.Fatalf("system should be allowed: %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("expected 2 calls, got %d", got)
	}
}

func TestScoreIsReviewerOnly(t *testing.T) {
	var calls int32
	srv := newSigningServer(t, &calls, "/internal/v1/diagnostics/score")

	client, err := New(srv.URL, testSecret)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.ScoreDiagnostic(context.Background(), SystemActor, "", map[string]string{})
	if !errors.Is(err, ErrActorNotAllowed) {
		t.Fatalf("system must not score, got %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 0 {
		t.Fatalf("a refused call must not reach the service, got %d calls", got)
	}
}
