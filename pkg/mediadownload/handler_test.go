package mediadownload

import (
	"context"
	"encoding/json"
	"errors"
	"net"
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

func TestValidateResolvedIP(t *testing.T) {
	cases := []struct {
		ip      string
		wantErr bool
	}{
		{"127.0.0.1", true},
		{"10.1.2.3", true},
		{"192.168.1.1", true},
		{"169.254.169.254", true}, // cloud metadata endpoint
		{"::1", true},
		{"0.0.0.0", true},
		{"224.0.0.1", true}, // multicast
		{"8.8.8.8", false},
		{"1.1.1.1", false},
	}

	for _, tc := range cases {
		err := validateResolvedIP(net.ParseIP(tc.ip))
		if tc.wantErr && err == nil {
			t.Fatalf("validateResolvedIP(%q) = nil, want error", tc.ip)
		}
		if !tc.wantErr && err != nil {
			t.Fatalf("validateResolvedIP(%q) = %v, want nil", tc.ip, err)
		}
	}
}

func TestSecureDialContext_BlocksDNSRebindingToPrivateIP(t *testing.T) {
	// Simulates an attacker-controlled hostname (passes validateRemoteURL,
	// which only inspects the literal URL string) that resolves to an
	// internal/cloud-metadata IP at connect time.
	lookup := func(_ context.Context, host string) ([]net.IP, error) {
		if host != "evil.example.com" {
			return nil, errors.New("unexpected host in test")
		}
		return []net.IP{net.ParseIP("169.254.169.254")}, nil
	}

	dial := secureDialContext(lookup, &net.Dialer{Timeout: time.Second})
	conn, err := dial(context.Background(), "tcp", "evil.example.com:443")
	if err == nil {
		if conn != nil {
			conn.Close()
		}
		t.Fatal("expected dial to be blocked for a hostname resolving to a private/metadata IP")
	}
}

func TestSecureDialContext_LookupErrorPropagates(t *testing.T) {
	lookup := func(_ context.Context, _ string) ([]net.IP, error) {
		return nil, errors.New("dns failure")
	}
	dial := secureDialContext(lookup, &net.Dialer{Timeout: time.Second})
	if _, err := dial(context.Background(), "tcp", "example.com:443"); err == nil {
		t.Fatal("expected lookup error to propagate")
	}
}

func TestSecureDialContext_NoAddressesReturnsError(t *testing.T) {
	lookup := func(_ context.Context, _ string) ([]net.IP, error) {
		return nil, nil
	}
	dial := secureDialContext(lookup, &net.Dialer{Timeout: time.Second})
	if _, err := dial(context.Background(), "tcp", "example.com:443"); err == nil {
		t.Fatal("expected error when resolver returns no addresses")
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

func TestHandleDownload_sniffsRealTypeWhenUpstreamIsGeneric(t *testing.T) {
	pngMagic := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0, 0, 0, 0}

	client := &http.Client{
		Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
			rec := httptest.NewRecorder()
			rec.Header().Set("Content-Type", "application/octet-stream")
			rec.WriteHeader(http.StatusOK)
			_, _ = rec.Write(pngMagic)
			return rec.Result(), nil
		}),
	}

	mux := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleDownload(w, r, client)
	})

	rec := httptest.NewRecorder()
	// Simulates the frontend bug: a fallback name like "image.png" with a
	// second (wrong) extension appended, and a CDN url without an extension.
	req := httptest.NewRequest(
		http.MethodGet,
		pathMediaDownload+"?url="+url.QueryEscape("https://cdn.example.com/output/abc123")+"&filename="+url.QueryEscape("image.png.bin"),
		nil,
	)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != "image/png" {
		t.Fatalf("content-type %q, want image/png (sniffed)", got)
	}
	if got := rec.Header().Get("Content-Disposition"); got != `attachment; filename="image.png"` {
		t.Fatalf("disposition %q, want a single .png extension", got)
	}
}

func TestHandleDownload_rejectsErrorPageServedAsMedia(t *testing.T) {
	client := &http.Client{
		Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
			rec := httptest.NewRecorder()
			rec.Header().Set("Content-Type", "application/json")
			rec.WriteHeader(http.StatusOK)
			_, _ = rec.Write([]byte(`{"error":"expired"}`))
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

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status %d, want 502 body %s", rec.Code, rec.Body.String())
	}
}

func TestNormalizeFilenameExtension(t *testing.T) {
	cases := []struct {
		filename    string
		contentType string
		want        string
	}{
		{"image.png.bin", "image/png", "image.png"},
		{"image.png", "image/webp", "image.webp"},
		{"video.mp4.mp4", "video/mp4", "video.mp4"},
		{"cybermate-media.bin", "video/mp4", "cybermate-media.mp4"},
		{"already-correct.jpeg", "image/jpeg", "already-correct.jpeg"},
	}

	for _, tc := range cases {
		if got := normalizeFilenameExtension(tc.filename, tc.contentType); got != tc.want {
			t.Fatalf("normalizeFilenameExtension(%q, %q) = %q, want %q", tc.filename, tc.contentType, got, tc.want)
		}
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}
