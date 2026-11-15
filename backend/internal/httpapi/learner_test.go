package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/s-chukwu/Praxis/backend/internal/store"
)

type fakeLearners struct {
	profile store.LearnerProfile
	patch   *store.LearnerPatch
	erased  bool
}

func (f *fakeLearners) GetLearnerProfile(ctx context.Context, userID string) (store.LearnerProfile, error) {
	return f.profile, nil
}

func (f *fakeLearners) UpdateLearnerProfile(ctx context.Context, userID string, p store.LearnerPatch) error {
	f.patch = &p
	return nil
}

func (f *fakeLearners) EraseLearner(ctx context.Context, userID string, actor store.Actor) error {
	f.erased = true
	return nil
}

type fakeAccounts struct {
	err    error
	closed string
}

func (f *fakeAccounts) SoftDeleteUser(ctx context.Context, userID string) error {
	f.closed = userID
	return f.err
}

func newLearnerRouter(l *fakeLearners, a *fakeAccounts) http.Handler {
	mux := http.NewServeMux()
	registerLearnerRoutes(mux, fakeAuth{}, l, a)
	return mux
}

func TestLearnerProfileReadsInTheFrontendShape(t *testing.T) {
	op, cv := "operator", "Five years in retail support."
	l := &fakeLearners{profile: store.LearnerProfile{FullName: "Amari", Programme: &op, OnboardingDone: true, CVText: &cv, ClaimedSkills: []string{"C2"}}}
	rec := call(t, newLearnerRouter(l, &fakeAccounts{}), http.MethodGet, "/v1/me/learner", candidateA, "candidate", "")
	expectStatus(t, rec, http.StatusOK, "get")
	for _, want := range []string{`"name":"Amari"`, `"program":"operator"`, `"onboarded":true`, `"cvSubmitted":true`, `"skills":{"C2":"confirmed"}`} {
		if !strings.Contains(rec.Body.String(), want) {
			t.Errorf("missing %s in %s", want, rec.Body.String())
		}
	}
	expectStatus(t, call(t, newLearnerRouter(l, &fakeAccounts{}), http.MethodGet, "/v1/me/learner", reviewerID, "reviewer", ""), http.StatusForbidden, "reviewer")
}

func TestLearnerPatchOnlyAllowsClaimsAndValidFields(t *testing.T) {
	l := &fakeLearners{}
	h := newLearnerRouter(l, &fakeAccounts{})
	body := `{"name":" Amari U ","program":"builder","skills":{"C2":"confirmed","C4":"verified","C5":"rejected"},"cvText":"My CV"}`
	expectStatus(t, call(t, h, http.MethodPatch, "/v1/me/learner", candidateA, "candidate", body), http.StatusNoContent, "patch")
	if l.patch == nil || *l.patch.FullName != "Amari U" || *l.patch.Programme != "builder" || *l.patch.CVText != "My CV" {
		t.Fatalf("unexpected patch: %+v", l.patch)
	}
	if got := *l.patch.ClaimedSkills; len(got) != 1 || got[0] != "C2" {
		t.Fatalf("only confirmed claims may be stored, never verified: %v", got)
	}

	for name, bad := range map[string]string{
		"unknown program": `{"program":"pilot"}`,
		"bad skill code":  `{"skills":{"X9":"confirmed"}}`,
		"empty name":      `{"name":"  "}`,
		"cv not text":     `{"cvText":42}`,
	} {
		expectStatus(t, call(t, h, http.MethodPatch, "/v1/me/learner", candidateA, "candidate", bad), http.StatusUnprocessableEntity, name)
	}
}

func TestLearnerCanDeleteTheirCVWithoutTheAccount(t *testing.T) {
	l := &fakeLearners{}
	expectStatus(t, call(t, newLearnerRouter(l, &fakeAccounts{}), http.MethodPatch, "/v1/me/learner", candidateA, "candidate", `{"cvText":null}`), http.StatusNoContent, "delete cv")
	if !l.patch.ClearCV || l.erased {
		t.Fatalf("deleting the CV must clear it and keep the account: %+v erased=%v", l.patch, l.erased)
	}
}

func TestDeleteAccountErasesDataThenClosesSignIn(t *testing.T) {
	l, a := &fakeLearners{}, &fakeAccounts{}
	expectStatus(t, call(t, newLearnerRouter(l, a), http.MethodDelete, "/v1/me", candidateA, "candidate", ""), http.StatusNoContent, "delete")
	if !l.erased || a.closed != candidateA {
		t.Fatalf("expected data erased and account closed, got erased=%v closed=%q", l.erased, a.closed)
	}

	l, a = &fakeLearners{}, &fakeAccounts{err: errors.New("supabase down")}
	rec := call(t, newLearnerRouter(l, a), http.MethodDelete, "/v1/me", candidateA, "candidate", "")
	expectStatus(t, rec, http.StatusBadGateway, "supabase down")
	if strings.Contains(rec.Body.String(), "supabase down") {
		t.Fatalf("internal error leaked: %s", rec.Body.String())
	}
	expectStatus(t, call(t, newLearnerRouter(&fakeLearners{}, &fakeAccounts{}), http.MethodDelete, "/v1/me", reviewerID, "reviewer", ""), http.StatusForbidden, "reviewer")
}
