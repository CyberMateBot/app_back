package usecase

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	repoModels "github.com/twelvepills-936/tgapp-/internal/repository/models"
	ucModels "github.com/twelvepills-936/tgapp-/internal/usecase/models"
)

// fakeAdminRepoStubs provides no-op admin repository methods for test fakes.
type fakeAdminRepoStubs struct{}

func (fakeAdminRepoStubs) GetAdminByEmail(ctx context.Context, tx pgx.Tx, email string) (repoModels.Admin, error) {
	return repoModels.Admin{}, errors.New("not impl")
}
func (fakeAdminRepoStubs) GetAdminByID(ctx context.Context, tx pgx.Tx, id int64) (repoModels.Admin, error) {
	return repoModels.Admin{}, errors.New("not impl")
}
func (fakeAdminRepoStubs) CreateAdmin(ctx context.Context, tx pgx.Tx, email, passwordHash string) (int64, error) {
	return 0, errors.New("not impl")
}
func (fakeAdminRepoStubs) CountAdmins(ctx context.Context, tx pgx.Tx) (int64, error) {
	return 0, nil
}
func (fakeAdminRepoStubs) GetAdminStats(ctx context.Context, tx pgx.Tx) (repoModels.AdminStats, error) {
	return repoModels.AdminStats{}, errors.New("not impl")
}
func (fakeAdminRepoStubs) ListAdminProfiles(ctx context.Context, tx pgx.Tx, search string, limit, offset int32) ([]repoModels.AdminProfile, int64, error) {
	return nil, 0, errors.New("not impl")
}
func (fakeAdminRepoStubs) GetAdminProfileByID(ctx context.Context, tx pgx.Tx, id int64) (repoModels.AdminProfile, error) {
	return repoModels.AdminProfile{}, errors.New("not impl")
}
func (fakeAdminRepoStubs) UpdateProfileActive(ctx context.Context, tx pgx.Tx, id int64, isActive bool) error {
	return errors.New("not impl")
}
func (fakeAdminRepoStubs) DeleteProfile(ctx context.Context, tx pgx.Tx, id int64) error {
	return errors.New("not impl")
}
func (fakeAdminRepoStubs) ListBroadcastTelegramIDs(ctx context.Context, tx pgx.Tx, activeOnly bool) ([]string, error) {
	return nil, errors.New("not impl")
}
func (fakeAdminRepoStubs) CreditProfileTokens(ctx context.Context, tx pgx.Tx, profileID, adminID int64, amount int64, reason string) (repoModels.TokenOperationResult, error) {
	return repoModels.TokenOperationResult{}, errors.New("not impl")
}
func (fakeAdminRepoStubs) DebitProfileTokens(ctx context.Context, tx pgx.Tx, profileID, adminID int64, amount int64, reason string) (repoModels.TokenOperationResult, error) {
	return repoModels.TokenOperationResult{}, errors.New("not impl")
}
func (fakeAdminRepoStubs) ListAdminEvents(ctx context.Context, tx pgx.Tx, limit int32) ([]repoModels.AdminEvent, error) {
	return nil, errors.New("not impl")
}
func (fakeAdminRepoStubs) GetAdminTransactionStats(ctx context.Context, tx pgx.Tx) (repoModels.AdminTransactionStats, error) {
	return repoModels.AdminTransactionStats{}, errors.New("not impl")
}
func (fakeAdminRepoStubs) ListAdminTokenTransactions(ctx context.Context, tx pgx.Tx, operation string, limit, offset int32) ([]repoModels.AdminTokenTransaction, int64, error) {
	return nil, 0, errors.New("not impl")
}
func (fakeAdminRepoStubs) CreateAdminBroadcast(ctx context.Context, tx pgx.Tx, adminID int64, message, target, parseMode string, sent, failed int64) (int64, error) {
	return 0, errors.New("not impl")
}
func (fakeAdminRepoStubs) ListAdminBroadcasts(ctx context.Context, tx pgx.Tx, limit, offset int32) ([]repoModels.AdminBroadcastRecord, int64, error) {
	return nil, 0, errors.New("not impl")
}
func (fakeAdminRepoStubs) GetAdminSettings(ctx context.Context, tx pgx.Tx) (map[string]json.RawMessage, error) {
	return nil, errors.New("not impl")
}
func (fakeAdminRepoStubs) UpsertAdminSetting(ctx context.Context, tx pgx.Tx, key string, value any) error {
	return errors.New("not impl")
}
func (fakeAdminRepoStubs) ListModelConfigs(ctx context.Context, tx pgx.Tx) (map[string]repoModels.ModelConfig, error) {
	return nil, errors.New("not impl")
}
func (fakeAdminRepoStubs) UpsertModelConfig(ctx context.Context, tx pgx.Tx, cfg repoModels.ModelConfig) error {
	return errors.New("not impl")
}
func (fakeAdminRepoStubs) ListHomeWidgets(ctx context.Context, tx pgx.Tx, activeOnly bool) ([]repoModels.HomeWidget, error) {
	return nil, errors.New("not impl")
}
func (fakeAdminRepoStubs) GetHomeWidgetByID(ctx context.Context, tx pgx.Tx, id int64) (repoModels.HomeWidget, error) {
	return repoModels.HomeWidget{}, errors.New("not impl")
}
func (fakeAdminRepoStubs) CreateHomeWidget(ctx context.Context, tx pgx.Tx, w repoModels.HomeWidget) (int64, error) {
	return 0, errors.New("not impl")
}
func (fakeAdminRepoStubs) UpdateHomeWidget(ctx context.Context, tx pgx.Tx, w repoModels.HomeWidget) error {
	return errors.New("not impl")
}
func (fakeAdminRepoStubs) DeleteHomeWidget(ctx context.Context, tx pgx.Tx, id int64) error {
	return errors.New("not impl")
}

