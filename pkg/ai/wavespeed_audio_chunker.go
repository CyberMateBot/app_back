package ai

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/twelvepills-936/tgapp-/pkg/config"
)

// perModelChunkChars is a conservative upper bound on how many characters we
// send to a single Wavespeed TTS call. Wavespeed models truncate or time-out on
// long inputs (Qwen3-TTS in particular caps at ~30s of audio, which is roughly
// 300–500 characters of natural speech). When the user prompt exceeds this,
// we split the text, generate each chunk separately, download the MP3s, and
// concatenate them into one file via ffmpeg — so the user gets a single
// playable audio_url with the full transcript spoken end-to-end.
var perModelChunkChars = map[string]int{
	"qwen3-tts":          500,
	"qwen3-tts-clone":    500,
	"omnivoice":          500,
	"elevenlabs-v3":      1500,
	"minimax-speech-2.6": 1500,
}

// isTTSModel reports whether the given model definition speaks text (as opposed
// to music-generation models like Mureka / ACE-Step where chunking makes no
// sense — the model isn't reading a script).
func isTTSModel(def mediaModelDef) bool {
	if isUnifiedQwen3TTS(def) {
		return true
	}
	switch def.ID {
	case "omnivoice", "elevenlabs-v3", "minimax-speech-2.6":
		return true
	}
	return false
}

// chunkLimitForModel returns the character budget for a single TTS request or
// zero when chunking should be disabled (unknown model → fall back to a single
// call, letting Wavespeed truncate as before).
func chunkLimitForModel(def mediaModelDef, slug string) int {
	if isUnifiedQwen3TTS(def) {
		if strings.HasSuffix(slug, "/voice-clone") {
			return perModelChunkChars["qwen3-tts-clone"]
		}
		return perModelChunkChars["qwen3-tts"]
	}
	if n, ok := perModelChunkChars[def.ID]; ok {
		return n
	}
	return 0
}

// sentenceBoundary matches the end of a natural sentence. We keep the trailing
// punctuation with the current chunk so speech synthesis retains the pause.
var sentenceBoundary = regexp.MustCompile(`([.!?…]+["»)\]]?\s+|[\n\r]+)`)

// splitTextIntoChunks breaks a long prompt into chunks that each fit under
// maxChars. Strategy (in order of preference):
//  1. Split at sentence boundaries — the natural pause matches TTS prosody.
//  2. If a sentence itself is too long, split at commas / semicolons.
//  3. Finally, hard-split on whitespace so no chunk ever exceeds the limit.
//
// The returned chunks are trimmed of surrounding whitespace and never empty.
func splitTextIntoChunks(text string, maxChars int) []string {
	text = strings.TrimSpace(text)
	if text == "" || maxChars <= 0 {
		return nil
	}
	if utf8.RuneCountInString(text) <= maxChars {
		return []string{text}
	}

	// Step 1 — sentence split.
	sentences := splitBySentences(text)

	// Merge adjacent sentences while under the budget so we don't waste
	// requests. If a single sentence overflows, split it further.
	var chunks []string
	var buf strings.Builder
	bufLen := 0

	flush := func() {
		trimmed := strings.TrimSpace(buf.String())
		if trimmed != "" {
			chunks = append(chunks, trimmed)
		}
		buf.Reset()
		bufLen = 0
	}

	for _, sentence := range sentences {
		sLen := utf8.RuneCountInString(sentence)

		if sLen > maxChars {
			flush()
			for _, sub := range splitSubSentence(sentence, maxChars) {
				chunks = append(chunks, sub)
			}
			continue
		}

		if bufLen+sLen > maxChars && bufLen > 0 {
			flush()
		}
		if buf.Len() > 0 {
			buf.WriteByte(' ')
			bufLen++
		}
		buf.WriteString(sentence)
		bufLen += sLen
	}
	flush()

	return chunks
}

func splitBySentences(text string) []string {
	indices := sentenceBoundary.FindAllStringIndex(text, -1)
	if len(indices) == 0 {
		return []string{text}
	}
	var sentences []string
	prev := 0
	for _, idx := range indices {
		end := idx[1]
		piece := strings.TrimSpace(text[prev:end])
		if piece != "" {
			sentences = append(sentences, piece)
		}
		prev = end
	}
	if prev < len(text) {
		tail := strings.TrimSpace(text[prev:])
		if tail != "" {
			sentences = append(sentences, tail)
		}
	}
	return sentences
}

// splitSubSentence breaks a single overlong sentence at commas / semicolons and,
// as a last resort, on whitespace, so every returned piece fits in maxChars.
func splitSubSentence(sentence string, maxChars int) []string {
	pieces := regexp.MustCompile(`([,;:—–]\s+)`).Split(sentence, -1)
	var out []string
	var buf strings.Builder
	bufLen := 0
	for _, p := range pieces {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		pl := utf8.RuneCountInString(p)
		if pl > maxChars {
			if buf.Len() > 0 {
				out = append(out, strings.TrimSpace(buf.String()))
				buf.Reset()
				bufLen = 0
			}
			out = append(out, hardWrapByRunes(p, maxChars)...)
			continue
		}
		if bufLen+pl+1 > maxChars && bufLen > 0 {
			out = append(out, strings.TrimSpace(buf.String()))
			buf.Reset()
			bufLen = 0
		}
		if buf.Len() > 0 {
			buf.WriteByte(' ')
			bufLen++
		}
		buf.WriteString(p)
		bufLen += pl
	}
	if buf.Len() > 0 {
		out = append(out, strings.TrimSpace(buf.String()))
	}
	return out
}

