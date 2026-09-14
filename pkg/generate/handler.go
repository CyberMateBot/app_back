package generate

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/twelvepills-936/tgapp-/pkg/ai"
	"github.com/twelvepills-936/tgapp-/pkg/billing"
	"github.com/twelvepills-936/tgapp-/pkg/prompthistory"
	"github.com/twelvepills-936/tgapp-/pkg/tokenguard"
)

// keepAliveInterval controls how often we flush a heartbeat byte to
// long-running generation responses so upstream reverse proxies
// (Timeweb / Cloudflare / etc.) don't drop the connection at their
// idle-timeout (~100-120s) while the LLM is still thinking.
const keepAliveInterval = 12 * time.Second

// runWithKeepAlive executes `exec` in a goroutine while writing the
// HTTP response prelude and flushing a single whitespace byte every
// keepAliveInterval. Whitespace before a JSON body is valid, so the
// client's JSON.parse still works. Once `exec` returns, we stop the
// heartbeat and let the caller finish writing the actual response.
//
// The status header is committed up-front (StatusOK) — that means
// even on failure the response is 200, so callers must indicate
// errors inside the JSON body. That's the trade-off for keeping the
// connection alive: the proxy stops caring about idle time as soon
// as we send *any* body byte, but browsers won't let us change the
// status code afterwards.
func runWithKeepAlive(ctx context.Context, w http.ResponseWriter, exec func() error) error {
	flusher, canFlush := w.(http.Flusher)
	if !canFlush {
		return exec()
	}

	// Advertise that we're going to stream chunked JSON. `X-Accel-
	// Buffering: no` disables nginx's default response buffering,
	// which otherwise swallows our heartbeats until end-of-response.
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Accel-Buffering", "no")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.WriteHeader(http.StatusOK)
	// Flush headers before we start ticking so the browser/proxy
	// have already seen response bytes when the LLM warms up.
	flusher.Flush()

	done := make(chan error, 1)
	go func() {
		done <- exec()
	}()

	ticker := time.NewTicker(keepAliveInterval)
	defer ticker.Stop()

	for {
		select {
		case err := <-done:
			return err
		case <-ticker.C:
			if _, writeErr := w.Write([]byte(" ")); writeErr != nil {
				// Client is already gone — cancel executor context
				// via the request context, then wait for it to
				// return so we don't leak the goroutine.
				return <-done
			}
			flusher.Flush()
		case <-ctx.Done():
			return <-done
		}
	}
}

const (
	pathGenerateModels = "/v1/generate/models"
	pathGenerateText   = "/v1/generate/text"
	pathGenerateImage  = "/v1/generate/image"
	pathGenerateVideo  = "/v1/generate/video"
	pathGenerateAudio  = "/v1/generate/audio"
	pathGenerate3D     = "/v1/generate/3d"
)

