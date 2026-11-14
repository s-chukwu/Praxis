package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"github.com/s-chukwu/Praxis/backend/internal/aiclient"
	"github.com/s-chukwu/Praxis/backend/internal/store"
	"github.com/s-chukwu/Praxis/backend/internal/workflow"
)

const (
	maxArtifacts     = 10
	maxArtifactChars = 20000
	maxNotesChars    = 2000
)

var competencyCode = regexp.MustCompile(`^[CB][1-9]$`)

// Every track has a fixed set of competencies, as in the allocation policy.
// A review has to score all of them, so no plan can ever come back incomplete.
var trackCompetencies = map[string][]string{
	"operator": {"C1", "C2", "C3", "C4", "C5", "C6"},
	"builder":  {"C1", "C2", "C3", "C4", "C5", "C6", "B1", "B2", "B3"},
}

// checkCompetencies returns a message when the reviewer's scores do not match the track.
func checkCompetencies(track string, scores map[string]int) string {
	required, ok := trackCompetencies[track]
	if !ok {
		return "the attempt has an unknown track: " + track
	}

	allowed := make(map[string]bool, len(required))
	var missing []string
	for _, code := range required {
		allowed[code] = true
		if _, scored := scores[code]; !scored {
			missing = append(missing, code)
		}
	}
	var extra []string
	for code := range scores {
		if !allowed[code] {
			extra = append(extra, code)
		}
	}
	sort.Strings(extra)

	switch {
	case len(missing) > 0:
		return "missing competency scores for " + strings.Join(missing, ", ")
	case len(extra) > 0:
		return "these competencies do not belong to the " + track + " track: " + strings.Join(extra, ", ")
	}
	return ""
}

// Scorer is the part of the AI client the finalize route needs.
type Scorer interface {
	ScoreDiagnostic(ctx context.Context, actor aiclient.Actor, requestID string, req any) (json.RawMessage, error)
}

// Finalizer is the part of the database layer that locks a review.
type Finalizer interface {
	FinalizeReview(ctx context.Context, reviewID string, reviewer store.Actor, in store.FinalizeInput) (store.Review, error)
}

type finalizeHandlers struct {
	reviews   ReviewStore
	finalizer Finalizer
	scorer    Scorer
}

// registerFinalizeRoute wires POST /v1/reviews/{id}/finalize. Only the reviewer who
// owns the draft can call it. Go builds the score request from the saved review,
// asks the AI service for the weighted score, and stores that result with the review.
func registerFinalizeRoute(mux *http.ServeMux, authn Authenticator, reviews ReviewStore, finalizer Finalizer, scorer Scorer) {
	h := &finalizeHandlers{reviews: reviews, finalizer: finalizer, scorer: scorer}
	reviewer := requireRole(authn, "reviewer")

	mux.Handle("POST /v1/reviews/{id}/finalize", reviewer(http.HandlerFunc(h.finalize)))
}

type dimensionBody struct {
	Score    int      `json:"score"`
	Evidence []string `json:"evidence"`
}

type artifactBody struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type scoreRequest struct {
	RequestID        string                   `json:"request_id"`
	AttemptID        string                   `json:"attempt_id"`
	CandidateID      string                   `json:"candidate_id"`
	ReviewerID       string                   `json:"reviewer_id"`
	ScenarioID       string                   `json:"scenario_id"`
	ScenarioVersion  string                   `json:"scenario_version"`
	RubricVersion    string                   `json:"rubric_version"`
	Dimensions       map[string]dimensionBody `json:"dimensions"`
	CriticalFindings []findingBody            `json:"critical_findings,omitempty"`
	Artifacts        []artifactBody           `json:"artifacts"`
}

type hashRef struct {
	SHA256 string `json:"sha256"`
}

type scoreResult struct {
	WeightedScore         float64         `json:"weighted_score"`
	ProposedOutcome       string          `json:"proposed_outcome"`
	OutcomeReasons        []string        `json:"outcome_reasons"`
	EvidenceIssues        json.RawMessage `json:"evidence_issues"`
	PendingCriticalEvents json.RawMessage `json:"pending_critical_events"`
}

