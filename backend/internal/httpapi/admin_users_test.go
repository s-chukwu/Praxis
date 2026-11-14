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
	attemptID := "55555555-5555-5555-5555-555555555555"
	return []store.AdminUser{
		{ID: candidateA, Email: "ada@example.com", FullName: "Ada", Role: "candidate",
			BaselineAttemptID: &attemptID, BaselineStatus: &submitted},
		{ID: candidateB, Email: "bola@example.com", FullName: "", Role: "candidate"},
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
	for _, want := range []string{
		`"email":"ada@example.com"`,
		`"baseline_attempt_id":"55555555-5555-5555-5555-555555555555"`,
		`"baseline_status":"submitted"`,
		`"baseline_attempt_id":null`,
		`"baseline_status":"none"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected %s in %s", want, body)
		}
	}
}

func TestAdminUserListAnswers503WithoutADatabase(t *testing.T) {
	mux := http.NewServeMux()
	registerAdminUsersRoute(mux, fakeAuth{}, nil)

	expectStatus(t, call(t, mux, http.MethodGet, "/v1/admin/users", adminID, "admin", ""), http.StatusServiceUnavailable, "no database")
}