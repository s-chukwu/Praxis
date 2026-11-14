package httpapi

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/s-chukwu/Praxis/backend/internal/store"
	"github.com/s-chukwu/Praxis/backend/internal/workflow"
)

func (f *fakeReviews) GetFinalizedReview(ctx context.Context, attemptID string) (store.Review, error) {
	for _, r := range f.reviews {
		if r.AttemptID == attemptID && r.FinalizedAt != nil {
			return r, nil
		}
	}
	return store.Review{}, store.ErrNotFound
}

type progressFixture struct {
	h        http.Handler
	attempts *fakeAttempts
	reviews  *fakeReviews
	plans    *fakePlans
}

func newProgressFixture() progressFixture {
	attempts := newFakeAttempts()
	reviews := &fakeReviews{attempts: attempts, reviews: map[string]store.Review{},
		scores: map[string][]store.DimensionScore{}, findings: map[string][]store.Finding{}}
	plans := &fakePlans{byReview: map[string]store.Plan{}}
	mux := http.NewServeMux()
	registerAttemptRoutes(mux, fakeAuth{}, attempts)
	registerProgressRoute(mux, fakeAuth{}, attempts, reviews, plans)
	return progressFixture{mux, attempts, reviews, plans}
}

func (f progressFixture) progress(t *testing.T, user string) string {
	t.Helper()
	rec := call(t, f.h, http.MethodGet, "/v1/me/progress", user, "candidate", "")
	expectStatus(t, rec, http.StatusOK, "progress")
	return rec.Body.String()
}

// setStatus moves candidate A's attempt straight to a state, skipping the workflow.
func (f progressFixture) setStatus(id string, s workflow.State) {
	a := f.attempts.items[id]
	a.Status = s
	f.attempts.items[id] = a
}

func TestProgressFollowsTheCandidateThroughEachStep(t *testing.T) {
	f := newProgressFixture()
	if body := f.progress(t, candidateA); !strings.Contains(body, `"baseline":null`) {
		t.Fatalf("no attempt yet should give an empty baseline, got %s", body)
	}

	id := mustCreate(t, f.h, candidateA)
	f.setStatus(id, workflow.Submitted)
	if body := f.progress(t, candidateA); !strings.Contains(body, `"status":"submitted"`) || !strings.Contains(body, `"result":null`) {
		t.Fatalf("submitted attempt should show status without a result, got %s", body)
	}

	now := time.Now()
	score, outcome := 75.0, "evidence_verified"
	reviewID := "99999999-9999-9999-9999-000000000001"
	f.reviews.reviews[reviewID] = store.Review{ID: reviewID, AttemptID: id, ReviewerID: reviewerID,
		FinalizedAt: &now, WeightedScore: &score, ProposedOutcome: &outcome}
	f.reviews.scores[reviewID] = []store.DimensionScore{{Dimension: "accuracy", Score: 3, Evidence: []string{"secret evidence quote"}}}
	f.setStatus(id, workflow.Reviewed)

	body := f.progress(t, candidateA)
	for _, want := range []string{`"weighted_score":75`, `"proposed_outcome":"evidence_verified"`, `"dimension":"accuracy","score":3`, `"plan":null`} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %s in %s", want, body)
		}
	}
	if strings.Contains(body, "secret evidence quote") {
		t.Fatalf("progress should not carry evidence quotes: %s", body)
	}

	f.plans.byReview[reviewID] = store.Plan{ID: "plan-1", ReviewID: reviewID, CandidateID: candidateA, Status: "active",
		Modules: []store.PlanModule{{Position: 1, CourseCode: "C4"}}}
	if body := f.progress(t, candidateA); !strings.Contains(body, `"course_code":"C4"`) {
		t.Fatalf("the plan should appear once it exists, got %s", body)
	}

	// An admin reopening the review hides the old result.
	f.setStatus(id, workflow.UnderReview)
	if body := f.progress(t, candidateA); !strings.Contains(body, `"result":null`) {
		t.Fatalf("a reopened review must hide the old result, got %s", body)
	}
}

func TestProgressIsPrivateToTheCandidate(t *testing.T) {
	f := newProgressFixture()
	mustCreate(t, f.h, candidateA)
	if body := f.progress(t, candidateB); !strings.Contains(body, `"baseline":null`) {
		t.Fatalf("candidate B must not see candidate A's attempt, got %s", body)
	}
	expectStatus(t, call(t, f.h, http.MethodGet, "/v1/me/progress", reviewerID, "reviewer", ""), http.StatusForbidden, "reviewer")
}
