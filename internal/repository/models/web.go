package models

import "time"

type WebAccount struct {
	ID           int64
	Email        string
	PasswordHash string
	DisplayName  string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type WebPrompt struct {
	ID           int64
	WebAccountID int64
	Prompt       string
	Category     string
	Model        string
	CreatedAt    time.Time
}

