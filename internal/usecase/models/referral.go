package models

import "fmt"

type ReferralItem struct {
	ProfileID           int64
	TelegramID          string
	Name                string
	Username            string
	Avatar              string
	CompletedTasksCount int64
	Earnings            int64
}

type ListReferralsOutput struct {
	Referrals      []ReferralItem
	TotalCount     int64
	TotalEarnings  int64
	CompletedTasks int64
}

func (i *ListReferralsOutput) ValidateTelegramID(telegramID string) error {
	if telegramID == "" {
		return fmt.Errorf("%w: telegram_id is required", ErrInvalidInput)
	}
	if len(telegramID) > 100 {
		return fmt.Errorf("%w: telegram_id too long", ErrInvalidInput)
	}
	return nil
}
