package httpapi

import (
	"context"
	"log"
	"net/http"

	"github.com/s-chukwu/Praxis/backend/internal/aiclient"
	"github.com/s-chukwu/Praxis/backend/internal/store"
	"github.com/s-chukwu/Praxis/backend/internal/workflow"
)

// Fixed pilot settings for the baseline attempt, used when the catalog does not
// name a version. The scenario version always comes from the catalog.
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
// The candidate then opens it with POST /v1/attempts/{id}/start.
func registerBaselineStartRoute(mux *http.ServeMux, authn Authenticator, catalog CatalogReader, starter BaselineStarter, scenarioID string) {
	h := &baselineStartHandlers{
		starter:  starter,
		scenario: &baselineHandlers{catalog: catalog, scenarioID: scenarioID},
	}
	mux.Handle("POST /v1/attempts/baseline", requireRole(authn, "candidate")(http.HandlerFunc(h.start)))
}

type startBaselineRequest struct {
	Program string `json:"program"`
}

func orDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
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

	// The body is optional so older callers keep working; without it the track is operator.
	body := startBaselineRequest{Program: baselineTrack}
	if r.ContentLength != 0 && !decodeBody(w, r, &body) {
		return
	}
	body.Program = orDefault(body.Program, baselineTrack)
	if _, known := trackCompetencies[body.Program]; !known {
		writeError(w, http.StatusUnprocessableEntity, "invalid_input", "program must be operator or builder")
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
		TargetTrack:         body.Program,
		RoleStandardVersion: orDefault(s.RoleStandardVersion, baselineRoleStandardVersion),
		ScenarioVersion:     s.Version,
		RubricVersion:       orDefault(s.RubricVersion, baselineRubricVersion),
		PolicyVersion:       orDefault(s.PolicyVersion, baselinePolicyVersion),
		ScenarioID:          s.ScenarioID,
	}, actorOf(caller))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if attempt.Status == workflow.Voided {
		writeError(w, http.StatusConflict, "baseline_voided",
			"your baseline task was closed by an admin, so contact your facilitator")
		return
	}

	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	writeJSON(w, status, toAttemptView(attempt, true))
}
