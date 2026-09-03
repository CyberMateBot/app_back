package models

import (
	"encoding/base64"
	"errors"
	"fmt"
)

// Error text below is intentionally a short, generic, lower-case phrase (not
// e.g. "ErrProfileNotFound") — these strings can end up verbatim in HTTP
// error responses via the gRPC-gateway, and Go-identifier-shaped messages
// were leaking internal type names to API callers.
var (
	ErrProfileNotFound          = errors.New("profile not found")
	ErrProfileAlreadyRegistered = errors.New("profile already registered")
	ErrInvalidInput             = errors.New("invalid input")
	// ErrInvalidTelegramSignature is returned when init_data_raw fails Telegram's
	// HMAC signature check, i.e. it was not actually issued by Telegram for this
	// bot. Without this check anyone could register arbitrary telegram ids.
	ErrInvalidTelegramSignature = errors.New("invalid telegram signature")
)

type RegisterByTelegramInput struct {
	InitDataRaw string
	StartParam  string
}

// Validate checks the input data
func (i *RegisterByTelegramInput) Validate() error {
	if i.InitDataRaw == "" {
		return fmt.Errorf("%w: init_data_raw is required", ErrInvalidInput)
	}

	if len(i.InitDataRaw) > 10000 {
		return fmt.Errorf("%w: init_data_raw too long", ErrInvalidInput)
	}

	// Проверка на корректность base64
	if _, err := base64.StdEncoding.DecodeString(i.InitDataRaw); err != nil {
		return fmt.Errorf("%w: init_data_raw is not valid base64", ErrInvalidInput)
	}

	return nil
}

type RegisterByTelegramOutput struct {
    ProfileID int64
}

type GetProfileOutput struct {
    Data ProfileUser
}

type ProfileUser struct {
    ID         int64
    Name       string
    TelegramID string
    Avatar     string
    Username   string
    Verified   bool
    Theme      string
}

const (
    ThemeLight = "light"
    ThemeDark  = "dark"
)

type UpdateProfileThemeInput struct {
    TelegramID string
    Theme      string
}

func (i *UpdateProfileThemeInput) Validate() error {
    if i.TelegramID == "" {
        return fmt.Errorf("%w: telegram_id is required", ErrInvalidInput)
    }
    if len(i.TelegramID) > 100 {
        return fmt.Errorf("%w: telegram_id too long", ErrInvalidInput)
    }
    if i.Theme != ThemeLight && i.Theme != ThemeDark {
        return fmt.Errorf("%w: theme must be light or dark", ErrInvalidInput)
    }
    return nil
}

type UpdateProfileThemeOutput struct {
    Theme string
}


