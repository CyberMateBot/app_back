package models

import (
	"errors"
	"time"
)

var ErrInsufficientTokens = errors.New("insufficient tokens")
type Admin struct {
	ID           int64
	Email        string
	PasswordHash string
	Role         string
	CreatedAt    time.Time
}

type AdminProfile struct {
	ID         int64
	TelegramID string
	Name       string
	Username   string
	IsActive   bool
	Tokens     int64
	CreatedAt  time.Time
}

type TokenOperationResult struct {
	ProfileID    int64
	BalanceAfter int64
}

type AdminStats struct {
	TotalUsers       int64
	ActiveUsersToday int64
	NewUsersToday    int64
	TotalMessages    int64
}