type scoreResponse struct {
	Rubric    hashRef     `json:"rubric"`
	Scenarios hashRef     `json:"scenarios"`
	Result    scoreResult `json:"result"`
}

type finalizeRequest struct {
	ReviewedCompetencies map[string]int `json:"reviewed_competencies"`
	Notes                string         `json:"notes"`
}

var errAnswersNotScorable = errors.New("answers must be an object of text values")

// buildArtifacts turns the candidate's saved answers into score artifacts.
// Each key is an artifact type (customer_reply, internal_note, and so on).
func buildArtifacts(answers json.RawMessage) ([]artifactBody, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(answers, &raw); err != nil {
		return nil, errAnswersNotScorable
	}

	keys := make([]string, 0, len(raw))
	for key := range raw {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	artifacts := make([]artifactBody, 0, len(keys))
	for _, key := range keys {
		var text string
		if err := json.Unmarshal(raw[key], &text); err != nil {
			return nil, errAnswersNotScorable
		}
		if strings.TrimSpace(text) == "" {
			continue
		}
		if len([]rune(text)) > maxArtifactChars {
			return nil, errAnswersNotScorable
		}
		artifacts = append(artifacts, artifactBody{Type: key, Text: text})
	}
	if len(artifacts) == 0 || len(artifacts) > maxArtifacts {
		return nil, errAnswersNotScorable
	}
	return artifacts, nil
}

func validateFinalizeBody(body finalizeRequest) string {
	if len(body.ReviewedCompetencies) == 0 {
		return "reviewed_competencies is required"
	}
	for code, score := range body.ReviewedCompetencies {
		if !competencyCode.MatchString(code) {
			return "competency codes look like C1 to C6 or B1 to B3"
		}
		if score < 0 || score > 4 {
			return "competency scores must be 0 to 4"
		}
	}
	if len([]rune(body.Notes)) > maxNotesChars {
		return "notes are too long"
	}
	return ""
}

func newRequestID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func writeAIError(w http.ResponseWriter, err error) {
	var apiErr *aiclient.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.Status {
		case http.StatusConflict:
			writeError(w, http.StatusConflict, "ai_version_conflict",
				"the AI service does not support the versions on this attempt")
			return
		case http.StatusUnprocessableEntity:
			writeError(w, http.StatusUnprocessableEntity, "ai_rejected",
				"the AI service rejected the review: "+apiErr.Message)
			return
		case http.StatusUnauthorized, http.StatusForbidden:
			log.Printf("ai service refused our call (signing or role problem): %v", err)
			writeError(w, http.StatusBadGateway, "ai_misconfigured",
				"the AI service refused the request, so this needs fixing on the server")
			return
		}
	}
	log.Printf("ai service failed: %v", err)
	writeError(w, http.StatusBadGateway, "ai_unavailable",
		"the AI service could not score this review right now")
}

