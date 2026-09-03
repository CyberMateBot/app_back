package models

import (
	"errors"
	"fmt"
	"strings"
)

const (
	UserFeedbackKindSuggestion = "suggestion"
	UserFeedbackKindBug        = "bug"
)

var ErrFeedbackMessageTooShort = errors.New("feedback message is too short")

type SubmitUserFeedbackInput struct {
	TelegramID string
	Kind       string
	Message    string
}

func (i *SubmitUserFeedbackInput) Validate() error {
	i.TelegramID = strings.TrimSpace(i.TelegramID)
	i.Kind = strings.ToLower(strings.TrimSpace(i.Kind))
	i.Message = strings.TrimSpace(i.Message)

	if i.TelegramID == "" {
		return fmt.Errorf("%w: telegram_id is required", ErrInvalidInput)
	}
	if i.Kind != UserFeedbackKindSuggestion && i.Kind != UserFeedbackKindBug {
		return fmt.Errorf("%w: invalid feedback kind", ErrInvalidInput)
	}
	if len([]rune(i.Message)) < 3 {
		return ErrFeedbackMessageTooShort
	}
	if len([]rune(i.Message)) > 2000 {
		return fmt.Errorf("%w: message is too long", ErrInvalidInput)
	}
	return nil
}

type SubmitUserFeedbackOutput struct {
	ID int64 `json:"id"`
}

type AdminListUserFeedbackInput struct {
	Page    int32
	PerPage int32
	Kind    string
}

func (i *AdminListUserFeedbackInput) Normalize() {
	if i.Page < 1 {
		i.Page = 1
	}
	if i.PerPage < 1 {
		i.PerPage = 20
	}
	if i.PerPage > 100 {
		i.PerPage = 100
	}
	i.Kind = strings.ToLower(strings.TrimSpace(i.Kind))
}

type AdminUserFeedbackItem struct {
	ID          int64  `json:"id"`
	User        string `json:"user"`
	Kind        string `json:"kind"`
	KindLabel   string `json:"kind_label"`
	Message     string `json:"message"`
	CreatedAt   string `json:"created_at"`
}

type AdminListUserFeedbackOutput struct {
	Data  []AdminUserFeedbackItem
	Total int64
}
