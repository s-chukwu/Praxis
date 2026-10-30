package httpapi

import (
	"context"
	"errors"
	"log"
	"net/http"
	"slices"

	"github.com/s-chukwu/Praxis/backend/internal/auth"
)

type ctxKey struct{}

func identityFrom(ctx context.Context) (auth.Identity, bool) {
	id, ok := ctx.Value(ctxKey{}).(auth.Identity)
	return id, ok
}

type errorBody struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, errorBody{Error: errorDetail{Code: code, Message: message}})
}

// requireRole checks the login token, looks up the user's role in the database
// and, when roles are listed, only lets those roles through.
func requireRole(authn *auth.Authenticator, allowed ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if authn == nil {
				writeError(w, http.StatusServiceUnavailable, "auth_not_configured",
					"login checking is not configured on this server")
				return
			}

			id, err := authn.Authenticate(r)
			switch {
			case err == nil:
			case errors.Is(err, auth.ErrMissingToken),
				errors.Is(err, auth.ErrInvalidToken),
				errors.Is(err, auth.ErrNoProfile):
				writeError(w, http.StatusUnauthorized, "unauthenticated",
					"missing, invalid or expired login token")
				return
			default:
				log.Printf("auth unavailable: %v", err)
				writeError(w, http.StatusServiceUnavailable, "auth_unavailable",
					"could not check the login right now")
				return
			}

			if len(allowed) > 0 && !slices.Contains(allowed, id.Role) {
				writeError(w, http.StatusForbidden, "forbidden",
					"your role is not allowed to do this")
				return
			}

			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, id)))
		})
	}
}

// me returns who the caller is. Useful for the frontend and for testing logins.
func me(w http.ResponseWriter, r *http.Request) {
	id, ok := identityFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "no identity")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"id": id.UserID, "role": id.Role})
}
