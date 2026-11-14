package httpapi

import (
	"context"
	"net/http"
	"time"

	"github.com/s-chukwu/Praxis/backend/internal/auth"
	"github.com/s-chukwu/Praxis/backend/internal/store"
)

// ReviewStore is what the review routes need from the database layer.
// *store.Store satisfies it, and tests use a fake.
type ReviewStore interface {
	GetAttempt(ctx context.Context, id string) (store.Attempt, error)
	StartReview(ctx context.Context, attemptID string, reviewer store.Actor) (store.Review, error)
	GetReview(ctx context.Context, id string) (store.Review, error)
	SaveDimensionScores(ctx context.Context, reviewID string, reviewer store.Actor, scores []store.ScoreInput) error
	SaveCriticalFindings(ctx context.Context, reviewID string, reviewer store.Actor, findings []store.FindingInput) error
	ListDimensionScores(ctx context.Context, reviewID string) ([]store.DimensionScore, error)
	ListCriticalFindings(ctx context.Context, reviewID string) ([]store.Finding, error)
	ListReviewerReviews(ctx context.Context, reviewerID string) ([]store.Review, error)
}

// reviewSummary is one row in a reviewer's "My reviews" list.
type reviewSummary struct {
	ID              string     `json:"id"`
	AttemptID       string     `json:"attempt_id"`
	FinalizedAt     *time.Time `json:"finalized_at"`
	WeightedScore   *float64   `json:"weighted_score"`
	ProposedOutcome *string    `json:"proposed_outcome"`
	CreatedAt       time.Time  `json:"created_at"`
}

type reviewHandlers struct {
	reviews ReviewStore
}

// registerReviewRoutes wires the review routes. Only reviewers can start a review
// or change scores and findings, and only on their own draft. Finalizing comes
// later, because it calls the AI service to compute the weighted score.
func registerReviewRoutes(mux *http.ServeMux, authn Authenticator, reviews ReviewStore) {
	h := &reviewHandlers{reviews: reviews}

	signedIn := requireRole(authn)
	reviewer := requireRole(authn, "reviewer")

	mux.Handle("GET /v1/reviews", reviewer(http.HandlerFunc(h.mine)))
	mux.Handle("POST /v1/attempts/{id}/review", reviewer(http.HandlerFunc(h.start)))
	mux.Handle("GET /v1/reviews/{id}", signedIn(http.HandlerFunc(h.get)))
	mux.Handle("PUT /v1/reviews/{id}/scores", reviewer(http.HandlerFunc(h.saveScores)))
	mux.Handle("PUT /v1/reviews/{id}/findings", reviewer(http.HandlerFunc(h.saveFindings)))
}

type scoreBody struct {
	Dimension string   `json:"dimension"`
	Score     int      `json:"score"`
	Evidence  []string `json:"evidence"`
}

type findingBody struct {
	EventType       string `json:"event_type"`
	SupportingQuote string `json:"supporting_quote"`
	PolicyReference string `json:"policy_reference"`
	Decision        string `json:"decision"`
}

type reviewView struct {
	ID                    string        `json:"id"`
	AttemptID             string        `json:"attempt_id"`
	ReviewerID            string        `json:"reviewer_id"`
	FinalizedAt           *time.Time    `json:"finalized_at"`
	CriticalSafetyFinding bool          `json:"critical_safety_finding"`
	Notes                 *string       `json:"notes"`
	WeightedScore         *float64      `json:"weighted_score"`
	ProposedOutcome       *string       `json:"proposed_outcome"`
	OutcomeReasons        []string      `json:"outcome_reasons"`
	Scores                []scoreBody   `json:"scores"`
	Findings              []findingBody `json:"findings"`
	CreatedAt             time.Time     `json:"created_at"`
}

func (h *reviewHandlers) ready(w http.ResponseWriter) bool {
	if h.reviews == nil {
		writeError(w, http.StatusServiceUnavailable, "database_not_configured",
			"the database is not configured on this server")
		return false
	}
	return true
}

func reviewPathID(w http.ResponseWriter, r *http.Request) (string, bool) {
	id := r.PathValue("id")
	if !uuidPattern.MatchString(id) {
		writeError(w, http.StatusNotFound, "not_found", "review not found")
		return "", false
	}
	return id, true
}

// load builds the full view of a review: the review, its scores and its findings.
func (h *reviewHandlers) load(ctx context.Context, review store.Review) (reviewView, error) {
	scores, err := h.reviews.ListDimensionScores(ctx, review.ID)
	if err != nil {
		return reviewView{}, err
	}
	findings, err := h.reviews.ListCriticalFindings(ctx, review.ID)
	if err != nil {
		return reviewView{}, err
	}

	view := reviewView{
		ID:                    review.ID,
		AttemptID:             review.AttemptID,
		ReviewerID:            review.ReviewerID,
		FinalizedAt:           review.FinalizedAt,
		CriticalSafetyFinding: review.CriticalSafetyFinding,
		Notes:                 review.Notes,
		WeightedScore:         review.WeightedScore,
		ProposedOutcome:       review.ProposedOutcome,
		OutcomeReasons:        review.OutcomeReasons,
		Scores:                make([]scoreBody, 0, len(scores)),
		Findings:              make([]findingBody, 0, len(findings)),
		CreatedAt:             review.CreatedAt,
	}
	if view.OutcomeReasons == nil {
		view.OutcomeReasons = []string{}
	}
	for _, s := range scores {
		view.Scores = append(view.Scores, scoreBody{Dimension: s.Dimension, Score: s.Score, Evidence: s.Evidence})
	}
	for _, f := range findings {
		view.Findings = append(view.Findings, findingBody{
			EventType:       f.EventType,
			SupportingQuote: f.SupportingQuote,
			PolicyReference: f.PolicyReference,
			Decision:        f.Decision,
		})
	}
	return view, nil
}

