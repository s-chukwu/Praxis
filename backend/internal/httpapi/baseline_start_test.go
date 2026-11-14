package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/s-chukwu/Praxis/backend/internal/aiclient"
	"github.com/s-chukwu/Praxis/backend/internal/store"
	"github.com/s-chukwu/Praxis/backend/internal/workflow"
)

// startFakeCatalog returns a fixed scenario catalog, like the AI service would.
type startFakeCatalog struct {
	scenarioID string
}

func (f startFakeCatalog) Catalog(ctx context.Context, actor aiclient.Actor, requestID string) (json.RawMessage, error) {
	return json.RawMessage(`{
		"current": {"scenarios": "1.1.0"},
		"scenario_banks": [{
			"version": "1.1.0",
			"scenarios": [{
				"id": "` + f.scenarioID + `",
				"stage": "baseline",
				"timebox_minutes": 45,
				"expected_artifacts": ["reply"],
				"candidate_brief": "Help the customer."
			}]
		}]

	}`), nil
}

// startFakeStarter behaves like the real store: one baseline attempt per candidate.
type startFakeStarter struct {
	byCandidate map[string]store.Attempt
	count       int
}

func newStartFakeStarter() *startFakeStarter {
	return &startFakeStarter{byCandidate: map[string]store.Attempt{}}
}

func (f *startFakeStarter) CreateOrGetBaselineAttempt(ctx context.Context, n store.NewAttempt, actor store.Actor) (store.Attempt, bool, error) {
	if existing, ok := f.byCandidate[n.CandidateID]; ok {
		return existing, false, nil
	}
	f.count++
	scenario := n.ScenarioID
	a := store.Attempt{
		ID:                  "00000000-0000-0000-0000-" + string(rune('a'+f.count)) + "00000000000",
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

	f.byCandidate[n.CandidateID] = a
	return a, true, nil
}

func newStartRouter(scenarioID string) (http.Handler, *startFakeStarter) {
	starter := newStartFakeStarter()
	mux := http.NewServeMux()
	registerBaselineStartRoute(mux, fakeAuth{}, startFakeCatalog{scenarioID: scenarioID}, starter, "CS-01-A")
	return mux, starter
}

func TestBaselineStartNeedsACandidateLogin(t *testing.T) {
	h, _ := newStartRouter("CS-01-A")

	expectStatus(t, call(t, h, http.MethodPost, "/v1/attempts/baseline", "", "", ""), http.StatusUnauthorized, "no login")
	expectStatus(t, call(t, h, http.MethodPost, "/v1/attempts/baseline", reviewerID, "reviewer", ""), http.StatusForbidden, "reviewer")
	expectStatus(t, call(t, h, http.MethodPost, "/v1/attempts/baseline", adminID, "admin", ""), http.StatusForbidden, "admin")
}

func TestBaselineStartCreatesOnceThenReturnsTheSameAttempt(t *testing.T) {
	h, starter := newStartRouter("CS-01-A")

	// First call creates the attempt.
	rec := call(t, h, http.MethodPost, "/v1/attempts/baseline", candidateA, "candidate", "")
	expectStatus(t, rec, http.StatusCreated, "first start")
	var first attemptView
	if err := json.Unmarshal(rec.Body.Bytes(), &first); err != nil {
		t.Fatal(err)
	}
	if first.Stage != "baseline" || first.Status != "assigned" || first.CandidateID != candidateA {
		t.Fatalf("unexpected attempt: %+v", first)
	}

	if first.ScenarioID == nil || *first.ScenarioID != "CS-01-A" || first.ScenarioVersion != "1.1.0" {
		t.Fatalf("scenario not set from the catalog: %+v", first)
	}

	// The candidate saves some work.
	saved := starter.byCandidate[candidateA]
	saved.Answers = json.RawMessage(`{"q1":"hello"}`)
	starter.byCandidate[candidateA] = saved

	// Second call returns the same attempt with the saved answers.
	rec = call(t, h, http.MethodPost, "/v1/attempts/baseline", candidateA, "candidate", "")
	expectStatus(t, rec, http.StatusOK, "second start")
	var second attemptView
	if err := json.Unmarshal(rec.Body.Bytes(), &second); err != nil {
		t.Fatal(err)
	}
	if second.ID != first.ID {
		t.Fatalf("expected the same attempt, got %s and %s", first.ID, second.ID)
	}
	if !strings.Contains(string(second.Answers), "hello") {
		t.Fatalf("saved answers must be kept, got %s", second.Answers)
	}
	if starter.count != 1 {
		t.Fatalf("only one attempt should exist, got %d", starter.count)
	}
}

func TestBaselineStartIsPerCandidate(t *testing.T) {
	h, _ := newStartRouter("CS-01-A")

	var a, b attemptView
	recA := call(t, h, http.MethodPost, "/v1/attempts/baseline", candidateA, "candidate", "")

	expectStatus(t, recA, http.StatusCreated, "candidate A")
	recB := call(t, h, http.MethodPost, "/v1/attempts/baseline", candidateB, "candidate", "")
	expectStatus(t, recB, http.StatusCreated, "candidate B")

	if err := json.Unmarshal(recA.Body.Bytes(), &a); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(recB.Body.Bytes(), &b); err != nil {
		t.Fatal(err)
	}
	if a.ID == b.ID {
		t.Fatalf("two candidates must get different attempts")
	}
}

func TestBaselineStartFailsCleanlyWhenTheScenarioIsMissing(t *testing.T) {
	// The catalog holds a different scenario than the one configured.
	h, starter := newStartRouter("CS-99-Z")

	rec := call(t, h, http.MethodPost, "/v1/attempts/baseline", candidateA, "candidate", "")
	expectStatus(t, rec, http.StatusBadGateway, "scenario missing")
	if starter.count != 0 {
		t.Fatalf("no attempt may be created when the scenario cannot be loaded")
	}
}

func TestBaselineStartAnswers503WhenNotConfigured(t *testing.T) {
	mux := http.NewServeMux()
	registerBaselineStartRoute(mux, fakeAuth{}, nil, nil, "CS-01-A")

	rec := call(t, mux, http.MethodPost, "/v1/attempts/baseline", candidateA, "candidate", "")
	expectStatus(t, rec, http.StatusServiceUnavailable, "nothing configured")

}
