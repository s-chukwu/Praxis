package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/s-chukwu/Praxis/backend/internal/auth"
	"github.com/s-chukwu/Praxis/backend/internal/store"
	"github.com/s-chukwu/Praxis/backend/internal/workflow"
)

const (
	candidateA = "11111111-1111-1111-1111-111111111111"
	candidateB = "22222222-2222-2222-2222-222222222222"
	reviewerID = "33333333-3333-3333-3333-333333333333"
	adminID    = "44444444-4444-4444-4444-444444444444"
)

// fakeAuth reads the caller from test headers instead of checking a real token.
type fakeAuth struct{}

func (fakeAuth) Authenticate(r *http.Request) (auth.Identity, error) {
	user, role := r.Header.Get("X-Test-User"), r.Header.Get("X-Test-Role")
	if user == "" || role == "" {
		return auth.Identity{}, auth.ErrMissingToken
	}
	return auth.Identity{UserID: user, Role: role}, nil
}

// fakeAttempts is an in memory stand in for the database layer. It applies the
// same state and role rules, so the tests see realistic answers.
type fakeAttempts struct {
	items map[string]store.Attempt
	count int
}

func newFakeAttempts() *fakeAttempts {
	return &fakeAttempts{items: map[string]store.Attempt{}}
}

func (f *fakeAttempts) CreateAttempt(ctx context.Context, n store.NewAttempt, actor store.Actor) (store.Attempt, error) {
	f.count++
	id := fmt.Sprintf("00000000-0000-0000-0000-%012d", f.count)
	scenario := n.ScenarioID
	a := store.Attempt{
		ID:                  id,
		CandidateID:         n.CandidateID,
		TargetTrack:         n.TargetTrack,
		RoleStandardVersion: n.RoleStandardVersion,
		ScenarioVersion:     n.ScenarioVersion,
		RubricVersion:       n.RubricVersion,
		PolicyVersion:       n.PolicyVersion,
		ScenarioID:          &scenario,
		Stage:               "baseline",
		Status:              workflow.Assigned,
		Answers:             json.RawMessage(`{}`),
	}
	f.items[id] = a
	return a, nil
}

func (f *fakeAttempts) GetAttempt(ctx context.Context, id string) (store.Attempt, error) {
	a, ok := f.items[id]
	if !ok {
		return store.Attempt{}, store.ErrNotFound
	}
	return a, nil
}

func (f *fakeAttempts) ListCandidateAttempts(ctx context.Context, candidateID string) ([]store.Attempt, error) {
	var out []store.Attempt
	for _, a := range f.items {
		if a.CandidateID == candidateID {
			out = append(out, a)
		}
	}
	return out, nil
}

func (f *fakeAttempts) ListReviewQueue(ctx context.Context) ([]store.Attempt, error) {
	var out []store.Attempt
	for _, a := range f.items {
		switch a.Status {
		case workflow.Submitted, workflow.UnderReview, workflow.ReviewPending:
			out = append(out, a)
		}
	}
	return out, nil
}

func (f *fakeAttempts) SaveAnswers(ctx context.Context, attemptID, candidateID string, answers json.RawMessage) (store.Attempt, error) {
	a, ok := f.items[attemptID]
	open := a.Status == workflow.Assigned || a.Status == workflow.InProgress
	if !ok || a.CandidateID != candidateID || !open {
		return store.Attempt{}, store.ErrNotEditable
	}
	a.Answers = answers
	f.items[attemptID] = a
	return a, nil
}

func (f *fakeAttempts) TransitionAttempt(ctx context.Context, attemptID string, to workflow.State, actor store.Actor) (store.Attempt, error) {
	a, ok := f.items[attemptID]
	if !ok {
		return store.Attempt{}, store.ErrNotFound
	}
	if err := workflow.CheckTransition(a.Status, to, actor.Role); err != nil {
		return store.Attempt{}, err
	}
	a.Status = to
	f.items[attemptID] = a
	return a, nil
}

func newTestRouter() (http.Handler, *fakeAttempts) {
	fake := newFakeAttempts()
	mux := http.NewServeMux()
	registerAttemptRoutes(mux, fakeAuth{}, fake)
	return mux, fake
}

