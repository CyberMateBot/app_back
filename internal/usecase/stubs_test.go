package usecase

import (
	"context"
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
func (fakeAdminUCStubs) AdminBroadcast(ctx context.Context, input ucModels.AdminBroadcastInput, messenger interface {
	Active() bool
	SendText(chatID int64, text, parseMode string) error
}) (ucModels.AdminBroadcastOutput, error) {
	return ucModels.AdminBroadcastOutput{}, errors.New("not impl")
}
