package httpapi

import (
	"context"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/s-chukwu/Praxis/backend/internal/aiclient"
	"github.com/s-chukwu/Praxis/backend/internal/store"
	"github.com/s-chukwu/Praxis/backend/internal/workflow"
)

var courseCodePattern = regexp.MustCompile(`^M-[A-Z-]{2,20}$`)

// CourseStore is the part of the database layer the course and final-task routes need.
type CourseStore interface {
	CompleteCourse(ctx context.Context, candidateID, courseCode string, actor store.Actor) (time.Time, error)
	PlanProgress(ctx context.Context, candidateID string) (total, completed int, err error)
}

type finalHandlers struct {
	attempts AttemptStore
	courses  CourseStore
	catalog  CatalogReader
}

// registerCourseRoutes wires course completion and the final task. The final task opens only after the
// baseline is reviewed and every course in the plan is complete (PM decision X-02).
func registerCourseRoutes(mux *http.ServeMux, authn Authenticator, attempts AttemptStore, courses CourseStore, catalog CatalogReader) {
	h := &finalHandlers{attempts: attempts, courses: courses, catalog: catalog}
	candidate := requireRole(authn, "candidate")
	mux.Handle("POST /v1/me/plan/courses/{code}/complete", candidate(http.HandlerFunc(h.complete)))
	mux.Handle("GET /v1/final/scenario", candidate(http.HandlerFunc(h.scenario)))
	mux.Handle("POST /v1/attempts/final", candidate(http.HandlerFunc(h.start)))
}

func (h *finalHandlers) ready(w http.ResponseWriter) bool {
	if h.attempts == nil || h.courses == nil {
		writeError(w, http.StatusServiceUnavailable, "database_not_configured", "the database is not configured on this server")
		return false
	}
	return true
}

func (h *finalHandlers) complete(w http.ResponseWriter, r *http.Request) {
	if !h.ready(w) {
		return
	}
	caller, ok := callerOf(w, r)
	if !ok {
		return
	}
	code := r.PathValue("code")
	if !courseCodePattern.MatchString(code) {
		writeError(w, http.StatusNotFound, "not_found", "course not found in your plan")
		return
	}
	done, err := h.courses.CompleteCourse(r.Context(), caller.UserID, code, actorOf(caller))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"course_code": code, "completed_at": done})
}

// eligible returns the reviewed baseline when the final task may open, or writes the reason it may not.
func (h *finalHandlers) eligible(w http.ResponseWriter, r *http.Request, candidateID string) (*store.Attempt, []store.Attempt, bool) {
	all, err := h.attempts.ListCandidateAttempts(r.Context(), candidateID)
	if err != nil {
		writeStoreError(w, err)
		return nil, nil, false
	}
	var baseline *store.Attempt
	for i := range all {
		if all[i].Stage == "baseline" && all[i].Status == workflow.Reviewed {
			baseline = &all[i]
		}
	}
	total, completed, err := h.courses.PlanProgress(r.Context(), candidateID)
	if err != nil {
		writeStoreError(w, err)
		return nil, nil, false
	}
	if baseline == nil || baseline.ScenarioID == nil || total == 0 || completed < total {
		writeError(w, http.StatusConflict, "final_locked", "the final task opens after your review and every course in your plan")
		return nil, nil, false
	}
	return baseline, all, true
}

func (h *finalHandlers) loadTwin(w http.ResponseWriter, r *http.Request, caller string, baseline *store.Attempt) (baselineScenario, bool) {
	if h.catalog == nil {
		writeError(w, http.StatusServiceUnavailable, "ai_not_configured", "the AI service is not configured on this server")
		return baselineScenario{}, false
	}
	raw, err := h.catalog.Catalog(r.Context(), aiclient.CandidateActor(caller), "")
	if err == nil {
		var s baselineScenario
		if s, err = pickFinal(raw, *baseline.ScenarioID, baseline.ScenarioVersion); err == nil {
			return s, true
		}
	}
	log.Printf("final scenario for %s could not be loaded: %v", *baseline.ScenarioID, err)
	writeError(w, http.StatusBadGateway, "scenario_unavailable", "the final task could not be loaded right now")
	return baselineScenario{}, false
}

func (h *finalHandlers) scenario(w http.ResponseWriter, r *http.Request) {
	if !h.ready(w) {
		return
	}
	caller, ok := callerOf(w, r)
	if !ok {
		return
	}
	baseline, _, ok := h.eligible(w, r, caller.UserID)
	if !ok {
		return
	}
	if s, ok := h.loadTwin(w, r, caller.UserID, baseline); ok {
		writeJSON(w, http.StatusOK, s)
	}
}

// start creates the final attempt on the baseline's pinned versions, so the two are comparable, and opens it.
func (h *finalHandlers) start(w http.ResponseWriter, r *http.Request) {
	if !h.ready(w) {
		return
	}
	caller, ok := callerOf(w, r)
	if !ok {
		return
	}
	baseline, all, ok := h.eligible(w, r, caller.UserID)
	if !ok {
		return
	}
	for i := range all {
		if all[i].Stage == "final" && all[i].Status != workflow.Voided {
			writeJSON(w, http.StatusOK, toAttemptView(all[i], true))
			return
		}
	}
	twin, ok := h.loadTwin(w, r, caller.UserID, baseline)
	if !ok {
		return
	}
	created, err := h.attempts.CreateAttempt(r.Context(), store.NewAttempt{
		CandidateID: caller.UserID, TargetTrack: baseline.TargetTrack, RoleStandardVersion: baseline.RoleStandardVersion,
		ScenarioVersion: baseline.ScenarioVersion, RubricVersion: baseline.RubricVersion, PolicyVersion: baseline.PolicyVersion,
		ScenarioID: twin.ScenarioID, Stage: "final",
	}, actorOf(caller))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	opened, err := h.attempts.TransitionAttempt(r.Context(), created.ID, workflow.InProgress, actorOf(caller))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toAttemptView(opened, true))
}
