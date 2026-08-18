// Package paymentsapi exposes the HTTP surface for YooKassa payments:
// starting a checkout from the Telegram Mini App and receiving YooKassa's
// webhook notifications.
package paymentsapi

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/twelvepills-936/tgapp-/internal"
	ucModels "github.com/twelvepills-936/tgapp-/internal/usecase/models"
	"github.com/twelvepills-936/tgapp-/pkg/tokenguard"
)

const (
	pathCheckout        = "/v1/billing/checkout"
	pathYooKassaWebhook = "/v1/payments/yookassa/webhook"

	maxCheckoutBodyBytes = 8 << 10  // 8 KiB
	maxWebhookBodyBytes  = 64 << 10 // 64 KiB, YooKassa payloads are small
)

// Wrap adds POST /v1/billing/checkout (Telegram-authenticated) and
// POST /v1/payments/yookassa/webhook (public, called by YooKassa).
func Wrap(next http.Handler, uc internal.UseCase, tokens *tokenguard.Guard) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == pathCheckout:
			handleCheckout(w, r, uc, tokens)
			return
		case r.Method == http.MethodPost && r.URL.Path == pathYooKassaWebhook:
			handleWebhook(w, r, uc)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type checkoutRequest struct {
	TelegramID  string `json:"telegramId"`
	InitDataRaw string `json:"initDataRaw"`
	Kind        string `json:"kind"`
	ItemID      string `json:"itemId"`
}

func handleCheckout(w http.ResponseWriter, r *http.Request, uc internal.UseCase, tokens *tokenguard.Guard) {
	var req checkoutRequest
	if err := decodeJSON(r, &req, maxCheckoutBodyBytes); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	identityErr := tokens.CheckIdentity(r.Context(), req.TelegramID, tokenguard.InitDataFromRequest(r, req.InitDataRaw))
	if tokenguard.WriteHTTPError(w, r, identityErr) {
		return
	}

	out, err := uc.StartCheckout(r.Context(), ucModels.StartCheckoutInput{
		TelegramID: req.TelegramID,
		Kind:       req.Kind,
		ItemID:     req.ItemID,
	})
	if err != nil {
		writeCheckoutError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, out)
}

func writeCheckoutError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ucModels.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, ucModels.ErrPaymentsDisabled):
		writeError(w, http.StatusServiceUnavailable, "payments are not available right now")
	case errors.Is(err, ucModels.ErrItemNotFound), errors.Is(err, ucModels.ErrProfileNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	default:
		slog.ErrorContext(r.Context(), "checkout failed", slog.Any("error", err))
		writeError(w, http.StatusInternalServerError, "failed to start checkout")
	}
}

type webhookNotification struct {
	Type   string `json:"type"`
	Event  string `json:"event"`
	Object struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	} `json:"object"`
}

// handleWebhook always re-verifies the payment via the YooKassa API before
// fulfilling it (see internal/usecase/payments.go), so the notification body
// itself only needs to tell us *which* payment to look up.
func handleWebhook(w http.ResponseWriter, r *http.Request, uc internal.UseCase) {
	var payload webhookNotification
	if err := decodeJSON(r, &payload, maxWebhookBodyBytes); err != nil {
		// Malformed body: nothing we can retry on, acknowledge so YooKassa stops resending it.
		w.WriteHeader(http.StatusOK)
		return
	}

	if payload.Object.ID == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if err := uc.HandleYooKassaWebhookNotification(r.Context(), payload.Object.ID); err != nil {
		slog.ErrorContext(r.Context(), "yookassa webhook processing failed",
			slog.Any("error", err), slog.String("payment_id", payload.Object.ID))
		// Non-2xx makes YooKassa retry the notification later.
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func decodeJSON(r *http.Request, dst any, maxBytes int64) error {
	if r.Body == nil {
		return errors.New("request body is required")
	}
	defer r.Body.Close()
	dec := json.NewDecoder(io.LimitReader(r.Body, maxBytes))
	return dec.Decode(dst)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
