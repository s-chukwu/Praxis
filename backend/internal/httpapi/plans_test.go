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
)

// fakePlanner stands in for the AI service. It records what Go sent.
type fakePlanner struct {
	response  string
	err       error
	calls     int
	lastActor aiclient.Actor
	lastBody  string
}

func (f *fakePlanner) LearningPlan(ctx context.Context, actor aiclient.Actor, requestID string, req any) (json.RawMessage, error) {
	f.calls++
	f.lastActor = actor
	encoded, _ := json.Marshal(req)
	f.lastBody = string(encoded)
	if f.err != nil {
		return nil, f.err
	}
	return json.RawMessage(f.response), nil
}

func planResponseJSON(status string) string {
	hash := strings.Repeat("c", 64)
	if status == "diagnostic_incomplete" {
		return fmt.Sprintf(`{"role_standard":{"version":"1.0.0","sha256":%q},"policy":{"version":"1.0.0","sha256":%q},
			"curriculum":{"version":"1.0.0","sha256":%q},
			"result":{"status":"diagnostic_incomplete","missing_competencies":["C5"],"modules":[],"reassessment_required":false}}`,
			hash, hash, hash)
	}
	return fmt.Sprintf(`{"role_standard":{"version":"1.0.0","sha256":%q},"policy":{"version":"1.0.0","sha256":%q},
		"curriculum":{"version":"1.0.0","sha256":%q},
		"result":{"status":"plan_ready","missing_competencies":[],"safety_gate":{"required":true},
		"modules":[
			{"order":1,"course_id":"M-ROLE","title":"Role orientation","covers":["C1","C3"],"reasons":[{"code":"role_orientation"}],"previously_completed":false},
			{"order":2,"course_id":"M-SAFE","title":"Safety and privacy","covers":["C4"],"reasons":[{"code":"safety_gate"}],"previously_completed":false}],
		"reassessment_required":true}}`, hash, hash, hash)
}

type fakePlans struct {
	byReview map[string]store.Plan
	verified []string
	count    int
}

func (f *fakePlans) CreatePlan(ctx context.Context, n store.NewPlan, actor store.Actor) (store.Plan, bool, error) {
	if existing, ok := f.byReview[n.ReviewID]; ok {
		return existing, false, nil
	}
	f.count++
	plan := store.Plan{
		ID:                   fmt.Sprintf("77777777-7777-7777-7777-%012d", f.count),
		CandidateID:          n.CandidateID,
		ReviewID:             n.ReviewID,
		TargetTrack:          n.TargetTrack,
		Status:               n.Status,
		RoleStandardVersion:  n.RoleStandardVersion,
		PolicyVersion:        n.PolicyVersion,
		CurriculumVersion:    n.CurriculumVersion,
		SafetyGateRequired:   n.SafetyGateRequired,
		ReassessmentRequired: n.ReassessmentRequired,
		CreatedAt:            time.Now(),
	}
	for _, m := range n.Modules {
		plan.Modules = append(plan.Modules, store.PlanModule{
			Position:            m.Position,
			CourseCode:          m.CourseCode,
			Title:               m.Title,
			Covers:              m.Covers,
			Reasons:             m.Reasons,
			PreviouslyCompleted: m.PreviouslyCompleted,
		})
	}
	f.byReview[n.ReviewID] = plan
	return plan, true, nil
}

func (f *fakePlans) GetPlanByReview(ctx context.Context, reviewID string) (store.Plan, error) {
	plan, ok := f.byReview[reviewID]
	if !ok {
		return store.Plan{}, store.ErrNotFound
	}
	return plan, nil
}

func (f *fakePlans) ListVerifiedModules(ctx context.Context, candidateID string) ([]string, error) {
	return f.verified, nil
}

