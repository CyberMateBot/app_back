package mediadownload

import (
	"bufio"
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

	// Some providers (and a mis-guessed client-supplied filename, e.g. a
	// double extension like "image.png.bin") leave us with a useless
	// Content-Type. Peek at the real bytes instead of trusting the name,
	// so the file we hand back always has a correct type + extension.
	reader := bufio.NewReaderSize(resp.Body, 512)
	contentType := strings.TrimSpace(resp.Header.Get("Content-Type"))
	if isGenericContentType(contentType) {
		if peeked, peekErr := reader.Peek(512); peekErr == nil || len(peeked) > 0 {
			if sniffed := sniffContentType(peeked); sniffed != "" {
				contentType = sniffed
			}
		}
	}
	if contentType == "" {
		contentType = contentTypeFromName(filename)
	}

	// If the provider (or its error page) returned something that clearly
	// isn't media, fail loudly instead of handing the client a "photo" that
	// is actually an HTML/JSON error body — that's what shows up on the
	// device as "unsupported format" / "file is corrupted".
	if looksLikeNonMedia(contentType) {
		writeError(w, http.StatusBadGateway, "media provider did not return a valid file")
		return
	}

	filename = normalizeFilenameExtension(filename, contentType)
	setDownloadHeaders(w, contentType, filename)

	limited := io.LimitReader(reader, maxDownloadBytes+1)
	written, err := io.Copy(w, limited)
	if err != nil || written > maxDownloadBytes {
		// Headers (and possibly part of the body) are already flushed, so we
		// can't switch to a JSON error response without corrupting the
		// stream — just stop writing.
		return
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
	case ".avif":
		return "image/avif"
	case ".heic", ".heif":
		return "image/heic"
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".mov":
		return "video/quicktime"
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".ogg":
		return "audio/ogg"
	default:
		return "application/octet-stream"
	}
}

// mediaExtensionByContentType is the inverse of contentTypeFromName, used to
// fix up a (possibly wrong/duplicated) client-supplied filename extension
// once we know the real content type of the bytes we're serving.
var mediaExtensionByContentType = map[string]string{
	"image/png":       ".png",
	"image/jpeg":      ".jpeg",
	"image/webp":      ".webp",
	"image/gif":       ".gif",
	"image/avif":      ".avif",
	"image/heic":      ".heic",
	"video/mp4":       ".mp4",
	"video/webm":      ".webm",
	"video/quicktime": ".mov",
	"audio/mpeg":      ".mp3",
	"audio/wav":       ".wav",
	"audio/x-wav":     ".wav",
	"audio/ogg":       ".ogg",
}

// knownMediaExtensions lets us strip a stale/duplicated extension (e.g. the
// "image.png.bin" shape that comes from a fallback name like "image.png"
// getting a second extension appended to it) before appending the real one.
var knownMediaExtensions = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".webp": true, ".gif": true,
	".avif": true, ".heic": true, ".heif": true, ".mp4": true, ".webm": true,
	".mov": true, ".mp3": true, ".wav": true, ".ogg": true, ".bin": true,
}

func baseContentType(contentType string) string {
	if idx := strings.Index(contentType, ";"); idx >= 0 {
		contentType = contentType[:idx]
	}
	return strings.ToLower(strings.TrimSpace(contentType))
}

func isGenericContentType(contentType string) bool {
	switch baseContentType(contentType) {
	case "", "application/octet-stream", "binary/octet-stream", "application/binary":
		return true
	default:
		return false
	}
}

// sniffContentType inspects the real bytes (via http.DetectContentType) and
// returns "" when the result isn't useful (still generic).
func sniffContentType(peeked []byte) string {
	if len(peeked) == 0 {
		return ""
	}
	detected := http.DetectContentType(peeked)
	if isGenericContentType(detected) {
		return ""
	}
	return detected
}

// looksLikeNonMedia flags content types that indicate the provider handed us
// an error/HTML/JSON page instead of the actual image/video/audio file.
func looksLikeNonMedia(contentType string) bool {
	switch baseContentType(contentType) {
	case "text/html", "text/plain", "application/json", "application/xml", "text/xml":
		return true
	default:
		return false
	}
}

// normalizeFilenameExtension strips any (possibly duplicated) known media
// extension from filename and appends the extension that actually matches
// contentType, so the file we hand the client always opens correctly
// regardless of what extension the caller originally guessed.
func normalizeFilenameExtension(filename, contentType string) string {
	wanted, ok := mediaExtensionByContentType[baseContentType(contentType)]
	if !ok {
		return filename
	}

	base := filename
	for i := 0; i < 4; i++ {
		ext := path.Ext(base)
		if ext == "" || !knownMediaExtensions[strings.ToLower(ext)] {
			break
		}
		stripped := base[:len(base)-len(ext)]
		if stripped == "" {
			break
		}
		base = stripped
	}

	if base == "" {
		base = "cybermate-media"
	}

	return base + wanted
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
