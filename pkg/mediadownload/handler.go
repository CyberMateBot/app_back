package mediadownload

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"
)

const (
	pathMediaDownload   = "/v1/media/download"
	pathPrepareDownload = "/v1/media/download/prepare"
	pathPreparedPrefix  = "/v1/media/download/prepared/"
	maxDownloadBytes    = 100 << 20 // 100 MiB
	maxPreparedBytes    = 12 << 20  // 12 MiB
	preparedTTL         = 10 * time.Minute
)

type preparedEntry struct {
	data      []byte
	mimeType  string
	filename  string
	expiresAt time.Time
}

var preparedDownloads sync.Map

// Wrap handles media download routes for browser and Telegram Mini App clients.
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

	go cleanupPreparedDownloads()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == pathMediaDownload:
			handleDownload(w, r, client)
			return
		case r.Method == http.MethodPost && r.URL.Path == pathPrepareDownload:
			handlePrepareDownload(w, r)
			return
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, pathPreparedPrefix):
			handlePreparedDownload(w, r)
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

	setDownloadHeaders(w, contentType, filename)

	limited := io.LimitReader(resp.Body, maxDownloadBytes+1)
	written, err := io.Copy(w, limited)
	if err != nil {
		return
	}
	if written > maxDownloadBytes {
		writeError(w, http.StatusBadGateway, "media file is too large")
	}
}

type prepareRequest struct {
	DataBase64 string `json:"dataBase64"`
	MimeType   string `json:"mimeType"`
	Filename   string `json:"filename"`
}

type prepareResponse struct {
	DownloadURL string `json:"downloadUrl"`
}

func handlePrepareDownload(w http.ResponseWriter, r *http.Request) {
	var req prepareRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	raw := strings.TrimSpace(req.DataBase64)
	if raw == "" {
		writeError(w, http.StatusBadRequest, "dataBase64 is required")
		return
	}

	data, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid dataBase64")
		return
	}
	if len(data) == 0 {
		writeError(w, http.StatusBadRequest, "empty file payload")
		return
	}
	if len(data) > maxPreparedBytes {
		writeError(w, http.StatusBadRequest, "file is too large")
		return
	}

	filename := sanitizeFilename(req.Filename)
	if filename == "" {
		filename = "cybermate-media.bin"
	}

	mimeType := strings.TrimSpace(req.MimeType)
	if mimeType == "" {
		mimeType = contentTypeFromName(filename)
	}

	id, err := newPreparedID()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to prepare download")
		return
	}

	preparedDownloads.Store(id, preparedEntry{
		data:      data,
		mimeType:  mimeType,
		filename:  filename,
		expiresAt: time.Now().Add(preparedTTL),
	})

	downloadURL := buildAbsoluteURL(r, pathPreparedPrefix+id+"/"+filename)
	writeJSON(w, http.StatusOK, prepareResponse{DownloadURL: downloadURL})
}

func handlePreparedDownload(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, pathPreparedPrefix)
	if slash := strings.Index(id, "/"); slash >= 0 {
		id = id[:slash]
	}

	id = sanitizePreparedID(id)
	if id == "" {
		writeError(w, http.StatusNotFound, "download not found")
		return
	}

	raw, ok := preparedDownloads.Load(id)
	if !ok {
		writeError(w, http.StatusNotFound, "download not found")
		return
	}

	entry := raw.(preparedEntry)
	if time.Now().After(entry.expiresAt) {
		preparedDownloads.Delete(id)
		writeError(w, http.StatusNotFound, "download expired")
		return
	}

	setDownloadHeaders(w, entry.mimeType, entry.filename)
	_, _ = w.Write(entry.data)
}

func setDownloadHeaders(w http.ResponseWriter, contentType, filename string) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.Header().Set("Cache-Control", "no-store")
}

func buildAbsoluteURL(r *http.Request, pathname string) string {
	scheme := "https"
	if proto := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); proto != "" {
		scheme = proto
	} else if r.TLS == nil {
		scheme = "http"
	}

	host := strings.TrimSpace(r.Host)
	if forwardedHost := strings.TrimSpace(r.Header.Get("X-Forwarded-Host")); forwardedHost != "" {
		host = forwardedHost
	}

	return fmt.Sprintf("%s://%s%s", scheme, host, pathname)
}

func newPreparedID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func sanitizePreparedID(id string) string {
	id = strings.TrimSpace(id)
	if id == "" {
		return ""
	}
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'f', r >= 'A' && r <= 'F', r >= '0' && r <= '9':
			return r
		default:
			return -1
		}
	}, id)
}

func cleanupPreparedDownloads() {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		preparedDownloads.Range(func(key, value any) bool {
			entry := value.(preparedEntry)
			if now.After(entry.expiresAt) {
				preparedDownloads.Delete(key)
			}
			return true
		})
	}
}

func decodeJSON(r *http.Request, dst any) error {
	if r.Body == nil {
		return errors.New("request body is required")
	}
	defer r.Body.Close()

	dec := json.NewDecoder(io.LimitReader(r.Body, maxPreparedBytes*2))
	if err := dec.Decode(dst); err != nil {
		return err
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
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
	writeJSON(w, status, map[string]string{"error": msg})
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
