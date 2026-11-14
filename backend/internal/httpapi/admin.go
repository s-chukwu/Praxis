package httpapi

import (
	"context"
	"errors"
	"net/http"

	"github.com/s-chukwu/Praxis/backend/internal/store"
)

// RoleStore is the part of the database layer that changes user roles.
type RoleStore interface {
	SetUserRole(ctx context.Context, userID, role string, actor store.Actor) error
}

type adminHandlers struct {
	roles RoleStore
}

// registerAdminRoutes wires the admin-only routes. Reviewers are invite-only (X-01),
// so this is the only way someone becomes a reviewer.
func registerAdminRoutes(mux *http.ServeMux, authn Authenticator, roles RoleStore) {
	h := &adminHandlers{roles: roles}
	mux.Handle("PUT /v1/admin/users/{id}/role", requireRole(authn, "admin")(http.HandlerFunc(h.setRole)))
}

type roleRequest struct {
	Role string `json:"role"`
}

func (h *adminHandlers) setRole(w http.ResponseWriter, r *http.Request) {
	if h.roles == nil {
		writeError(w, http.StatusServiceUnavailable, "database_not_configured",
			"the database is not configured on this server")
		return
	}
	caller, ok := callerOf(w, r)
	if !ok {
		return
	}
	userID := r.PathValue("id")
	if !uuidPattern.MatchString(userID) {
		writeError(w, http.StatusNotFound, "not_found", "user not found")
		return
	}

	var body roleRequest
	if !decodeBody(w, r, &body) {
		return
	}

	err := h.roles.SetUserRole(r.Context(), userID, body.Role, actorOf(caller))
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, map[string]string{"user_id": userID, "role": body.Role})
	case errors.Is(err, store.ErrInvalidInput):
		writeError(w, http.StatusUnprocessableEntity, "invalid_input", "role must be candidate or reviewer")
	case errors.Is(err, store.ErrSelfRoleChange):
		writeError(w, http.StatusForbidden, "forbidden", "admin roles cannot be changed here")
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "user not found")
	default:
		writeStoreError(w, err)
	}
}
