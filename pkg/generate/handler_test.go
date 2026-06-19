package generate

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/twelvepills-936/tgapp-/pkg/ai"
	"github.com/twelvepills-936/tgapp-/pkg/config"
)

func TestWrap_GenerateText_NotConfigured(t *testing.T) {
	svc := ai.NewService(config.ConfigAI{})
	mux := Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}), svc, nil, nil)

	body, _ := json.Marshal(map[string]string{"prompt": "hello"})
	req := httptest.NewRequest(http.MethodPost, pathGenerateText, bytes.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}
}

func TestWrap_GenerateText_MissingPrompt(t *testing.T) {
	svc := ai.NewService(config.ConfigAI{YandexAPIKey: "k", YandexFolderID: "f"})
	mux := Wrap(http.NotFoundHandler(), svc, nil, nil)

	req := httptest.NewRequest(http.MethodPost, pathGenerateText, bytes.NewReader([]byte(`{}`)))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestWrap_PassThrough(t *testing.T) {
	svc := ai.NewService(config.ConfigAI{})
	mux := Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), svc, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/app/links", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
}
