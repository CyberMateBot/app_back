package applinks

import (
	"context"
	"errors"

	"github.com/twelvepills-936/tgapp-/internal"
	ucModels "github.com/twelvepills-936/tgapp-/internal/usecase/models"
)

type fakeReferralsUC struct{}

func (f *fakeReferralsUC) GetUser(ctx context.Context, input ucModels.GetUserInput) (ucModels.GetUserOutput, error) {
	return ucModels.GetUserOutput{}, errors.New("not impl")
}
func (f *fakeReferralsUC) RegisterByTelegram(ctx context.Context, input ucModels.RegisterByTelegramInput) (ucModels.RegisterByTelegramOutput, error) {
	return ucModels.RegisterByTelegramOutput{}, errors.New("not impl")
}
func (f *fakeReferralsUC) GetUserByTelegramID(ctx context.Context, telegramID string) (ucModels.GetProfileOutput, error) {
	return ucModels.GetProfileOutput{}, errors.New("not impl")
}
func (f *fakeReferralsUC) UpdateProfileTheme(ctx context.Context, input ucModels.UpdateProfileThemeInput) (ucModels.UpdateProfileThemeOutput, error) {
	return ucModels.UpdateProfileThemeOutput{}, errors.New("not impl")
}
func (f *fakeReferralsUC) ListReferralsByTelegramID(ctx context.Context, telegramID string) (ucModels.ListReferralsOutput, error) {
	return ucModels.ListReferralsOutput{
		Referrals: []ucModels.ReferralItem{
			{TelegramID: "999", Name: "Friend", CompletedTasksCount: 1, Earnings: 50},
		},
		TotalCount:     1,
		TotalEarnings:  50,
		CompletedTasks: 1,
	}, nil
}
func (f *fakeReferralsUC) RegisterWebAccount(ctx context.Context, input ucModels.RegisterWebAccountInput) (ucModels.AuthTokensOutput, error) {
	return ucModels.AuthTokensOutput{}, errors.New("not impl")
}
func (f *fakeReferralsUC) LoginWebAccount(ctx context.Context, input ucModels.LoginWebAccountInput) (ucModels.AuthTokensOutput, error) {
	return ucModels.AuthTokensOutput{}, errors.New("not impl")
}
func (f *fakeReferralsUC) GetWebAccount(ctx context.Context, webAccountID int64) (ucModels.GetWebAccountOutput, error) {
	return ucModels.GetWebAccountOutput{}, errors.New("not impl")
}
func (f *fakeReferralsUC) CreateWebPrompt(ctx context.Context, input ucModels.CreateWebPromptInput) (ucModels.CreateWebPromptOutput, error) {
	return ucModels.CreateWebPromptOutput{}, errors.New("not impl")
}
func (f *fakeReferralsUC) ListWebPrompts(ctx context.Context, input ucModels.ListWebPromptsInput) (ucModels.ListWebPromptsOutput, error) {
	return ucModels.ListWebPromptsOutput{}, errors.New("not impl")
}

