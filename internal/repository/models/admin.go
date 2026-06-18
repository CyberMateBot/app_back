package models

import "time"

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
	CreatedAt  time.Time
}

type AdminStats struct {
	TotalUsers       int64
	ActiveUsersToday int64
	NewUsersToday    int64
	TotalMessages    int64
}
