package auth

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const (
	testIssuer = "https://project.supabase.co/auth/v1"
	testUser   = "11111111-1111-1111-1111-111111111111"
	testKid    = "key-1"
)

var testNow = time.Unix(1_700_000_000, 0)

func b64(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

func newTestVerifier(t *testing.T) (*Verifier, *ecdsa.PrivateKey) {
	t.Helper()

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		doc := map[string]any{
			"keys": []map[string]string{{
				"kty": "EC",
				"crv": "P-256",
				"kid": testKid,
				"alg": "ES256",
				"x":   b64(priv.X.FillBytes(make([]byte, 32))),
				"y":   b64(priv.Y.FillBytes(make([]byte, 32))),
			}},
		}
		json.NewEncoder(w).Encode(doc)
	}))
	t.Cleanup(srv.Close)

	return &Verifier{
		keys:   NewKeySet(srv.URL),
		issuer: testIssuer,
		now:    func() time.Time { return testNow },
	}, priv
}

func signToken(t *testing.T, priv *ecdsa.PrivateKey, kid, alg string, claims map[string]any) string {
	t.Helper()

	header, _ := json.Marshal(map[string]string{"alg": alg, "kid": kid, "typ": "JWT"})
	payload, _ := json.Marshal(claims)
	signingInput := b64(header) + "." + b64(payload)

	digest := sha256.Sum256([]byte(signingInput))
	r, s, err := ecdsa.Sign(rand.Reader, priv, digest[:])
	if err != nil {
		t.Fatal(err)
	}
	sig := append(r.FillBytes(make([]byte, 32)), s.FillBytes(make([]byte, 32))...)
	return signingInput + "." + b64(sig)
}

func goodClaims() map[string]any {
	return map[string]any{
		"sub":  testUser,
		"iss":  testIssuer,
		"aud":  "authenticated",
		"role": "authenticated",
		"exp":  testNow.Add(time.Hour).Unix(),
	}
}

func TestVerifyAcceptsValidToken(t *testing.T) {
	v, priv := newTestVerifier(t)
	token := signToken(t, priv, testKid, "ES256", goodClaims())

	claims, err := v.Verify(context.Background(), token)
	if err != nil {
		t.Fatalf("expected a valid token, got %v", err)
	}
	if claims.Subject != testUser {
		t.Fatalf("unexpected subject %q", claims.Subject)
	}
}

func TestVerifyRejectsBadTokens(t *testing.T) {
	v, priv := newTestVerifier(t)

	expired := goodClaims()
	expired["exp"] = testNow.Add(-time.Minute).Unix()

	wrongIssuer := goodClaims()
	wrongIssuer["iss"] = "https://evil.example.com/auth/v1"

	anon := goodClaims()
	anon["role"] = "anon"

	wrongAudience := goodClaims()
	wrongAudience["aud"] = "something-else"

	otherKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	valid := signToken(t, priv, testKid, "ES256", goodClaims())
	forgedPayload := signToken(t, priv, testKid, "ES256", map[string]any{
		"sub": "22222222-2222-2222-2222-222222222222", "iss": testIssuer,
		"aud": "authenticated", "role": "authenticated", "exp": testNow.Add(time.Hour).Unix(),
	})

	cases := map[string]string{
		"expired":             signToken(t, priv, testKid, "ES256", expired),
		"wrong issuer":        signToken(t, priv, testKid, "ES256", wrongIssuer),
		"anon role":           signToken(t, priv, testKid, "ES256", anon),
		"wrong audience":      signToken(t, priv, testKid, "ES256", wrongAudience),
		"wrong alg":           signToken(t, priv, testKid, "HS256", goodClaims()),
		"unknown kid":         signToken(t, priv, "other-key", "ES256", goodClaims()),
		"signed by other key": signToken(t, otherKey, testKid, "ES256", goodClaims()),
		"not a token":         "hello",
		"empty":               "",
	}
	// Valid header and signature from one token glued to a payload from another.
	cases["tampered payload"] = splitJoin(valid, forgedPayload)

	for name, token := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := v.Verify(context.Background(), token)
			if !errors.Is(err, ErrInvalidToken) {
				t.Fatalf("expected ErrInvalidToken, got %v", err)
			}
		})
	}
}

// splitJoin takes the header and signature of a and the payload of b.
func splitJoin(a, b string) string {
	pa := splitDots(a)
	pb := splitDots(b)
	return pa[0] + "." + pb[1] + "." + pa[2]
}

func splitDots(token string) []string {
	parts := make([]string, 0, 3)
	start := 0
	for i := 0; i < len(token); i++ {
		if token[i] == '.' {
			parts = append(parts, token[start:i])
			start = i + 1
		}
	}
	return append(parts, token[start:])
}

func TestNormalizeBaseURL(t *testing.T) {
	cases := map[string]string{
		"https://abc.supabase.co/rest/v1/": "https://abc.supabase.co",
		"https://abc.supabase.co/":         "https://abc.supabase.co",
		"https://abc.supabase.co":          "https://abc.supabase.co",
	}
	for input, want := range cases {
		if got := NormalizeBaseURL(input); got != want {
			t.Fatalf("NormalizeBaseURL(%q) = %q, want %q", input, got, want)
		}
	}
}
