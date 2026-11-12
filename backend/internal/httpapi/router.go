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
	var reviews ReviewStore
	var finalizer Finalizer
	if db != nil {
		attempts = db
		reviews = db
		finalizer = db
	}

	// Without a valid shared secret the AI client cannot sign calls, so the routes
	// that need it answer 503 instead of failing at startup.
	var scorer Scorer
	if client, err := aiclient.New(cfg.AIServiceURL, cfg.AIServiceSharedSecret); err == nil {
		scorer = client
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", health)
	mux.Handle("GET /v1/me", requireRole(authn)(http.HandlerFunc(me)))
	registerAttemptRoutes(mux, authn, attempts)
	registerReviewRoutes(mux, authn, reviews)
	registerFinalizeRoute(mux, authn, reviews, finalizer, scorer)

	return cors(cfg.CORSAllowedOrigin, mux)
}
