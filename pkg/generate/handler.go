package generate

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/twelvepills-936/tgapp-/pkg/ai"
	"github.com/twelvepills-936/tgapp-/pkg/prompthistory"
)

const (
	pathGenerateModels = "/v1/generate/models"
	pathGenerateText   = "/v1/generate/text"
	pathGenerateImage  = "/v1/generate/image"
	pathGenerateVideo  = "/v1/generate/video"
	pathGenerateAudio  = "/v1/generate/audio"
)

// Wrap adds POST /v1/generate/text and POST /v1/generate/image.
func Wrap(next http.Handler, svc *ai.Service, history *prompthistory.Store) http.Handler {
	if svc == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == pathGenerateModels:
			writeJSON(w, http.StatusOK, svc.ListModels())
			return
		case r.Method == http.MethodPost && r.URL.Path == pathGenerateText:
			handleText(w, r, svc, history)
			return
		case r.Method == http.MethodPost && r.URL.Path == pathGenerateImage:
			handleImage(w, r, svc, history)
			return
		case r.Method == http.MethodPost && r.URL.Path == pathGenerateVideo:
			handleVideo(w, r, svc, history)
			return
		case r.Method == http.MethodPost && r.URL.Path == pathGenerateAudio:
			handleAudio(w, r, svc)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type textGenerateRequest struct {
	ai.TextRequest
	TelegramID string `json:"telegramId"`
	SessionID  string `json:"sessionId"`
	Category   string `json:"category"`
}

type textGenerateResponse struct {
	Text   string              `json:"text"`
	Model  string              `json:"model"`
	Format string              `json:"format,omitempty"`
	Item   *prompthistory.Item `json:"item,omitempty"`
}

func handleText(w http.ResponseWriter, r *http.Request, svc *ai.Service, history *prompthistory.Store) {
	var req textGenerateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	out, err := svc.GenerateText(r.Context(), req.TextRequest)
	if err != nil {
		writeServiceError(w, r, "generate text", err)
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

	writeJSON(w, http.StatusOK, resp)
}

type imageGenerateResponse struct {
	ImageURL    string              `json:"image_url,omitempty"`
	ImageURLs   []string            `json:"image_urls,omitempty"`
	ImageBase64 string              `json:"image_base64,omitempty"`
	Model       string              `json:"model"`
	Item        *prompthistory.Item `json:"item,omitempty"`
}

func handleImage(w http.ResponseWriter, r *http.Request, svc *ai.Service, history *prompthistory.Store) {
	var req ai.ImageRequest
	if err := decodeImageJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, formatDecodeError(err))
		return
	}

	out, err := svc.GenerateImage(r.Context(), req)
	if err != nil {
		writeServiceError(w, r, "generate image", err)
		return
	}

	finalized := ai.FinalizeImageResponse(out)
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

func handleVideo(w http.ResponseWriter, r *http.Request, svc *ai.Service, history *prompthistory.Store) {
	var req ai.VideoRequest
	if err := decodeVideoJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, formatDecodeError(err))
		return
	}

	out, err := svc.GenerateVideo(r.Context(), req)
	if err != nil {
		writeServiceError(w, r, "generate video", err)
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

func handleAudio(w http.ResponseWriter, r *http.Request, svc *ai.Service) {
	var req ai.AudioRequest
	if err := decodeAudioJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, formatDecodeError(err))
		return
	}

	out, err := svc.GenerateAudio(r.Context(), req)
	if err != nil {
		writeServiceError(w, r, "generate audio", err)
		return
	}

	writeJSON(w, http.StatusOK, out)
}

const (
	maxJSONBodyDefault = 1 << 20       // 1 MiB — text / small payloads
	maxJSONBodyImage   = 12 << 20      // 12 MiB — imageBase64 (до ~8 MiB binary)
	maxJSONBodyVideo   = 2 << 20       // 2 MiB — video metadata
	maxJSONBodyAudio   = 12 << 20      // 12 MiB — audioBase64 for voice clone
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
