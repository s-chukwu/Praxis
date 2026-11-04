package auth

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"strings"
	"time"
)

var ErrInvalidToken = errors.New("invalid token")

type Claims struct {
	Subject   string
	Role      string
	ExpiresAt time.Time
}

type Verifier struct {
	keys   *KeySet
	issuer string
	now    func() time.Time
}

// NormalizeBaseURL turns pasted values like https://xyz.supabase.co/rest/v1/
// into https://xyz.supabase.co
func NormalizeBaseURL(raw string) string {
	base := strings.TrimSpace(raw)
	base = strings.TrimRight(base, "/")
	base = strings.TrimSuffix(base, "/rest/v1")
	return strings.TrimRight(base, "/")
}

func NewVerifier(supabaseURL string) *Verifier {
	base := NormalizeBaseURL(supabaseURL)
	return &Verifier{
		keys:   NewKeySet(base + "/auth/v1/.well-known/jwks.json"),
		issuer: base + "/auth/v1",
		now:    time.Now,
	}
}

type tokenHeader struct {
	Alg string `json:"alg"`
	Kid string `json:"kid"`
}

type tokenPayload struct {
	Sub  string          `json:"sub"`
	Iss  string          `json:"iss"`
	Aud  json.RawMessage `json:"aud"`
	Exp  int64           `json:"exp"`
	Role string          `json:"role"`
}

// Verify checks the signature, issuer, audience and expiry of a Supabase
// access token. Only ES256 is accepted, so "none" and shared secret tokens fail.
func (v *Verifier) Verify(ctx context.Context, token string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	var header tokenHeader
	if err := decodeSegment(parts[0], &header); err != nil {
		return nil, ErrInvalidToken
	}
	if header.Alg != "ES256" || header.Kid == "" {
		return nil, ErrInvalidToken
	}

	key, err := v.keys.Key(ctx, header.Kid)
	if err != nil {
		if errors.Is(err, ErrKeysUnavailable) {
			return nil, err
		}
		return nil, ErrInvalidToken
	}

	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || len(sig) != 64 {
		return nil, ErrInvalidToken
	}
	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	r := new(big.Int).SetBytes(sig[:32])
	s := new(big.Int).SetBytes(sig[32:])
	if !ecdsa.Verify(key, digest[:], r, s) {
		return nil, ErrInvalidToken
	}

	var payload tokenPayload
	if err := decodeSegment(parts[1], &payload); err != nil {
		return nil, ErrInvalidToken
	}
	if payload.Iss != v.issuer || !audienceContains(payload.Aud, "authenticated") {
		return nil, ErrInvalidToken
	}
	if payload.Role != "authenticated" || payload.Sub == "" || payload.Exp == 0 {
		return nil, ErrInvalidToken
	}
	expires := time.Unix(payload.Exp, 0)
	if !v.now().Before(expires) {
		return nil, ErrInvalidToken
	}

	return &Claims{Subject: payload.Sub, Role: payload.Role, ExpiresAt: expires}, nil
}

func decodeSegment(segment string, out any) error {
	raw, err := base64.RawURLEncoding.DecodeString(segment)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, out)
}

func audienceContains(raw json.RawMessage, want string) bool {
	var single string
	if err := json.Unmarshal(raw, &single); err == nil {
		return single == want
	}
	var many []string
	if err := json.Unmarshal(raw, &many); err == nil {
		for _, aud := range many {
			if aud == want {
				return true
			}
		}
	}
	return false
}
