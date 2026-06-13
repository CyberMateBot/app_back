package mediadownload

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestValidateRemoteURL_blocksPrivateHosts(t *testing.T) {
	cases := []string{
		"http://127.0.0.1/file.png",
		"http://localhost/secret",
		"http://10.0.0.1/a.png",
		"ftp://example.com/a.png",
	}

	for _, raw := range cases {
		parsed, err := http.NewRequest(http.MethodGet, raw, nil)
		if err != nil {
			t.Fatalf("parse %q: %v", raw, err)
		}
		if err := validateRemoteURL(parsed.URL); err == nil {
			t.Fatalf("expected error for %q", raw)
		}
	}
}

func TestValidateRemoteURL_allowsPublicHTTPS(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "https://cdn.example.com/image.png", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateRemoteURL(req.URL); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSanitizeFilename(t *testing.T) {
	if got := sanitizeFilename(`../../evil/name.png`); got != "name.png" {
		t.Fatalf("got %q", got)
	}
	if got := sanitizeFilename(`safe-file_1.mp4`); got != "safe-file_1.mp4" {
		t.Fatalf("got %q", got)
	}
}

func TestHandleDownload_proxiesUpstream(t *testing.T) {
	client := &http.Client{
		Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
			rec := httptest.NewRecorder()
			rec.Header().Set("Content-Type", "image/png")
			rec.WriteHeader(http.StatusOK)
			_, _ = rec.Write([]byte("png-bytes"))
			return rec.Result(), nil
		}),
	}

	mux := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleDownload(w, r, client)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodGet,
		pathMediaDownload+"?url="+url.QueryEscape("https://cdn.example.com/out.png")+"&filename=test.png",
		nil,
	)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Disposition"); got != `attachment; filename="test.png"` {
		t.Fatalf("disposition %q", got)
	}
	if rec.Body.String() != "png-bytes" {
		t.Fatalf("body %q", rec.Body.String())
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}
