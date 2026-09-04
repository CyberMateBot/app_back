package siteapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/twelvepills-936/tgapp-/pkg/config"
)

func TestWrap_BlocksWeakJWTSecretInProduction(t *testing.T) {
	t.Setenv("ENVIRONMENT", "production")

	// uc is nil: the weak-secret gate must reject the request before any
	// usecase method is invoked, so a nil usecase is safe here.
	handler := Wrap(http.NotFoundHandler(), nil, config.ConfigJWT{Secret: "your-super-secret-jwt-key-change-this"})

	for _, path := range []string{"/v1/site/auth/register", "/v1/site/auth/login", "/v1/site/auth/me", "/v1/site/prompts"} {
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{}`))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("%s: status = %d, want 503 when JWT secret is the well-known default", path, rec.Code)
		}
	}
}

func TestWrap_AllowsPublicModelsEvenWithWeakSecretInProduction(t *testing.T) {
	t.Setenv("ENVIRONMENT", "production")

	handler := Wrap(http.NotFoundHandler(), nil, config.ConfigJWT{Secret: "your-super-secret-jwt-key-change-this"})

	req := httptest.NewRequest(http.MethodGet, "/v1/site/models", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 for the public model catalog", rec.Code)
	}
}

func TestWrap_AllowsStrongSecretInProduction(t *testing.T) {
	t.Setenv("ENVIRONMENT", "production")

	handler := Wrap(http.NotFoundHandler(), nil, config.ConfigJWT{Secret: "a-sufficiently-long-random-production-secret"})

	req := httptest.NewRequest(http.MethodGet, "/v1/site/auth/me", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Falls through to requireAuth and fails with 401 (missing header), not
	// 503 — proving the weak-secret gate did not trigger.
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401 (auth required, not 503 config error)", rec.Code)
	}
}

func TestWrap_WeakSecretAllowedOutsideProduction(t *testing.T) {
	t.Setenv("ENVIRONMENT", "development")

	handler := Wrap(http.NotFoundHandler(), nil, config.ConfigJWT{Secret: "your-super-secret-jwt-key-change-this"})

	req := httptest.NewRequest(http.MethodGet, "/v1/site/auth/me", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401 in development even with the default secret", rec.Code)
	}
}
