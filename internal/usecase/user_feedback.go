package usecase

import (
	"context"
	"fmt"
	"time"

	repoModels "github.com/twelvepills-936/tgapp-/internal/repository/models"
	ucModels "github.com/twelvepills-936/tgapp-/internal/usecase/models"
)

func (uc *useCase) SubmitUserFeedback(ctx context.Context, input ucModels.SubmitUserFeedbackInput) (ucModels.SubmitUserFeedbackOutput, error) {
	if err := input.Validate(); err != nil {
		return ucModels.SubmitUserFeedbackOutput{}, err
	}

	profile, err := uc.repo.GetProfileByTelegramID(ctx, nil, input.TelegramID)
	if err != nil {
		return ucModels.SubmitUserFeedbackOutput{}, err
	}

	id, err := uc.repo.CreateUserFeedback(ctx, nil, profile.ID, input.Kind, input.Message)
	if err != nil {
		return ucModels.SubmitUserFeedbackOutput{}, err
	}

	return ucModels.SubmitUserFeedbackOutput{ID: id}, nil
}

func (uc *useCase) ListAdminUserFeedback(ctx context.Context, input ucModels.AdminListUserFeedbackInput) (ucModels.AdminListUserFeedbackOutput, error) {
	input.Normalize()
	offset := (input.Page - 1) * input.PerPage

	items, total, err := uc.repo.ListAdminUserFeedback(ctx, nil, input.Kind, input.PerPage, offset)
	if err != nil {
		return ucModels.AdminListUserFeedbackOutput{}, err
	}

	out := ucModels.AdminListUserFeedbackOutput{
		Data:  make([]ucModels.AdminUserFeedbackItem, 0, len(items)),
		Total: total,
	}
	for _, item := range items {
		out.Data = append(out.Data, mapAdminUserFeedback(item))
	}
	return out, nil
}

func (uc *useCase) DeleteAdminUserFeedback(ctx context.Context, id int64) error {
	if id < 1 {
		return fmt.Errorf("%w: invalid feedback id", ucModels.ErrInvalidInput)
	}
	return uc.repo.DeleteAdminUserFeedback(ctx, nil, id)
}

func mapAdminUserFeedback(item repoModels.UserFeedback) ucModels.AdminUserFeedbackItem {
	kindLabel := "Нововведение"
	if item.Kind == repoModels.UserFeedbackKindBug {
		kindLabel = "Баг"
	}
	return ucModels.AdminUserFeedbackItem{
		ID:        item.ID,
		User:      item.UserName,
		Kind:      item.Kind,
		KindLabel: kindLabel,
		Message:   item.Message,
		CreatedAt: item.CreatedAt.UTC().Format(time.RFC3339),
	}
}
