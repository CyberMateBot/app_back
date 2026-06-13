package mediadownload

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

const (
	pathMediaDownload = "/v1/media/download"
	maxDownloadBytes  = 100 << 20 // 100 MiB
)

// Wrap handles GET /v1/media/download — proxies remote media for browser downloads (CORS-safe).
func Wrap(next http.Handler) http.Handler {
	client := &http.Client{
		Timeout: 3 * time.Minute,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return errors.New("too many redirects")
			}
			return validateRemoteURL(req.URL)
		},
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == pathMediaDownload {
			handleDownload(w, r, client)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func handleDownload(w http.ResponseWriter, r *http.Request, client *http.Client) {
	rawURL := strings.TrimSpace(r.URL.Query().Get("url"))
	if rawURL == "" {
		writeError(w, http.StatusBadRequest, "url query parameter is required")
		return
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid url")
		return
	}
	if err := validateRemoteURL(parsed); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	filename := sanitizeFilename(r.URL.Query().Get("filename"))
	if filename == "" {
		filename = guessFilename(parsed)
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, parsed.String(), nil)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid url")
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		writeError(w, http.StatusBadGateway, "failed to fetch media")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		writeError(w, http.StatusBadGateway, fmt.Sprintf("upstream returned status %d", resp.StatusCode))
		return
	}

	contentType := strings.TrimSpace(resp.Header.Get("Content-Type"))
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = contentTypeFromName(filename)
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.Header().Set("Cache-Control", "no-store")

	limited := io.LimitReader(resp.Body, maxDownloadBytes+1)
	written, err := io.Copy(w, limited)
	if err != nil {
		return
	}
	if written > maxDownloadBytes {
		writeError(w, http.StatusBadGateway, "media file is too large")
	}
}

func validateRemoteURL(parsed *url.URL) error {
	if parsed == nil {
		return errors.New("invalid url")
	}

	scheme := strings.ToLower(strings.TrimSpace(parsed.Scheme))
	if scheme != "https" && scheme != "http" {
		return errors.New("only http(s) urls are allowed")
	}

	host := strings.TrimSpace(parsed.Hostname())
	if host == "" {
		return errors.New("invalid url host")
	}

	if strings.EqualFold(host, "localhost") {
		return errors.New("localhost urls are not allowed")
	}

	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
			return errors.New("private network urls are not allowed")
		}
	}

	return nil
}

func sanitizeFilename(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	name = path.Base(name)
	name = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '.', r == '-', r == '_':
			return r
		default:
			return '_'
		}
	}, name)
	return strings.Trim(name, "._")
}

func guessFilename(parsed *url.URL) string {
	if parsed == nil {
		return "cybermate-media.bin"
	}
	base := path.Base(parsed.Path)
	base = sanitizeFilename(base)
	if base == "" || base == "." {
		return "cybermate-media.bin"
	}
	return base
}

func contentTypeFromName(filename string) string {
	switch strings.ToLower(path.Ext(filename)) {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	default:
		return "application/octet-stream"
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// FetchURL downloads a validated remote URL (for tests and reuse).
func FetchURL(ctx context.Context, rawURL string, client *http.Client) (*http.Response, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return nil, err
	}
	if err := validateRemoteURL(parsed); err != nil {
		return nil, err
	}
	if client == nil {
		client = http.DefaultClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return nil, err
	}
	return client.Do(req)
}
