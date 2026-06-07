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

var _ internal.UseCase = (*fakeReferralsUC)(nil)
