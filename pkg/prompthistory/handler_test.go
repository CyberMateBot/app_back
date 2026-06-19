package prompthistory

import (
	"net/http"
	"net/http/httptest"
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
