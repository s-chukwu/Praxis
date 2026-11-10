package aiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
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

const cvMarker = "Quokka Ledger zanzibar.q@example.org hunter2"

func TestSkillsSuggestReturnsTheServiceAnswerUntouched(t *testing.T) { // success
	const answer = `{"request_id":"r1","result":{"suggestions":[{"competency":"C4","evidence_quote":"Escalated suspected account takeovers","confidence":"high","status":"suggested"}],"dropped":0,"model":"gemini/test","requires_candidate_confirmation":true}}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(answer))
	}))
	defer srv.Close()

	client, err := New(srv.URL, testSecret)
	if err != nil {
		t.Fatal(err)
	}
	data, err := client.SkillsSuggest(context.Background(), "", validSkillsRequest())
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if !json.Valid(data) {
		t.Fatalf("the answer must be valid JSON, got %s", data)
	}
	for _, want := range []string{`"evidence_quote":"Escalated suspected account takeovers"`, `"status":"suggested"`, `"requires_candidate_confirmation":true`} {
		if !strings.Contains(string(data), want) {
			t.Errorf("answer lost %s: %s", want, data)
		}
	}
}

func TestSkillsSuggestRefusesEveryBlankConsentBeforeAnythingElse(t *testing.T) { // consent required
	var calls int32
	srv := newSigningServer(t, &calls, "/internal/v1/skills/suggest")
	client, err := New(srv.URL, testSecret)
	if err != nil {
		t.Fatal(err)
	}

	for _, blank := range []string{"", " ", "\t\n"} {
		req := validSkillsRequest()
		req.ConsentRef = blank
		req.CandidateID = "" // also invalid: consent must be the error reported
		req.Text = ""
		if _, err := client.SkillsSuggest(context.Background(), "", req); !errors.Is(err, ErrNoConsent) {
			t.Errorf("consent_ref %q must be refused with ErrNoConsent, got %v", blank, err)
		}
	}
	if got := atomic.LoadInt32(&calls); got != 0 {
		t.Fatalf("a call without consent must never reach the service, got %d calls", got)
	}
}

func TestSkillsSuggestDeadlineIsManualReview(t *testing.T) { // 503 / timeout fallback
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
	}))
	defer srv.Close()

	client, err := New(srv.URL, testSecret)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err = client.SkillsSuggest(ctx, "", validSkillsRequest())
	if !NeedsManualReview(err) {
		t.Fatalf("a context deadline must mean manual review, got %v", err)
	}
}

func TestNeedsManualReviewSeesThroughWrapping(t *testing.T) { // 503 fallback
	down := &APIError{Status: http.StatusServiceUnavailable, Code: "model_unavailable"}
	if !NeedsManualReview(fmt.Errorf("suggest skills for candidate: %w", down)) {
		t.Fatal("a wrapped 503 model_unavailable must still mean manual review")
	}
	if NeedsManualReview(&APIError{Status: http.StatusInternalServerError, Code: "internal_error"}) {
		t.Fatal("a 500 is a bug, not a model outage")
	}
}

func TestSkillsSuggestNeverLogsOrReturnsCVText(t *testing.T) { // no CV text in logs
	var logs bytes.Buffer
	log.SetOutput(&logs)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	down := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"error":{"code":"model_unavailable","message":"model down","details":{"reason":"timeout"}}}`))
	}))
	defer down.Close()
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
	}))
	defer slow.Close()

	withCV := func() SkillsRequest {
		req := validSkillsRequest()
		req.Text = cvMarker
		return req
	}
	var errs []error

	c503, _ := New(down.URL, testSecret)
	_, err := c503.SkillsSuggest(context.Background(), "", withCV())
	errs = append(errs, err)

	cSlow, _ := New(slow.URL, testSecret)
	cSlow.slow = &http.Client{Timeout: 50 * time.Millisecond}
	_, err = cSlow.SkillsSuggest(context.Background(), "", withCV())
	errs = append(errs, err)

	noConsent := withCV()
	noConsent.ConsentRef = ""
	_, err = c503.SkillsSuggest(context.Background(), "", noConsent)
	errs = append(errs, err)

	tooLong := withCV()
	tooLong.Text += strings.Repeat("a", maxCVTextRunes)
	_, err = c503.SkillsSuggest(context.Background(), "", tooLong)
	errs = append(errs, err)

	for i, err := range errs {
		if err == nil {
			t.Fatalf("case %d: expected an error", i)
		}
		for _, part := range strings.Fields(cvMarker) {
			if strings.Contains(err.Error(), part) {
				t.Errorf("case %d: error text leaks CV content %q: %v", i, part, err)
			}
			if strings.Contains(logs.String(), part) {
				t.Errorf("case %d: log output leaks CV content %q", i, part)
			}
		}
	}
}
