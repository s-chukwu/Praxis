package httpapi

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/s-chukwu/Praxis/backend/internal/store"
)

type adminFakeLister struct{}

func (adminFakeLister) ListUsers(ctx context.Context) ([]store.AdminUser, error) {
	submitted := "submitted"
	return []store.AdminUser{
		{ID: candidateA, FullName: "Ada", Role: "candidate", BaselineStatus: &submitted},
		{ID: candidateB, FullName: "Bola", Role: "candidate", BaselineStatus: nil},
	}, nil
}

func TestAdminUserListIsAdminOnly(t *testing.T) {
	mux := http.NewServeMux()
	registerAdminUsersRoute(mux, fakeAuth{}, adminFakeLister{})

	expectStatus(t, call(t, mux, http.MethodGet, "/v1/admin/users", "", "", ""), http.StatusUnauthorized, "no login")
	expectStatus(t, call(t, mux, http.MethodGet, "/v1/admin/users", candidateA, "candidate", ""), http.StatusForbidden, "candidate")
	expectStatus(t, call(t, mux, http.MethodGet, "/v1/admin/users", reviewerID, "reviewer", ""), http.StatusForbidden, "reviewer")

	rec := call(t, mux, http.MethodGet, "/v1/admin/users", adminID, "admin", "")
	expectStatus(t, rec, http.StatusOK, "admin")
	body := rec.Body.String()

	if !strings.Contains(body, `"baseline_status":"submitted"`) || !strings.Contains(body, `"baseline_status":"none"`) {
		t.Fatalf("expected submitted and none statuses, got %s", body)
	}
}

func TestAdminUserListAnswers503WithoutADatabase(t *testing.T) {
	mux := http.NewServeMux()
	registerAdminUsersRoute(mux, fakeAuth{}, nil)

	expectStatus(t, call(t, mux, http.MethodGet, "/v1/admin/users", adminID, "admin", ""), http.StatusServiceUnavailable, "no database")
}