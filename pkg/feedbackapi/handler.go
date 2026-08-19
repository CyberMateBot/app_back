package feedbackapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/twelvepills-936/tgapp-/internal"
	ucModels "github.com/twelvepills-936/tgapp-/internal/usecase/models"
	"github.com/twelvepills-936/tgapp-/pkg/tokenguard"
)

const (
	pathFeedback     = "/v1/feedback"
	maxFeedbackBytes = 4 << 10
)

func Wrap(next http.Handler, uc internal.UseCase, tokens *tokenguard.Guard) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == pathFeedback {
			handleSubmit(w, r, uc, tokens)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type feedbackRequest struct {
	TelegramID  string `json:"telegramId"`
	InitDataRaw string `json:"initDataRaw"`
	Kind        string `json:"kind"`
	Message     string `json:"message"`
}

type feedbackResponse struct {
	ID int64 `json:"id"`
}

func handleSubmit(w http.ResponseWriter, r *http.Request, uc internal.UseCase, tokens *tokenguard.Guard) {
	var req feedbackRequest
	if err := decodeJSON(r, &req, maxFeedbackBytes); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if tokens != nil {
		identityErr := tokens.CheckIdentity(r.Context(), req.TelegramID, tokenguard.InitDataFromRequest(r, req.InitDataRaw))
		if tokenguard.WriteHTTPError(w, r, identityErr) {
			return
		}
	}

	out, err := uc.SubmitUserFeedback(r.Context(), ucModels.SubmitUserFeedbackInput{
		TelegramID: req.TelegramID,
		Kind:       req.Kind,
		Message:    req.Message,
	})
	if err != nil {
		switch {
		case errors.Is(err, ucModels.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, ucModels.ErrFeedbackMessageTooShort):
			writeError(w, http.StatusBadRequest, "message is too short")
		default:
			writeError(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	writeJSON(w, http.StatusOK, feedbackResponse{ID: out.ID})
}

func decodeJSON(r *http.Request, dst any, maxBytes int64) error {
	body, err := io.ReadAll(io.LimitReader(r.Body, maxBytes+1))
	if err != nil {
		return err
	}
	if int64(len(body)) > maxBytes {
		return errors.New("request body too large")
	}
	if len(body) == 0 {
		return errors.New("empty request body")
	}
	return json.Unmarshal(body, dst)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
