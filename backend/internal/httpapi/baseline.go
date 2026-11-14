package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/s-chukwu/Praxis/backend/internal/aiclient"
)

const catalogTTL = 10 * time.Minute

// CatalogReader is the part of the AI client the baseline scenario route needs.
type CatalogReader interface {
	Catalog(ctx context.Context, actor aiclient.Actor, requestID string) (json.RawMessage, error)
}

// baselineScenario is everything a candidate may see. The grading key lives in a
// separate assessor packet, so nothing here can reveal how answers are scored.
type baselineScenario struct {
	ScenarioID     string   `json:"scenario_id"`
	Version        string   `json:"version"`
	Brief          string   `json:"brief"`
	TimeboxMinutes int      `json:"timebox_minutes"`
	Fields         []string `json:"fields"`

	// Versions pinned on the attempt when it is created; never sent to the browser.
	RubricVersion       string `json:"-"`
	RoleStandardVersion string `json:"-"`
	PolicyVersion       string `json:"-"`
}

type catalogBody struct {
	Current struct {
		Scenarios        string `json:"scenarios"`
		RoleStandard     string `json:"role_standard"`
		AllocationPolicy string `json:"allocation_policy"`
	} `json:"current"`
	ScenarioBanks []struct {
		Version   string `json:"version"`
		Scenarios []struct {
			ID                string   `json:"id"`
			Stage             string   `json:"stage"`
			TimeboxMinutes    int      `json:"timebox_minutes"`
			ExpectedArtifacts []string `json:"expected_artifacts"`
			CandidateBrief    string   `json:"candidate_brief"`
			RubricVersion     string   `json:"rubric_version"`
		} `json:"scenarios"`
	} `json:"scenario_banks"`
}

var errScenarioMissing = errors.New("baseline scenario not found in the current scenario bank")

type baselineHandlers struct {
	catalog    CatalogReader
	scenarioID string

	mu       sync.Mutex
	cached   *baselineScenario
	cachedAt time.Time
}

// registerBaselineRoute wires GET /v1/baseline/scenario for candidates.
func registerBaselineRoute(mux *http.ServeMux, authn Authenticator, catalog CatalogReader, scenarioID string) {
	h := &baselineHandlers{catalog: catalog, scenarioID: scenarioID}
	mux.Handle("GET /v1/baseline/scenario", requireRole(authn, "candidate")(http.HandlerFunc(h.get)))
}

// pick copies only the candidate-safe fields, and refuses anything that is not a
// baseline scenario so a misconfigured ID can never leak the final-task twin.
func pick(raw json.RawMessage, scenarioID string) (baselineScenario, error) {
	var body catalogBody
	if err := json.Unmarshal(raw, &body); err != nil {
		return baselineScenario{}, err
	}
	for _, bank := range body.ScenarioBanks {
		if bank.Version != body.Current.Scenarios {
			continue
		}
		for _, s := range bank.Scenarios {
			if s.ID == scenarioID && s.Stage == "baseline" && s.CandidateBrief != "" {
				return baselineScenario{
					ScenarioID:     s.ID,
					Version:        bank.Version,
					Brief:          s.CandidateBrief,
					TimeboxMinutes: s.TimeboxMinutes,
					Fields:         s.ExpectedArtifacts,

					RubricVersion:       s.RubricVersion,
					RoleStandardVersion: body.Current.RoleStandard,
					PolicyVersion:       body.Current.AllocationPolicy,
				}, nil
			}
		}
	}
	return baselineScenario{}, errScenarioMissing
}

func (h *baselineHandlers) load(ctx context.Context, caller aiclient.Actor) (baselineScenario, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.cached != nil && time.Since(h.cachedAt) < catalogTTL {
		return *h.cached, nil
	}
	raw, err := h.catalog.Catalog(ctx, caller, "")
	if err != nil {
		return baselineScenario{}, err
	}
	s, err := pick(raw, h.scenarioID)
	if err != nil {
		return baselineScenario{}, err
	}
	h.cached, h.cachedAt = &s, time.Now()
	return s, nil
}

func (h *baselineHandlers) get(w http.ResponseWriter, r *http.Request) {
	if h.catalog == nil {
		writeError(w, http.StatusServiceUnavailable, "ai_not_configured",
			"the AI service is not configured on this server")
		return
	}
	caller, ok := callerOf(w, r)
	if !ok {
		return
	}

	s, err := h.load(r.Context(), aiclient.CandidateActor(caller.UserID))
	if err != nil {
		log.Printf("baseline scenario %q could not be loaded: %v", h.scenarioID, err)
		writeError(w, http.StatusBadGateway, "scenario_unavailable",
			"the baseline task could not be loaded right now")
		return
	}
	writeJSON(w, http.StatusOK, s)
}