func call(t *testing.T, h http.Handler, method, path, user, role, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if user != "" {
		req.Header.Set("X-Test-User", user)
		req.Header.Set("X-Test-Role", role)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func createBody(candidate string) string {
	return fmt.Sprintf(`{"candidate_id":%q,"target_track":"operator","role_standard_version":"1.0.0",
		"scenario_version":"1.1.0","rubric_version":"0.2.0","policy_version":"1.0.0","scenario_id":"CS-01-A"}`, candidate)
}

func mustCreate(t *testing.T, h http.Handler, candidate string) string {
	t.Helper()

	rec := call(t, h, http.MethodPost, "/v1/attempts", reviewerID, "reviewer", createBody(candidate))
	if rec.Code != http.StatusCreated {
		t.Fatalf("create attempt: status %d, body %s", rec.Code, rec.Body.String())
	}
	var view attemptView
	if err := json.Unmarshal(rec.Body.Bytes(), &view); err != nil {
		t.Fatal(err)
	}
	return view.ID
}

func expectStatus(t *testing.T, rec *httptest.ResponseRecorder, want int, what string) {
	t.Helper()
	if rec.Code != want {
		t.Fatalf("%s: expected %d, got %d (%s)", what, want, rec.Code, rec.Body.String())
	}
}

func TestAttemptRoutesNeedALogin(t *testing.T) {
	h, _ := newTestRouter()

	expectStatus(t, call(t, h, http.MethodGet, "/v1/attempts", "", "", ""), http.StatusUnauthorized, "no login")
}

func TestOnlyStaffCanCreateAttempts(t *testing.T) {
	h, _ := newTestRouter()

	rec := call(t, h, http.MethodPost, "/v1/attempts", candidateA, "candidate", createBody(candidateA))
	expectStatus(t, rec, http.StatusForbidden, "candidate creating an attempt")

	rec = call(t, h, http.MethodPost, "/v1/attempts", reviewerID, "employer", createBody(candidateA))
	expectStatus(t, rec, http.StatusForbidden, "employer creating an attempt")

	rec = call(t, h, http.MethodPost, "/v1/attempts", reviewerID, "reviewer", `{"candidate_id":"not-a-uuid"}`)
	expectStatus(t, rec, http.StatusUnprocessableEntity, "bad candidate id and missing fields")

	rec = call(t, h, http.MethodPost, "/v1/attempts", reviewerID, "reviewer", `{"surprise":true}`)
	expectStatus(t, rec, http.StatusBadRequest, "unknown fields are rejected")

	bad := strings.Replace(createBody(candidateA), `"scenario_id":"CS-01-A"`, `"scenario_id":" "`, 1)
	rec = call(t, h, http.MethodPost, "/v1/attempts", adminID, "admin", bad)
	expectStatus(t, rec, http.StatusUnprocessableEntity, "blank scenario id")

	mustCreate(t, h, candidateA)
}

func TestCandidatesSeeOnlyTheirOwnAttempts(t *testing.T) {
	h, _ := newTestRouter()
	id := mustCreate(t, h, candidateA)

	expectStatus(t, call(t, h, http.MethodGet, "/v1/attempts/"+id, candidateA, "candidate", ""), http.StatusOK, "owner reads own attempt")
	expectStatus(t, call(t, h, http.MethodGet, "/v1/attempts/"+id, candidateB, "candidate", ""), http.StatusNotFound, "another candidate reads it")
	expectStatus(t, call(t, h, http.MethodGet, "/v1/attempts/"+id, adminID, "admin", ""), http.StatusOK, "admin reads it")

	rec := call(t, h, http.MethodGet, "/v1/attempts", candidateB, "candidate", "")
	expectStatus(t, rec, http.StatusOK, "list for another candidate")
	if !strings.Contains(rec.Body.String(), `"attempts":[]`) {
		t.Fatalf("another candidate must see an empty list, got %s", rec.Body.String())
	}

	expectStatus(t, call(t, h, http.MethodGet, "/v1/attempts/not-a-uuid", candidateA, "candidate", ""), http.StatusNotFound, "bad id")
}

func TestReviewersSeeAttemptsOnlyFromSubmission(t *testing.T) {
	h, _ := newTestRouter()
	id := mustCreate(t, h, candidateA)

	expectStatus(t, call(t, h, http.MethodGet, "/v1/attempts/"+id, reviewerID, "reviewer", ""), http.StatusNotFound, "reviewer before submission")

	expectStatus(t, call(t, h, http.MethodPost, "/v1/attempts/"+id+"/start", candidateA, "candidate", ""), http.StatusOK, "start")
	expectStatus(t, call(t, h, http.MethodPost, "/v1/attempts/"+id+"/submit", candidateA, "candidate", ""), http.StatusOK, "submit")

	expectStatus(t, call(t, h, http.MethodGet, "/v1/attempts/"+id, reviewerID, "reviewer", ""), http.StatusOK, "reviewer after submission")

	rec := call(t, h, http.MethodGet, "/v1/attempts", reviewerID, "reviewer", "")
	expectStatus(t, rec, http.StatusOK, "review queue")
	if !strings.Contains(rec.Body.String(), id) {
		t.Fatalf("the submitted attempt should be in the review queue, got %s", rec.Body.String())
	}
}

func TestCandidateFlowAndLocks(t *testing.T) {
	h, _ := newTestRouter()
	id := mustCreate(t, h, candidateA)
	answers := `{"answers":{"q1":"hello"}}`

	expectStatus(t, call(t, h, http.MethodPut, "/v1/attempts/"+id+"/answers", candidateA, "candidate", answers), http.StatusOK, "save answers")
	expectStatus(t, call(t, h, http.MethodPut, "/v1/attempts/"+id+"/answers", candidateA, "candidate", `{"answers":"text"}`), http.StatusUnprocessableEntity, "answers must be an object")
	expectStatus(t, call(t, h, http.MethodPut, "/v1/attempts/"+id+"/answers", candidateB, "candidate", answers), http.StatusConflict, "another candidate saves answers")

	// Skipping a step is refused.
	expectStatus(t, call(t, h, http.MethodPost, "/v1/attempts/"+id+"/submit", candidateA, "candidate", ""), http.StatusConflict, "submit before start")

	expectStatus(t, call(t, h, http.MethodPost, "/v1/attempts/"+id+"/start", candidateB, "candidate", ""), http.StatusNotFound, "another candidate starts it")
	expectStatus(t, call(t, h, http.MethodPost, "/v1/attempts/"+id+"/start", candidateA, "candidate", ""), http.StatusOK, "start")
	expectStatus(t, call(t, h, http.MethodPost, "/v1/attempts/"+id+"/submit", candidateA, "candidate", ""), http.StatusOK, "submit")

	// Locked after submit.
	expectStatus(t, call(t, h, http.MethodPut, "/v1/attempts/"+id+"/answers", candidateA, "candidate", answers), http.StatusConflict, "answers after submit")

	// Only an admin can void.
	expectStatus(t, call(t, h, http.MethodPost, "/v1/attempts/"+id+"/void", candidateA, "candidate", ""), http.StatusForbidden, "candidate voids")
	expectStatus(t, call(t, h, http.MethodPost, "/v1/attempts/"+id+"/void", reviewerID, "reviewer", ""), http.StatusForbidden, "reviewer voids")

	rec := call(t, h, http.MethodPost, "/v1/attempts/"+id+"/void", adminID, "admin", "")
	expectStatus(t, rec, http.StatusOK, "admin voids")
	if !strings.Contains(rec.Body.String(), `"status":"voided"`) {
		t.Fatalf("attempt should be voided, got %s", rec.Body.String())
	}

	// Voided is final.
	expectStatus(t, call(t, h, http.MethodPost, "/v1/attempts/"+id+"/start", candidateA, "candidate", ""), http.StatusConflict, "start after void")
}

func TestAttemptRoutesAnswer503WithoutADatabase(t *testing.T) {
	mux := http.NewServeMux()
	registerAttemptRoutes(mux, fakeAuth{}, nil)

	rec := call(t, mux, http.MethodGet, "/v1/attempts", candidateA, "candidate", "")
	expectStatus(t, rec, http.StatusServiceUnavailable, "no database")
}
