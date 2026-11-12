package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/s-chukwu/Praxis/backend/internal/aiclient"
	"github.com/s-chukwu/Praxis/backend/internal/store"
	"github.com/s-chukwu/Praxis/backend/internal/workflow"
)

// fakeScorer stands in for the AI service. It records what Go sent.
type fakeScorer struct {
	response  string
	err       error
	calls     int
	lastActor aiclient.Actor
	lastBody  string
}

func (f *fakeScorer) ScoreDiagnostic(ctx context.Context, actor aiclient.Actor, requestID string, req any) (json.RawMessage, error) {
	f.calls++
	f.lastActor = actor
	encoded, _ := json.Marshal(req)
	f.lastBody = string(encoded)
	if f.err != nil {
		return nil, f.err
	}
	return json.RawMessage(f.response), nil
}

func scoreResponseJSON(outcome string) string {
	return fmt.Sprintf(`{"rubric":{"sha256":%q},"scenarios":{"sha256":%q},
		"result":{"weighted_score":82.5,"proposed_outcome":%q,"outcome_reasons":["meets_pilot_threshold"],
		"evidence_issues":[],"pending_critical_events":[]}}`,
		strings.Repeat("a", 64), strings.Repeat("b", 64), outcome)
}

// fakeFinalizer applies the same rules as the database layer: own draft only,
// no pending findings, and the attempt moves to reviewed.
type fakeFinalizer struct {
	reviews *fakeReviews
}

func (f fakeFinalizer) FinalizeReview(ctx context.Context, reviewID string, reviewer store.Actor, in store.FinalizeInput) (store.Review, error) {
	if err := f.reviews.ownDraft(reviewID, reviewer); err != nil {
		return store.Review{}, err
	}
	for _, x := range f.reviews.findings[reviewID] {
		if x.Decision == "pending" {
			return store.Review{}, store.ErrPendingFindings
		}
	}

	r := f.reviews.reviews[reviewID]
	now := time.Now()
	score := in.WeightedScore
	outcome := in.ProposedOutcome
	r.FinalizedAt = &now
	r.ReviewedCompetencies = in.ReviewedCompetencies
	r.WeightedScore = &score
	r.ProposedOutcome = &outcome
	r.OutcomeReasons = in.OutcomeReasons
	f.reviews.reviews[reviewID] = r

	a := f.reviews.attempts.items[r.AttemptID]
	a.Status = workflow.Reviewed
	f.reviews.attempts.items[r.AttemptID] = a
	return r, nil
}

func newFinalizeTestRouter(scorer Scorer) (http.Handler, *fakeReviews) {
	attempts := newFakeAttempts()
	reviews := &fakeReviews{
		attempts: attempts,
		reviews:  map[string]store.Review{},
		scores:   map[string][]store.DimensionScore{},
		findings: map[string][]store.Finding{},
	}
	mux := http.NewServeMux()
	registerAttemptRoutes(mux, fakeAuth{}, attempts)
	registerReviewRoutes(mux, fakeAuth{}, reviews)
	registerFinalizeRoute(mux, fakeAuth{}, reviews, fakeFinalizer{reviews: reviews}, scorer)
	return mux, reviews
}

const goodAnswers = `{"answers":{"customer_reply":"Thank you for writing. I can see one captured payment and one pending entry.","internal_note":"Escalated to billing for confirmation."}}`

// reviewWithScores creates an attempt with the given answers, submits it, starts a
// review as the reviewer and saves one dimension score. It returns the attempt and review ids.
func reviewWithScores(t *testing.T, h http.Handler, answers string) (string, string) {
	t.Helper()

	attemptID := mustCreate(t, h, candidateA)
	expectStatus(t, call(t, h, http.MethodPut, "/v1/attempts/"+attemptID+"/answers", candidateA, "candidate", answers), http.StatusOK, "save answers")
	expectStatus(t, call(t, h, http.MethodPost, "/v1/attempts/"+attemptID+"/start", candidateA, "candidate", ""), http.StatusOK, "start")
	expectStatus(t, call(t, h, http.MethodPost, "/v1/attempts/"+attemptID+"/submit", candidateA, "candidate", ""), http.StatusOK, "submit")

	reviewID := startedReview(t, h, attemptID)
	expectStatus(t, call(t, h, http.MethodPut, "/v1/reviews/"+reviewID+"/scores", reviewerID, "reviewer", scoresJSON), http.StatusOK, "save scores")
	return attemptID, reviewID
}

const finalizeJSON = `{"reviewed_competencies":{"C1":3,"C2":4},"notes":"Reviewed against the rubric."}`

