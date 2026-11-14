package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/s-chukwu/Praxis/backend/internal/auth"
	"github.com/s-chukwu/Praxis/backend/internal/store"
	"github.com/s-chukwu/Praxis/backend/internal/workflow"
)

const maxBodyBytes = 256 << 10

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F-]{36}$`)

// AttemptStore is what the attempt routes need from the database layer.
// *store.Store satisfies it, and tests use a fake.
type AttemptStore interface {
	CreateAttempt(ctx context.Context, n store.NewAttempt, actor store.Actor) (store.Attempt, error)
	CreateOrGetBaselineAttempt(ctx context.Context, n store.NewAttempt, actor store.Actor) (store.Attempt, bool, error)
	GetAttempt(ctx context.Context, id string) (store.Attempt, error)
	ListCandidateAttempts(ctx context.Context, candidateID string) ([]store.Attempt, error)
	ListReviewQueue(ctx context.Context) ([]store.Attempt, error)
	SaveAnswers(ctx context.Context, attemptID, candidateID string, answers json.RawMessage) (store.Attempt, error)
	TransitionAttempt(ctx context.Context, attemptID string, to workflow.State, actor store.Actor) (store.Attempt, error)
}

type attemptHandlers struct {
	attempts AttemptStore
}

// registerAttemptRoutes wires the attempt routes. Each route first checks the
// caller's role (requireRole), then the handler checks ownership. The Go backend
// uses the service role key, so these checks are what protects the data.
func registerAttemptRoutes(mux *http.ServeMux, authn Authenticator, attempts AttemptStore) {
	h := &attemptHandlers{attempts: attempts}

	signedIn := requireRole(authn)
	candidate := requireRole(authn, "candidate")
	staff := requireRole(authn, "reviewer", "admin")
	admin := requireRole(authn, "admin")

	mux.Handle("POST /v1/attempts", staff(http.HandlerFunc(h.create)))
	mux.Handle("GET /v1/attempts", signedIn(http.HandlerFunc(h.list)))
	mux.Handle("GET /v1/attempts/{id}", signedIn(http.HandlerFunc(h.get)))
	mux.Handle("PUT /v1/attempts/{id}/answers", candidate(http.HandlerFunc(h.saveAnswers)))
	mux.Handle("POST /v1/attempts/{id}/start", candidate(h.move(workflow.InProgress, true)))
	mux.Handle("POST /v1/attempts/{id}/submit", candidate(h.move(workflow.Submitted, true)))
	mux.Handle("POST /v1/attempts/{id}/void", admin(h.move(workflow.Voided, false)))
}

type attemptView struct {
	ID                  string          `json:"id"`
	CandidateID         string          `json:"candidate_id"`
	TargetTrack         string          `json:"target_track"`
	Stage               string          `json:"stage"`
	Status              string          `json:"status"`
	ScenarioID          *string         `json:"scenario_id"`
	RoleStandardVersion string          `json:"role_standard_version"`
	ScenarioVersion     string          `json:"scenario_version"`
	RubricVersion       string          `json:"rubric_version"`
	PolicyVersion       string          `json:"policy_version"`
	Answers             json.RawMessage `json:"answers,omitempty"`
	SubmittedAt         *time.Time      `json:"submitted_at"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
}

func toAttemptView(a store.Attempt, withAnswers bool) attemptView {
	v := attemptView{
		ID:                  a.ID,
		CandidateID:         a.CandidateID,
		TargetTrack:         a.TargetTrack,
		Stage:               a.Stage,
		Status:              string(a.Status),
		ScenarioID:          a.ScenarioID,
		RoleStandardVersion: a.RoleStandardVersion,
		ScenarioVersion:     a.ScenarioVersion,
		RubricVersion:       a.RubricVersion,
		PolicyVersion:       a.PolicyVersion,
		SubmittedAt:         a.SubmittedAt,
		CreatedAt:           a.CreatedAt,
		UpdatedAt:           a.UpdatedAt,
	}
	if withAnswers {
		v.Answers = a.Answers
	}
	return v
}