func (f *fakeReferralsUC) BootstrapAdmin(ctx context.Context) error { return nil }
func (f *fakeReferralsUC) AdminLogin(ctx context.Context, input ucModels.AdminLoginInput) (ucModels.AdminLoginOutput, error) {
	return ucModels.AdminLoginOutput{}, errors.New("not impl")
}
func (f *fakeReferralsUC) GetAdmin(ctx context.Context, adminID int64) (ucModels.AdminUser, error) {
	return ucModels.AdminUser{}, errors.New("not impl")
}
func (f *fakeReferralsUC) GetAdminStats(ctx context.Context) (ucModels.AdminStatsOutput, error) {
	return ucModels.AdminStatsOutput{}, errors.New("not impl")
}
func (f *fakeReferralsUC) ListAdminUsers(ctx context.Context, input ucModels.AdminListUsersInput) (ucModels.AdminListUsersOutput, error) {
	return ucModels.AdminListUsersOutput{}, errors.New("not impl")
}
func (f *fakeReferralsUC) GetAdminUser(ctx context.Context, userID int64) (ucModels.AdminUserItem, error) {
	return ucModels.AdminUserItem{}, errors.New("not impl")
}
func (f *fakeReferralsUC) UpdateAdminUserActive(ctx context.Context, input ucModels.AdminUpdateUserInput) (ucModels.AdminUserItem, error) {
	return ucModels.AdminUserItem{}, errors.New("not impl")
}
func (f *fakeReferralsUC) DeleteAdminUser(ctx context.Context, userID int64) error {
	return errors.New("not impl")
}
func (f *fakeReferralsUC) AdminCreditTokens(ctx context.Context, input ucModels.AdminTokenChangeInput) (ucModels.AdminTokenChangeOutput, error) {
	return ucModels.AdminTokenChangeOutput{}, errors.New("not impl")
}
func (f *fakeReferralsUC) AdminDebitTokens(ctx context.Context, input ucModels.AdminTokenChangeInput) (ucModels.AdminTokenChangeOutput, error) {
	return ucModels.AdminTokenChangeOutput{}, errors.New("not impl")
}
func (f *fakeReferralsUC) AdminBroadcast(ctx context.Context, input ucModels.AdminBroadcastInput, messenger interface {
	Active() bool
	SendText(chatID int64, text, parseMode string) error
}) (ucModels.AdminBroadcastOutput, error) {
	return ucModels.AdminBroadcastOutput{}, errors.New("not impl")
}
func (f *fakeReferralsUC) ListAdminEvents(ctx context.Context, limit int32) (ucModels.AdminListEventsOutput, error) {
	return ucModels.AdminListEventsOutput{}, errors.New("not impl")
}
func (f *fakeReferralsUC) ListAdminTransactions(ctx context.Context, input ucModels.AdminListTransactionsInput) (ucModels.AdminListTransactionsOutput, error) {
	return ucModels.AdminListTransactionsOutput{}, errors.New("not impl")
}
func (f *fakeReferralsUC) ListAdminBroadcasts(ctx context.Context, input ucModels.AdminListBroadcastsInput) (ucModels.AdminListBroadcastsOutput, error) {
	return ucModels.AdminListBroadcastsOutput{}, errors.New("not impl")
}
func (f *fakeReferralsUC) GetAdminSettings(ctx context.Context) (ucModels.AdminSettingsOutput, error) {
	return ucModels.AdminSettingsOutput{}, errors.New("not impl")
}
func (f *fakeReferralsUC) UpdateAdminSettings(ctx context.Context, input ucModels.AdminUpdateSettingsInput) (ucModels.AdminSettingsOutput, error) {
	return ucModels.AdminSettingsOutput{}, errors.New("not impl")
}
func (f *fakeReferralsUC) ListAdminModels(ctx context.Context) (ucModels.AdminListModelsOutput, error) {
	return ucModels.AdminListModelsOutput{}, errors.New("not impl")
}
func (f *fakeReferralsUC) UpdateAdminModel(ctx context.Context, input ucModels.AdminUpdateModelInput) (ucModels.AdminModelItem, error) {
	return ucModels.AdminModelItem{}, errors.New("not impl")
}
func (f *fakeReferralsUC) ListAdminSubscriptionPlans(ctx context.Context) (ucModels.AdminListSubscriptionPlansOutput, error) {
	return ucModels.AdminListSubscriptionPlansOutput{}, errors.New("not impl")
}
func (f *fakeReferralsUC) UpdateAdminSubscriptionPlans(ctx context.Context, input ucModels.AdminUpdateSubscriptionPlansInput) (ucModels.AdminListSubscriptionPlansOutput, error) {
	return ucModels.AdminListSubscriptionPlansOutput{}, errors.New("not impl")
}
func (f *fakeReferralsUC) ListAdminCoinPacks(ctx context.Context) (ucModels.AdminListCoinPacksOutput, error) {
	return ucModels.AdminListCoinPacksOutput{}, errors.New("not impl")
}
func (f *fakeReferralsUC) UpdateAdminCoinPacks(ctx context.Context, input ucModels.AdminUpdateCoinPacksInput) (ucModels.AdminListCoinPacksOutput, error) {
	return ucModels.AdminListCoinPacksOutput{}, errors.New("not impl")
}
func (f *fakeReferralsUC) GetPublicBillingCatalog(ctx context.Context) (ucModels.PublicBillingCatalogOutput, error) {
	return ucModels.PublicBillingCatalogOutput{}, errors.New("not impl")
}
func (f *fakeReferralsUC) GetSubscriptionByTelegramID(ctx context.Context, telegramID string) (ucModels.SubscriptionStateOutput, error) {
	return ucModels.SubscriptionStateOutput{}, errors.New("not impl")
}
func (f *fakeReferralsUC) AdminSetUserSubscription(ctx context.Context, input ucModels.AdminSetSubscriptionInput) (ucModels.AdminSubscriptionOutput, error) {
	return ucModels.AdminSubscriptionOutput{}, errors.New("not impl")
}
func (f *fakeReferralsUC) AdminClearUserSubscription(ctx context.Context, userID int64) (ucModels.AdminSubscriptionOutput, error) {
	return ucModels.AdminSubscriptionOutput{}, errors.New("not impl")
}
func (f *fakeReferralsUC) ListHomeWidgets(ctx context.Context) (ucModels.ListHomeWidgetsOutput, error) {
	return ucModels.ListHomeWidgetsOutput{}, errors.New("not impl")
}
func (f *fakeReferralsUC) ListAdminHomeWidgets(ctx context.Context) (ucModels.ListHomeWidgetsOutput, error) {
	return ucModels.ListHomeWidgetsOutput{}, errors.New("not impl")
}
func (f *fakeReferralsUC) CreateAdminHomeWidget(ctx context.Context, input ucModels.AdminCreateHomeWidgetInput) (ucModels.AdminHomeWidgetOutput, error) {
	return ucModels.AdminHomeWidgetOutput{}, errors.New("not impl")
}
func (f *fakeReferralsUC) UpdateAdminHomeWidget(ctx context.Context, input ucModels.AdminUpdateHomeWidgetInput) (ucModels.AdminHomeWidgetOutput, error) {
	return ucModels.AdminHomeWidgetOutput{}, errors.New("not impl")
}
func (f *fakeReferralsUC) DeleteAdminHomeWidget(ctx context.Context, id int64) error {
	return errors.New("not impl")
}

var _ internal.UseCase = (*fakeReferralsUC)(nil)
