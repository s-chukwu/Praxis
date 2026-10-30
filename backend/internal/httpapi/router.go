package httpapi

import (
	"net/http"

	"github.com/s-chukwu/Praxis/backend/internal/auth"
	"github.com/s-chukwu/Praxis/backend/internal/config"
)

func NewRouter(cfg config.Config) http.Handler {
	authn := auth.New(cfg.SupabaseURL, cfg.SupabaseServiceRoleKey)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", health)
	mux.Handle("GET /v1/me", requireRole(authn)(http.HandlerFunc(me)))

	return cors(cfg.CORSAllowedOrigin, mux)
}