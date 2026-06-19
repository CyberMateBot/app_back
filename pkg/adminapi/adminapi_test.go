package adminapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/twelvepills-936/tgapp-/internal"
	ucModels "github.com/twelvepills-936/tgapp-/internal/usecase/models"
	"github.com/twelvepills-936/tgapp-/pkg/config"
)

type fakeAdminUC struct{}

func (f *fakeAdminUC) GetUser(ctx context.Context, input ucModels.GetUserInput) (ucModels.GetUserOutput, error) {
	return ucModels.GetUserOutput{}, errors.New("not impl")
}
func (f *fakeAdminUC) RegisterByTelegram(ctx context.Context, input ucModels.RegisterByTelegramInput) (ucModels.RegisterByTelegramOutput, error) {
	return ucModels.RegisterByTelegramOutput{}, errors.New("not impl")
}
func (f *fakeAdminUC) GetUserByTelegramID(ctx context.Context, telegramID string) (ucModels.GetProfileOutput, error) {
	return ucModels.GetProfileOutput{}, errors.New("not impl")
}
func (f *fakeAdminUC) UpdateProfileTheme(ctx context.Context, input ucModels.UpdateProfileThemeInput) (ucModels.UpdateProfileThemeOutput, error) {
	return ucModels.UpdateProfileThemeOutput{}, errors.New("not impl")
}
func (f *fakeAdminUC) ListReferralsByTelegramID(ctx context.Context, telegramID string) (ucModels.ListReferralsOutput, error) {
	return ucModels.ListReferralsOutput{}, errors.New("not impl")
}
func (f *fakeAdminUC) RegisterWebAccount(ctx context.Context, input ucModels.RegisterWebAccountInput) (ucModels.AuthTokensOutput, error) {
	return ucModels.AuthTokensOutput{}, errors.New("not impl")
}
func (f *fakeAdminUC) LoginWebAccount(ctx context.Context, input ucModels.LoginWebAccountInput) (ucModels.AuthTokensOutput, error) {
	return ucModels.AuthTokensOutput{}, errors.New("not impl")
}
func (f *fakeAdminUC) GetWebAccount(ctx context.Context, webAccountID int64) (ucModels.GetWebAccountOutput, error) {
	return ucModels.GetWebAccountOutput{}, errors.New("not impl")
}
func (f *fakeAdminUC) CreateWebPrompt(ctx context.Context, input ucModels.CreateWebPromptInput) (ucModels.CreateWebPromptOutput, error) {
	return ucModels.CreateWebPromptOutput{}, errors.New("not impl")
}
func (f *fakeAdminUC) ListWebPrompts(ctx context.Context, input ucModels.ListWebPromptsInput) (ucModels.ListWebPromptsOutput, error) {
	return ucModels.ListWebPromptsOutput{}, errors.New("not impl")
}
func (f *fakeAdminUC) BootstrapAdmin(ctx context.Context) error { return nil }
func (f *fakeAdminUC) AdminLogin(ctx context.Context, input ucModels.AdminLoginInput) (ucModels.AdminLoginOutput, error) {
	return ucModels.AdminLoginOutput{
		Token: "test-token",
		Admin: ucModels.AdminUser{ID: 1, Email: input.Email},
	}, nil
}
func (f *fakeAdminUC) GetAdmin(ctx context.Context, adminID int64) (ucModels.AdminUser, error) {
	return ucModels.AdminUser{ID: 1, Email: "admin@test.com"}, nil
}
func (f *fakeAdminUC) GetAdminStats(ctx context.Context) (ucModels.AdminStatsOutput, error) {
	return ucModels.AdminStatsOutput{TotalUsers: 10, ActiveUsersToday: 3, NewUsersToday: 1, TotalMessages: 42}, nil
}
func (f *fakeAdminUC) ListAdminUsers(ctx context.Context, input ucModels.AdminListUsersInput) (ucModels.AdminListUsersOutput, error) {
	return ucModels.AdminListUsersOutput{}, errors.New("not impl")
}
func (f *fakeAdminUC) GetAdminUser(ctx context.Context, userID int64) (ucModels.AdminUserItem, error) {
	return ucModels.AdminUserItem{}, errors.New("not impl")
}
func (f *fakeAdminUC) UpdateAdminUserActive(ctx context.Context, input ucModels.AdminUpdateUserInput) (ucModels.AdminUserItem, error) {
	return ucModels.AdminUserItem{}, errors.New("not impl")
}
func (f *fakeAdminUC) DeleteAdminUser(ctx context.Context, userID int64) error {
	return errors.New("not impl")
}
func (f *fakeAdminUC) AdminCreditTokens(ctx context.Context, input ucModels.AdminTokenChangeInput) (ucModels.AdminTokenChangeOutput, error) {
	return ucModels.AdminTokenChangeOutput{}, errors.New("not impl")
}
func (f *fakeAdminUC) AdminDebitTokens(ctx context.Context, input ucModels.AdminTokenChangeInput) (ucModels.AdminTokenChangeOutput, error) {
	return ucModels.AdminTokenChangeOutput{}, errors.New("not impl")
}
func (f *fakeAdminUC) AdminBroadcast(ctx context.Context, input ucModels.AdminBroadcastInput, messenger interface {
	Active() bool
	SendText(chatID int64, text, parseMode string) error
}) (ucModels.AdminBroadcastOutput, error) {
	return ucModels.AdminBroadcastOutput{Sent: 1}, nil
}

var _ internal.UseCase = (*fakeAdminUC)(nil)

func TestWrap_AdminLogin(t *testing.T) {
	t.Parallel()

	mux := Wrap(http.NotFoundHandler(), &fakeAdminUC{}, config.ConfigJWT{Secret: "test"}, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/auth/login", strings.NewReader(`{"email":"admin@test.com","password":"secret"}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var resp loginResp
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Token == "" || resp.Admin.Email != "admin@test.com" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}