func actorOf(id auth.Identity) store.Actor {
	return store.Actor{ID: id.UserID, Role: workflow.Role(id.Role)}
}

// canView mirrors the database rules: candidates see their own attempts,
// reviewers see attempts from submission onwards, admins see everything.
func canView(id auth.Identity, a store.Attempt) bool {
	switch id.Role {
	case "admin":
		return true
	case "candidate":
		return a.CandidateID == id.UserID
	case "reviewer":
		switch a.Status {
		case workflow.Submitted, workflow.UnderReview, workflow.ReviewPending, workflow.Reviewed:
			return true
		}
	}
	return false
}

func (h *attemptHandlers) ready(w http.ResponseWriter) bool {
	if h.attempts == nil {
		writeError(w, http.StatusServiceUnavailable, "database_not_configured",
			"the database is not configured on this server")
		return false
	}
	return true
}

func callerOf(w http.ResponseWriter, r *http.Request) (auth.Identity, bool) {
	id, ok := identityFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "no identity")
	}
	return id, ok
}

func pathID(w http.ResponseWriter, r *http.Request) (string, bool) {
	id := r.PathValue("id")
	if !uuidPattern.MatchString(id) {
		writeError(w, http.StatusNotFound, "not_found", "attempt not found")
		return "", false
	}
	return id, true
}

func decodeBody(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "the request body is not valid for this route")
		return false
	}
	return true
}

func writeStoreError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "attempt not found")
	case errors.Is(err, store.ErrNotEditable):
		writeError(w, http.StatusConflict, "not_editable", "the attempt was not found or can no longer be edited")
	case errors.Is(err, store.ErrWrongState):
		writeError(w, http.StatusConflict, "wrong_state", "the attempt is not in a state that allows this")
	case errors.Is(err, workflow.ErrIllegalTransition):
		writeError(w, http.StatusConflict, "illegal_transition", "the attempt cannot move to that state")
	case errors.Is(err, workflow.ErrRoleNotAllowed):
		writeError(w, http.StatusForbidden, "forbidden", "your role is not allowed to do this")
	case errors.Is(err, workflow.ErrSelfReview):
		writeError(w, http.StatusForbidden, "forbidden", "a reviewer cannot review their own attempt")
	case errors.Is(err, store.ErrReviewNotYours):
		writeError(w, http.StatusNotFound, "not_found", "review not found")
	case errors.Is(err, store.ErrReviewTaken):
		writeError(w, http.StatusConflict, "review_taken", "another reviewer is already reviewing this attempt")
	case errors.Is(err, store.ErrReviewFinalized):
		writeError(w, http.StatusConflict, "review_finalized", "the review is already finalized")
	case errors.Is(err, store.ErrPendingFindings):
		writeError(w, http.StatusConflict, "pending_findings", "every critical finding must be confirmed or dismissed first")
	case errors.Is(err, store.ErrNoScores):
		writeError(w, http.StatusConflict, "no_scores", "a review needs at least one dimension score")
	case errors.Is(err, store.ErrInvalidInput):
		writeError(w, http.StatusUnprocessableEntity, "invalid_input", err.Error())
	default:
		log.Printf("attempt route failed: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "something went wrong")
	}
}

type createAttemptRequest struct {
	CandidateID         string `json:"candidate_id"`
	TargetTrack         string `json:"target_track"`
	RoleStandardVersion string `json:"role_standard_version"`
	ScenarioVersion     string `json:"scenario_version"`
	RubricVersion       string `json:"rubric_version"`
	PolicyVersion       string `json:"policy_version"`
	ScenarioID          string `json:"scenario_id"`
	Stage               string `json:"stage"`
}