// Wrap adds POST /v1/generate/text and POST /v1/generate/image.
func Wrap(next http.Handler, svc *ai.Service, history *prompthistory.Store, tokens *tokenguard.Guard) http.Handler {
	if svc == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == pathGenerateModels:
			writeJSON(w, http.StatusOK, svc.ListModels())
			return
		case r.Method == http.MethodPost && r.URL.Path == pathGenerateText:
			handleText(w, r, svc, history, tokens)
			return
		case r.Method == http.MethodPost && r.URL.Path == pathGenerateImage:
			handleImage(w, r, svc, history, tokens)
			return
		case r.Method == http.MethodPost && r.URL.Path == pathGenerateVideo:
			handleVideo(w, r, svc, history, tokens)
			return
		case r.Method == http.MethodPost && r.URL.Path == pathGenerateAudio:
			handleAudio(w, r, svc, tokens)
			return
		case r.Method == http.MethodPost && r.URL.Path == pathGenerate3D:
			handle3D(w, r, svc, history, tokens)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type textGenerateRequest struct {
	ai.TextRequest
	TelegramID  string `json:"telegramId"`
	InitDataRaw string `json:"initDataRaw"`
	SessionID   string `json:"sessionId"`
	Category    string `json:"category"`
}

type textGenerateResponse struct {
	Text   string              `json:"text"`
	Model  string              `json:"model"`
	Format string              `json:"format,omitempty"`
	Item   *prompthistory.Item `json:"item,omitempty"`
}

func handleText(w http.ResponseWriter, r *http.Request, svc *ai.Service, history *prompthistory.Store, tokens *tokenguard.Guard) {
	var req textGenerateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Fast-fail obvious errors BEFORE the keep-alive wrapper commits
	// the 200 status header. Empty prompt / missing provider config
	// should be plain HTTP errors so clients (and tests) can rely on
	// the status code.
	if strings.TrimSpace(req.Prompt) == "" && strings.TrimSpace(req.Text) == "" {
		writeError(w, http.StatusBadRequest, ai.ErrPromptEmpty.Error())
		return
	}
	if !svc.HasTextProvider() {
		writeError(w, http.StatusServiceUnavailable, ai.ErrNotConfigured.Error())
		return
	}

	if err := ensureGenerationAccess(w, r, tokens, req.TelegramID, tokenguard.InitDataFromRequest(r, req.InitDataRaw), req.Model, "text"); err != nil {
		return
	}

	// Run the actual LLM call under a keep-alive wrapper. LLM
	// completions can take 30–180s; without heartbeats the reverse
	// proxy silently kills the connection at ~2min idle, which
	// users saw as the "попробуйте ещё раз" error.
	var out ai.TextResponse
	genErr := runWithKeepAlive(r.Context(), w, func() error {
		var innerErr error
		out, innerErr = svc.GenerateText(r.Context(), req.TextRequest)
		return innerErr
	})

	// From here on the response headers/status are already flushed,
	// so we can only communicate outcomes via the JSON body. Any
	// downstream helper that also calls w.WriteHeader will be a
	// no-op — that's OK, we just need the body to be valid JSON.
	if genErr != nil {
		writeStreamedError(r.Context(), w, "generate text", genErr)
		return
	}

	if err := chargeGeneration(r.Context(), tokens, req.TelegramID, out.Model, "text"); err != nil {
		writeStreamedError(r.Context(), w, "charge generation", err)
		return
	}

	resp := textGenerateResponse{
		Text:   out.Text,
		Model:  out.Model,
		Format: out.Format,
	}

	if history != nil {
		prompt := strings.TrimSpace(req.Prompt)
		if prompt == "" {
			prompt = strings.TrimSpace(req.Text)
		}
		category := strings.TrimSpace(req.Category)
		if category == "" {
			category = "text"
		}
		if item, saveErr := history.SaveAfterGenerate(
			r.Context(),
			req.TelegramID,
			prompt,
			out.Text,
			category,
			req.Model,
			req.SessionID,
		); saveErr != nil {
			slog.WarnContext(r.Context(), "failed to save prompt history after text generation", slog.Any("error", saveErr))
		} else if item != nil {
			resp.Item = item
		}
	}

	// Body was pre-flushed; just encode the final payload.
	_ = json.NewEncoder(w).Encode(resp)
}

// writeStreamedError encodes an error into a JSON body when the
// response headers have already been committed (200 OK). Clients
// look for the `error` field on 200 responses too.
func writeStreamedError(ctx context.Context, w http.ResponseWriter, op string, err error) {
	slog.ErrorContext(ctx, op, slog.Any("error", err))
	msg := "generation failed"
	var pe *ai.ProviderError
	switch {
	case errors.Is(err, ai.ErrPromptEmpty):
		msg = err.Error()
	case errors.Is(err, ai.ErrNotConfigured):
		msg = err.Error()
	case errors.As(err, &pe):
		msg = pe.Error()
	default:
		if em := err.Error(); em != "" {
			msg = em
		}
	}
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

type imageGenerateResponse struct {
	ImageURL    string              `json:"image_url,omitempty"`
	ImageURLs   []string            `json:"image_urls,omitempty"`
	ImageBase64 string              `json:"image_base64,omitempty"`
	Model       string              `json:"model"`
	Item        *prompthistory.Item `json:"item,omitempty"`
}

func handleImage(w http.ResponseWriter, r *http.Request, svc *ai.Service, history *prompthistory.Store, tokens *tokenguard.Guard) {
	var req ai.ImageRequest
	if err := decodeImageJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, formatDecodeError(err))
		return
	}
	if err := ensureImageGenerationAccess(w, r, tokens, req); err != nil {
		return
	}

	out, err := svc.GenerateImage(r.Context(), req)
	if err != nil {
		writeServiceError(w, r, "generate image", err)
		return
	}

	finalized := ai.FinalizeImageResponse(out)
	if err := chargeImageGeneration(r.Context(), tokens, req, finalized.Model); err != nil {
		tokenguard.WriteHTTPError(w, r, err)
		return
	}
	resp := imageGenerateResponse{
		ImageURL:    finalized.ImageURL,
		ImageURLs:   finalized.ImageURLs,
		ImageBase64: finalized.ImageBase64,
		Model:       finalized.Model,
	}

	if history != nil {
		prompt := strings.TrimSpace(req.Prompt)
		if prompt == "" {
			prompt = strings.TrimSpace(req.Text)
		}
		response := strings.TrimSpace(finalized.ImageURL)
		if response == "" && len(finalized.ImageURLs) > 0 {
			response = strings.TrimSpace(finalized.ImageURLs[0])
		}
		category := strings.TrimSpace(req.Category)
		if category == "" {
			category = strings.TrimSpace(req.Model)
		}
		if category == "" {
			category = "image"
		}
		if item, saveErr := history.SaveAfterGenerate(
			r.Context(),
			req.TelegramID,
			prompt,
			response,
			category,
			req.Model,
			req.SessionID,
		); saveErr != nil {
			slog.WarnContext(r.Context(), "failed to save prompt history after image generation", slog.Any("error", saveErr))
		} else if item != nil {
			resp.Item = item
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

type videoGenerateResponse struct {
	VideoURL string              `json:"video_url"`
	Model    string              `json:"model"`
	Item     *prompthistory.Item `json:"item,omitempty"`
}

func handleVideo(w http.ResponseWriter, r *http.Request, svc *ai.Service, history *prompthistory.Store, tokens *tokenguard.Guard) {
	var req ai.VideoRequest
	if err := decodeVideoJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, formatDecodeError(err))
		return
	}
	if err := ensureVideoGenerationAccess(w, r, tokens, req); err != nil {
		return
	}

	out, err := svc.GenerateVideo(r.Context(), req)
	if err != nil {
		writeServiceError(w, r, "generate video", err)
		return
	}

	if err := chargeVideoGeneration(r.Context(), tokens, req, out.Model); err != nil {
		tokenguard.WriteHTTPError(w, r, err)
		return
	}

	resp := videoGenerateResponse{
		VideoURL: out.VideoURL,
		Model:    out.Model,
	}

	if history != nil {
		prompt := strings.TrimSpace(req.Prompt)
		if prompt == "" {
			prompt = strings.TrimSpace(req.Text)
		}
		category := strings.TrimSpace(req.Category)
		if category == "" {
			category = strings.TrimSpace(req.Model)
		}
		if category == "" {
			category = "video"
		}
		if item, saveErr := history.SaveAfterGenerate(
			r.Context(),
			req.TelegramID,
			prompt,
			strings.TrimSpace(out.VideoURL),
			category,
			req.Model,
			req.SessionID,
		); saveErr != nil {
			slog.WarnContext(r.Context(), "failed to save prompt history after video generation", slog.Any("error", saveErr))
		} else if item != nil {
			resp.Item = item
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

type audioGenerateRequest struct {
	ai.AudioRequest
}

func handleAudio(w http.ResponseWriter, r *http.Request, svc *ai.Service, tokens *tokenguard.Guard) {
	var req audioGenerateRequest
	if err := decodeAudioJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, formatDecodeError(err))
		return
	}
	if err := ensureAudioGenerationAccess(w, r, tokens, req.AudioRequest); err != nil {
		return
	}

	out, err := svc.GenerateAudio(r.Context(), req.AudioRequest)
	if err != nil {
		writeServiceError(w, r, "generate audio", err)
		return
	}

	if err := chargeAudioGeneration(r.Context(), tokens, req.AudioRequest, out.Model); err != nil {
		tokenguard.WriteHTTPError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, out)
}

type threeDGenerateResponse struct {
	ModelURL string              `json:"model_url"`
	Model    string              `json:"model"`
	Item     *prompthistory.Item `json:"item,omitempty"`
}

func handle3D(w http.ResponseWriter, r *http.Request, svc *ai.Service, history *prompthistory.Store, tokens *tokenguard.Guard) {
	var req ai.ThreeDRequest
	if err := decode3DJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, formatDecodeError(err))
		return
	}
	if err := ensureThreeDGenerationAccess(w, r, tokens, req); err != nil {
		return
	}

	out, err := svc.GenerateThreeD(r.Context(), req)
	if err != nil {
		writeServiceError(w, r, "generate 3d", err)
		return
	}

	if err := chargeThreeDGeneration(r.Context(), tokens, req, out.Model); err != nil {
		tokenguard.WriteHTTPError(w, r, err)
		return
	}

	resp := threeDGenerateResponse{
		ModelURL: out.ModelURL,
		Model:    out.Model,
	}

	if history != nil {
		prompt := strings.TrimSpace(req.Prompt)
		if prompt == "" {
			prompt = strings.TrimSpace(req.Text)
		}
		category := strings.TrimSpace(req.Category)
		if category == "" {
			category = strings.TrimSpace(req.Model)
		}
		if category == "" {
			category = "3d"
		}
		if item, saveErr := history.SaveAfterGenerate(
			r.Context(),
			req.TelegramID,
			prompt,
			strings.TrimSpace(out.ModelURL),
			category,
			req.Model,
			req.SessionID,
		); saveErr != nil {
			slog.WarnContext(r.Context(), "failed to save prompt history after 3d generation", slog.Any("error", saveErr))
		} else if item != nil {
			resp.Item = item
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

const (
	maxJSONBodyDefault = 1 << 20  // 1 MiB — text / small payloads
	maxJSONBodyImage   = 12 << 20 // 12 MiB — imageBase64 (до ~8 MiB binary)
	maxJSONBodyVideo   = 24 << 20 // 24 MiB — videoBase64 (до ~16 MiB binary)
	maxJSONBodyAudio   = 12 << 20 // 12 MiB — audioBase64 for voice clone
	maxJSONBody3D      = 24 << 20 // 24 MiB — multiview image uploads
)

func decodeJSON(r *http.Request, dst any) error {
	return decodeJSONWithLimit(r, dst, maxJSONBodyDefault)
}

func decodeImageJSON(r *http.Request, dst any) error {
	return decodeJSONWithLimit(r, dst, maxJSONBodyImage)
}

func decodeVideoJSON(r *http.Request, dst any) error {
	return decodeJSONWithLimit(r, dst, maxJSONBodyVideo)
}

func decodeAudioJSON(r *http.Request, dst any) error {
	return decodeJSONWithLimit(r, dst, maxJSONBodyAudio)
}

func decode3DJSON(r *http.Request, dst any) error {
	return decodeJSONWithLimit(r, dst, maxJSONBody3D)
}

func decodeJSONWithLimit(r *http.Request, dst any, maxBytes int64) error {
	if r.Body == nil {
		return errors.New("request body is required")
	}
	defer r.Body.Close()
	dec := json.NewDecoder(io.LimitReader(r.Body, maxBytes))
	if err := dec.Decode(dst); err != nil {
		return err
	}
	return nil
}

func formatDecodeError(err error) string {
	if errors.Is(err, io.EOF) || strings.Contains(err.Error(), "unexpected EOF") {
		return "request body too large or truncated (max image payload ~12 MiB)"
	}
	return err.Error()
}

func ensureTokens(w http.ResponseWriter, r *http.Request, tokens *tokenguard.Guard, telegramID, initDataRaw string) error {
	return ensureGenerationAccess(w, r, tokens, telegramID, initDataRaw, "", "")
}

func ensureGenerationAccess(w http.ResponseWriter, r *http.Request, tokens *tokenguard.Guard, telegramID, initDataRaw, modelID, category string) error {
	if tokens == nil {
		return nil
	}
	err := tokens.CheckAccessForModel(r.Context(), telegramID, initDataRaw, modelID, category)
	if tokenguard.WriteHTTPError(w, r, err) {
		return err
	}
	return nil
}

func ensureImageGenerationAccess(w http.ResponseWriter, r *http.Request, tokens *tokenguard.Guard, req ai.ImageRequest) error {
	if tokens == nil {
		return nil
	}
	price := tokens.ResolveImageGenerationPrice(r.Context(), req.Model, billing.ImageGenerationParams{
		ModelID:     req.Model,
		Resolution:  req.Resolution,
		Quality:     req.Quality,
		Size:        req.Size,
		AspectRatio: req.AspectRatio,
		NumImages:   req.NumImages,
		WebSearch:   req.WebSearch,
		ImageSearch: req.ImageSearch,
	})
	err := tokens.CheckAccessForModelPrice(
		r.Context(),
		req.TelegramID,
		tokenguard.InitDataFromRequest(r, req.InitDataRaw),
		req.Model,
		"image",
		price,
	)
	if tokenguard.WriteHTTPError(w, r, err) {
		return err
	}
	return nil
}

func ensureVideoGenerationAccess(w http.ResponseWriter, r *http.Request, tokens *tokenguard.Guard, req ai.VideoRequest) error {
	if tokens == nil {
		return nil
	}
	params := videoBillingParams(req, req.Model)
	price := tokens.ResolveVideoGenerationPrice(r.Context(), req.Model, params)
	err := tokens.CheckAccessForModelPrice(
		r.Context(),
		req.TelegramID,
		tokenguard.InitDataFromRequest(r, req.InitDataRaw),
		req.Model,
		"video",
		price,
	)
	if tokenguard.WriteHTTPError(w, r, err) {
		return err
	}
	return nil
}

func videoBillingParams(req ai.VideoRequest, modelID string) billing.VideoGenerationParams {
	modelID = strings.TrimSpace(modelID)
	if modelID == "" {
		modelID = strings.TrimSpace(req.Model)
	}
	p := billing.VideoGenerationParams{
		ModelID:    modelID,
		Duration:   req.Duration,
		ExtendBy:   req.ExtendBy,
		Resolution: req.Resolution,
	}
	if req.Sound != nil {
		p.Sound = *req.Sound
	}
	if req.GenerateAudio != nil {
		p.GenerateAudio = *req.GenerateAudio
	} else if strings.HasPrefix(modelID, "seedance-v1.5") {
		p.GenerateAudio = true
	}
	if req.TurboMode != nil {
		p.TurboMode = *req.TurboMode
	} else if modelID == "seedance-v2-video-edit" {
		switch strings.ToLower(strings.TrimSpace(req.Resolution)) {
		case "720p", "1080p":
			p.TurboMode = true
		}
	}
	return p
}

func ensureAudioGenerationAccess(w http.ResponseWriter, r *http.Request, tokens *tokenguard.Guard, req ai.AudioRequest) error {
	if tokens == nil {
		return nil
	}
	params := audioBillingParams(req, req.Model)
	price := tokens.ResolveAudioGenerationPrice(r.Context(), req.Model, params)
	err := tokens.CheckAccessForModelPrice(
		r.Context(),
		req.TelegramID,
		tokenguard.InitDataFromRequest(r, req.InitDataRaw),
		req.Model,
		"audio",
		price,
	)
	if tokenguard.WriteHTTPError(w, r, err) {
		return err
	}
	return nil
}

func audioBillingParams(req ai.AudioRequest, modelID string) billing.AudioGenerationParams {
	modelID = strings.TrimSpace(modelID)
	if modelID == "" {
		modelID = strings.TrimSpace(req.Model)
	}
	voiceClone := isQwen3TTSVoiceClone(req, modelID)
	billingModel := modelID
	if modelID == "qwen3-tts-clone" {
		billingModel = "qwen3-tts"
		voiceClone = true
	}
	songs := req.NumberOfSongs
	if songs < 1 {
		songs = 1
	}
	return billing.AudioGenerationParams{
		ModelID:       billingModel,
		TextLength:    audioPromptLength(req),
		VoiceClone:    voiceClone,
		Duration:      req.Duration,
		NumberOfSongs: songs,
	}
}

func audioPromptLength(req ai.AudioRequest) int {
	text := strings.TrimSpace(req.Prompt)
	if text == "" {
		text = strings.TrimSpace(req.Text)
	}
	return utf8.RuneCountInString(text)
}

func isQwen3TTSVoiceClone(req ai.AudioRequest, modelID string) bool {
	if strings.TrimSpace(modelID) != "qwen3-tts" {
		return false
	}
	if strings.TrimSpace(req.AudioBase64) != "" {
		return true
	}
	if strings.TrimSpace(req.SourceAudioURL) != "" || strings.TrimSpace(req.AudioURL) != "" {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(req.Mode), "clone")
}

func ensureThreeDGenerationAccess(w http.ResponseWriter, r *http.Request, tokens *tokenguard.Guard, req ai.ThreeDRequest) error {
	if tokens == nil {
		return nil
	}
	params := threeDBillingParams(req, req.Model)
	price := tokens.ResolveThreeDGenerationPrice(r.Context(), req.Model, params)
	err := tokens.CheckAccessForModelPrice(
		r.Context(),
		req.TelegramID,
		tokenguard.InitDataFromRequest(r, req.InitDataRaw),
		req.Model,
		"3d",
		price,
	)
	if tokenguard.WriteHTTPError(w, r, err) {
		return err
	}
	return nil
}

func threeDBillingParams(req ai.ThreeDRequest, modelID string) billing.ThreeDGenerationParams {
	modelID = strings.TrimSpace(modelID)
	if modelID == "" {
		modelID = strings.TrimSpace(req.Model)
	}
	p := billing.ThreeDGenerationParams{
		ModelID:         modelID,
		TextureQuality:  req.TextureQuality,
		GeometryQuality: req.GeometryQuality,
		GenerateType:    req.GenerateType,
		Tier:            req.Tier,
		Addons:          req.Addons,
	}
	if req.Texture != nil {
		p.Texture = *req.Texture
		p.TextureSet = true
	}
	if req.Quad != nil {
		p.Quad = *req.Quad
	}
	if req.EnablePBR != nil {
		p.EnablePBR = *req.EnablePBR
		p.EnablePBRSet = true
	}
	return p
}

// chargeImageGeneration debits CyberCoins for a completed image generation.
// Callers MUST treat a non-nil error as fatal for the request (write an error
// response and discard the generated result) instead of returning the
// content for free — see chargeGeneration for the full rationale.
func chargeImageGeneration(ctx context.Context, tokens *tokenguard.Guard, req ai.ImageRequest, modelID string) error {
	if tokens == nil {
		return nil
	}
	model := strings.TrimSpace(modelID)
	if model == "" {
		model = strings.TrimSpace(req.Model)
	}
	price := tokens.ResolveImageGenerationPrice(ctx, model, billing.ImageGenerationParams{
		ModelID:     model,
		Resolution:  req.Resolution,
		Quality:     req.Quality,
		Size:        req.Size,
		AspectRatio: req.AspectRatio,
		NumImages:   req.NumImages,
		WebSearch:   req.WebSearch,
		ImageSearch: req.ImageSearch,
	})
	if err := tokens.ChargeForGenerationPrice(ctx, req.TelegramID, model, "image", price); err != nil {
		slog.WarnContext(ctx, "failed to charge generation; discarding result",
			slog.String("telegram_id", req.TelegramID),
			slog.String("model", model),
			slog.String("category", "image"),
			slog.Int("price", price),
			slog.Any("error", err),
		)
		return err
	}
	return nil
}

func chargeAudioGeneration(ctx context.Context, tokens *tokenguard.Guard, req ai.AudioRequest, modelID string) error {
	if tokens == nil {
		return nil
	}
	model := strings.TrimSpace(modelID)
	if model == "" {
		model = strings.TrimSpace(req.Model)
	}
	params := audioBillingParams(req, model)
	price := tokens.ResolveAudioGenerationPrice(ctx, model, params)
	if err := tokens.ChargeForGenerationPrice(ctx, req.TelegramID, model, "audio", price); err != nil {
		slog.WarnContext(ctx, "failed to charge generation; discarding result",
			slog.String("telegram_id", req.TelegramID),
			slog.String("model", model),
			slog.String("category", "audio"),
			slog.Int("price", price),
			slog.Any("error", err),
		)
		return err
	}
	return nil
}

func chargeVideoGeneration(ctx context.Context, tokens *tokenguard.Guard, req ai.VideoRequest, modelID string) error {
	if tokens == nil {
		return nil
	}
	model := strings.TrimSpace(modelID)
	if model == "" {
		model = strings.TrimSpace(req.Model)
	}
	params := videoBillingParams(req, model)
	price := tokens.ResolveVideoGenerationPrice(ctx, model, params)
	if err := tokens.ChargeForGenerationPrice(ctx, req.TelegramID, model, "video", price); err != nil {
		slog.WarnContext(ctx, "failed to charge generation; discarding result",
			slog.String("telegram_id", req.TelegramID),
			slog.String("model", model),
			slog.String("category", "video"),
			slog.Int("price", price),
			slog.Any("error", err),
		)
		return err
	}
	return nil
}

func chargeThreeDGeneration(ctx context.Context, tokens *tokenguard.Guard, req ai.ThreeDRequest, modelID string) error {
	if tokens == nil {
		return nil
	}
	model := strings.TrimSpace(modelID)
	if model == "" {
		model = strings.TrimSpace(req.Model)
	}
	params := threeDBillingParams(req, model)
	price := tokens.ResolveThreeDGenerationPrice(ctx, model, params)
	if err := tokens.ChargeForGenerationPrice(ctx, req.TelegramID, model, "3d", price); err != nil {
		slog.WarnContext(ctx, "failed to charge generation; discarding result",
			slog.String("telegram_id", req.TelegramID),
			slog.String("model", model),
			slog.String("category", "3d"),
			slog.Int("price", price),
			slog.Any("error", err),
		)
		return err
	}
	return nil
}

// chargeGeneration debits CyberCoins for a completed generation. Charging
// happens after the (already-paid-for-by-us) provider call succeeds, so a
// concurrent burst of requests can all pass the pre-generation balance check
// before any of them actually debits the wallet. The atomic, row-locked debit
// in debitProfileByTelegram guarantees only requests the user can actually
// afford succeed here — callers MUST propagate a non-nil error back as an
// HTTP error and MUST NOT hand the generated content to the client, or the
// remaining (unpaid) requests in the burst would receive paid output for
// free.
func chargeGeneration(ctx context.Context, tokens *tokenguard.Guard, telegramID, modelID, category string) error {
	if tokens == nil {
		return nil
	}
	model := strings.TrimSpace(modelID)
	if model == "" {
		model = strings.TrimSpace(category)
	}
	if err := tokens.ChargeForGeneration(ctx, telegramID, model, category); err != nil {
		slog.WarnContext(ctx, "failed to charge generation; discarding result",
			slog.String("telegram_id", telegramID),
			slog.String("model", model),
			slog.String("category", category),
			slog.Any("error", err),
		)
		return err
	}
	return nil
}

func writeServiceError(w http.ResponseWriter, r *http.Request, op string, err error) {
	switch {
	case errors.Is(err, ai.ErrPromptEmpty):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, ai.ErrNotConfigured):
		writeError(w, http.StatusServiceUnavailable, err.Error())
	default:
		var pe *ai.ProviderError
		if errors.As(err, &pe) {
			slog.ErrorContext(r.Context(), op+" provider error",
				slog.String("provider", pe.Provider),
				slog.Int("status", pe.Status),
				slog.String("message", pe.Message),
			)
			writeError(w, http.StatusBadGateway, pe.Error())
			return
		}
		slog.ErrorContext(r.Context(), op, slog.Any("error", err))
		writeError(w, http.StatusInternalServerError, "generation failed")
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
