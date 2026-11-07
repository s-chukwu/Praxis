package httpapi

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/s-chukwu/Praxis/backend/internal/auth"
)

const (
	testUserID     = "11111111-1111-1111-1111-111111111111"
	testKid        = "key-1"
	testServiceKey = "service-role-key"
)

// fakeSupabase stands in for Supabase: it publishes a signing key (JWKS) and
// answers profile lookups. Tests change its fields to simulate failures.
type fakeSupabase struct {
	t    *testing.T
	srv  *httptest.Server
	priv *ecdsa.PrivateKey

	// profileRole is returned by the profiles endpoint. Empty means "no row".
	profileRole string
	// profileStatus, when non-zero, makes the profiles endpoint fail with it.
	profileStatus int
	// jwksStatus, when non-zero, makes the JWKS endpoint fail with it.
	jwksStatus int
	// lastAPIKey records the apikey header the profiles endpoint received.
	lastAPIKey string
}

func newFakeSupabase(t *testing.T) *fakeSupabase {
	t.Helper()

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	f := &fakeSupabase{t: t, priv: priv, profileRole: "learner"}

	mux := http.NewServeMux()
	mux.HandleFunc("/auth/v1/.well-known/jwks.json", func(w http.ResponseWriter, r *http.Request) {
		if f.jwksStatus != 0 {
			w.WriteHeader(f.jwksStatus)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{
			"keys": []map[string]string{{
				"kty": "EC",
				"crv": "P-256",
				"kid": testKid,
				"x":   b64(priv.X.FillBytes(make([]byte, 32))),
				"y":   b64(priv.Y.FillBytes(make([]byte, 32))),
			}},
		})
	})
	mux.HandleFunc("/rest/v1/profiles", func(w http.ResponseWriter, r *http.Request) {
		f.lastAPIKey = r.Header.Get("apikey")
		if f.profileStatus != 0 {
			w.WriteHeader(f.profileStatus)
			return
		}
		rows := []map[string]string{}
		if f.profileRole != "" {
			rows = append(rows, map[string]string{"role": f.profileRole})
		}
		json.NewEncoder(w).Encode(rows)
	})

	f.srv = httptest.NewServer(mux)
	t.Cleanup(f.srv.Close)
	return f
}

func (f *fakeSupabase) authenticator() *auth.Authenticator {
	f.t.Helper()
	a := auth.New(f.srv.URL, testServiceKey)
	if a == nil {
		f.t.Fatal("auth.New returned nil for a configured server")
	}
	return a
}

// token signs a valid access token. Pass overrides to change or remove claims
// (a nil value deletes the claim).
func (f *fakeSupabase) token(overrides map[string]any) string {
	f.t.Helper()
	claims := map[string]any{
		"sub":  testUserID,
		"iss":  f.srv.URL + "/auth/v1",
		"aud":  "authenticated",
		"role": "authenticated",
		"exp":  time.Now().Add(time.Hour).Unix(),
	}
	for k, v := range overrides {
		if v == nil {
			delete(claims, k)
		} else {
			claims[k] = v
		}
	}
	return signToken(f.t, f.priv, testKid, "ES256", claims)
}

func b64(data []byte) string { return base64.RawURLEncoding.EncodeToString(data) }

func signToken(t *testing.T, priv *ecdsa.PrivateKey, kid, alg string, claims map[string]any) string {
	t.Helper()

	header, _ := json.Marshal(map[string]string{"alg": alg, "kid": kid, "typ": "JWT"})
	payload, _ := json.Marshal(claims)
	input := b64(header) + "." + b64(payload)

	digest := sha256.Sum256([]byte(input))
	r, s, err := ecdsa.Sign(rand.Reader, priv, digest[:])
	if err != nil {
		t.Fatal(err)
	}
	sig := append(r.FillBytes(make([]byte, 32)), s.FillBytes(make([]byte, 32))...)
	return input + "." + b64(sig)
}

// silenceLogs hides the expected "auth unavailable" log lines during a test.
func silenceLogs(t *testing.T) {
	t.Helper()
	log.SetOutput(io.Discard)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })
}

func decodeError(t *testing.T, rec *httptest.ResponseRecorder) errorDetail {
	t.Helper()
	var body errorBody
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode error body: %v (raw: %s)", err, rec.Body.String())
	}
	return body.Error
}

func serve(h http.Handler, bearer string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if bearer != "" {
		req.Header.Set("Authorization", bearer)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// okHandler records whether it ran and the identity it saw.
func okHandler(called *bool, got *auth.Identity) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*called = true
		if id, ok := identityFrom(r.Context()); ok {
			*got = id
		}
		w.WriteHeader(http.StatusOK)
	})
}

func TestRequireRole_NotConfigured(t *testing.T) {
	called := false
	var id auth.Identity
	h := requireRole(nil)(okHandler(&called, &id))

	rec := serve(h, "Bearer whatever")

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}
	if got := decodeError(t, rec); got.Code != "auth_not_configured" {
		t.Errorf("code = %q, want auth_not_configured", got.Code)
	}
	if called {
		t.Error("next handler must not run when auth is not configured")
	}
}