func TestFinalizeBuildsTheScoreRequestAndStoresTheResult(t *testing.T) {
	scorer := &fakeScorer{response: scoreResponseJSON("evidence_verified")}
	h, _ := newFinalizeTestRouter(scorer)
	attemptID, reviewID := reviewWithScores(t, h, goodAnswers)

	rec := call(t, h, http.MethodPost, "/v1/reviews/"+reviewID+"/finalize", reviewerID, "reviewer", finalizeJSON)
	expectStatus(t, rec, http.StatusOK, "finalize")
	if !strings.Contains(rec.Body.String(), `"proposed_outcome":"evidence_verified"`) || strings.Contains(rec.Body.String(), `"finalized_at":null`) {
		t.Fatalf("the review should be finalized with the AI outcome, got %s", rec.Body.String())
	}

	if scorer.calls != 1 {
		t.Fatalf("the AI service should be called once, got %d", scorer.calls)
	}
	if scorer.lastActor != aiclient.ReviewerActor(reviewerID) {
		t.Fatalf("the call must be signed as the reviewer, got %q", scorer.lastActor)
	}
	for _, want := range []string{
		`"reviewer_id":"` + reviewerID + `"`,
		`"candidate_id":"` + candidateA + `"`,
		`"scenario_id":"CS-01-A"`,
		`"type":"customer_reply"`,
		`"type":"internal_note"`,
		`"accuracy":{"score":3`,
	} {
		if !strings.Contains(scorer.lastBody, want) {
			t.Errorf("the score request is missing %s, got %s", want, scorer.lastBody)
		}
	}

	attempt := call(t, h, http.MethodGet, "/v1/attempts/"+attemptID, adminID, "admin", "")
	if !strings.Contains(attempt.Body.String(), `"status":"reviewed"`) {
		t.Fatalf("the attempt should be reviewed, got %s", attempt.Body.String())
	}

	// A second finalize is refused.
	again := call(t, h, http.MethodPost, "/v1/reviews/"+reviewID+"/finalize", reviewerID, "reviewer", finalizeJSON)
	expectStatus(t, again, http.StatusConflict, "second finalize")
}

func TestOnlyTheOwningReviewerCanFinalize(t *testing.T) {
	scorer := &fakeScorer{response: scoreResponseJSON("evidence_verified")}
	h, _ := newFinalizeTestRouter(scorer)
	_, reviewID := reviewWithScores(t, h, goodAnswers)
	path := "/v1/reviews/" + reviewID + "/finalize"

	expectStatus(t, call(t, h, http.MethodPost, path, reviewerB, "reviewer", finalizeJSON), http.StatusNotFound, "another reviewer")
	expectStatus(t, call(t, h, http.MethodPost, path, candidateA, "candidate", finalizeJSON), http.StatusForbidden, "candidate")
	expectStatus(t, call(t, h, http.MethodPost, path, adminID, "admin", finalizeJSON), http.StatusForbidden, "admin")
	expectStatus(t, call(t, h, http.MethodPost, path, "", "", finalizeJSON), http.StatusUnauthorized, "no login")

	if scorer.calls != 0 {
		t.Fatalf("refused callers must not reach the AI service, got %d calls", scorer.calls)
	}
}

func TestFinalizeRefusesBadCompetencies(t *testing.T) {
	scorer := &fakeScorer{response: scoreResponseJSON("evidence_verified")}
	h, _ := newFinalizeTestRouter(scorer)
	_, reviewID := reviewWithScores(t, h, goodAnswers)
	path := "/v1/reviews/" + reviewID + "/finalize"

	for name, body := range map[string]string{
		"score above 4": `{"reviewed_competencies":{"C1":9}}`,
		"unknown code":  `{"reviewed_competencies":{"X1":3}}`,
		"empty":         `{"reviewed_competencies":{}}`,
	} {
		rec := call(t, h, http.MethodPost, path, reviewerID, "reviewer", body)
		expectStatus(t, rec, http.StatusUnprocessableEntity, name)
	}
	expectStatus(t, call(t, h, http.MethodPost, path, reviewerID, "reviewer", `{"surprise":1}`), http.StatusBadRequest, "unknown fields")

	if scorer.calls != 0 {
		t.Fatalf("bad input must not reach the AI service, got %d calls", scorer.calls)
	}
}

func TestFinalizeNeedsScorableAnswers(t *testing.T) {
	scorer := &fakeScorer{response: scoreResponseJSON("evidence_verified")}
	h, _ := newFinalizeTestRouter(scorer)
	_, reviewID := reviewWithScores(t, h, `{"answers":{"customer_reply":5}}`)

	rec := call(t, h, http.MethodPost, "/v1/reviews/"+reviewID+"/finalize", reviewerID, "reviewer", finalizeJSON)
	expectStatus(t, rec, http.StatusUnprocessableEntity, "answers that are not text")
	if !strings.Contains(rec.Body.String(), "answers_not_scorable") || scorer.calls != 0 {
		t.Fatalf("expected answers_not_scorable without calling the AI service, got %s", rec.Body.String())
	}
}

