package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/s-chukwu/Praxis/backend/internal/aiclient"
)

const catalogJSON = `{"current":{"scenarios":"1.2.0"},"scenario_banks":[
	{"version":"1.1.0","scenarios":[{"id":"CS-01-A","stage":"baseline","timebox_minutes":9,
		"expected_artifacts":["customer_reply"],"candidate_brief":"Old brief."}]},
	{"version":"1.2.0","scenarios":[
		{"id":"CS-01-A","stage":"baseline","pair":"A","rubric_version":"0.2.0","timebox_minutes":12,
			"expected_artifacts":["customer_reply","internal_note","verification_steps","ai_interaction"],
			"candidate_brief":"A customer was debited twice."},
		{"id":"CS-01-B","stage":"final","timebox_minutes":12,
			"expected_artifacts":["customer_reply"],"candidate_brief":"Final task brief."}]}]}`

type fakeCatalog struct {
	response  string
	err       error
	calls     int
	lastActor aiclient.Actor
}

func (f *fakeCatalog) Catalog(ctx context.Context, actor aiclient.Actor, requestID string) (json.RawMessage, error) {
	f.calls++
	f.lastActor = actor
	if f.err != nil {
		return nil, f.err
	}
	return json.RawMessage(f.response), nil
}

func newBaselineRouter(catalog CatalogReader, scenarioID string) http.Handler {
	mux := http.NewServeMux()
	registerBaselineRoute(mux, fakeAuth{}, catalog, scenarioID)
	return mux
}

func TestBaselineScenarioReturnsOnlyCandidateSafeFields(t *testing.T) {
	catalog := &fakeCatalog{response: catalogJSON}
	h := newBaselineRouter(catalog, "CS-01-A")

	rec := call(t, h, http.MethodGet, "/v1/baseline/scenario", candidateA, "candidate", "")
	expectStatus(t, rec, http.StatusOK, "get scenario")

	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 5 {
		t.Fatalf("expected exactly 5 fields, got %v", got)
	}
	body := rec.Body.String()
	for _, want := range []string{`"scenario_id":"CS-01-A"`, `"version":"1.2.0"`, `"timebox_minutes":12`,
		`"brief":"A customer was debited twice."`, `"ai_interaction"`} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %s in %s", want, body)
		}
	}
	if catalog.lastActor != aiclient.CandidateActor(candidateA) {
		t.Fatalf("the call must be signed as the candidate, got %q", catalog.lastActor)
	}
}

func TestBaselineScenarioIsForCandidatesOnly(t *testing.T) {
	h := newBaselineRouter(&fakeCatalog{response: catalogJSON}, "CS-01-A")
	expectStatus(t, call(t, h, http.MethodGet, "/v1/baseline/scenario", reviewerID, "reviewer", ""), http.StatusForbidden, "reviewer")
	expectStatus(t, call(t, h, http.MethodGet, "/v1/baseline/scenario", "", "", ""), http.StatusUnauthorized, "anonymous")
}

func TestBaselineScenarioRefusesNonBaselineOrUnknownIDs(t *testing.T) {
	for _, id := range []string{"CS-01-B", "CS-99-A"} {
		h := newBaselineRouter(&fakeCatalog{response: catalogJSON}, id)
		rec := call(t, h, http.MethodGet, "/v1/baseline/scenario", candidateA, "candidate", "")
		expectStatus(t, rec, http.StatusBadGateway, id)
		if strings.Contains(rec.Body.String(), "Final task brief") {
			t.Fatalf("the final task brief leaked for %s", id)
		}
	}
}

func TestBaselineScenarioHandlesMissingOrFailingAIService(t *testing.T) {
	h := newBaselineRouter(nil, "CS-01-A")
	expectStatus(t, call(t, h, http.MethodGet, "/v1/baseline/scenario", candidateA, "candidate", ""), http.StatusServiceUnavailable, "no client")

	h = newBaselineRouter(&fakeCatalog{err: errors.New("connection refused")}, "CS-01-A")
	rec := call(t, h, http.MethodGet, "/v1/baseline/scenario", candidateA, "candidate", "")
	expectStatus(t, rec, http.StatusBadGateway, "ai down")
	if strings.Contains(rec.Body.String(), "connection refused") {
		t.Fatalf("internal error details leaked: %s", rec.Body.String())
	}
}

func TestBaselineScenarioIsCached(t *testing.T) {
	catalog := &fakeCatalog{response: catalogJSON}
	h := newBaselineRouter(catalog, "CS-01-A")
	for i := 0; i < 3; i++ {
		expectStatus(t, call(t, h, http.MethodGet, "/v1/baseline/scenario", candidateA, "candidate", ""), http.StatusOK, "get")
	}
	if catalog.calls != 1 {
		t.Fatalf("expected one catalog call, got %d", catalog.calls)
	}
}
