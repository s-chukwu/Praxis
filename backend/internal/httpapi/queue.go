package httpapi

import (
	"net/http"
	"time"
)

// reviewSLA is how long a submitted attempt may wait before it counts as overdue.
const reviewSLA = 48 * time.Hour

// queueItem is one row of the reviewer queue. Candidates appear as a short
// reference, not a name, so reviewers mark the work and not the person.
type queueItem struct {
	AttemptID    string     `json:"attempt_id"`
	CandidateRef string     `json:"candidate_ref"`
	Program      string     `json:"program"`
	Stage        string     `json:"stage"`
	Status       string     `json:"status"`
	SubmittedAt  *time.Time `json:"submitted_at"`
	DueAt        *time.Time `json:"due_at"`
	Priority     string     `json:"priority"`
	Mine         bool       `json:"mine"`
}

type queueHandlers struct {
	attempts AttemptStore
	reviews  ReviewStore
	now      func() time.Time
}

func registerQueueRoute(mux *http.ServeMux, authn Authenticator, attempts AttemptStore, reviews ReviewStore) {
	h := &queueHandlers{attempts: attempts, reviews: reviews, now: time.Now}
	mux.Handle("GET /v1/review-queue", requireRole(authn, "reviewer", "admin")(http.HandlerFunc(h.list)))
}

// priority is overdue past the SLA, due_soon in its last 12 hours, otherwise normal.
func priority(due *time.Time, now time.Time) string {
	switch {
	case due == nil:
		return "normal"
	case now.After(*due):
		return "overdue"
	case due.Sub(now) <= 12*time.Hour:
		return "due_soon"
	}
	return "normal"
}

func (h *queueHandlers) list(w http.ResponseWriter, r *http.Request) {
	if h.attempts == nil || h.reviews == nil {
		writeError(w, http.StatusServiceUnavailable, "database_not_configured",
			"the database is not configured on this server")
		return
	}
	caller, ok := callerOf(w, r)
	if !ok {
		return
	}

	attempts, err := h.attempts.ListReviewQueue(r.Context())
	if err != nil {
		writeStoreError(w, err)
		return
	}
	held, err := h.reviews.ListReviewerReviews(r.Context(), caller.UserID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	mine := make(map[string]bool, len(held))
	for _, rv := range held {
		if rv.FinalizedAt == nil {
			mine[rv.AttemptID] = true
		}
	}

	now := h.now()
	items := make([]queueItem, 0, len(attempts))
	for _, a := range attempts {
		var due *time.Time
		if a.SubmittedAt != nil {
			d := a.SubmittedAt.Add(reviewSLA)
			due = &d
		}
		items = append(items, queueItem{
			AttemptID:    a.ID,
			CandidateRef: "C-" + a.CandidateID[:8],
			Program:      a.TargetTrack,
			Stage:        a.Stage,
			Status:       string(a.Status),
			SubmittedAt:  a.SubmittedAt,
			DueAt:        due,
			Priority:     priority(due, now),
			Mine:         mine[a.ID],
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "sla_hours": int(reviewSLA.Hours())})
}
