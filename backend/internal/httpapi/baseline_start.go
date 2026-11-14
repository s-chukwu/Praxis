package httpapi

import (
	"log"
	"net/http"

	"github.com/s-chukwu/Praxis/backend/internal/aiclient"
	"github.com/s-chukwu/Praxis/backend/internal/store"
	"github.com/s-chukwu/Praxis/backend/internal/workflow"
)

type startBaselineRequest struct {
	Program string `json:"program"`
}

// start gives the candidate their baseline attempt, opened and ready for answers.
// Calling it again returns the same attempt.
func (h *baselineHandlers) start(w http.ResponseWriter, r *http.Request) {
	if h.attempts == nil {
		writeError(w, http.StatusServiceUnavailable, "database_not_configured",
			"the database is not configured on this server")
		return
	}
	if h.catalog == nil {
		writeError(w, http.StatusServiceUnavailable, "ai_not_configured",
			"the AI service is not configured on this server")
		return
	}
	caller, ok := callerOf(w, r)
	if !ok {
		return
	}
	var body startBaselineRequest
	if !decodeBody(w, r, &body) {
		return
	}
	if _, known := trackCompetencies[body.Program]; !known {
		writeError(w, http.StatusUnprocessableEntity, "invalid_input", "program must be operator or builder")
		return
	}

	s, err := h.load(r.Context(), aiclient.CandidateActor(caller.UserID))
	if err != nil {
		log.Printf("baseline start could not load scenario %q: %v", h.scenarioID, err)
		writeError(w, http.StatusBadGateway, "scenario_unavailable", "the baseline task could not be loaded right now")
		return
	}
	// The database allows one baseline per candidate, so this is safe across servers and double clicks.
	got, created, err := h.attempts.CreateOrGetBaselineAttempt(r.Context(), store.NewAttempt{
		CandidateID:         caller.UserID,
		TargetTrack:         body.Program,
		RoleStandardVersion: s.RoleStandardVersion,
		ScenarioVersion:     s.Version,
		RubricVersion:       s.RubricVersion,
		PolicyVersion:       s.PolicyVersion,
		ScenarioID:          s.ScenarioID,
		Stage:               "baseline",
	}, actorOf(caller))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if got.Status == workflow.Voided {
		writeError(w, http.StatusConflict, "baseline_voided", "your baseline task was closed by an admin, so contact your facilitator")
		return
	}
	attempt := &got
	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}

	if attempt.Status == workflow.Assigned {
		opened, err := h.attempts.TransitionAttempt(r.Context(), attempt.ID, workflow.InProgress, actorOf(caller))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		attempt = &opened
	}
	writeJSON(w, status, toAttemptView(*attempt, true))
}