// fakeAdminUCStubs provides no-op admin usecase methods for test fakes.
type fakeAdminUCStubs struct{}

func (fakeAdminUCStubs) BootstrapAdmin(ctx context.Context) error { return nil }
func (fakeAdminUCStubs) AdminLogin(ctx context.Context, input ucModels.AdminLoginInput) (ucModels.AdminLoginOutput, error) {
	return ucModels.AdminLoginOutput{}, errors.New("not impl")
}
func (fakeAdminUCStubs) GetAdmin(ctx context.Context, adminID int64) (ucModels.AdminUser, error) {
	return ucModels.AdminUser{}, errors.New("not impl")
}
func (fakeAdminUCStubs) GetAdminStats(ctx context.Context) (ucModels.AdminStatsOutput, error) {
	return ucModels.AdminStatsOutput{}, errors.New("not impl")
}
func (fakeAdminUCStubs) ListAdminUsers(ctx context.Context, input ucModels.AdminListUsersInput) (ucModels.AdminListUsersOutput, error) {
	return ucModels.AdminListUsersOutput{}, errors.New("not impl")
}
func (fakeAdminUCStubs) GetAdminUser(ctx context.Context, userID int64) (ucModels.AdminUserItem, error) {
	return ucModels.AdminUserItem{}, errors.New("not impl")
}
func (fakeAdminUCStubs) UpdateAdminUserActive(ctx context.Context, input ucModels.AdminUpdateUserInput) (ucModels.AdminUserItem, error) {
	return ucModels.AdminUserItem{}, errors.New("not impl")
}
func (fakeAdminUCStubs) DeleteAdminUser(ctx context.Context, userID int64) error {
	return errors.New("not impl")
}
func (fakeAdminUCStubs) AdminCreditTokens(ctx context.Context, input ucModels.AdminTokenChangeInput) (ucModels.AdminTokenChangeOutput, error) {
	return ucModels.AdminTokenChangeOutput{}, errors.New("not impl")
}
func (fakeAdminUCStubs) AdminDebitTokens(ctx context.Context, input ucModels.AdminTokenChangeInput) (ucModels.AdminTokenChangeOutput, error) {
	return ucModels.AdminTokenChangeOutput{}, errors.New("not impl")
}
func (fakeAdminUCStubs) AdminBroadcast(ctx context.Context, input ucModels.AdminBroadcastInput, messenger interface {
	Active() bool
	SendText(chatID int64, text, parseMode string) error
}) (ucModels.AdminBroadcastOutput, error) {
	return ucModels.AdminBroadcastOutput{}, errors.New("not impl")
}
func (fakeAdminUCStubs) ListAdminEvents(ctx context.Context, limit int32) (ucModels.AdminListEventsOutput, error) {
	return ucModels.AdminListEventsOutput{}, errors.New("not impl")
}
func (fakeAdminUCStubs) ListAdminTransactions(ctx context.Context, input ucModels.AdminListTransactionsInput) (ucModels.AdminListTransactionsOutput, error) {
	return ucModels.AdminListTransactionsOutput{}, errors.New("not impl")
}
func (fakeAdminUCStubs) ListAdminBroadcasts(ctx context.Context, input ucModels.AdminListBroadcastsInput) (ucModels.AdminListBroadcastsOutput, error) {
	return ucModels.AdminListBroadcastsOutput{}, errors.New("not impl")
}
func (fakeAdminUCStubs) GetAdminSettings(ctx context.Context) (ucModels.AdminSettingsOutput, error) {
	return ucModels.AdminSettingsOutput{}, errors.New("not impl")
}
func (fakeAdminUCStubs) UpdateAdminSettings(ctx context.Context, input ucModels.AdminUpdateSettingsInput) (ucModels.AdminSettingsOutput, error) {
	return ucModels.AdminSettingsOutput{}, errors.New("not impl")
}
func (fakeAdminUCStubs) ListAdminModels(ctx context.Context) (ucModels.AdminListModelsOutput, error) {
	return ucModels.AdminListModelsOutput{}, errors.New("not impl")
}
func (fakeAdminUCStubs) UpdateAdminModel(ctx context.Context, input ucModels.AdminUpdateModelInput) (ucModels.AdminModelItem, error) {
	return ucModels.AdminModelItem{}, errors.New("not impl")
}
func (fakeAdminUCStubs) ListAdminSubscriptionPlans(ctx context.Context) (ucModels.AdminListSubscriptionPlansOutput, error) {
	return ucModels.AdminListSubscriptionPlansOutput{}, errors.New("not impl")
}
func (fakeAdminUCStubs) UpdateAdminSubscriptionPlans(ctx context.Context, input ucModels.AdminUpdateSubscriptionPlansInput) (ucModels.AdminListSubscriptionPlansOutput, error) {
	return ucModels.AdminListSubscriptionPlansOutput{}, errors.New("not impl")
}
func (fakeAdminUCStubs) ListAdminCoinPacks(ctx context.Context) (ucModels.AdminListCoinPacksOutput, error) {
	return ucModels.AdminListCoinPacksOutput{}, errors.New("not impl")
}
func (fakeAdminUCStubs) UpdateAdminCoinPacks(ctx context.Context, input ucModels.AdminUpdateCoinPacksInput) (ucModels.AdminListCoinPacksOutput, error) {
	return ucModels.AdminListCoinPacksOutput{}, errors.New("not impl")
}
func (fakeAdminUCStubs) GetPublicBillingCatalog(ctx context.Context) (ucModels.PublicBillingCatalogOutput, error) {
	return ucModels.PublicBillingCatalogOutput{}, errors.New("not impl")
}
func (fakeAdminUCStubs) ListHomeWidgets(ctx context.Context) (ucModels.ListHomeWidgetsOutput, error) {
	return ucModels.ListHomeWidgetsOutput{}, errors.New("not impl")
}
func (fakeAdminUCStubs) ListAdminHomeWidgets(ctx context.Context) (ucModels.ListHomeWidgetsOutput, error) {
	return ucModels.ListHomeWidgetsOutput{}, errors.New("not impl")
}
func (fakeAdminUCStubs) CreateAdminHomeWidget(ctx context.Context, input ucModels.AdminCreateHomeWidgetInput) (ucModels.AdminHomeWidgetOutput, error) {
	return ucModels.AdminHomeWidgetOutput{}, errors.New("not impl")
}
func (fakeAdminUCStubs) UpdateAdminHomeWidget(ctx context.Context, input ucModels.AdminUpdateHomeWidgetInput) (ucModels.AdminHomeWidgetOutput, error) {
	return ucModels.AdminHomeWidgetOutput{}, errors.New("not impl")
}
func (fakeAdminUCStubs) DeleteAdminHomeWidget(ctx context.Context, id int64) error {
	return errors.New("not impl")
}
