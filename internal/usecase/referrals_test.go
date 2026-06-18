package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	repoModels "github.com/twelvepills-936/tgapp-/internal/repository/models"
	"github.com/twelvepills-936/tgapp-/pkg/config"
)

type fakeRepoReferrals struct {
	fakeAdminRepoStubs
	profiles map[string]repoModels.Profile
	referrals []repoModels.Referral
	addCalls  [][2]int64
}

func (f *fakeRepoReferrals) DBBeginTransaction(ctx context.Context) (pgx.Tx, error) { return &fakeTx{}, nil }
func (f *fakeRepoReferrals) ReadUser(ctx context.Context, id int64, dbTx pgx.Tx) (repoModels.User, error) {
	return repoModels.User{}, nil
}
func (f *fakeRepoReferrals) CreateProfile(ctx context.Context, tx pgx.Tx, p repoModels.Profile) (int64, error) {
	p.ID = int64(len(f.profiles) + 1)
	f.profiles[p.TelegramID] = p
	return p.ID, nil
}
func (f *fakeRepoReferrals) GetProfileByTelegramID(ctx context.Context, tx pgx.Tx, telegramID string) (repoModels.Profile, error) {
	if p, ok := f.profiles[telegramID]; ok {
		return p, nil
	}
	return repoModels.Profile{}, pgx.ErrNoRows
}
func (f *fakeRepoReferrals) UpdateProfileTheme(ctx context.Context, tx pgx.Tx, telegramID, theme string) error {
	return errors.New("not impl")
}
func (f *fakeRepoReferrals) CreateWalletForUser(ctx context.Context, tx pgx.Tx, profileID int64) (int64, error) {
	return 1, nil
}
func (f *fakeRepoReferrals) AddReferral(ctx context.Context, tx pgx.Tx, referrerProfileID int64, refereeProfileID int64) error {
	f.addCalls = append(f.addCalls, [2]int64{referrerProfileID, refereeProfileID})
	return nil
}
func (f *fakeRepoReferrals) ListReferralsByReferrerProfileID(ctx context.Context, tx pgx.Tx, referrerProfileID int64) ([]repoModels.Referral, error) {
	_ = ctx
	_ = tx
	_ = referrerProfileID
	return f.referrals, nil
}
func (f *fakeRepoReferrals) CreateWebAccount(ctx context.Context, tx pgx.Tx, a repoModels.WebAccount) (int64, error) {
	return 0, errors.New("not impl")
}
func (f *fakeRepoReferrals) GetWebAccountByEmail(ctx context.Context, tx pgx.Tx, email string) (repoModels.WebAccount, error) {
	return repoModels.WebAccount{}, errors.New("not impl")
}
func (f *fakeRepoReferrals) GetWebAccountByID(ctx context.Context, tx pgx.Tx, id int64) (repoModels.WebAccount, error) {
	return repoModels.WebAccount{}, errors.New("not impl")
}
func (f *fakeRepoReferrals) CreateWebPrompt(ctx context.Context, tx pgx.Tx, p repoModels.WebPrompt) (int64, error) {
	return 0, errors.New("not impl")
}
func (f *fakeRepoReferrals) ListWebPrompts(ctx context.Context, tx pgx.Tx, webAccountID int64, limit int32, offset int32) ([]repoModels.WebPrompt, error) {
	return nil, errors.New("not impl")
}

func TestListReferralsByTelegramID_OK(t *testing.T) {
	repo := &fakeRepoReferrals{
		profiles: map[string]repoModels.Profile{
			"111": {ID: 1, TelegramID: "111", Name: "Referrer"},
		},
		referrals: []repoModels.Referral{
			{
				RefereeProfileID:    2,
				RefereeTelegramID:   "222",
				RefereeName:         "Friend",
				RefereeUsername:     "friend",
				CompletedTasksCount: 3,
				Earnings:            150,
			},
		},
	}
	uc := NewUseCase(repo, config.ConfigJWT{})

	out, err := uc.ListReferralsByTelegramID(context.Background(), "111")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.TotalCount != 1 || out.TotalEarnings != 150 || out.CompletedTasks != 3 {
		t.Fatalf("unexpected summary: %+v", out)
	}
	if len(out.Referrals) != 1 || out.Referrals[0].TelegramID != "222" {
		t.Fatalf("unexpected referrals: %+v", out.Referrals)
	}
}
