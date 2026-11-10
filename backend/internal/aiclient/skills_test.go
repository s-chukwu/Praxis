package aiclient

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func validSkillsRequest() SkillsRequest {
	return SkillsRequest{
		RequestID:           "r1",
		CandidateID:         "11111111-1111-1111-1111-111111111111",
		ConsentRef:          "22222222-2222-2222-2222-222222222222",
		RoleStandardVersion: "1.0.0",
		Text:                "Five years in customer support, handling escalations and refunds.",
	}
}

func TestSkillsSuggestSignsAsSystem(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		body, _ := io.ReadAll(r.Body)
		timestamp, _ := strconv.ParseInt(r.Header.Get("X-Praxis-Timestamp"), 10, 64)
		want := sign(testSecret, timestamp, r.Method, r.URL.RequestURI(), r.Header.Get("X-Praxis-Actor"), body)
		if r.Header.Get("X-Praxis-Signature") != want {
			t.Errorf("bad signature")
		}
		if r.Header.Get("X-Praxis-Actor") != "system" {
			t.Errorf("skills/suggest must be called as system, got %q", r.Header.Get("X-Praxis-Actor"))
		}
		if r.URL.Path != "/internal/v1/skills/suggest" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		if !strings.Contains(string(body), `"consent_ref":"22222222-2222-2222-2222-222222222222"`) {
			t.Errorf("consent_ref must be sent, body was %s", body)
		}
		w.Write([]byte(`{"suggestions":[]}`))
	}))
	defer srv.Close()

	client, err := New(srv.URL, testSecret)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.SkillsSuggest(context.Background(), "", validSkillsRequest()); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("expected 1 call, got %d", got)
	}
}

func TestSkillsSuggestNeedsConsentAndValidText(t *testing.T) {
	var calls int32
	srv := newSigningServer(t, &calls, "/internal/v1/skills/suggest")

	client, err := New(srv.URL, testSecret)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	noConsent := validSkillsRequest()
	noConsent.ConsentRef = " "
	if _, err := client.SkillsSuggest(ctx, "", noConsent); !errors.Is(err, ErrNoConsent) {
		t.Fatalf("a missing consent must be refused, got %v", err)
	}

	noID := validSkillsRequest()
	noID.CandidateID = ""
	if _, err := client.SkillsSuggest(ctx, "", noID); !errors.Is(err, ErrMissingID) {
		t.Fatalf("a missing candidate id must be refused, got %v", err)
	}

	empty := validSkillsRequest()
	empty.Text = ""
	if _, err := client.SkillsSuggest(ctx, "", empty); !errors.Is(err, ErrInvalidText) {
		t.Fatalf("empty text must be refused, got %v", err)
	}

	tooLong := validSkillsRequest()
	tooLong.Text = strings.Repeat("a", maxCVTextRunes+1)
	if _, err := client.SkillsSuggest(ctx, "", tooLong); !errors.Is(err, ErrInvalidText) {
		t.Fatalf("text over 20000 characters must be refused, got %v", err)
	}

	if got := atomic.LoadInt32(&calls); got != 0 {
		t.Fatalf("refused calls must not reach the service, got %d calls", got)
	}
}

func TestSkillsSuggestModelUnavailableIsManualReviewAndNotRetried(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"error":{"code":"model_unavailable","message":"model down","details":{"reason":"timeout"}}}`))
	}))
	defer srv.Close()

	client, err := New(srv.URL, testSecret)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.SkillsSuggest(context.Background(), "", validSkillsRequest())
	if !NeedsManualReview(err) {
		t.Fatalf("503 model_unavailable must mean manual review, got %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("skills/suggest must never be retried, got %d calls", got)
	}
}

func TestSkillsSuggestOwnTimeoutIsManualReview(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		time.Sleep(300 * time.Millisecond)
		w.Write([]byte(`{"suggestions":[]}`))
	}))
	defer srv.Close()

	client, err := New(srv.URL, testSecret)
	if err != nil {
		t.Fatal(err)
	}
	client.slow = &http.Client{Timeout: 50 * time.Millisecond}

	_, err = client.SkillsSuggest(context.Background(), "", validSkillsRequest())
	if !NeedsManualReview(err) {
		t.Fatalf("a timeout must mean manual review, got %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("a timeout must not be retried, got %d calls", got)
	}
}

func TestOtherFailuresAreNotManualReview(t *testing.T) {
	if NeedsManualReview(nil) {
		t.Fatal("no error is not a manual review")
	}
	authError := &APIError{Status: http.StatusServiceUnavailable, Code: "service_auth_not_configured"}
	if NeedsManualReview(authError) {
		t.Fatal("a configuration error is not a model outage")
	}
	forbidden := &APIError{Status: http.StatusForbidden, Code: "forbidden"}
	if NeedsManualReview(forbidden) {
		t.Fatal("a 403 is not a model outage")
	}
}
