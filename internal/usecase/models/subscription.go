package models

import (
	"fmt"
	"time"
)

// ExpiringSoonDays controls when a subscription is flagged as "expiring soon".
const ExpiringSoonDays = 3

// SubscriptionStateOutput is the resolved subscription view for a user.
// Expiry is already applied: an expired paid plan resolves to "free".
type SubscriptionStateOutput struct {
	PlanID       string `json:"plan_id"`
	PlanName     string `json:"plan_name"`
	PlanRank     int    `json:"plan_rank"`
	IsPaid       bool   `json:"is_paid"`
	Coins        int64  `json:"coins"`
	StartedAt    string `json:"started_at,omitempty"`
	ExpiresAt    string `json:"expires_at,omitempty"`
	DaysLeft     int    `json:"days_left"`
	HoursLeft    int    `json:"hours_left"`
	IsActive     bool   `json:"is_active"`
	ExpiringSoon bool   `json:"expiring_soon"`
	Expired      bool   `json:"expired"`
}

// AdminSetSubscriptionInput grants/edits a user's subscription.
type AdminSetSubscriptionInput struct {
	UserID       int64
	AdminID      int64
	PlanID       string
	DurationDays int        // when > 0: expires_at = now + DurationDays
	ExpiresAt    *time.Time // explicit expiry (overrides DurationDays)
	NoExpiry     bool       // explicit lifetime (no expiry)
	GrantCoins   bool       // credit the plan's monthly coins
}

func (i *AdminSetSubscriptionInput) Validate() error {
	if i.UserID <= 0 {
		return fmt.Errorf("%w: user_id invalid", ErrInvalidInput)
	}
	if i.PlanID == "" {
		return fmt.Errorf("%w: plan_id is required", ErrInvalidInput)
	}
	if i.DurationDays < 0 || i.DurationDays > 3650 {
		return fmt.Errorf("%w: duration_days out of range", ErrInvalidInput)
	}
	return nil
}

// AdminSubscriptionOutput is returned after admin grant/clear operations.
type AdminSubscriptionOutput struct {
	User         AdminUserItem
	Subscription SubscriptionStateOutput
	CoinsGranted int64
}