func (h *finalizeHandlers) finalize(w http.ResponseWriter, r *http.Request) {
	if h.reviews == nil || h.finalizer == nil {
		writeError(w, http.StatusServiceUnavailable, "database_not_configured",
			"the database is not configured on this server")
		return
	}
	if h.scorer == nil {
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

	var body finalizeRequest
	if !decodeBody(w, r, &body) {
		return
	}
	if msg := validateFinalizeBody(body); msg != "" {
		writeError(w, http.StatusUnprocessableEntity, "invalid_input", msg)
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
	if review.FinalizedAt != nil {
		writeStoreError(w, store.ErrReviewFinalized)
		return
	}

	attempt, err := h.reviews.GetAttempt(ctx, review.AttemptID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if attempt.Status != workflow.UnderReview && attempt.Status != workflow.ReviewPending {
		writeStoreError(w, fmt.Errorf("%w: attempt is %s", store.ErrWrongState, attempt.Status))
		return
	}
	if attempt.ScenarioID == nil || strings.TrimSpace(*attempt.ScenarioID) == "" {
		writeError(w, http.StatusUnprocessableEntity, "invalid_input", "the attempt has no scenario id")
		return
	}

	if msg := checkCompetencies(attempt.TargetTrack, body.ReviewedCompetencies); msg != "" {
		writeError(w, http.StatusUnprocessableEntity, "invalid_input", msg)
		return
	}

	scores, err := h.reviews.ListDimensionScores(ctx, reviewID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if len(scores) == 0 {
		writeStoreError(w, store.ErrNoScores)
		return
	}
	findings, err := h.reviews.ListCriticalFindings(ctx, reviewID)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	artifacts, err := buildArtifacts(attempt.Answers)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "answers_not_scorable",
			"the candidate's answers must be text values with at least one non-empty answer")
		return
	}

	requestID := newRequestID()
	payload := scoreRequest{
		RequestID:       requestID,
		AttemptID:       attempt.ID,
		CandidateID:     attempt.CandidateID,
		ReviewerID:      caller.UserID,
		ScenarioID:      *attempt.ScenarioID,
		ScenarioVersion: attempt.ScenarioVersion,
		RubricVersion:   attempt.RubricVersion,
		Dimensions:      make(map[string]dimensionBody, len(scores)),
		Artifacts:       artifacts,
	}
	for _, s := range scores {
		payload.Dimensions[s.Dimension] = dimensionBody{Score: s.Score, Evidence: s.Evidence}
	}
	hasConfirmed := false
	for _, f := range findings {
		payload.CriticalFindings = append(payload.CriticalFindings, findingBody{
			EventType:       f.EventType,
			SupportingQuote: f.SupportingQuote,
			PolicyReference: f.PolicyReference,
			Decision:        f.Decision,
		})
		if f.Decision == "confirmed" {
			hasConfirmed = true
		}
	}

	raw, err := h.scorer.ScoreDiagnostic(ctx, aiclient.ReviewerActor(caller.UserID), requestID, payload)
	if err != nil {
		writeAIError(w, err)
		return
	}
	var scored scoreResponse
	if err := json.Unmarshal(raw, &scored); err != nil {
		log.Printf("unreadable score response: %v", err)
		writeError(w, http.StatusBadGateway, "ai_unavailable", "the AI service sent an answer we could not read")
		return
	}
	result := scored.Result

	// The AI service holds the review back when a quote does not match the answers
	// or a finding is still pending. The reviewer has to fix that first.
	if result.ProposedOutcome == "review_pending" {
		writeJSON(w, http.StatusConflict, map[string]any{"error": map[string]any{
			"code":    "review_pending",
			"message": "evidence quotes or critical findings need attention before this review can be finalized",
			"details": map[string]any{
				"evidence_issues":         result.EvidenceIssues,
				"pending_critical_events": result.PendingCriticalEvents,
			},
		}})
		return
	}

	// Second guard, in Go: a confirmed critical finding can never end as a pass.
	if hasConfirmed && result.ProposedOutcome == "evidence_verified" {
		log.Printf("ai service passed a review that has a confirmed critical finding (review %s)", reviewID)
		writeError(w, http.StatusConflict, "inconsistent_result",
			"a review with a confirmed critical finding cannot be a pass, so it was not finalized")
		return
	}

	competencies, err := json.Marshal(body.ReviewedCompetencies)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	final, err := h.finalizer.FinalizeReview(ctx, reviewID, actorOf(caller), store.FinalizeInput{
		ReviewedCompetencies: competencies,
		Notes:                body.Notes,
		WeightedScore:        result.WeightedScore,
		ProposedOutcome:      result.ProposedOutcome,
		OutcomeReasons:       result.OutcomeReasons,
		RubricSHA256:         scored.Rubric.SHA256,
		ScenarioSHA256:       scored.Scenarios.SHA256,
	})
	if err != nil {
		writeStoreError(w, err)
		return
	}

	view, err := (&reviewHandlers{reviews: h.reviews}).load(ctx, final)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, view)
}
