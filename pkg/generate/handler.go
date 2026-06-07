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
			handleImage(w, r, svc)
			return
		case r.Method == http.MethodPost && r.URL.Path == pathGenerateVideo:
			handleVideo(w, r, svc)
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

func handleImage(w http.ResponseWriter, r *http.Request, svc *ai.Service) {
	var req ai.ImageRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	out, err := svc.GenerateImage(r.Context(), req)
	if err != nil {
		writeServiceError(w, r, "generate image", err)
		return
	}

	writeJSON(w, http.StatusOK, ai.FinalizeImageResponse(out))
}

func handleVideo(w http.ResponseWriter, r *http.Request, svc *ai.Service) {
	var req ai.VideoRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	out, err := svc.GenerateVideo(r.Context(), req)
	if err != nil {
		writeServiceError(w, r, "generate video", err)
		return
	}

	writeJSON(w, http.StatusOK, out)
}

func decodeJSON(r *http.Request, dst any) error {
	if r.Body == nil {
		return errors.New("request body is required")
	}
	defer r.Body.Close()
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	if err := dec.Decode(dst); err != nil {
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