func TestRequireRole_Unauthenticated(t *testing.T) {
	fake := newFakeSupabase(t)

	// A token signed by some other key, so the signature check fails.
	otherKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	forged := signToken(t, otherKey, testKid, "ES256", map[string]any{
		"sub": testUserID, "iss": fake.srv.URL + "/auth/v1",
		"aud": "authenticated", "role": "authenticated",
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	tests := []struct {
		name   string
		bearer string
	}{
		{"no Authorization header", ""},
		{"wrong scheme", "Basic abc123"},
		{"empty bearer token", "Bearer "},
		{"token is not a JWT", "Bearer not-a-jwt"},
		{"signed with an unknown key", "Bearer " + forged},
		{"expired token", "Bearer " + fake.token(map[string]any{"exp": time.Now().Add(-time.Hour).Unix()})},
		{"wrong issuer", "Bearer " + fake.token(map[string]any{"iss": "https://evil.example/auth/v1"})},
		{"wrong audience", "Bearer " + fake.token(map[string]any{"aud": "someone-else"})},
		{"anon role token", "Bearer " + fake.token(map[string]any{"role": "anon"})},
		{"missing subject", "Bearer " + fake.token(map[string]any{"sub": nil})},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			var id auth.Identity
			h := requireRole(fake.authenticator())(okHandler(&called, &id))

			rec := serve(h, tc.bearer)

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401", rec.Code)
			}
			if got := decodeError(t, rec); got.Code != "unauthenticated" {
				t.Errorf("code = %q, want unauthenticated", got.Code)
			}
			if called {
				t.Error("next handler must not run for an unauthenticated request")
			}
		})
	}
}

func TestRequireRole_NoProfile(t *testing.T) {
	fake := newFakeSupabase(t)
	fake.profileRole = "" // valid token, but no profiles row for the user

	called := false
	var id auth.Identity
	h := requireRole(fake.authenticator())(okHandler(&called, &id))

	rec := serve(h, "Bearer "+fake.token(nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
	if got := decodeError(t, rec); got.Code != "unauthenticated" {
		t.Errorf("code = %q, want unauthenticated", got.Code)
	}
	if called {
		t.Error("next handler must not run when the user has no profile")
	}
}

func TestRequireRole_ServiceUnavailable(t *testing.T) {
	silenceLogs(t)

	tests := []struct {
		name  string
		setup func(f *fakeSupabase)
	}{
		{"profile lookup fails", func(f *fakeSupabase) { f.profileStatus = http.StatusInternalServerError }},
		{"signing keys cannot be fetched", func(f *fakeSupabase) { f.jwksStatus = http.StatusInternalServerError }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fake := newFakeSupabase(t)
			token := fake.token(nil)
			tc.setup(fake)

			called := false
			var id auth.Identity
			h := requireRole(fake.authenticator())(okHandler(&called, &id))

			rec := serve(h, "Bearer "+token)

			// An outage must read as 503, never as "you are logged out" (401).
			if rec.Code != http.StatusServiceUnavailable {
				t.Fatalf("status = %d, want 503", rec.Code)
			}
			if got := decodeError(t, rec); got.Code != "auth_unavailable" {
				t.Errorf("code = %q, want auth_unavailable", got.Code)
			}
			if called {
				t.Error("next handler must not run when auth is unavailable")
			}
		})
	}
}

func TestRequireRole_RoleChecks(t *testing.T) {
	tests := []struct {
		name       string
		userRole   string
		allowed    []string
		wantStatus int
	}{
		{"no role list allows any authenticated user", "learner", nil, http.StatusOK},
		{"role in list is allowed", "admin", []string{"admin", "assessor"}, http.StatusOK},
		{"role not in list is forbidden", "learner", []string{"admin", "assessor"}, http.StatusForbidden},
		{"role match is case sensitive", "Admin", []string{"admin"}, http.StatusForbidden},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fake := newFakeSupabase(t)
			fake.profileRole = tc.userRole

			called := false
			var id auth.Identity
			h := requireRole(fake.authenticator(), tc.allowed...)(okHandler(&called, &id))

			rec := serve(h, "Bearer "+fake.token(nil))

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			if called != (tc.wantStatus == http.StatusOK) {
				t.Errorf("next called = %v, want %v", called, tc.wantStatus == http.StatusOK)
			}
			if tc.wantStatus == http.StatusForbidden {
				if got := decodeError(t, rec); got.Code != "forbidden" {
					t.Errorf("code = %q, want forbidden", got.Code)
				}
			}
		})
	}
}

func TestRequireRole_PutsIdentityInContext(t *testing.T) {
	fake := newFakeSupabase(t)
	fake.profileRole = "assessor"

	called := false
	var got auth.Identity
	h := requireRole(fake.authenticator())(okHandler(&called, &got))

	rec := serve(h, "Bearer "+fake.token(nil))

	if rec.Code != http.StatusOK || !called {
		t.Fatalf("status = %d, called = %v; want 200 and true", rec.Code, called)
	}
	if want := (auth.Identity{UserID: testUserID, Role: "assessor"}); got != want {
		t.Errorf("identity = %+v, want %+v", got, want)
	}
	// The role must come from the database, using the server's service key.
	if fake.lastAPIKey != testServiceKey {
		t.Errorf("profile lookup apikey = %q, want the service key", fake.lastAPIKey)
	}
}

func TestIdentityFrom_EmptyContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if _, ok := identityFrom(req.Context()); ok {
		t.Error("identityFrom on a bare context should report ok = false")
	}
}

func TestMe(t *testing.T) {
	t.Run("returns the caller's id and role", func(t *testing.T) {
		fake := newFakeSupabase(t)
		fake.profileRole = "learner"
		h := requireRole(fake.authenticator())(http.HandlerFunc(me))

		rec := serve(h, "Bearer "+fake.token(nil))

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		var got map[string]string
		if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if got["id"] != testUserID || got["role"] != "learner" {
			t.Errorf("body = %v", got)
		}
	})

	t.Run("401 when called without an identity", func(t *testing.T) {
		// Guards against someone mounting /v1/me without requireRole.
		rec := httptest.NewRecorder()
		me(rec, httptest.NewRequest(http.MethodGet, "/v1/me", nil))

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", rec.Code)
		}
		if got := decodeError(t, rec); got.Code != "unauthenticated" {
			t.Errorf("code = %q, want unauthenticated", got.Code)
		}
	})
}