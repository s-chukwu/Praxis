package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/s-chukwu/Praxis/backend/internal/config"
)

const testOrigin = "http://localhost:3000"

func do(h http.Handler, method, path, bearer string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	if bearer != "" {
		req.Header.Set("Authorization", bearer)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestRouter_Health(t *testing.T) {
	// /health must work even with no Supabase configuration at all.
	h := NewRouter(config.Config{CORSAllowedOrigin: testOrigin})

	rec := do(h, http.MethodGet, "/health", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var got healthResponse
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if got.Status != "ok" {
		t.Errorf("status field = %q, want ok", got.Status)
	}
}

func TestRouter_RoutingRules(t *testing.T) {
	h := NewRouter(config.Config{CORSAllowedOrigin: testOrigin})

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{"unknown path is 404", http.MethodGet, "/nope", http.StatusNotFound},
		{"unversioned /me is 404", http.MethodGet, "/me", http.StatusNotFound},
		{"POST /health is 405", http.MethodPost, "/health", http.StatusMethodNotAllowed},
		{"POST /v1/me is 405", http.MethodPost, "/v1/me", http.StatusMethodNotAllowed},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := do(h, tc.method, tc.path, "")
			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
		})
	}
}

func TestRouter_CORSOnEveryResponse(t *testing.T) {
	h := NewRouter(config.Config{CORSAllowedOrigin: testOrigin})

	for _, path := range []string{"/health", "/v1/me", "/nope"} {
		t.Run(path, func(t *testing.T) {
			rec := do(h, http.MethodGet, path, "")
			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != testOrigin {
				t.Errorf("Allow-Origin = %q, want %q", got, testOrigin)
			}
		})
	}
}

func TestRouter_PreflightSkipsAuth(t *testing.T) {
	// Browsers send OPTIONS without an Authorization header, so it must not
	// be rejected by the auth check.
	h := NewRouter(config.Config{CORSAllowedOrigin: testOrigin})

	rec := do(h, http.MethodOptions, "/v1/me", "")

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
}

func TestRouter_MeWithoutSupabaseConfig(t *testing.T) {
	// The local stack starts without keys; protected routes must say so.
	h := NewRouter(config.Config{CORSAllowedOrigin: testOrigin})

	rec := do(h, http.MethodGet, "/v1/me", "Bearer anything")

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}
	if got := decodeError(t, rec); got.Code != "auth_not_configured" {
		t.Errorf("code = %q, want auth_not_configured", got.Code)
	}
}

func TestRouter_MeEndToEnd(t *testing.T) {
	fake := newFakeSupabase(t)
	fake.profileRole = "learner"
	h := NewRouter(config.Config{
		CORSAllowedOrigin:      testOrigin,
		SupabaseURL:            fake.srv.URL,
		SupabaseServiceRoleKey: testServiceKey,
	})

	t.Run("valid token returns identity", func(t *testing.T) {
		rec := do(h, http.MethodGet, "/v1/me", "Bearer "+fake.token(nil))

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body: %s)", rec.Code, rec.Body.String())
		}
		var got map[string]string
		if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if got["id"] != testUserID || got["role"] != "learner" {
			t.Errorf("body = %v", got)
		}
	})

	t.Run("missing token is 401", func(t *testing.T) {
		rec := do(h, http.MethodGet, "/v1/me", "")
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want 401", rec.Code)
		}
	})
}