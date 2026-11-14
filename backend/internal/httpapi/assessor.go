package httpapi

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/s-chukwu/Praxis/backend/internal/aiclient"
)

// PacketReader is the part of the AI client that fetches the grading key.
type PacketReader interface {
	AssessorPacket(ctx context.Context, actor aiclient.Actor, requestID, scenarioID, version string) (json.RawMessage, error)
}

type assessorHandlers struct {
	reviews ReviewStore
	packets PacketReader
}

// registerAssessorRoute wires GET /v1/attempts/{id}/assessor-packet. The packet holds
// the grading key, so only the reviewer holding a review of that attempt, or an admin, gets it.
func registerAssessorRoute(mux *http.ServeMux, authn Authenticator, reviews ReviewStore, packets PacketReader) {
	h := &assessorHandlers{reviews: reviews, packets: packets}
	mux.Handle("GET /v1/attempts/{id}/assessor-packet", requireRole(authn, "reviewer", "admin")(http.HandlerFunc(h.get)))
}

func (h *assessorHandlers) holdsReview(ctx context.Context, reviewerID, attemptID string) (bool, error) {
	items, err := h.reviews.ListReviewerReviews(ctx, reviewerID)
	if err != nil {
		return false, err
	}
	for _, r := range items {
		if r.AttemptID == attemptID {
			return true, nil
		}
	}
	return false, nil
}

func (h *assessorHandlers) get(w http.ResponseWriter, r *http.Request) {
	if h.reviews == nil {
		writeError(w, http.StatusServiceUnavailable, "database_not_configured",
			"the database is not configured on this server")
		return
	}
	if h.packets == nil {
		writeError(w, http.StatusServiceUnavailable, "ai_not_configured",
			"the AI service is not configured on this server")
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

	actor := aiclient.AdminActor(caller.UserID)
	if caller.Role == "reviewer" {
		holds, err := h.holdsReview(r.Context(), caller.UserID, attemptID)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		// 404, not 403, so reviewers cannot probe which attempts exist.
		if !holds {
			writeError(w, http.StatusNotFound, "not_found", "attempt not found")
			return
		}
		actor = aiclient.ReviewerActor(caller.UserID)
	}

	attempt, err := h.reviews.GetAttempt(r.Context(), attemptID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if attempt.ScenarioID == nil || *attempt.ScenarioID == "" {
		writeError(w, http.StatusConflict, "no_scenario", "this attempt has no scenario to review against")
		return
	}

	packet, err := h.packets.AssessorPacket(r.Context(), actor, "", *attempt.ScenarioID, attempt.ScenarioVersion)
	if err != nil {
		log.Printf("assessor packet for attempt %s failed: %v", attemptID, err)
		writeError(w, http.StatusBadGateway, "ai_unavailable", "the marking guide could not be loaded right now")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, packet)
}
