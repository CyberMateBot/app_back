package mediadownload

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
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

func TestHandlePrepareDownload_returnsDownloadURL(t *testing.T) {
	payload := `{"dataBase64":"aGVsbG8=","mimeType":"text/plain","filename":"test.txt"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, pathPrepareDownload, strings.NewReader(payload))
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Host = "api.example.com"

	handlePrepareDownload(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}

	var resp prepareResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(resp.DownloadURL, "https://api.example.com/v1/media/download/prepared/") {
		t.Fatalf("unexpected url %q", resp.DownloadURL)
	}
}

func TestHandlePreparedDownload_servesAttachment(t *testing.T) {
	id, err := newPreparedID()
	if err != nil {
		t.Fatal(err)
	}

	preparedDownloads.Store(id, preparedEntry{
		data:      []byte("hello"),
		mimeType:  "text/plain",
		filename:  "test.txt",
		expiresAt: time.Now().Add(preparedTTL),
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, pathPreparedPrefix+id+"/test.txt", nil)
	handlePreparedDownload(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	if got := rec.Header().Get("Content-Disposition"); got != `attachment; filename="test.txt"` {
		t.Fatalf("disposition %q", got)
	}
	if rec.Body.String() != "hello" {
		t.Fatalf("body %q", rec.Body.String())
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}