func TestFinalizeStopsWhenTheAIHoldsTheReviewBack(t *testing.T) {
	scorer := &fakeScorer{response: scoreResponseJSON("review_pending")}
	h, _ := newFinalizeTestRouter(scorer)
	_, reviewID := reviewWithScores(t, h, goodAnswers)

	rec := call(t, h, http.MethodPost, "/v1/reviews/"+reviewID+"/finalize", reviewerID, "reviewer", finalizeJSON)
	expectStatus(t, rec, http.StatusConflict, "review_pending outcome")
	if !strings.Contains(rec.Body.String(), `"code":"review_pending"`) {
		t.Fatalf("expected the review_pending code, got %s", rec.Body.String())
	}

	still := call(t, h, http.MethodGet, "/v1/reviews/"+reviewID, reviewerID, "reviewer", "")
	if !strings.Contains(still.Body.String(), `"finalized_at":null`) {
		t.Fatalf("the review must stay open, got %s", still.Body.String())
	}
}

func TestFinalizeNeverPassesAConfirmedCriticalFinding(t *testing.T) {
	scorer := &fakeScorer{response: scoreResponseJSON("evidence_verified")}
	h, _ := newFinalizeTestRouter(scorer)
	_, reviewID := reviewWithScores(t, h, goodAnswers)

	confirmed := `{"findings":[{"event_type":"unauthorized_action","supporting_quote":"I can help you reset that password now.","policy_reference":"POL-1","decision":"confirmed"}]}`
	expectStatus(t, call(t, h, http.MethodPut, "/v1/reviews/"+reviewID+"/findings", reviewerID, "reviewer", confirmed), http.StatusOK, "save finding")

	rec := call(t, h, http.MethodPost, "/v1/reviews/"+reviewID+"/finalize", reviewerID, "reviewer", finalizeJSON)
	expectStatus(t, rec, http.StatusConflict, "pass with a confirmed finding")
	if !strings.Contains(rec.Body.String(), "inconsistent_result") {
		t.Fatalf("expected inconsistent_result, got %s", rec.Body.String())
	}

	still := call(t, h, http.MethodGet, "/v1/reviews/"+reviewID, reviewerID, "reviewer", "")
	if !strings.Contains(still.Body.String(), `"finalized_at":null`) {
		t.Fatalf("the review must stay open, got %s", still.Body.String())
	}
}

func TestFinalizeTurnsAIFailuresIntoClearAnswers(t *testing.T) {
	cases := map[string]struct {
		err      error
		status   int
		wantCode string
	}{
		"service unavailable": {&aiclient.APIError{Status: 503, Code: "service_auth_not_configured"}, http.StatusBadGateway, "ai_unavailable"},
		"signing refused":     {&aiclient.APIError{Status: 401, Code: "unauthenticated"}, http.StatusBadGateway, "ai_misconfigured"},
		"rejected review":     {&aiclient.APIError{Status: 422, Code: "invalid_input", Message: "a rubric dimension is missing"}, http.StatusUnprocessableEntity, "ai_rejected"},
		"version conflict":    {&aiclient.APIError{Status: 409, Code: "unsupported_version"}, http.StatusConflict, "ai_version_conflict"},
		"network failure":     {context.DeadlineExceeded, http.StatusBadGateway, "ai_unavailable"},
	}

	for name, c := range cases {
		scorer := &fakeScorer{err: c.err}
		h, _ := newFinalizeTestRouter(scorer)
		_, reviewID := reviewWithScores(t, h, goodAnswers)

		rec := call(t, h, http.MethodPost, "/v1/reviews/"+reviewID+"/finalize", reviewerID, "reviewer", finalizeJSON)
		expectStatus(t, rec, c.status, name)
		if !strings.Contains(rec.Body.String(), c.wantCode) {
			t.Errorf("%s: expected %s, got %s", name, c.wantCode, rec.Body.String())
		}
	}
}

func TestFinalizeAnswers503WhenTheAIServiceIsNotConfigured(t *testing.T) {
	h, _ := newFinalizeTestRouter(nil)
	_, reviewID := reviewWithScores(t, h, goodAnswers)

	rec := call(t, h, http.MethodPost, "/v1/reviews/"+reviewID+"/finalize", reviewerID, "reviewer", finalizeJSON)
	expectStatus(t, rec, http.StatusServiceUnavailable, "no AI service")
	if !strings.Contains(rec.Body.String(), "ai_not_configured") {
		t.Fatalf("expected ai_not_configured, got %s", rec.Body.String())
	}
}