// create assigns a new attempt to a candidate. Reviewers and admins only.
func (h *attemptHandlers) create(w http.ResponseWriter, r *http.Request) {
	if !h.ready(w) {
		return
	}
	caller, ok := callerOf(w, r)
	if !ok {
		return
	}

	var body createAttemptRequest
	if !decodeBody(w, r, &body) {
		return
	}

	required := []string{body.TargetTrack, body.RoleStandardVersion, body.ScenarioVersion,
		body.RubricVersion, body.PolicyVersion, body.ScenarioID}
	valid := uuidPattern.MatchString(body.CandidateID)
	for _, value := range required {
		if strings.TrimSpace(value) == "" {
			valid = false
		}
	}
	if body.Stage != "" && body.Stage != "baseline" && body.Stage != "final" {
		valid = false
	}
	if !valid {
		writeError(w, http.StatusUnprocessableEntity, "invalid_input",
			"candidate_id must be a uuid, every version field is required, and stage must be baseline or final")
		return
	}

	created, err := h.attempts.CreateAttempt(r.Context(), store.NewAttempt{
		CandidateID:         body.CandidateID,
		TargetTrack:         body.TargetTrack,
		RoleStandardVersion: body.RoleStandardVersion,
		ScenarioVersion:     body.ScenarioVersion,
		RubricVersion:       body.RubricVersion,
		PolicyVersion:       body.PolicyVersion,
		ScenarioID:          body.ScenarioID,
		Stage:               body.Stage,
	}, actorOf(caller))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toAttemptView(created, false))
}

// list returns a candidate's own attempts, or the review queue for staff.
func (h *attemptHandlers) list(w http.ResponseWriter, r *http.Request) {
	if !h.ready(w) {
		return
	}
	caller, ok := callerOf(w, r)
	if !ok {
		return
	}

	var items []store.Attempt
	var err error
	if caller.Role == "candidate" {
		items, err = h.attempts.ListCandidateAttempts(r.Context(), caller.UserID)
	} else {
		items, err = h.attempts.ListReviewQueue(r.Context())
	}
	if err != nil {
		writeStoreError(w, err)
		return
	}

	views := make([]attemptView, 0, len(items))
	for _, a := range items {
		views = append(views, toAttemptView(a, false))
	}
	writeJSON(w, http.StatusOK, map[string]any{"attempts": views})
}

// get returns one attempt. Someone who may not see it gets a 404, not a 403,
// so the API does not reveal which attempts exist.
func (h *attemptHandlers) get(w http.ResponseWriter, r *http.Request) {
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

	found, err := h.attempts.GetAttempt(r.Context(), attemptID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if !canView(caller, found) {
		writeError(w, http.StatusNotFound, "not_found", "attempt not found")
		return
	}
	writeJSON(w, http.StatusOK, toAttemptView(found, true))
}

type answersRequest struct {
	Answers json.RawMessage `json:"answers"`
}

// saveAnswers stores the candidate's answers while the attempt is open.
func (h *attemptHandlers) saveAnswers(w http.ResponseWriter, r *http.Request) {
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

	var body answersRequest
	if !decodeBody(w, r, &body) {
		return
	}
	if trimmed := bytes.TrimSpace(body.Answers); len(trimmed) == 0 || trimmed[0] != '{' {
		writeError(w, http.StatusUnprocessableEntity, "invalid_input", "answers must be a JSON object")
		return
	}

	saved, err := h.attempts.SaveAnswers(r.Context(), attemptID, caller.UserID, body.Answers)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toAttemptView(saved, true))
}

// move returns a handler that moves an attempt to a new state. With ownerOnly,
// the candidate must own the attempt. The state and role rules are checked again
// inside the database transaction, so a race cannot slip past them.
func (h *attemptHandlers) move(to workflow.State, ownerOnly bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		if ownerOnly {
			found, err := h.attempts.GetAttempt(r.Context(), attemptID)
			if err != nil {
				writeStoreError(w, err)
				return
			}
			if found.CandidateID != caller.UserID {
				writeError(w, http.StatusNotFound, "not_found", "attempt not found")
				return
			}
		}

		updated, err := h.attempts.TransitionAttempt(r.Context(), attemptID, to, actorOf(caller))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, toAttemptView(updated, true))
	})
}