// canViewReview: the reviewer who owns it, an admin, or the candidate once it is finalized.
func (h *reviewHandlers) canViewReview(ctx context.Context, caller auth.Identity, review store.Review) bool {
	switch caller.Role {
	case "admin":
		return true
	case "reviewer":
		return review.ReviewerID == caller.UserID
	case "candidate":
		if review.FinalizedAt == nil {
			return false
		}
		attempt, err := h.reviews.GetAttempt(ctx, review.AttemptID)
		return err == nil && attempt.CandidateID == caller.UserID
	}
	return false
}

func (h *reviewHandlers) respondWithReview(w http.ResponseWriter, r *http.Request, review store.Review, status int) {
	view, err := h.load(r.Context(), review)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, status, view)
}

// start begins a review of a submitted attempt. A reviewer can never review their own attempt.
func (h *reviewHandlers) start(w http.ResponseWriter, r *http.Request) {
	if !h.ready(w) {
		return
	}
	caller, ok := callerOf(w, r)
	if !ok {
		return
	}
	attemptID, ok := pathID(w, r)
	if !ok {
		return
	}

	review, err := h.reviews.StartReview(r.Context(), attemptID, actorOf(caller))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	h.respondWithReview(w, r, review, http.StatusCreated)
}

// mine lists the caller's own reviews, so a reviewer can resume a draft.
func (h *reviewHandlers) mine(w http.ResponseWriter, r *http.Request) {
	if !h.ready(w) {
		return
	}
	caller, ok := callerOf(w, r)
	if !ok {
		return
	}

	items, err := h.reviews.ListReviewerReviews(r.Context(), caller.UserID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	out := make([]reviewSummary, 0, len(items))
	for _, rv := range items {
		out = append(out, reviewSummary{
			ID:              rv.ID,
			AttemptID:       rv.AttemptID,
			FinalizedAt:     rv.FinalizedAt,
			WeightedScore:   rv.WeightedScore,
			ProposedOutcome: rv.ProposedOutcome,
			CreatedAt:       rv.CreatedAt,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"reviews": out})
}

// get returns a review. Someone who may not see it gets a 404, not a 403.
func (h *reviewHandlers) get(w http.ResponseWriter, r *http.Request) {
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

	review, err := h.reviews.GetReview(r.Context(), reviewID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if !h.canViewReview(r.Context(), caller, review) {
		writeError(w, http.StatusNotFound, "not_found", "review not found")
		return
	}
	h.respondWithReview(w, r, review, http.StatusOK)
}

type scoresRequest struct {
	Scores []scoreBody `json:"scores"`
}

// saveScores saves or replaces the reviewer's dimension scores on their own draft review.
func (h *reviewHandlers) saveScores(w http.ResponseWriter, r *http.Request) {
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

	var body scoresRequest
	if !decodeBody(w, r, &body) {
		return
	}
	inputs := make([]store.ScoreInput, 0, len(body.Scores))
	for _, s := range body.Scores {
		inputs = append(inputs, store.ScoreInput{Dimension: s.Dimension, Score: s.Score, Evidence: s.Evidence})
	}

	if err := h.reviews.SaveDimensionScores(r.Context(), reviewID, actorOf(caller), inputs); err != nil {
		writeStoreError(w, err)
		return
	}
	h.respondAfterSave(w, r, reviewID)
}

type findingsRequest struct {
	Findings []findingBody `json:"findings"`
}

// saveFindings replaces the critical findings on the reviewer's own draft review.
func (h *reviewHandlers) saveFindings(w http.ResponseWriter, r *http.Request) {
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

	var body findingsRequest
	if !decodeBody(w, r, &body) {
		return
	}
	inputs := make([]store.FindingInput, 0, len(body.Findings))
	for _, f := range body.Findings {
		inputs = append(inputs, store.FindingInput{
			EventType:       f.EventType,
			SupportingQuote: f.SupportingQuote,
			PolicyReference: f.PolicyReference,
			Decision:        f.Decision,
		})
	}

	if err := h.reviews.SaveCriticalFindings(r.Context(), reviewID, actorOf(caller), inputs); err != nil {
		writeStoreError(w, err)
		return
	}
	h.respondAfterSave(w, r, reviewID)
}

func (h *reviewHandlers) respondAfterSave(w http.ResponseWriter, r *http.Request, reviewID string) {
	review, err := h.reviews.GetReview(r.Context(), reviewID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	h.respondWithReview(w, r, review, http.StatusOK)
}
