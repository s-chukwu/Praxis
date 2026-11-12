package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/s-chukwu/Praxis/backend/internal/aiclient"
	"github.com/s-chukwu/Praxis/backend/internal/store"
)

// PlanStore is what the learning plan routes need from the database layer.
// *store.Store satisfies it, and tests use a fake.
type PlanStore interface {
	CreatePlan(ctx context.Context, n store.NewPlan, actor store.Actor) (store.Plan, bool, error)
	GetPlanByReview(ctx context.Context, reviewID string) (store.Plan, error)
	ListVerifiedModules(ctx context.Context, candidateID string) ([]string, error)
}

// Planner is the part of the AI client the plan route needs.
type Planner interface {
	LearningPlan(ctx context.Context, actor aiclient.Actor, requestID string, req any) (json.RawMessage, error)
}

type planHandlers struct {
	reviews ReviewStore
	plans   PlanStore
	planner Planner
}

// registerPlanRoutes wires the learning plan routes. The reviewer who finalized a
// review asks for its plan. The candidate reads their own plan once it exists.
func registerPlanRoutes(mux *http.ServeMux, authn Authenticator, reviews ReviewStore, plans PlanStore, planner Planner) {
	h := &planHandlers{reviews: reviews, plans: plans, planner: planner}

	signedIn := requireRole(authn)
	reviewer := requireRole(authn, "reviewer")

	mux.Handle("POST /v1/reviews/{id}/learning-plan", reviewer(http.HandlerFunc(h.create)))
	mux.Handle("GET /v1/reviews/{id}/learning-plan", signedIn(http.HandlerFunc(h.get)))
}

type planRequest struct {
	RequestID                string          `json:"request_id"`
	CandidateID              string          `json:"candidate_id"`
	ReviewID                 string          `json:"review_id"`
	TargetTrack              string          `json:"target_track"`
	RoleStandardVersion      string          `json:"role_standard_version"`
	PolicyVersion            string          `json:"policy_version"`
	ReviewedCompetencies     json.RawMessage `json:"reviewed_competencies"`
	CriticalSafetyFinding    bool            `json:"critical_safety_finding"`
	VerifiedCompletedModules []string        `json:"verified_completed_modules,omitempty"`
}

type planRef struct {
	Version string `json:"version"`
	SHA256  string `json:"sha256"`
}

type planEnvelope struct {
	RoleStandard planRef         `json:"role_standard"`
	Policy       planRef         `json:"policy"`
	Curriculum   planRef         `json:"curriculum"`
	Result       json.RawMessage `json:"result"`
}

type safetyGate struct {
	Required bool `json:"required"`
}

type planModuleResult struct {
	Order               int             `json:"order"`
	CourseID            string          `json:"course_id"`
	Title               string          `json:"title"`
	Covers              []string        `json:"covers"`
	Reasons             json.RawMessage `json:"reasons"`
	PreviouslyCompleted bool            `json:"previously_completed"`
}

type planResult struct {
	Status               string             `json:"status"`
	MissingCompetencies  []string           `json:"missing_competencies"`
	SafetyGate           safetyGate         `json:"safety_gate"`
	Modules              []planModuleResult `json:"modules"`
	ReassessmentRequired bool               `json:"reassessment_required"`
}

type planModuleView struct {
	Position            int             `json:"position"`
	CourseCode          string          `json:"course_code"`
	Title               string          `json:"title"`
	Covers              []string        `json:"covers"`
	Reasons             json.RawMessage `json:"reasons"`
	PreviouslyCompleted bool            `json:"previously_completed"`
	CompletedAt         *time.Time      `json:"completed_at"`
	VerifiedAt          *time.Time      `json:"verified_at"`
}

type planView struct {
	ID                   string           `json:"id"`
	CandidateID          string           `json:"candidate_id"`
	ReviewID             string           `json:"review_id"`
	TargetTrack          string           `json:"target_track"`
	Status               string           `json:"status"`
	RoleStandardVersion  string           `json:"role_standard_version"`
	PolicyVersion        string           `json:"policy_version"`
	CurriculumVersion    string           `json:"curriculum_version"`
	SafetyGateRequired   bool             `json:"safety_gate_required"`
	ReassessmentRequired bool             `json:"reassessment_required"`
	Modules              []planModuleView `json:"modules"`
	CreatedAt            time.Time        `json:"created_at"`
}

func toPlanView(p store.Plan) planView {
	view := planView{
		ID:                   p.ID,
		CandidateID:          p.CandidateID,
		ReviewID:             p.ReviewID,
		TargetTrack:          p.TargetTrack,
		Status:               p.Status,
		RoleStandardVersion:  p.RoleStandardVersion,
		PolicyVersion:        p.PolicyVersion,
		CurriculumVersion:    p.CurriculumVersion,
		SafetyGateRequired:   p.SafetyGateRequired,
		ReassessmentRequired: p.ReassessmentRequired,
		Modules:              make([]planModuleView, 0, len(p.Modules)),
		CreatedAt:            p.CreatedAt,
	}
	for _, m := range p.Modules {
		covers := m.Covers
		if covers == nil {
			covers = []string{}
		}
		reasons := m.Reasons
		if len(reasons) == 0 {
			reasons = json.RawMessage(`[]`)
		}
		view.Modules = append(view.Modules, planModuleView{
			Position:            m.Position,
			CourseCode:          m.CourseCode,
			Title:               m.Title,
			Covers:              covers,
			Reasons:             reasons,
			PreviouslyCompleted: m.PreviouslyCompleted,
			CompletedAt:         m.CompletedAt,
			VerifiedAt:          m.VerifiedAt,
		})
	}
	return view
}

