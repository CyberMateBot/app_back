package usecase

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	ucModels "github.com/twelvepills-936/tgapp-/internal/usecase/models"
	"github.com/twelvepills-936/tgapp-/pkg/applinks"
)

func (uc *useCase) ListReferralsByTelegramID(ctx context.Context, telegramID string) (ucModels.ListReferralsOutput, error) {
	var out ucModels.ListReferralsOutput
	if err := out.ValidateTelegramID(telegramID); err != nil {
		return out, err
	}

	referrer, err := uc.repo.GetProfileByTelegramID(ctx, nil, telegramID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return out, ucModels.ErrProfileNotFound
		}
		return out, err
	}

	items, err := uc.repo.ListReferralsByReferrerProfileID(ctx, nil, referrer.ID)
	if err != nil {
		return out, err
	}

	out.Referrals = make([]ucModels.ReferralItem, 0, len(items))
	for _, item := range items {
		out.Referrals = append(out.Referrals, ucModels.ReferralItem{
			ProfileID:           item.RefereeProfileID,
			TelegramID:          item.RefereeTelegramID,
			Name:                item.RefereeName,
			Username:            item.RefereeUsername,
			Avatar:              item.RefereeAvatar,
			CompletedTasksCount: item.CompletedTasksCount,
			Earnings:            item.Earnings,
		})
		out.TotalCount++
		out.TotalEarnings += item.Earnings
		out.CompletedTasks += item.CompletedTasksCount
	}
	return out, nil
}

func (uc *useCase) linkReferral(ctx context.Context, tx pgx.Tx, refereeProfileID int64, refereeTelegramID, startParam string) error {
	referrerID := applinks.ParseReferralStartParam(startParam, "")
	if referrerID == "" {
		referrerID = startParam
	}
	if referrerID == "" || referrerID == refereeTelegramID {
		return nil
	}

	ref, refErr := uc.repo.GetProfileByTelegramID(ctx, tx, referrerID)
	if refErr != nil || ref.ID == 0 || ref.ID == refereeProfileID {
		return nil
	}
	return uc.repo.AddReferral(ctx, tx, ref.ID, refereeProfileID)
}
