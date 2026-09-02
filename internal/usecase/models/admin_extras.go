package models

import "strings"

type AdminEventItem struct {
	ID      string
	Time    string
	User    string
	Action  string
	Details string
}

type AdminListEventsOutput struct {
	Data []AdminEventItem
}

type AdminListTransactionsInput struct {
	Page      int32
	PerPage   int32
	Operation string
}

func (i *AdminListTransactionsInput) Normalize() {
	if i.Page < 1 {
		i.Page = 1
	}
	if i.PerPage < 1 {
		i.PerPage = 20
	}
	if i.PerPage > 100 {
		i.PerPage = 100
	}
	i.Operation = strings.ToLower(strings.TrimSpace(i.Operation))
}

type AdminTransactionItem struct {
	ID          int64
	User        string
	Type        string
	TypeLabel   string
	Amount      int64
	AmountLabel string
	Method      string
	MethodLabel string
	CreatedAt   string
	Status      string
	StatusLabel string
}

type AdminTransactionStatsOutput struct {
	CreditsMonth    int64
	DebitsMonth     int64
	OperationsMonth int64
	AvgAmount       int64
}

type AdminListTransactionsOutput struct {
	Stats AdminTransactionStatsOutput
	Data  []AdminTransactionItem
	Total int64
}

type AdminListBroadcastsInput struct {
	Page    int32
	PerPage int32
}

func (i *AdminListBroadcastsInput) Normalize() {
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

type AdminBroadcastItem struct {
	ID          int64
	Message     string
	Target      string
	TargetLabel string
	Sent        int64
	Failed      int64
	CreatedAt   string
	Status      string
	StatusLabel string
}

type AdminListBroadcastsOutput struct {
	Data  []AdminBroadcastItem
	Total int64
}

type AdminSettingsOutput struct {
	RegistrationBonus     int64   `json:"registration_bonus"`
	ReferralBonus         int64   `json:"referral_bonus"`
	ReferralRefereeBonus  int64   `json:"referral_referee_bonus"`
	TokenExpiryDays       int64   `json:"token_expiry_days"`
	MaintenanceMode       bool    `json:"maintenance_mode"`
	YookassaEnabled       bool    `json:"yookassa_enabled"`
	TelegramStarsEnabled  bool    `json:"telegram_stars_enabled"`
	CoinRateRub           float64 `json:"coin_rate_rub"`
}

type AdminUpdateSettingsInput struct {
	RegistrationBonus    *int64   `json:"registration_bonus"`
	ReferralBonus        *int64   `json:"referral_bonus"`
	ReferralRefereeBonus *int64   `json:"referral_referee_bonus"`
	TokenExpiryDays      *int64   `json:"token_expiry_days"`
	MaintenanceMode      *bool    `json:"maintenance_mode"`
	YookassaEnabled      *bool    `json:"yookassa_enabled"`
	TelegramStarsEnabled *bool    `json:"telegram_stars_enabled"`
	CoinRateRub          *float64 `json:"coin_rate_rub"`
}

type AdminModelItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Provider string `json:"provider"`
	Category string `json:"category"`
	Price    int64  `json:"price"`
	Enabled  bool   `json:"enabled"`
}

type AdminListModelsOutput struct {
	Data []AdminModelItem
}

type AdminUpdateModelInput struct {
	ModelID string
	Price   *int64
	Enabled *bool
}
