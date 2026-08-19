package models

import "time"

type AdminEvent struct {
	ID      string
	Time    time.Time
	User    string
	Action  string
	Details string
}

type AdminTokenTransaction struct {
	ID          int64
	UserName    string
	Operation   string
	Amount      int64
	Reason      string
	CreatedAt   time.Time
	Source      string
	Status      string
	PaymentKind string
	AmountRub   float64
}

type AdminTransactionStats struct {
	CreditsMonth    int64
	DebitsMonth     int64
	OperationsMonth int64
	AvgAmount       int64
}

type AdminBroadcastRecord struct {
	ID          int64
	Message     string
	Target      string
	SentCount   int64
	FailedCount int64
	CreatedAt   time.Time
}

type ModelConfig struct {
	ModelID    string
	Category   string
	Name       string
	Provider   string
	PriceCoins int64
	Enabled    bool
}
