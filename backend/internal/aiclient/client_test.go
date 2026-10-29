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
)

var testSecret = strings.Repeat("s", 32)

func TestCanonicalFormat(t *testing.T) {
	got := canonical(1700000000, "GET", "/internal/v1/catalog", "system", nil)
	want := "v1\n1700000000\nGET\n/internal/v1/catalog\nsystem\n" +
		"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if got != want {
		t.Fatalf("canonical mismatch\n got: %q\nwant: %q", got, want)
	}
}

func TestClientSignsRequests(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		timestamp, err := strconv.ParseInt(r.Header.Get("X-Praxis-Timestamp"), 10, 64)
		if err != nil {
			t.Errorf("bad timestamp header: %v", err)
		}
		actor := r.Header.Get("X-Praxis-Actor")
		want := sign(testSecret, timestamp, r.Method, r.URL.RequestURI(), actor, body)
		if got := r.Header.Get("X-Praxis-Signature"); got != want {
			t.Errorf("signature mismatch: got %s want %s", got, want)
		}
		if actor != "reviewer:abc" {
			t.Errorf("unexpected actor %q", actor)
		}
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	client, err := New(srv.URL, testSecret)
	if err != nil {
		t.Fatal(err)
	}
	payload := map[string]string{"hello": "world"}
	if _, err := client.LearningPlan(context.Background(), ReviewerActor("abc"), "req-12345678", payload); err != nil {
		t.Fatal(err)
	}
}

func TestRetriesServerErrorsButNotClientErrors(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if r.URL.Path == "/internal/v1/catalog" && n < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"error":{"code":"service_auth_not_configured","message":"down"}}`))
			return
		}
		if r.URL.Path == "/internal/v1/learning-plans" {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte(`{"error":{"code":"invalid_input","message":"bad","request_id":"r1"}}`))
			return
		}
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	client, err := New(srv.URL, testSecret)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := client.Catalog(context.Background(), SystemActor, ""); err != nil {
		t.Fatalf("expected success after retries: %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Fatalf("expected 3 calls for 5xx retries, got %d", got)
	}

	atomic.StoreInt32(&calls, 0)
	_, err = client.LearningPlan(context.Background(), SystemActor, "", map[string]string{})
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Code != "invalid_input" || apiErr.Status != 422 {
		t.Fatalf("expected invalid_input APIError, got %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("4xx must not be retried, got %d calls", got)
	}
}

func TestNewRejectsShortSecret(t *testing.T) {
	if _, err := New("http://ai:8000", "too-short"); err == nil {
		t.Fatal("expected an error for a short secret")
	}
}
