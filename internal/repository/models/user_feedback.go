package models

import "time"

const (
	UserFeedbackKindSuggestion = "suggestion"
	UserFeedbackKindBug        = "bug"
)

type UserFeedback struct {
	ID        int64
	ProfileID int64
	Kind      string
	Message   string
	CreatedAt time.Time
	UserName  string
}
