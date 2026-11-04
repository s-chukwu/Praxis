package auth

import (
	"errors"
	"net/http"
	"strings"
)

var ErrMissingToken = errors.New("missing bearer token")

// Identity is who is calling the API and what role the database says they have.
type Identity struct {
	UserID string
	Role   string
}

type Authenticator struct {
	verifier *Verifier
	profiles *Profiles
}

// New returns nil when Supabase is not configured, so the local stack still
// starts without keys. Protected routes answer 503 in that case.
func New(supabaseURL, serviceKey string) *Authenticator {
	if strings.TrimSpace(supabaseURL) == "" || serviceKey == "" {
		return nil
	}
	return &Authenticator{
		verifier: NewVerifier(supabaseURL),
		profiles: NewProfiles(supabaseURL, serviceKey),
	}
}

func (a *Authenticator) Authenticate(r *http.Request) (Identity, error) {
	token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !ok || strings.TrimSpace(token) == "" {
		return Identity{}, ErrMissingToken
	}

	claims, err := a.verifier.Verify(r.Context(), strings.TrimSpace(token))
	if err != nil {
		return Identity{}, err
	}

	role, err := a.profiles.Role(r.Context(), claims.Subject)
	if err != nil {
		return Identity{}, err
	}
	return Identity{UserID: claims.Subject, Role: role}, nil
}
