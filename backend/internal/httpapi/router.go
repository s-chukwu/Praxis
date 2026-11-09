package httpapi

import (
	"net/http"

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
	if db != nil {
		attempts = db
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", health)
	mux.Handle("GET /v1/me", requireRole(authn)(http.HandlerFunc(me)))
	registerAttemptRoutes(mux, authn, attempts)

	return cors(cfg.CORSAllowedOrigin, mux)
}
