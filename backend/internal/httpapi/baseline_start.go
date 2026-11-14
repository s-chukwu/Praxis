package httpapi

import (
	"context"
	"log"
	"net/http"

	"github.com/s-chukwu/Praxis/backend/internal/aiclient"
	"github.com/s-chukwu/Praxis/backend/internal/store"
)

// Fixed pilot settings for the baseline attempt. The scenario version comes from
// the catalog; these three are the versions the baseline is graded against.
const (
	baselineTrack               = "operator"
	baselineRoleStandardVersion = "1.0.0"
	baselineRubricVersion       = "0.2.0"
	baselinePolicyVersion       = "1.0.0"
)

// BaselineStarter is what the start route needs from the database layer.
// It is separate from AttemptStore so existing fakes in tests keep working.
type BaselineStarter interface {
	CreateOrGetBaselineAttempt(ctx context.Context, n store.NewAttempt, actor store.Actor) (store.Attempt, bool, error)
}

type baselineStartHandlers struct {
	starter  BaselineStarter
	scenario *baselineHandlers
}

// registerBaselineStartRoute wires POST /v1/attempts/baseline for candidates.

// It creates the candidate's one baseline attempt, or returns the existing one.
func registerBaselineStartRoute(mux *http.ServeMux, authn Authenticator, catalog CatalogReader, starter BaselineStarter, scenarioID string) {
	h := &baselineStartHandlers{
		starter:  starter,
		scenario: &baselineHandlers{catalog: catalog, scenarioID: scenarioID},
	}
	mux.Handle("POST /v1/attempts/baseline", requireRole(authn, "candidate")(http.HandlerFunc(h.start)))
}

func (h *baselineStartHandlers) start(w http.ResponseWriter, r *http.Request) {
	if h.starter == nil {
		writeError(w, http.StatusServiceUnavailable, "database_not_configured",
			"the database is not configured on this server")
		return
	}
	if h.scenario.catalog == nil {
		writeError(w, http.StatusServiceUnavailable, "ai_not_configured",
			"the AI service is not configured on this server")
		return
	}
	caller, ok := callerOf(w, r)
	if !ok {
		return
	}

	s, err := h.scenario.load(r.Context(), aiclient.CandidateActor(caller.UserID))
	if err != nil {
		log.Printf("baseline scenario %q could not be loaded: %v", h.scenario.scenarioID, err)
		writeError(w, http.StatusBadGateway, "scenario_unavailable",
			"the baseline task could not be loaded right now")
		return
	}


	attempt, created, err := h.starter.CreateOrGetBaselineAttempt(r.Context(), store.NewAttempt{
		CandidateID:         caller.UserID,
		TargetTrack:         baselineTrack,
		RoleStandardVersion: baselineRoleStandardVersion,
		ScenarioVersion:     s.Version,
		RubricVersion:       baselineRubricVersion,
		PolicyVersion:       baselinePolicyVersion,
		ScenarioID:          s.ScenarioID,
	}, actorOf(caller))
	if err != nil {
		writeStoreError(w, err)
		return
	}

	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	writeJSON(w, status, toAttemptView(attempt, true))
}