// newPlanTestRouter wires attempts, reviews, finalize and plan routes with fakes.
func newPlanTestRouter(scorer Scorer, planner Planner) (http.Handler, *fakePlans) {
	attempts := newFakeAttempts()
	reviews := &fakeReviews{
		attempts: attempts,
		reviews:  map[string]store.Review{},
		scores:   map[string][]store.DimensionScore{},
		findings: map[string][]store.Finding{},
	}
	plans := &fakePlans{byReview: map[string]store.Plan{}, verified: []string{"M-ROLE"}}

	mux := http.NewServeMux()
	registerAttemptRoutes(mux, fakeAuth{}, attempts)
	registerReviewRoutes(mux, fakeAuth{}, reviews)
	registerFinalizeRoute(mux, fakeAuth{}, reviews, fakeFinalizer{reviews: reviews}, scorer)
	registerPlanRoutes(mux, fakeAuth{}, reviews, plans, planner)
	return mux, plans
}

// finalizedReview takes an attempt all the way to a finalized review.
func finalizedReview(t *testing.T, h http.Handler) string {
	t.Helper()

	_, reviewID := reviewWithScores(t, h, goodAnswers)
	rec := call(t, h, http.MethodPost, "/v1/reviews/"+reviewID+"/finalize", reviewerID, "reviewer", finalizeJSON)
	expectStatus(t, rec, http.StatusOK, "finalize")
	return reviewID
}

func TestPlanIsBuiltFromTheFinalizedReview(t *testing.T) {
	scorer := &fakeScorer{response: scoreResponseJSON("needs_practice")}
	planner := &fakePlanner{response: planResponseJSON("plan_ready")}
	h, plans := newPlanTestRouter(scorer, planner)
	reviewID := finalizedReview(t, h)
	path := "/v1/reviews/" + reviewID + "/learning-plan"

	rec := call(t, h, http.MethodPost, path, reviewerID, "reviewer", "")
	expectStatus(t, rec, http.StatusCreated, "create plan")
	for _, want := range []string{`"status":"plan_ready"`, `"course_code":"M-ROLE"`, `"course_code":"M-SAFE"`, `"safety_gate_required":true`} {
		if !strings.Contains(rec.Body.String(), want) {
			t.Errorf("the plan is missing %s, got %s", want, rec.Body.String())
		}
	}

	if planner.lastActor != aiclient.ReviewerActor(reviewerID) {
		t.Fatalf("the call must be signed as the reviewer, got %q", planner.lastActor)
	}
	for _, want := range []string{
		`"candidate_id":"` + candidateA + `"`,
		`"target_track":"operator"`,
		`"reviewed_competencies":{"C1":3,"C2":4}`,
		`"verified_completed_modules":["M-ROLE"]`,
	} {
		if !strings.Contains(planner.lastBody, want) {
			t.Errorf("the plan request is missing %s, got %s", want, planner.lastBody)
		}
	}
	if len(plans.byReview) != 1 {
		t.Fatalf("one plan should be stored, got %d", len(plans.byReview))
	}

	// Asking again returns the same plan without calling the AI service again.
	again := call(t, h, http.MethodPost, path, reviewerID, "reviewer", "")
	expectStatus(t, again, http.StatusOK, "second request")
	if planner.calls != 1 {
		t.Fatalf("the AI service must be called once, got %d calls", planner.calls)
	}
}

func TestPlanNeedsAFinalizedReviewAndTheOwningReviewer(t *testing.T) {
	scorer := &fakeScorer{response: scoreResponseJSON("needs_practice")}
	planner := &fakePlanner{response: planResponseJSON("plan_ready")}
	h, _ := newPlanTestRouter(scorer, planner)

	// An open draft review cannot get a plan.
	_, draftID := reviewWithScores(t, h, goodAnswers)
	rec := call(t, h, http.MethodPost, "/v1/reviews/"+draftID+"/learning-plan", reviewerID, "reviewer", "")
	expectStatus(t, rec, http.StatusConflict, "plan before finalizing")

	// Roles on a finalized review.
	h2, _ := newPlanTestRouter(scorer, planner)
	id := finalizedReview(t, h2)
	path := "/v1/reviews/" + id + "/learning-plan"

	expectStatus(t, call(t, h2, http.MethodPost, path, reviewerB, "reviewer", ""), http.StatusNotFound, "another reviewer")
	expectStatus(t, call(t, h2, http.MethodPost, path, candidateA, "candidate", ""), http.StatusForbidden, "candidate")
	expectStatus(t, call(t, h2, http.MethodPost, path, adminID, "admin", ""), http.StatusForbidden, "admin")
	expectStatus(t, call(t, h2, http.MethodPost, path, "", "", ""), http.StatusUnauthorized, "no login")

	if planner.calls != 0 {
		t.Fatalf("refused callers must not reach the AI service, got %d calls", planner.calls)
	}
}

