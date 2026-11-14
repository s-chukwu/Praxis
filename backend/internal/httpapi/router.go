package httpapi

import (
	"net/http"

	"github.com/s-chukwu/Praxis/backend/internal/aiclient"
	"github.com/s-chukwu/Praxis/backend/internal/auth"
	"github.com/s-chukwu/Praxis/backend/internal/config"
	"github.com/s-chukwu/Praxis/backend/internal/store"
)

// NewRouter builds the routes without a database, so the attempt routes answer 503.
// It keeps the original signature, which the existing router tests use.
func NewRouter(cfg config.Config) http.Handler {
	return NewRouterWithStore(cfg, nil)
}

// NewRouterWithStore builds every route. db may be nil (no DATABASE_URL), in which
// case the attempt routes answer 503 instead of failing at startup.
func NewRouterWithStore(cfg config.Config, db *store.Store) http.Handler {
	var authn Authenticator
	if real := auth.New(cfg.SupabaseURL, cfg.SupabaseServiceRoleKey); real != nil {
		authn = real
	}

	var attempts AttemptStore
	var baselineStarter BaselineStarter
	var reviews ReviewStore
	var finalizer Finalizer
	var plans PlanStore
	var roles RoleStore
	var userLister UserLister

	if db != nil {
		roles = db
		userLister = db
		attempts = db
		baselineStarter = db
		reviews = db
		finalizer = db
		plans = db
	}

	// Without a valid shared secret the AI client cannot sign calls, so the routes
	// that need it answer 503 instead of failing at startup.
	var scorer Scorer
	var planner Planner
	var catalog CatalogReader
	var packets PacketReader
	if client, err := aiclient.New(cfg.AIServiceURL, cfg.AIServiceSharedSecret); err == nil {
		scorer = client
		planner = client
		catalog = client
		packets = client
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", health)

	// Temporary route for testing Sentry error reporting.
	mux.HandleFunc("GET /debug/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("sentry test")
	})


	mux.Handle("GET /v1/me", requireRole(authn)(http.HandlerFunc(me)))
	registerAttemptRoutes(mux, authn, attempts)
	registerReviewRoutes(mux, authn, reviews)
	registerFinalizeRoute(mux, authn, reviews, finalizer, scorer)
	registerPlanRoutes(mux, authn, reviews, plans, planner)
	registerBaselineRoute(mux, authn, catalog, cfg.BaselineScenarioID)
	registerBaselineStartRoute(mux, authn, catalog, baselineStarter, cfg.BaselineScenarioID)
	registerAssessorRoute(mux, authn, reviews, packets)
	registerQueueRoute(mux, authn, attempts, reviews)
	registerAdminRoutes(mux, authn, roles)
	registerAdminUsersRoute(mux, authn, userLister)

	return cors(cfg.CORSAllowedOrigin, mux)
}