package aiclient

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"
)

// The model takes 5 to 8 seconds per call, so this endpoint gets 30 seconds.
const skillsTimeout = 30 * time.Second

const maxCVTextRunes = 20000

var (
	ErrNoConsent   = errors.New("cv_analysis consent is required: consent_ref is empty")
	ErrInvalidText = errors.New("cv text must be 1 to 20000 characters")
	ErrMissingID   = errors.New("candidate_id is required")
)

// SkillsRequest is the body of POST /internal/v1/skills/suggest.
type SkillsRequest struct {
	RequestID           string `json:"request_id"`
	CandidateID         string `json:"candidate_id"`
	ConsentRef          string `json:"consent_ref"`
	RoleStandardVersion string `json:"role_standard_version"`
	Text                string `json:"text"`
}

// SkillsSuggest asks the AI service to suggest competencies from CV text.
// Only the system may call it. Go must have recorded the candidate's cv_analysis
// consent first and pass that consent's id as ConsentRef. The call has a 30 second
// timeout and is never retried: the service already retries the model once.
// Use NeedsManualReview on the error: if it is true, send the CV to a person.
// Suggestions are claims for the candidate to confirm, never scores.
func (c *Client) SkillsSuggest(ctx context.Context, requestID string, req SkillsRequest) (json.RawMessage, error) {
	if strings.TrimSpace(req.ConsentRef) == "" {
		return nil, ErrNoConsent
	}
	if strings.TrimSpace(req.CandidateID) == "" {
		return nil, ErrMissingID
	}
	if n := len([]rune(req.Text)); n == 0 || n > maxCVTextRunes {
		return nil, ErrInvalidText
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	// A single attempt on the slow client: the second return value, "retry", is ignored on purpose.
	data, _, err := c.once(ctx, c.slow, http.MethodPost, "/internal/v1/skills/suggest", SystemActor, requestID, body)
	return data, err
}

// NeedsManualReview is true when the model could not answer: the service said
// 503 model_unavailable, or our own 30 second timeout ran out. It must never be
// treated as "the candidate has no skills".
func NeedsManualReview(err error) bool {
	if err == nil {
		return false
	}

	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Status == http.StatusServiceUnavailable && apiErr.Code == "model_unavailable"
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	return errors.Is(err, context.DeadlineExceeded)
}