// hardWrapByRunes is a whitespace-boundary fallback for pathological cases —
// a single unbroken run of characters longer than maxChars.
func hardWrapByRunes(text string, maxChars int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		// No whitespace at all — slice by rune count.
		var out []string
		runes := []rune(text)
		for i := 0; i < len(runes); i += maxChars {
			end := i + maxChars
			if end > len(runes) {
				end = len(runes)
			}
			out = append(out, string(runes[i:end]))
		}
		return out
	}

	var out []string
	var buf strings.Builder
	bufLen := 0
	for _, w := range words {
		wl := utf8.RuneCountInString(w)
		if wl > maxChars {
			if buf.Len() > 0 {
				out = append(out, buf.String())
				buf.Reset()
				bufLen = 0
			}
			runes := []rune(w)
			for i := 0; i < len(runes); i += maxChars {
				end := i + maxChars
				if end > len(runes) {
					end = len(runes)
				}
				out = append(out, string(runes[i:end]))
			}
			continue
		}
		if bufLen+wl+1 > maxChars && bufLen > 0 {
			out = append(out, buf.String())
			buf.Reset()
			bufLen = 0
		}
		if buf.Len() > 0 {
			buf.WriteByte(' ')
			bufLen++
		}
		buf.WriteString(w)
		bufLen += wl
	}
	if buf.Len() > 0 {
		out = append(out, buf.String())
	}
	return out
}

// generateChunkedWavespeedAudio synthesizes each chunk with a fresh Wavespeed
// call, downloads every resulting mp3, concatenates them via ffmpeg, and
// re-uploads the merged file so the frontend receives a single URL.
//
// If any step of the merge fails (ffmpeg missing, chunk fetch fails, upload
// fails) we transparently fall back to returning the first chunk — the user
// still hears speech (though only the beginning) rather than an error.
func generateChunkedWavespeedAudio(
	ctx context.Context,
	cfg config.ConfigAI,
	chunks []string,
	req AudioRequest,
	def mediaModelDef,
) (AudioResponse, error) {
	if len(chunks) == 0 {
		return AudioResponse{}, &ProviderError{Provider: "wavespeed", Message: "no text to synthesize"}
	}

	slug := resolveWavespeedAudioSlug(def, req)
	urls := make([]string, 0, len(chunks))
	for _, chunk := range chunks {
		input := buildWavespeedAudioInput(def, chunk, req, slug)
		got, err := runWavespeedModel(ctx, cfg, slug, input)
		if err != nil {
			return AudioResponse{}, err
		}
		if len(got) == 0 || strings.TrimSpace(got[0]) == "" {
			return AudioResponse{}, &ProviderError{Provider: "wavespeed", Message: "empty audio chunk"}
		}
		urls = append(urls, strings.TrimSpace(got[0]))
	}

	if len(urls) == 1 {
		return AudioResponse{AudioURL: urls[0], Model: wavespeedAudioModelID(def, slug)}, nil
	}

	mergedURL, err := concatMP3AndUpload(ctx, cfg, urls)
	if err != nil {
		// Fall back to a single chunk. Better a truncated response than an
		// error message — users can retry with shorter input if they care.
		return AudioResponse{AudioURL: urls[0], Model: wavespeedAudioModelID(def, slug)}, nil
	}
	return AudioResponse{AudioURL: mergedURL, Model: wavespeedAudioModelID(def, slug)}, nil
}

// concatMP3AndUpload downloads each mp3 URL, concatenates them with ffmpeg
// using the concat demuxer (stream-copy — no re-encode, fast + lossless), and
// re-uploads the resulting file through Wavespeed's uploader so we get a
// public URL usable from the Mini App.
func concatMP3AndUpload(ctx context.Context, cfg config.ConfigAI, urls []string) (string, error) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return "", fmt.Errorf("ffmpeg not available: %w", err)
	}

	workDir, err := os.MkdirTemp("", "cm-tts-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(workDir)

	// Download each part.
	var listBuf strings.Builder
	for i, u := range urls {
		partPath := filepath.Join(workDir, fmt.Sprintf("part-%02d.mp3", i))
		if err := downloadFile(ctx, u, partPath); err != nil {
			return "", err
		}
		// ffmpeg concat demuxer wants POSIX-style paths, single-quoted.
		listBuf.WriteString(fmt.Sprintf("file '%s'\n", filepath.ToSlash(partPath)))
	}

	listPath := filepath.Join(workDir, "list.txt")
	if err := os.WriteFile(listPath, []byte(listBuf.String()), 0o600); err != nil {
		return "", err
	}

	outPath := filepath.Join(workDir, "merged.mp3")
	ffCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ffCtx,
		"ffmpeg", "-y",
		"-f", "concat", "-safe", "0",
		"-i", listPath,
		"-c", "copy",
		outPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("ffmpeg concat failed: %w: %s", err, truncate(string(out), 200))
	}

	// Upload merged mp3 via Wavespeed so it becomes publicly accessible.
	type uploadResult struct {
		url string
		err error
	}
	ch := make(chan uploadResult, 1)
	go func() {
		client := newWavespeedClient(cfg)
		u, err := client.Upload(outPath)
		ch <- uploadResult{url: u, err: err}
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case res := <-ch:
		if res.err != nil {
			return "", fmt.Errorf("upload merged audio: %w", res.err)
		}
		trimmed := strings.TrimSpace(res.url)
		if trimmed == "" {
			return "", fmt.Errorf("upload merged audio: empty url")
		}
		return trimmed, nil
	}
}

// downloadFile streams a URL to disk. Uses a bounded timeout so a slow chunk
// download can't stall the whole TTS request.
func downloadFile(ctx context.Context, url, path string) error {
	dlCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(dlCtx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("download %s: status %d", url, resp.StatusCode)
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(f, resp.Body); err != nil {
		return err
	}
	return nil
}
