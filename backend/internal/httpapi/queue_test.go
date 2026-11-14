package httpapi

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/s-chukwu/Praxis/backend/internal/store"
)

func TestPriorityFollowsTheSLA(t *testing.T) {
	now := time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC)
	at := func(d time.Duration) *time.Time { v := now.Add(d); return &v }
	cases := map[string]*time.Time{"normal": at(30 * time.Hour), "due_soon": at(6 * time.Hour), "overdue": at(-time.Minute)}
	for want, due := range cases {
		if got := priority(due, now); got != want {
			t.Errorf("due %v: want %s, got %s", due, want, got)
		}
	}
	if priority(nil, now) != "normal" {
		t.Error("no submitted time should be normal")
	}
}

func TestReviewQueueShowsWhatReviewersNeedWithoutNames(t *testing.T) {
	attempts := newFakeAttempts()
	reviews := &fakeReviews{attempts: attempts, reviews: map[string]store.Review{},
		scores: map[string][]store.DimensionScore{}, findings: map[string][]store.Finding{}}
	mux := http.NewServeMux()
	registerAttemptRoutes(mux, fakeAuth{}, attempts)
	registerReviewRoutes(mux, fakeAuth{}, reviews)
	registerQueueRoute(mux, fakeAuth{}, attempts, reviews)

	attemptID := submittedAttempt(t, mux)
	a := attempts.items[attemptID]
	submitted := time.Now().Add(-50 * time.Hour)
	a.SubmittedAt = &submitted
	attempts.items[attemptID] = a
	startedReview(t, mux, attemptID)

	rec := call(t, mux, http.MethodGet, "/v1/review-queue", reviewerID, "reviewer", "")
	expectStatus(t, rec, http.StatusOK, "queue")
	body := rec.Body.String()
	for _, want := range []string{`"candidate_ref":"C-11111111"`, `"program":"operator"`, `"status":"under_review"`,
		`"priority":"overdue"`, `"mine":true`, `"sla_hours":48`} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %s in %s", want, body)
		}
	}
	if strings.Contains(body, candidateA) || strings.Contains(body, "answers") {
		t.Fatalf("the queue must not expose the full candidate id or answers: %s", body)
	}

	rec = call(t, mux, http.MethodGet, "/v1/review-queue", "44444444-4444-4444-4444-444444444444", "reviewer", "")
	if !strings.Contains(rec.Body.String(), `"mine":false`) {
		t.Fatalf("another reviewer should see the case as not theirs: %s", rec.Body.String())
	}
	expectStatus(t, call(t, mux, http.MethodGet, "/v1/review-queue", candidateA, "candidate", ""), http.StatusForbidden, "candidate")
}
