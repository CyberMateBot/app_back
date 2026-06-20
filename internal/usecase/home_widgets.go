package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	repoModels "github.com/twelvepills-936/tgapp-/internal/repository/models"
	ucModels "github.com/twelvepills-936/tgapp-/internal/usecase/models"
)

func mapHomeWidget(w repoModels.HomeWidget) ucModels.HomeWidgetItem {
	return ucModels.HomeWidgetItem{
		ID:              w.ID,
		SortOrder:       w.SortOrder,
		TagText:         w.TagText,
		TagBg:           w.TagBg,
		TagColor:        w.TagColor,
		Title:           w.Title,
		Description:     w.Description,
		BackgroundStyle: w.BackgroundStyle,
		ImageURL:        w.ImageURL,
		IsActive:        w.IsActive,
	}
}

func (uc *useCase) ListHomeWidgets(ctx context.Context) (ucModels.ListHomeWidgetsOutput, error) {
	rows, err := uc.repo.ListHomeWidgets(ctx, nil, true)
	if err != nil {
		return ucModels.ListHomeWidgetsOutput{}, err
	}
	items := make([]ucModels.HomeWidgetItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapHomeWidget(row))
	}
	return ucModels.ListHomeWidgetsOutput{Data: items}, nil
}

func (uc *useCase) ListAdminHomeWidgets(ctx context.Context) (ucModels.ListHomeWidgetsOutput, error) {
	rows, err := uc.repo.ListHomeWidgets(ctx, nil, false)
	if err != nil {
		return ucModels.ListHomeWidgetsOutput{}, err
	}
	items := make([]ucModels.HomeWidgetItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapHomeWidget(row))
	}
	return ucModels.ListHomeWidgetsOutput{Data: items}, nil
}

func (uc *useCase) CreateAdminHomeWidget(ctx context.Context, input ucModels.AdminCreateHomeWidgetInput) (ucModels.AdminHomeWidgetOutput, error) {
	input.Normalize()
	if input.Title == "" {
		return ucModels.AdminHomeWidgetOutput{}, ucModels.ErrInvalidInput
	}

	id, err := uc.repo.CreateHomeWidget(ctx, nil, repoModels.HomeWidget{
		SortOrder:       input.SortOrder,
		TagText:         input.TagText,
		TagBg:           input.TagBg,
		TagColor:        input.TagColor,
		Title:           input.Title,
		Description:     input.Description,
		BackgroundStyle: input.BackgroundStyle,
		ImageURL:        input.ImageURL,
		IsActive:        input.IsActive,
	})
	if err != nil {
		return ucModels.AdminHomeWidgetOutput{}, err
	}

	w, err := uc.repo.GetHomeWidgetByID(ctx, nil, id)
	if err != nil {
		return ucModels.AdminHomeWidgetOutput{}, err
	}
	return ucModels.AdminHomeWidgetOutput{HomeWidgetItem: mapHomeWidget(w)}, nil
}

func (uc *useCase) UpdateAdminHomeWidget(ctx context.Context, input ucModels.AdminUpdateHomeWidgetInput) (ucModels.AdminHomeWidgetOutput, error) {
	if input.ID < 1 {
		return ucModels.AdminHomeWidgetOutput{}, ucModels.ErrInvalidInput
	}

	current, err := uc.repo.GetHomeWidgetByID(ctx, nil, input.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ucModels.AdminHomeWidgetOutput{}, ucModels.ErrHomeWidgetNotFound
		}
		return ucModels.AdminHomeWidgetOutput{}, err
	}

	if input.SortOrder != nil {
		current.SortOrder = *input.SortOrder
	}
	if input.TagText != nil {
		current.TagText = strings.TrimSpace(*input.TagText)
	}
	if input.TagBg != nil {
		current.TagBg = strings.TrimSpace(*input.TagBg)
		if current.TagBg == "" {
			current.TagBg = "rgba(60,200,100,0.85)"
		}
	}
	if input.TagColor != nil {
		current.TagColor = strings.TrimSpace(*input.TagColor)
		if current.TagColor == "" {
			current.TagColor = "#06291a"
		}
	}
	if input.Title != nil {
		current.Title = strings.TrimSpace(*input.Title)
		if current.Title == "" {
			return ucModels.AdminHomeWidgetOutput{}, ucModels.ErrInvalidInput
		}
	}
	if input.Description != nil {
		current.Description = strings.TrimSpace(*input.Description)
	}
	if input.BackgroundStyle != nil {
		current.BackgroundStyle = strings.TrimSpace(*input.BackgroundStyle)
	}
	if input.ImageURL != nil {
		current.ImageURL = strings.TrimSpace(*input.ImageURL)
	}
	if input.IsActive != nil {
		current.IsActive = *input.IsActive
	}
	if current.BackgroundStyle == "" && current.ImageURL == "" {
		current.BackgroundStyle = "linear-gradient(135deg,#1a1030,#2a1840)"
	}

	if err := uc.repo.UpdateHomeWidget(ctx, nil, current); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ucModels.AdminHomeWidgetOutput{}, ucModels.ErrHomeWidgetNotFound
		}
		return ucModels.AdminHomeWidgetOutput{}, err
	}

	updated, err := uc.repo.GetHomeWidgetByID(ctx, nil, input.ID)
	if err != nil {
		return ucModels.AdminHomeWidgetOutput{}, err
	}
	return ucModels.AdminHomeWidgetOutput{HomeWidgetItem: mapHomeWidget(updated)}, nil
}

func (uc *useCase) DeleteAdminHomeWidget(ctx context.Context, id int64) error {
	if id < 1 {
		return ucModels.ErrInvalidInput
	}
	err := uc.repo.DeleteHomeWidget(ctx, nil, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return ucModels.ErrHomeWidgetNotFound
	}
	return err
}
