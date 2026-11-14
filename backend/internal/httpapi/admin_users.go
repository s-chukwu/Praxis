package httpapi

import (
	"context"
	"net/http"

	"github.com/s-chukwu/Praxis/backend/internal/store"
)

// UserLister is separate from RoleStore so existing fakes keep working.
type UserLister interface {
	ListUsers(ctx context.Context) ([]store.AdminUser, error)
}

type adminUserView struct {
	ID             string `json:"id"`
	FullName       string `json:"full_name"`
	Role           string `json:"role"`
	BaselineStatus string `json:"baseline_status"` // "none" when there is no baseline attempt
}

// registerAdminUsersRoute wires GET /v1/admin/users, admin only.
func registerAdminUsersRoute(mux *http.ServeMux, authn Authenticator, users UserLister) {
	mux.Handle("GET /v1/admin/users", requireRole(authn, "admin")(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if users == nil {
				writeError(w, http.StatusServiceUnavailable, "database_not_configured",
					"the database is not configured on this server")
				return
			}
			list, err := users.ListUsers(r.Context())
			if err != nil {

				writeStoreError(w, err)
				return
			}
			views := make([]adminUserView, 0, len(list))
			for _, u := range list {
				status := "none"
				if u.BaselineStatus != nil {
					status = *u.BaselineStatus
				}
				views = append(views, adminUserView{
					ID: u.ID, FullName: u.FullName, Role: u.Role, BaselineStatus: status,
				})
			}
			writeJSON(w, http.StatusOK, map[string]any{"users": views})
		})))
}