func TestIncompleteReviewsDoNotGetAPlan(t *testing.T) {
	scorer := &fakeScorer{response: scoreResponseJSON("needs_practice")}
	planner := &fakePlanner{response: planResponseJSON("diagnostic_incomplete")}
	h, plans := newPlanTestRouter(scorer, planner)
	reviewID := finalizedReview(t, h)

	rec := call(t, h, http.MethodPost, "/v1/reviews/"+reviewID+"/learning-plan", reviewerID, "reviewer", "")
	expectStatus(t, rec, http.StatusConflict, "diagnostic incomplete")
	if !strings.Contains(rec.Body.String(), "diagnostic_incomplete") || !strings.Contains(rec.Body.String(), "C5") {
		t.Fatalf("the answer should name the missing competency, got %s", rec.Body.String())
	}
	if len(plans.byReview) != 0 {
		t.Fatal("no plan may be stored for an incomplete review")
	}
}

func TestPlanAIFailuresAreClear(t *testing.T) {
	scorer := &fakeScorer{response: scoreResponseJSON("needs_practice")}
	planner := &fakePlanner{err: &aiclient.APIError{Status: 503, Code: "service_auth_not_configured"}}
	h, plans := newPlanTestRouter(scorer, planner)
	reviewID := finalizedReview(t, h)

	rec := call(t, h, http.MethodPost, "/v1/reviews/"+reviewID+"/learning-plan", reviewerID, "reviewer", "")
	expectStatus(t, rec, http.StatusBadGateway, "AI service down")
	if len(plans.byReview) != 0 {
		t.Fatal("no plan may be stored when the AI service fails")
	}
}

func TestWhoCanReadAPlan(t *testing.T) {
	scorer := &fakeScorer{response: scoreResponseJSON("needs_practice")}
	planner := &fakePlanner{response: planResponseJSON("plan_ready")}
	h, _ := newPlanTestRouter(scorer, planner)
	reviewID := finalizedReview(t, h)
	path := "/v1/reviews/" + reviewID + "/learning-plan"

	expectStatus(t, call(t, h, http.MethodGet, path, candidateA, "candidate", ""), http.StatusNotFound, "no plan yet")
	expectStatus(t, call(t, h, http.MethodPost, path, reviewerID, "reviewer", ""), http.StatusCreated, "create plan")

	expectStatus(t, call(t, h, http.MethodGet, path, candidateA, "candidate", ""), http.StatusOK, "owning candidate")
	expectStatus(t, call(t, h, http.MethodGet, path, reviewerID, "reviewer", ""), http.StatusOK, "owning reviewer")
	expectStatus(t, call(t, h, http.MethodGet, path, adminID, "admin", ""), http.StatusOK, "admin")
	expectStatus(t, call(t, h, http.MethodGet, path, candidateB, "candidate", ""), http.StatusNotFound, "another candidate")
	expectStatus(t, call(t, h, http.MethodGet, path, reviewerB, "reviewer", ""), http.StatusNotFound, "another reviewer")
}

func TestPlanRoutesAnswer503WhenNotConfigured(t *testing.T) {
	scorer := &fakeScorer{response: scoreResponseJSON("needs_practice")}
	h, _ := newPlanTestRouter(scorer, nil)
	reviewID := finalizedReview(t, h)

	rec := call(t, h, http.MethodPost, "/v1/reviews/"+reviewID+"/learning-plan", reviewerID, "reviewer", "")
	expectStatus(t, rec, http.StatusServiceUnavailable, "no AI service")
	if !strings.Contains(rec.Body.String(), "ai_not_configured") {
		t.Fatalf("expected ai_not_configured, got %s", rec.Body.String())
	}

	mux := http.NewServeMux()
	registerPlanRoutes(mux, fakeAuth{}, nil, nil, nil)
	expectStatus(t, call(t, mux, http.MethodGet, "/v1/reviews/"+reviewID+"/learning-plan", reviewerID, "reviewer", ""),
		http.StatusServiceUnavailable, "no database")
}
