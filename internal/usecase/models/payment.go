package models

import (
	"errors"
	"fmt"
	"strings"
)

const (
	PaymentKindCoinPack     = "coin_pack"
	PaymentKindSubscription = "subscription"
)

var (
	// ErrPaymentsDisabled is returned when YooKassa credentials are missing or
	// the admin has switched payments off in the admin panel.
	ErrPaymentsDisabled = errors.New("payments are not available right now")
	// ErrItemNotFound is returned when the requested plan_id/pack_id does not
	// exist (or is disabled) in the current billing catalog.
	ErrItemNotFound = errors.New("item not found")
	// ErrPaymentNotFound is returned when a webhook references an unknown payment.
	ErrPaymentNotFound = errors.New("payment not found")
)

// StartCheckoutInput starts a YooKassa redirect checkout for one catalog item.
type StartCheckoutInput struct {
	TelegramID string
	Kind       string // "coin_pack" | "subscription"
	ItemID     string // pack_id or plan_id
}

func (i *StartCheckoutInput) Validate() error {
	if strings.TrimSpace(i.TelegramID) == "" {
		return fmt.Errorf("%w: telegram_id is required", ErrInvalidInput)
	}
	kind := strings.TrimSpace(i.Kind)
	if kind != PaymentKindCoinPack && kind != PaymentKindSubscription {
		return fmt.Errorf("%w: kind must be %q or %q", ErrInvalidInput, PaymentKindCoinPack, PaymentKindSubscription)
	}
	if strings.TrimSpace(i.ItemID) == "" {
		return fmt.Errorf("%w: item_id is required", ErrInvalidInput)
	}
	return nil
}

// StartCheckoutOutput is returned to the client to open the YooKassa payment page.
type StartCheckoutOutput struct {
	PaymentID       int64   `json:"payment_id"`
	ConfirmationURL string  `json:"confirmation_url"`
	AmountRub       float64 `json:"amount_rub"`
	Coins           int64   `json:"coins"`
}
