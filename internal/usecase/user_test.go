package usecase

import (
    "context"
    "errors"
    "testing"

    "github.com/jackc/pgx/v5"
    "github.com/twelvepills-936/tgapp-/internal"
    repoModels "github.com/twelvepills-936/tgapp-/internal/repository/models"
    ucModels "github.com/twelvepills-936/tgapp-/internal/usecase/models"
    "github.com/twelvepills-936/tgapp-/pkg/config"
)

type fakeRepo struct{}

func (f *fakeRepo) DBBeginTransaction(ctx context.Context) (pgx.Tx, error) { return nil, nil }
func (f *fakeRepo) ReadUser(ctx context.Context, id int64, _ pgx.Tx) (repoModels.User, error) {
    if id == 42 { return repoModels.User{ID: 42, Name: "John", Surname: "Doe"}, nil }
    return repoModels.User{}, repoModels.ErrUserIsNotFound
}
// CyberMate methods (unused in this test)
func (f *fakeRepo) CreateProfile(ctx context.Context, tx pgx.Tx, p repoModels.Profile) (int64, error) { return 0, nil }
func (f *fakeRepo) GetProfileByTelegramID(ctx context.Context, tx pgx.Tx, telegramID string) (repoModels.Profile, error) { return repoModels.Profile{}, errors.New("not impl") }
func (f *fakeRepo) CreateWalletForUser(ctx context.Context, tx pgx.Tx, profileID int64) (int64, error) { return 0, nil }
func (f *fakeRepo) AddReferral(ctx context.Context, tx pgx.Tx, referrerProfileID int64, refereeProfileID int64) error { return nil }
func (f *fakeRepo) ListReferralsByReferrerProfileID(ctx context.Context, tx pgx.Tx, referrerProfileID int64) ([]repoModels.Referral, error) {
    return nil, errors.New("not impl")
}
func (f *fakeRepo) UpdateProfileTheme(ctx context.Context, tx pgx.Tx, telegramID, theme string) error { return nil }

// Web site methods (unused in this test)
func (f *fakeRepo) CreateWebAccount(ctx context.Context, tx pgx.Tx, a repoModels.WebAccount) (int64, error) { return 0, errors.New("not impl") }
func (f *fakeRepo) GetWebAccountByEmail(ctx context.Context, tx pgx.Tx, email string) (repoModels.WebAccount, error) { return repoModels.WebAccount{}, errors.New("not impl") }
func (f *fakeRepo) GetWebAccountByID(ctx context.Context, tx pgx.Tx, id int64) (repoModels.WebAccount, error) { return repoModels.WebAccount{}, errors.New("not impl") }
func (f *fakeRepo) CreateWebPrompt(ctx context.Context, tx pgx.Tx, p repoModels.WebPrompt) (int64, error) { return 0, errors.New("not impl") }
func (f *fakeRepo) ListWebPrompts(ctx context.Context, tx pgx.Tx, webAccountID int64, limit int32, offset int32) ([]repoModels.WebPrompt, error) {
    return nil, errors.New("not impl")
}

func TestGetUser_OK(t *testing.T) {
    var _ internal.Repository = (*fakeRepo)(nil)
    uc := NewUseCase(&fakeRepo{}, config.ConfigJWT{})
    out, err := uc.GetUser(context.Background(), ucModels.GetUserInput{UserID: 42})
    if err != nil { t.Fatalf("unexpected error: %v", err) }
    if out.Data.ID != 42 || out.Data.Name != "John" || out.Data.Surname != "Doe" {
        t.Fatalf("unexpected output: %+v", out)
    }
}

func TestGetUser_NotFound(t *testing.T) {
    uc := NewUseCase(&fakeRepo{}, config.ConfigJWT{})
    _, err := uc.GetUser(context.Background(), ucModels.GetUserInput{UserID: 99})
    if !errors.Is(err, ucModels.ErrUserIsNotFound) {
        t.Fatalf("expected ErrUserIsNotFound, got %v", err)
    }
}


