package internal

import (
	"context"

	"github.com/jackc/pgx/v5"
	rcModels "github.com/twelvepills-936/tgapp-/internal/rest/client/models"
	repoModels "github.com/twelvepills-936/tgapp-/internal/repository/models"
	ucModels "github.com/twelvepills-936/tgapp-/internal/usecase/models"
)

type Repository interface {
	DBBeginTransaction(ctx context.Context) (pgx.Tx, error)

	ReadUser(ctx context.Context, id int64, dbTx pgx.Tx) (user repoModels.User, err error)

	// CyberMate repositories
	CreateProfile(ctx context.Context, tx pgx.Tx, p repoModels.Profile) (int64, error)
	GetProfileByTelegramID(ctx context.Context, tx pgx.Tx, telegramID string) (repoModels.Profile, error)
	UpdateProfileTheme(ctx context.Context, tx pgx.Tx, telegramID, theme string) error
	CreateWalletForUser(ctx context.Context, tx pgx.Tx, profileID int64) (int64, error)
	AddReferral(ctx context.Context, tx pgx.Tx, referrerProfileID int64, refereeProfileID int64) error
	ListReferralsByReferrerProfileID(ctx context.Context, tx pgx.Tx, referrerProfileID int64) ([]repoModels.Referral, error)

	// Web site repositories
	CreateWebAccount(ctx context.Context, tx pgx.Tx, a repoModels.WebAccount) (int64, error)
	GetWebAccountByEmail(ctx context.Context, tx pgx.Tx, email string) (repoModels.WebAccount, error)
	GetWebAccountByID(ctx context.Context, tx pgx.Tx, id int64) (repoModels.WebAccount, error)
	CreateWebPrompt(ctx context.Context, tx pgx.Tx, p repoModels.WebPrompt) (int64, error)
	ListWebPrompts(ctx context.Context, tx pgx.Tx, webAccountID int64, limit int32, offset int32) ([]repoModels.WebPrompt, error)

	// Admin panel repositories
	GetAdminByEmail(ctx context.Context, tx pgx.Tx, email string) (repoModels.Admin, error)
	GetAdminByID(ctx context.Context, tx pgx.Tx, id int64) (repoModels.Admin, error)
	CreateAdmin(ctx context.Context, tx pgx.Tx, email, passwordHash string) (int64, error)
	CountAdmins(ctx context.Context, tx pgx.Tx) (int64, error)
	GetAdminStats(ctx context.Context, tx pgx.Tx) (repoModels.AdminStats, error)
	ListAdminProfiles(ctx context.Context, tx pgx.Tx, search string, limit, offset int32) ([]repoModels.AdminProfile, int64, error)
	GetAdminProfileByID(ctx context.Context, tx pgx.Tx, id int64) (repoModels.AdminProfile, error)
	UpdateProfileActive(ctx context.Context, tx pgx.Tx, id int64, isActive bool) error
	DeleteProfile(ctx context.Context, tx pgx.Tx, id int64) error
	ListBroadcastTelegramIDs(ctx context.Context, tx pgx.Tx, activeOnly bool) ([]string, error)
}

type UseCase interface {
	GetUser(ctx context.Context, input ucModels.GetUserInput) (output ucModels.GetUserOutput, err error)

	// CyberMate usecases
	RegisterByTelegram(ctx context.Context, input ucModels.RegisterByTelegramInput) (ucModels.RegisterByTelegramOutput, error)
	GetUserByTelegramID(ctx context.Context, telegramID string) (ucModels.GetProfileOutput, error)
	UpdateProfileTheme(ctx context.Context, input ucModels.UpdateProfileThemeInput) (ucModels.UpdateProfileThemeOutput, error)
	ListReferralsByTelegramID(ctx context.Context, telegramID string) (ucModels.ListReferralsOutput, error)

	// Web site usecases
	RegisterWebAccount(ctx context.Context, input ucModels.RegisterWebAccountInput) (ucModels.AuthTokensOutput, error)
	LoginWebAccount(ctx context.Context, input ucModels.LoginWebAccountInput) (ucModels.AuthTokensOutput, error)
	GetWebAccount(ctx context.Context, webAccountID int64) (ucModels.GetWebAccountOutput, error)
	CreateWebPrompt(ctx context.Context, input ucModels.CreateWebPromptInput) (ucModels.CreateWebPromptOutput, error)
	ListWebPrompts(ctx context.Context, input ucModels.ListWebPromptsInput) (ucModels.ListWebPromptsOutput, error)

	// Admin panel usecases
	BootstrapAdmin(ctx context.Context) error
	AdminLogin(ctx context.Context, input ucModels.AdminLoginInput) (ucModels.AdminLoginOutput, error)
	GetAdmin(ctx context.Context, adminID int64) (ucModels.AdminUser, error)
	GetAdminStats(ctx context.Context) (ucModels.AdminStatsOutput, error)
	ListAdminUsers(ctx context.Context, input ucModels.AdminListUsersInput) (ucModels.AdminListUsersOutput, error)
	GetAdminUser(ctx context.Context, userID int64) (ucModels.AdminUserItem, error)
	UpdateAdminUserActive(ctx context.Context, input ucModels.AdminUpdateUserInput) (ucModels.AdminUserItem, error)
	DeleteAdminUser(ctx context.Context, userID int64) error
	AdminBroadcast(ctx context.Context, input ucModels.AdminBroadcastInput, messenger interface {
		Active() bool
		SendText(chatID int64, text, parseMode string) error
	}) (ucModels.AdminBroadcastOutput, error)
}

type Client interface {
	PostingsToCancel(ctx context.Context, token string, req rcModels.PostingsToCancelReq) (rcModels.PostingsToCancelResp, error)
	PostingsCancelResponse(ctx context.Context, token string, req rcModels.PostingsCancelResponseReq) (rcModels.PostingsCancelResponseResp, error)
}
