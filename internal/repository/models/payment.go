package models

import "time"

// Payment statuses mirror the CHECK constraint on the payments table.
const (
	PaymentStatusPending   = "pending"
	PaymentStatusSucceeded = "succeeded"
	PaymentStatusCanceled  = "canceled"
	PaymentStatusRefunded  = "refunded"
)

// Payment kinds mirror the CHECK constraint on the payments table.
const (
	PaymentKindCoinPack     = "coin_pack"
	PaymentKindSubscription = "subscription"
)

// Payment is a single YooKassa checkout attempt tied to a profile.
type Payment struct {
	ID                int64
	ProfileID         int64
	Provider          string
	ProviderPaymentID string
	IdempotenceKey    string
	Kind              string
	ItemID            string
	AmountRub         float64
	Coins             int64
	Status            string
	ConfirmationURL   string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
