package models

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrAdminNotFound     = errors.New("ErrAdminNotFound")
	ErrAdminUserNotFound = errors.New("ErrAdminUserNotFound")
	ErrBroadcastNotReady = errors.New("ErrBroadcastNotReady")
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
	ID         int64
	TelegramID int64
	Username   string
	FirstName  string
	LastName   string
	IsActive   bool
	CreatedAt  string
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
