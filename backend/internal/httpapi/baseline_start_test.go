package httpapi

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/s-chukwu/Praxis/backend/internal/workflow"
)

var startCatalogJSON = strings.Replace(catalogJSON, `"current":{"scenarios":"1.2.0"}`,
	`"current":{"scenarios":"1.2.0","role_standard":"1.0.0","allocation_policy":"1.0.0"}`, 1)

func newStartRouter(catalog CatalogReader) (http.Handler, *fakeAttempts) {
	attempts := newFakeAttempts()
	mux := http.NewServeMux()
	registerAttemptRoutes(mux, fakeAuth{}, attempts)
	registerBaselineRoutes(mux, fakeAuth{}, catalog, attempts, "CS-01-A")
	return mux, attempts
}

func TestStartBaselineCreatesAnOpenAttemptWithPinnedVersions(t *testing.T) {
	h, attempts := newStartRouter(&fakeCatalog{response: startCatalogJSON})

	rec := call(t, h, http.MethodPost, "/v1/baseline/start", candidateA, "candidate", `{"program":"operator"}`)
	expectStatus(t, rec, http.StatusCreated, "first start")
	for _, want := range []string{`"status":"in_progress"`, `"scenario_id":"CS-01-A"`, `"scenario_version":"1.2.0"`,
		`"rubric_version":"0.2.0"`, `"role_standard_version":"1.0.0"`, `"policy_version":"1.0.0"`, `"target_track":"operator"`} {
		if !strings.Contains(rec.Body.String(), want) {
			t.Errorf("missing %s in %s", want, rec.Body.String())
		}
	}
	if len(attempts.items) != 1 {
		t.Fatalf("expected one attempt, got %d", len(attempts.items))
	}
}

func TestStartBaselineAgainReturnsTheSameAttempt(t *testing.T) {
	h, attempts := newStartRouter(&fakeCatalog{response: startCatalogJSON})
	expectStatus(t, call(t, h, http.MethodPost, "/v1/baseline/start", candidateA, "candidate", `{"program":"operator"}`), http.StatusCreated, "first")
	expectStatus(t, call(t, h, http.MethodPost, "/v1/baseline/start", candidateA, "candidate", `{"program":"operator"}`), http.StatusOK, "again")
	expectStatus(t, call(t, h, http.MethodPost, "/v1/baseline/start", candidateB, "candidate", `{"program":"builder"}`), http.StatusCreated, "other candidate")
	if len(attempts.items) != 2 {
		t.Fatalf("expected one attempt per candidate, got %d", len(attempts.items))
	}

	// A voided baseline does not block a fresh one.
	for id, a := range attempts.items {
		if a.CandidateID == candidateA {
			a.Status = workflow.Voided
			attempts.items[id] = a
		}
	}
	expectStatus(t, call(t, h, http.MethodPost, "/v1/baseline/start", candidateA, "candidate", `{"program":"operator"}`), http.StatusCreated, "after void")
}

func TestStartBaselineRefusesBadInputAndOtherRoles(t *testing.T) {
	h, attempts := newStartRouter(&fakeCatalog{response: startCatalogJSON})
	expectStatus(t, call(t, h, http.MethodPost, "/v1/baseline/start", candidateA, "candidate", `{"program":"pilot"}`), http.StatusUnprocessableEntity, "unknown program")
	expectStatus(t, call(t, h, http.MethodPost, "/v1/baseline/start", candidateA, "candidate", `{}`), http.StatusUnprocessableEntity, "no program")
	expectStatus(t, call(t, h, http.MethodPost, "/v1/baseline/start", reviewerID, "reviewer", `{"program":"operator"}`), http.StatusForbidden, "reviewer")
	if len(attempts.items) != 0 {
		t.Fatalf("refused calls must not create attempts, got %d", len(attempts.items))
	}
}

func TestStartBaselineDoesNotCreateAnythingWhenTheScenarioFails(t *testing.T) {
	h, attempts := newStartRouter(&fakeCatalog{err: errors.New("connection refused")})
	rec := call(t, h, http.MethodPost, "/v1/baseline/start", candidateA, "candidate", `{"program":"operator"}`)
	expectStatus(t, rec, http.StatusBadGateway, "ai down")
	if len(attempts.items) != 0 || strings.Contains(rec.Body.String(), "connection refused") {
		t.Fatalf("expected no attempt and no leaked detail, got %d attempts, %s", len(attempts.items), rec.Body.String())
	}
}
