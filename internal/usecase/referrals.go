package usecase

import (
	"context"
	"errors"
	"log/slog"

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

func (uc *useCase) linkReferral(ctx context.Context, tx pgx.Tx, refereeProfileID int64, refereeTelegramID, startParam string) (bool, error) {
	referrerID := applinks.ParseReferralStartParam(startParam, "")
	if referrerID == "" {
		referrerID = startParam
	}
	if referrerID == "" || referrerID == refereeTelegramID {
		return false, nil
	}

	ref, refErr := uc.repo.GetProfileByTelegramID(ctx, tx, referrerID)
	if refErr != nil || ref.ID == 0 || ref.ID == refereeProfileID {
		return false, nil
	}
	if err := uc.repo.AddReferral(ctx, tx, ref.ID, refereeProfileID); err != nil {
		return false, err
	}
	return true, nil
}

// creditReferralBonusOnSubscription pays out the referrer bonus the first
// time an invited profile actually pays for a subscription. Called from the
// YooKassa webhook handler AFTER the subscription is persisted, inside the
// same transaction so the bonus is atomic with the purchase.
//
// This deliberately does *not* trigger on any billable action (e.g. AI
// generations that spend the 20-coin registration bonus) — that let cheap
// non-paying accounts unlock a valuable referrer bonus for free. Requiring
// a real ruble-denominated purchase guarantees the platform earns revenue
// before it pays out the bonus, so a referral chain can't be a net loss.
//
// Idempotent: the referrals row's completed_tasks_count flag prevents
// double-payment across retried webhooks.
func (uc *useCase) creditReferralBonusOnSubscription(ctx context.Context, tx pgx.Tx, refereeProfileID int64) error {
	referralID, referrerProfileID, err := uc.repo.LockPendingReferralByReferee(ctx, tx, refereeProfileID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// No inviter, or referrer was already paid — nothing to do.
			return nil
		}
		return err
	}

	settings, err := uc.GetAdminSettings(ctx)
	if err != nil {
		return err
	}
	bonus := settings.ReferralBonus
	if bonus <= 0 {
		// Referrer program is turned off in admin settings — still mark
		// the referral as "processed" so we don't keep re-checking it.
		return uc.repo.MarkReferralCompleted(ctx, tx, referralID, 0)
	}

	if _, err := uc.repo.CreditProfileTokens(ctx, tx, referrerProfileID, 0, bonus, "referral_bonus"); err != nil {
		return err
	}

	if err := uc.repo.MarkReferralCompleted(ctx, tx, referralID, bonus); err != nil {
		return err
	}

	slog.InfoContext(ctx, "referral bonus credited on subscription purchase",
		slog.Int64("referrer_profile_id", referrerProfileID),
		slog.Int64("referee_profile_id", refereeProfileID),
		slog.Int64("bonus", bonus),
	)
	return nil
}
