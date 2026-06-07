package models

type Referral struct {
	ID                  int64
	RefereeProfileID    int64
	RefereeTelegramID   string
	RefereeName         string
	RefereeUsername     string
	RefereeAvatar       string
	CompletedTasksCount int64
	Earnings            int64
}
