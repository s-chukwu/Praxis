package httpapi

import (
	"context"
	"net/http"
	"testing"

	"github.com/s-chukwu/Praxis/backend/internal/store"
)

// fakeRoles applies the same rules as store.SetUserRole.
type fakeRoles struct {
	roles map[string]string
}

func (f *fakeRoles) SetUserRole(ctx context.Context, userID, role string, actor store.Actor) error {
	if role != "candidate" && role != "reviewer" {
		return store.ErrInvalidInput
	}
	if userID == actor.ID || f.roles[userID] == "admin" {
		return store.ErrSelfRoleChange
	}
	if _, ok := f.roles[userID]; !ok {
		return store.ErrNotFound
	}
	f.roles[userID] = role
	return nil
}

func TestAdminsCanMakeSomeoneAReviewer(t *testing.T) {
	roles := &fakeRoles{roles: map[string]string{candidateA: "candidate", adminID: "admin"}}
	mux := http.NewServeMux()
	registerAdminRoutes(mux, fakeAuth{}, roles)
	path := "/v1/admin/users/" + candidateA + "/role"

	expectStatus(t, call(t, mux, http.MethodPut, path, adminID, "admin", `{"role":"reviewer"}`), http.StatusOK, "admin grants reviewer")
	if roles.roles[candidateA] != "reviewer" {
		t.Fatalf("role should be reviewer, got %s", roles.roles[candidateA])
	}
	expectStatus(t, call(t, mux, http.MethodPut, path, adminID, "admin", `{"role":"candidate"}`), http.StatusOK, "admin removes reviewer")
}

func TestRoleChangesAreLockedDown(t *testing.T) {
	roles := &fakeRoles{roles: map[string]string{candidateA: "candidate", adminID: "admin"}}
	mux := http.NewServeMux()
	registerAdminRoutes(mux, fakeAuth{}, roles)
	path := "/v1/admin/users/" + candidateA + "/role"

	expectStatus(t, call(t, mux, http.MethodPut, path, reviewerID, "reviewer", `{"role":"reviewer"}`), http.StatusForbidden, "reviewer")
	expectStatus(t, call(t, mux, http.MethodPut, path, candidateA, "candidate", `{"role":"reviewer"}`), http.StatusForbidden, "candidate promotes self")
	expectStatus(t, call(t, mux, http.MethodPut, path, adminID, "admin", `{"role":"admin"}`), http.StatusUnprocessableEntity, "grant admin")
	expectStatus(t, call(t, mux, http.MethodPut, "/v1/admin/users/"+adminID+"/role", adminID, "admin", `{"role":"candidate"}`), http.StatusForbidden, "admin demotes self")
	expectStatus(t, call(t, mux, http.MethodPut, "/v1/admin/users/"+candidateB+"/role", adminID, "admin", `{"role":"reviewer"}`), http.StatusNotFound, "unknown user")
	expectStatus(t, call(t, mux, http.MethodPut, "/v1/admin/users/not-a-uuid/role", adminID, "admin", `{"role":"reviewer"}`), http.StatusNotFound, "bad id")
	expectStatus(t, call(t, mux, http.MethodPut, path, adminID, "admin", `{"role":"reviewer","extra":1}`), http.StatusBadRequest, "unknown field")
	if roles.roles[candidateA] != "candidate" {
		t.Fatalf("no refused call may change the role, got %s", roles.roles[candidateA])
	}
}