func (h *planHandlers) ready(w http.ResponseWriter) bool {
	if h.reviews == nil || h.plans == nil {
		writeError(w, http.StatusServiceUnavailable, "database_not_configured",
			"the database is not configured on this server")
		return false
	}
	return true
}

// create builds a learning plan from a finalized review. It sends the reviewer's
// competency scores to the AI service and stores the plan it returns. Asking twice
// for the same review returns the same plan without calling the AI service again.
func (h *planHandlers) create(w http.ResponseWriter, r *http.Request) {
	if !h.ready(w) {
		return
	}
	if h.planner == nil {
		writeError(w, http.StatusServiceUnavailable, "ai_not_configured",
			"the AI service is not configured on this server")
		return
	}
	caller, ok := callerOf(w, r)
	if !ok {
		return
	}
	reviewID, ok := reviewPathID(w, r)
	if !ok {
		return
	}
	ctx := r.Context()

	review, err := h.reviews.GetReview(ctx, reviewID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if review.ReviewerID != caller.UserID {
		writeError(w, http.StatusNotFound, "not_found", "review not found")
		return
	}
	if review.FinalizedAt == nil {
		writeError(w, http.StatusConflict, "review_not_finalized",
			"a learning plan can only be created from a finalized review")
		return
	}

	existing, err := h.plans.GetPlanByReview(ctx, reviewID)
	if err == nil {
		writeJSON(w, http.StatusOK, toPlanView(existing))
		return
	}
	if !errors.Is(err, store.ErrNotFound) {
		writeStoreError(w, err)
		return
	}

	attempt, err := h.reviews.GetAttempt(ctx, review.AttemptID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	verified, err := h.plans.ListVerifiedModules(ctx, attempt.CandidateID)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	requestID := newRequestID()
	raw, err := h.planner.LearningPlan(ctx, aiclient.ReviewerActor(caller.UserID), requestID, planRequest{
		RequestID:                requestID,
		CandidateID:              attempt.CandidateID,
		ReviewID:                 review.ID,
		TargetTrack:              attempt.TargetTrack,
		RoleStandardVersion:      attempt.RoleStandardVersion,
		PolicyVersion:            attempt.PolicyVersion,
		ReviewedCompetencies:     review.ReviewedCompetencies,
		CriticalSafetyFinding:    review.CriticalSafetyFinding,
		VerifiedCompletedModules: verified,
	})
	if err != nil {
		writeAIError(w, err)
		return
	}

	var envelope planEnvelope
	var result planResult
	if err := json.Unmarshal(raw, &envelope); err != nil || json.Unmarshal(envelope.Result, &result) != nil {
		log.Printf("unreadable learning plan response (review %s)", reviewID)
		writeError(w, http.StatusBadGateway, "ai_unavailable", "the AI service sent an answer we could not read")
		return
	}

	switch result.Status {
	case "plan_ready":
	case "diagnostic_incomplete":
		writeJSON(w, http.StatusConflict, map[string]any{"error": map[string]any{
			"code":    "diagnostic_incomplete",
			"message": "the review does not cover every competency of this track, so no plan was created",
			"details": map[string]any{"missing_competencies": result.MissingCompetencies},
		}})
		return
	default:
		log.Printf("unknown learning plan status %q (review %s)", result.Status, reviewID)
		writeError(w, http.StatusBadGateway, "ai_unavailable", "the AI service sent an unexpected plan status")
		return
	}

	modules := make([]store.NewPlanModule, 0, len(result.Modules))
	for _, m := range result.Modules {
		modules = append(modules, store.NewPlanModule{
			Position:            m.Order,
			CourseCode:          m.CourseID,
			Title:               m.Title,
			Covers:              m.Covers,
			Reasons:             m.Reasons,
			PreviouslyCompleted: m.PreviouslyCompleted,
		})
	}

	plan, created, err := h.plans.CreatePlan(ctx, store.NewPlan{
		CandidateID:          attempt.CandidateID,
		ReviewID:             review.ID,
		RequestID:            requestID,
		TargetTrack:          attempt.TargetTrack,
		Status:               result.Status,
		RoleStandardVersion:  envelope.RoleStandard.Version,
		PolicyVersion:        envelope.Policy.Version,
		CurriculumVersion:    envelope.Curriculum.Version,
		RoleStandardSHA256:   envelope.RoleStandard.SHA256,
		PolicySHA256:         envelope.Policy.SHA256,
		CurriculumSHA256:     envelope.Curriculum.SHA256,
		SafetyGateRequired:   result.SafetyGate.Required,
		ReassessmentRequired: result.ReassessmentRequired,
		Result:               envelope.Result,
		Modules:              modules,
	}, actorOf(caller))
	if err != nil {
		writeStoreError(w, err)
		return
	}

	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	writeJSON(w, status, toPlanView(plan))
}

// get returns the learning plan of a review. It follows the same visibility rule as
// the review: the reviewer who owns it, an admin, or the candidate once it is finalized.
func (h *planHandlers) get(w http.ResponseWriter, r *http.Request) {
	if !h.ready(w) {
		return
	}
	caller, ok := callerOf(w, r)
	if !ok {
		return
	}
	reviewID, ok := reviewPathID(w, r)
	if !ok {
		return
	}
	ctx := r.Context()

	review, err := h.reviews.GetReview(ctx, reviewID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if !(&reviewHandlers{reviews: h.reviews}).canViewReview(ctx, caller, review) {
		writeError(w, http.StatusNotFound, "not_found", "review not found")
		return
	}

	plan, err := h.plans.GetPlanByReview(ctx, reviewID)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "no_plan", "there is no learning plan for this review yet")
		return
	}
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toPlanView(plan))
}
