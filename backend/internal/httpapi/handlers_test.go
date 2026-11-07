package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealth(t *testing.T) {
	rec := httptest.NewRecorder()
	health(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var got healthResponse
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	want := healthResponse{Status: "ok", Service: "praxis-backend", Version: "0.1.0"}
	if got != want {
		t.Errorf("body = %+v, want %+v", got, want)
	}
}

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	writeJSON(rec, http.StatusCreated, map[string]string{"hello": "world"})

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"hello":"world"}` {
		t.Errorf("body = %s", got)
	}
}

func TestWriteError(t *testing.T) {
	rec := httptest.NewRecorder()
	writeError(rec, http.StatusTeapot, "some_code", "some message")

	if rec.Code != http.StatusTeapot {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusTeapot)
	}
	var got errorBody
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if got.Error.Code != "some_code" || got.Error.Message != "some message" {
		t.Errorf("body = %+v", got)
	}
}

func TestCORS(t *testing.T) {
	const origin = "http://localhost:3000"

	tests := []struct {
		name         string
		method       string
		wantStatus   int
		wantNextCall bool
	}{
		{"GET passes through to next handler", http.MethodGet, http.StatusTeapot, true},
		{"POST passes through to next handler", http.MethodPost, http.StatusTeapot, true},
		{"OPTIONS preflight is answered without calling next", http.MethodOptions, http.StatusNoContent, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusTeapot)
			})

			rec := httptest.NewRecorder()
			cors(origin, next).ServeHTTP(rec, httptest.NewRequest(tc.method, "/anything", nil))

			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			if called != tc.wantNextCall {
				t.Errorf("next called = %v, want %v", called, tc.wantNextCall)
			}

			// CORS headers must be present on every response, including preflight.
			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != origin {
				t.Errorf("Allow-Origin = %q, want %q", got, origin)
			}
			if got := rec.Header().Get("Access-Control-Allow-Methods"); !strings.Contains(got, "OPTIONS") {
				t.Errorf("Allow-Methods = %q, want it to include OPTIONS", got)
			}
			headers := rec.Header().Get("Access-Control-Allow-Headers")
			for _, h := range []string{"Content-Type", "Authorization"} {
				if !strings.Contains(headers, h) {
					t.Errorf("Allow-Headers = %q, want it to include %s", headers, h)
				}
			}
		})
	}
}