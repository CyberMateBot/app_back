package service

import (
    "context"
    "errors"
    "testing"

    "github.com/twelvepills-936/tgapp-/internal"
    ucModels "github.com/twelvepills-936/tgapp-/internal/usecase/models"
    api "github.com/twelvepills-936/tgapp-/pkg/api"
)

type fakeUC struct{}

func (f *fakeUC) GetUser(ctx context.Context, input ucModels.GetUserInput) (ucModels.GetUserOutput, error) {
    if input.UserID == 7 { return ucModels.GetUserOutput{Data: ucModels.User{ID:7, Name:"N", Surname:"S"}}, nil }
    return ucModels.GetUserOutput{}, ucModels.ErrUserIsNotFound
}
func (f *fakeUC) RegisterByTelegram(ctx context.Context, input ucModels.RegisterByTelegramInput) (ucModels.RegisterByTelegramOutput, error) {
    if input.InitDataRaw == "dup" { return ucModels.RegisterByTelegramOutput{}, ucModels.ErrProfileAlreadyRegistered }
    if input.InitDataRaw == "" { return ucModels.RegisterByTelegramOutput{}, errors.New("bad") }
    return ucModels.RegisterByTelegramOutput{ProfileID: 1}, nil
}
func (f *fakeUC) GetUserByTelegramID(ctx context.Context, telegramID string) (ucModels.GetProfileOutput, error) {
    if telegramID == "x" { return ucModels.GetProfileOutput{}, ucModels.ErrProfileNotFound }
    return ucModels.GetProfileOutput{Data: ucModels.ProfileUser{ID:1, Name:"A", Theme: ucModels.ThemeLight}}, nil
}
func (f *fakeUC) UpdateProfileTheme(ctx context.Context, input ucModels.UpdateProfileThemeInput) (ucModels.UpdateProfileThemeOutput, error) {
    if input.TelegramID == "x" { return ucModels.UpdateProfileThemeOutput{}, ucModels.ErrProfileNotFound }
    if input.Theme != ucModels.ThemeLight && input.Theme != ucModels.ThemeDark {
        return ucModels.UpdateProfileThemeOutput{}, ucModels.ErrInvalidInput
    }
    return ucModels.UpdateProfileThemeOutput{Theme: input.Theme}, nil
}

// Web site methods (unused in service tests)
func (f *fakeUC) RegisterWebAccount(ctx context.Context, input ucModels.RegisterWebAccountInput) (ucModels.AuthTokensOutput, error) {
    return ucModels.AuthTokensOutput{}, errors.New("not impl")
}
func (f *fakeUC) LoginWebAccount(ctx context.Context, input ucModels.LoginWebAccountInput) (ucModels.AuthTokensOutput, error) {
    return ucModels.AuthTokensOutput{}, errors.New("not impl")
}
func (f *fakeUC) GetWebAccount(ctx context.Context, webAccountID int64) (ucModels.GetWebAccountOutput, error) {
    return ucModels.GetWebAccountOutput{}, errors.New("not impl")
}
func (f *fakeUC) CreateWebPrompt(ctx context.Context, input ucModels.CreateWebPromptInput) (ucModels.CreateWebPromptOutput, error) {
    return ucModels.CreateWebPromptOutput{}, errors.New("not impl")
}
func (f *fakeUC) ListWebPrompts(ctx context.Context, input ucModels.ListWebPromptsInput) (ucModels.ListWebPromptsOutput, error) {
    return ucModels.ListWebPromptsOutput{}, errors.New("not impl")
}
func (f *fakeUC) ListReferralsByTelegramID(ctx context.Context, telegramID string) (ucModels.ListReferralsOutput, error) {
    if telegramID == "x" {
        return ucModels.ListReferralsOutput{}, ucModels.ErrProfileNotFound
    }
    return ucModels.ListReferralsOutput{
        Referrals: []ucModels.ReferralItem{
            {TelegramID: "999", Name: "Friend", CompletedTasksCount: 2, Earnings: 100},
        },
        TotalCount:     1,
        TotalEarnings:  100,
        CompletedTasks: 2,
    }, nil
}

func (f *fakeUC) BootstrapAdmin(ctx context.Context) error { return nil }
func (f *fakeUC) AdminLogin(ctx context.Context, input ucModels.AdminLoginInput) (ucModels.AdminLoginOutput, error) {
    return ucModels.AdminLoginOutput{}, errors.New("not impl")
}
func (f *fakeUC) GetAdmin(ctx context.Context, adminID int64) (ucModels.AdminUser, error) {
    return ucModels.AdminUser{}, errors.New("not impl")
}
func (f *fakeUC) GetAdminStats(ctx context.Context) (ucModels.AdminStatsOutput, error) {
    return ucModels.AdminStatsOutput{}, errors.New("not impl")
}
func (f *fakeUC) ListAdminUsers(ctx context.Context, input ucModels.AdminListUsersInput) (ucModels.AdminListUsersOutput, error) {
    return ucModels.AdminListUsersOutput{}, errors.New("not impl")
}
func (f *fakeUC) GetAdminUser(ctx context.Context, userID int64) (ucModels.AdminUserItem, error) {
    return ucModels.AdminUserItem{}, errors.New("not impl")
}
func (f *fakeUC) UpdateAdminUserActive(ctx context.Context, input ucModels.AdminUpdateUserInput) (ucModels.AdminUserItem, error) {
    return ucModels.AdminUserItem{}, errors.New("not impl")
}
func (f *fakeUC) DeleteAdminUser(ctx context.Context, userID int64) error {
    return errors.New("not impl")
}
func (f *fakeUC) AdminCreditTokens(ctx context.Context, input ucModels.AdminTokenChangeInput) (ucModels.AdminTokenChangeOutput, error) {
    return ucModels.AdminTokenChangeOutput{}, errors.New("not impl")
}
func (f *fakeUC) AdminDebitTokens(ctx context.Context, input ucModels.AdminTokenChangeInput) (ucModels.AdminTokenChangeOutput, error) {
    return ucModels.AdminTokenChangeOutput{}, errors.New("not impl")
}
func (f *fakeUC) AdminBroadcast(ctx context.Context, input ucModels.AdminBroadcastInput, messenger interface {
    Active() bool
    SendText(chatID int64, text, parseMode string) error
}) (ucModels.AdminBroadcastOutput, error) {
    return ucModels.AdminBroadcastOutput{}, errors.New("not impl")
}

var _ internal.UseCase = (*fakeUC)(nil)

func TestService_GetUser_OK(t *testing.T) {
    s := NewService(&fakeUC{})
    out, err := s.GetUser(context.Background(), &api.GetUserRequest{UserId: 7})
    if err != nil { t.Fatalf("unexpected error: %v", err) }
    if out.GetData().GetId() != 7 { t.Fatalf("unexpected id: %d", out.GetData().GetId()) }
}

func TestService_RegisterByTelegram_AlreadyExists(t *testing.T) {
    s := NewService(&fakeUC{})
    _, err := s.RegisterByTelegram(context.Background(), &api.RegisterByTelegramRequest{InitDataRaw: "dup"})
    if err == nil { t.Fatalf("expected error") }
}

func TestService_GetUserByTelegramId_NotFound(t *testing.T) {
    s := NewService(&fakeUC{})
    _, err := s.GetUserByTelegramId(context.Background(), &api.GetUserByTelegramIdRequest{TelegramId: "x"})
    if err == nil { t.Fatalf("expected not found") }
}


