package httpapi

import (
	"net/http"

	"github.com/s-chukwu/Praxis/backend/internal/config"
)

func NewRouter(cfg config.Config) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", health)

	return cors(cfg.CORSAllowedOrigin, mux)
}