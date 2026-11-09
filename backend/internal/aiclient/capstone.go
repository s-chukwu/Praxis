package aiclient

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

const (
	SuitePractice = "practice"
	SuiteHeldOut  = "held_out"
)

var ErrInvalidSuite = errors.New("suite must be practice or held_out")

// CapstoneRequest is the body of POST /internal/v1/capstones/evaluate.
// Config is the candidate's bot configuration, passed through untouched.
type CapstoneRequest struct {
	RequestID   string          `json:"request_id"`
	AttemptID   string          `json:"attempt_id"`
	CandidateID string          `json:"candidate_id"`
	Suite       string          `json:"suite"`
	Config      json.RawMessage `json:"config"`
}

// CapstoneEvaluate runs a candidate's bot configuration against the practice or
// held out cases. Allowed callers: candidate (own id, practice only), reviewer and
// system. The held out suite is server only, so a candidate is refused here before
// anything is sent. Go should start a held out run once per locked attempt.
func (c *Client) CapstoneEvaluate(ctx context.Context, actor Actor, requestID string, req CapstoneRequest) (json.RawMessage, error) {
	if err := requireActor(actor, "candidate", "reviewer", "system"); err != nil {
		return nil, err
	}
	if req.Suite != SuitePractice && req.Suite != SuiteHeldOut {
		return nil, ErrInvalidSuite
	}
	if actor.kind() == "candidate" {
		if req.Suite != SuitePractice || actor != CandidateActor(req.CandidateID) {
			return nil, ErrActorNotAllowed
		}
	}
	return c.do(ctx, http.MethodPost, "/internal/v1/capstones/evaluate", actor, requestID, req)
}
