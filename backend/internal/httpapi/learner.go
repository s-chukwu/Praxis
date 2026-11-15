package httpapi

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/s-chukwu/Praxis/backend/internal/store"
)

const (
	maxCVChars           = 20000
	maxNameChars         = 120
	maxAvailabilityChars = 120
)

// LearnerStore is the part of the database layer the learner profile routes need.
type LearnerStore interface {
	GetLearnerProfile(ctx context.Context, userID string) (store.LearnerProfile, error)
	UpdateLearnerProfile(ctx context.Context, userID string, patch store.LearnerPatch) error
	EraseLearner(ctx context.Context, userID string, actor store.Actor) error
}

// AccountCloser closes the sign-in account itself, at Supabase.
type AccountCloser interface {
	SoftDeleteUser(ctx context.Context, userID string) error
}

type learnerHandlers struct {
	learners LearnerStore
	accounts AccountCloser
}

// registerLearnerRoutes wires the learner's own profile. These are the routes the frontend
// calls in live mode, so their shape matches the frontend's Learner type.
func registerLearnerRoutes(mux *http.ServeMux, authn Authenticator, learners LearnerStore, accounts AccountCloser) {
	h := &learnerHandlers{learners: learners, accounts: accounts}
	candidate := requireRole(authn, "candidate")
	mux.Handle("GET /v1/me/learner", candidate(http.HandlerFunc(h.get)))
	mux.Handle("PATCH /v1/me/learner", candidate(http.HandlerFunc(h.patch)))
	mux.Handle("DELETE /v1/me", candidate(http.HandlerFunc(h.erase)))
}

type learnerView struct {
	Name         string            `json:"name"`
	Program      *string           `json:"program,omitempty"`
	Availability *string           `json:"availability,omitempty"`
	Onboarded    bool              `json:"onboarded"`
	CVSubmitted  bool              `json:"cvSubmitted"`
	CVText       *string           `json:"cvText,omitempty"`
	Skills       map[string]string `json:"skills"`
}

func (h *learnerHandlers) ready(w http.ResponseWriter) bool {
	if h.learners == nil {
		writeError(w, http.StatusServiceUnavailable, "database_not_configured", "the database is not configured on this server")
		return false
	}
	return true
}

func (h *learnerHandlers) get(w http.ResponseWriter, r *http.Request) {
	if !h.ready(w) {
		return
	}
	caller, ok := callerOf(w, r)
	if !ok {
		return
	}
	p, err := h.learners.GetLearnerProfile(r.Context(), caller.UserID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	skills := make(map[string]string, len(p.ClaimedSkills))
	for _, c := range p.ClaimedSkills {
		skills[c] = "confirmed"
	}
	writeJSON(w, http.StatusOK, learnerView{
		Name: p.FullName, Program: p.Programme, Availability: p.Availability, Onboarded: p.OnboardingDone,
		CVSubmitted: p.CVText != nil, CVText: p.CVText, Skills: skills,
	})
}

// learnerPatchBody mirrors the frontend's saveLearner. CVText uses RawMessage so "null" (delete the CV)
// can be told apart from the field being absent (leave it alone).
type learnerPatchBody struct {
	Name         *string           `json:"name"`
	Program      *string           `json:"program"`
	Availability *string           `json:"availability"`
	Onboarded    *bool             `json:"onboarded"`
	CVSubmitted  *bool             `json:"cvSubmitted"`
	CVText       json.RawMessage   `json:"cvText"`
	Skills       map[string]string `json:"skills"`
}

func (h *learnerHandlers) patch(w http.ResponseWriter, r *http.Request) {
	if !h.ready(w) {
		return
	}
	caller, ok := callerOf(w, r)
	if !ok {
		return
	}
	var body learnerPatchBody
	if !decodeBody(w, r, &body) {
		return
	}

	patch := store.LearnerPatch{Programme: body.Program, Availability: body.Availability, OnboardingDone: body.Onboarded}
	bad := func(msg string) { writeError(w, http.StatusUnprocessableEntity, "invalid_input", msg) }

	if body.Name != nil {
		name := strings.TrimSpace(*body.Name)
		if name == "" || len([]rune(name)) > maxNameChars {
			bad("name must be 1 to 120 characters")
			return
		}
		patch.FullName = &name
	}
	if body.Program != nil {
		if _, known := trackCompetencies[*body.Program]; !known {
			bad("program must be operator or builder")
			return
		}
	}
	if body.Availability != nil && len([]rune(*body.Availability)) > maxAvailabilityChars {
		bad("availability must be at most 120 characters")
		return
	}
	switch text := strings.TrimSpace(string(body.CVText)); {
	case text == "":
	case text == "null":
		patch.ClearCV = true
	default:
		var cv string
		if err := json.Unmarshal(body.CVText, &cv); err != nil || len([]rune(cv)) > maxCVChars {
			bad("cvText must be text of at most 20000 characters, or null to delete it")
			return
		}
		if strings.TrimSpace(cv) == "" {
			patch.ClearCV = true
		} else {
			patch.CVText = &cv
		}
	}
	if body.Skills != nil {
		// A learner can only claim a skill. Verified comes from a reviewer, never from this route.
		claimed := []string{}
		for code, status := range body.Skills {
			if !competencyCode.MatchString(code) {
				bad("skills must use competency codes such as C1 or B2")
				return
			}
			if status == "confirmed" {
				claimed = append(claimed, code)
			}
		}
		patch.ClaimedSkills = &claimed
	}

	if err := h.learners.UpdateLearnerProfile(r.Context(), caller.UserID, patch); err != nil {
		writeStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// erase deletes the learner's personal data, then closes the sign-in account. The data goes first, so a
// failure at Supabase leaves a still-working account the learner can retry from, never orphaned data.
func (h *learnerHandlers) erase(w http.ResponseWriter, r *http.Request) {
	if !h.ready(w) {
		return
	}
	if h.accounts == nil {
		writeError(w, http.StatusServiceUnavailable, "auth_not_configured", "account deletion is not configured on this server")
		return
	}
	caller, ok := callerOf(w, r)
	if !ok {
		return
	}
	if err := h.learners.EraseLearner(r.Context(), caller.UserID, actorOf(caller)); err != nil {
		writeStoreError(w, err)
		return
	}
	if err := h.accounts.SoftDeleteUser(r.Context(), caller.UserID); err != nil {
		log.Printf("account %s: data erased but sign-in not closed: %v", caller.UserID, err)
		writeError(w, http.StatusBadGateway, "account_not_closed", "your data was deleted but your account could not be closed, so try again")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
