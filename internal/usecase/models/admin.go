package models

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrAdminNotFound        = errors.New("ErrAdminNotFound")
	ErrAdminUserNotFound    = errors.New("ErrAdminUserNotFound")
	ErrBroadcastNotReady    = errors.New("ErrBroadcastNotReady")
	ErrInsufficientTokens   = errors.New("ErrInsufficientTokens")
)

type AdminLoginInput struct {
	Email    string
	Password string
}

func (i *AdminLoginInput) Validate() error {
	if strings.TrimSpace(i.Email) == "" {
		return fmt.Errorf("%w: email is required", ErrInvalidInput)
	}
	if i.Password == "" {
		return fmt.Errorf("%w: password is required", ErrInvalidInput)
	}
	return nil
}

type AdminLoginOutput struct {
	Token string
	Admin AdminUser
}

type AdminUser struct {
	ID    int64
	Email string
}

type AdminStatsOutput struct {
	TotalUsers       int64
	ActiveUsersToday int64
	NewUsersToday    int64
	TotalMessages    int64
}

type AdminListUsersInput struct {
	Page    int32
	PerPage int32
	Search  string
}

func (i *AdminListUsersInput) Normalize() {
	if i.Page < 1 {
		i.Page = 1
	}
	if i.PerPage < 1 {
		i.PerPage = 20
	}
	if i.PerPage > 100 {
		i.PerPage = 100
	}
}

type AdminUserItem struct {
	ID                  int64
	TelegramID          int64
	Username            string
	FirstName           string
	LastName            string
	IsActive            bool
	Tokens              int64
	CreatedAt           string
	SubscriptionPlanID  string
	SubscriptionPlan    string // localized/display name
	SubscriptionExpires string // RFC3339, empty when free/no expiry
	SubscriptionDaysLeft int
}

type AdminListUsersOutput struct {
	Data  []AdminUserItem
	Total int64
}

type AdminUpdateUserInput struct {
	UserID   int64
	IsActive bool
}

type AdminBroadcastInput struct {
	AdminID   int64
	Message   string
	Target    string // all | active
	ParseMode string
}

func (i *AdminBroadcastInput) Validate() error {
	if strings.TrimSpace(i.Message) == "" {
		return fmt.Errorf("%w: message is required", ErrInvalidInput)
	}
	if i.Target != "all" && i.Target != "active" {
		return fmt.Errorf("%w: target must be all or active", ErrInvalidInput)
	}
	return nil
}

type AdminBroadcastOutput struct {
	Sent   int64
	Failed int64
}

type AdminTokenChangeInput struct {
	UserID  int64
	AdminID int64
	Amount  int64
	Reason  string
}

func (i *AdminTokenChangeInput) Validate() error {
	if i.UserID <= 0 {
		return fmt.Errorf("%w: user_id invalid", ErrInvalidInput)
	}
	if i.AdminID <= 0 {
		return fmt.Errorf("%w: admin_id invalid", ErrInvalidInput)
	}
	if i.Amount <= 0 {
		return fmt.Errorf("%w: amount must be greater than 0", ErrInvalidInput)
	}
	if len(strings.TrimSpace(i.Reason)) > 255 {
		return fmt.Errorf("%w: reason is too long", ErrInvalidInput)
	}
	return nil
}

type AdminTokenChangeOutput struct {
	UserID    int64
	Tokens    int64
	Delta     int64
	Operation string
}
