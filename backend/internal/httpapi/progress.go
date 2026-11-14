package httpapi

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/s-chukwu/Praxis/backend/internal/store"
	"github.com/s-chukwu/Praxis/backend/internal/workflow"
)

// ProgressReviews is the part of the database layer the progress route needs for results.
type ProgressReviews interface {
	GetFinalizedReview(ctx context.Context, attemptID string) (store.Review, error)
	ListDimensionScores(ctx context.Context, reviewID string) ([]store.DimensionScore, error)
}

type progressAttempt struct {
	AttemptID   string     `json:"attempt_id"`
	Status      string     `json:"status"`
	SubmittedAt *time.Time `json:"submitted_at"`
}

type progressScore struct {
	Dimension string `json:"dimension"`
	Score     int    `json:"score"`
}

type progressResult struct {
	ReviewID        string          `json:"review_id"`
	WeightedScore   *float64        `json:"weighted_score"`
	ProposedOutcome *string         `json:"proposed_outcome"`
	Scores          []progressScore `json:"scores"`
	FinalizedAt     *time.Time      `json:"finalized_at"`
}

// progressView is everything the candidate's dashboard needs in one call.
// Each part is null until that step has happened.
type progressView struct {
	Baseline *progressAttempt `json:"baseline"`
	Result   *progressResult  `json:"result"`
	Plan     *planView        `json:"plan"`
}

type progressHandlers struct {
	attempts AttemptStore
	reviews  ProgressReviews
	plans    PlanStore
}

func registerProgressRoute(mux *http.ServeMux, authn Authenticator, attempts AttemptStore, reviews ProgressReviews, plans PlanStore) {
	h := &progressHandlers{attempts: attempts, reviews: reviews, plans: plans}
	mux.Handle("GET /v1/me/progress", requireRole(authn, "candidate")(http.HandlerFunc(h.get)))
}

func (h *progressHandlers) get(w http.ResponseWriter, r *http.Request) {
	if h.attempts == nil || h.reviews == nil || h.plans == nil {
		writeError(w, http.StatusServiceUnavailable, "database_not_configured",
			"the database is not configured on this server")
		return
	}
	caller, ok := callerOf(w, r)
	if !ok {
		return
	}
	ctx := r.Context()

	attempts, err := h.attempts.ListCandidateAttempts(ctx, caller.UserID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	view := progressView{}
	var baseline *store.Attempt
	for i := range attempts {
		if attempts[i].Stage == "baseline" && attempts[i].Status != workflow.Voided {
			baseline = &attempts[i]
			break
		}
	}
	if baseline == nil {
		writeJSON(w, http.StatusOK, view)
		return
	}
	view.Baseline = &progressAttempt{AttemptID: baseline.ID, Status: string(baseline.Status), SubmittedAt: baseline.SubmittedAt}

	// A result only counts while the attempt is reviewed; an admin reopening it hides the old one.
	if baseline.Status != workflow.Reviewed {
		writeJSON(w, http.StatusOK, view)
		return
	}
	review, err := h.reviews.GetFinalizedReview(ctx, baseline.ID)
	if errors.Is(err, store.ErrNotFound) {
		writeJSON(w, http.StatusOK, view)
		return
	}
	if err != nil {
		writeStoreError(w, err)
		return
	}
	scores, err := h.reviews.ListDimensionScores(ctx, review.ID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	result := &progressResult{ReviewID: review.ID, WeightedScore: review.WeightedScore,
		ProposedOutcome: review.ProposedOutcome, FinalizedAt: review.FinalizedAt, Scores: make([]progressScore, 0, len(scores))}
	for _, s := range scores {
		result.Scores = append(result.Scores, progressScore{Dimension: s.Dimension, Score: s.Score})
	}
	view.Result = result

	plan, err := h.plans.GetPlanByReview(ctx, review.ID)
	switch {
	case err == nil:
		p := toPlanView(plan)
		view.Plan = &p
	case !errors.Is(err, store.ErrNotFound):
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, view)
}
