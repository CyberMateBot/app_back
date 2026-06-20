package prompthistory

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWrap_NilStore_PassThrough(t *testing.T) {
	mux := Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}), nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/prompts/history/telegram/123", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusTeapot {
		t.Fatalf("status = %d, want 418", rec.Code)
	}
}

func TestWrap_DeleteTopicRoute(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	mux := Wrap(next, &Store{}, nil)

	req := httptest.NewRequest(http.MethodPost, "/v1/prompts/history/delete", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code == http.StatusTeapot {
		t.Fatalf("delete route fell through to next handler, status = %d", rec.Code)
	}

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (missing telegramId)", rec.Code)
	}
}
