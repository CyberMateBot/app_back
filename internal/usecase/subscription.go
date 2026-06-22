package usecase

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	ucModels "github.com/twelvepills-936/tgapp-/internal/usecase/models"
	"github.com/twelvepills-936/tgapp-/pkg/billing"
)

// GetSubscriptionByTelegramID resolves the effective subscription state for a user.
// Expiry is applied: an expired paid plan resolves to "free".
func (uc *useCase) GetSubscriptionByTelegramID(ctx context.Context, telegramID string) (ucModels.SubscriptionStateOutput, error) {
	telegramID = strings.TrimSpace(telegramID)
	if telegramID == "" {
		return ucModels.SubscriptionStateOutput{}, fmt.Errorf("%w: telegram_id is required", ucModels.ErrInvalidInput)
	}

	plans, err := uc.loadSubscriptionPlans(ctx)
	if err != nil {
		return ucModels.SubscriptionStateOutput{}, err
	}

	sub, err := uc.repo.GetUserSubscriptionByTelegramID(ctx, nil, telegramID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uc.buildSubscriptionState(billing.PlanFree, nil, nil, plans), nil
		}
		return ucModels.SubscriptionStateOutput{}, err
	}

	return uc.buildSubscriptionState(sub.PlanID, &sub.StartedAt, sub.ExpiresAt, plans), nil
}

// AdminSetUserSubscription grants or edits a subscription for a user with any duration.
func (uc *useCase) AdminSetUserSubscription(ctx context.Context, input ucModels.AdminSetSubscriptionInput) (ucModels.AdminSubscriptionOutput, error) {
	if err := input.Validate(); err != nil {
		return ucModels.AdminSubscriptionOutput{}, err
	}
	planID := strings.ToLower(strings.TrimSpace(input.PlanID))
	if !billing.IsValidPlanID(planID) {
		return ucModels.AdminSubscriptionOutput{}, fmt.Errorf("%w: unknown plan_id %q", ucModels.ErrInvalidInput, planID)
	}

	plans, err := uc.loadSubscriptionPlans(ctx)
	if err != nil {
		return ucModels.AdminSubscriptionOutput{}, err
	}
	plan, found := findPlan(plans, planID)
	if !found {
		return ucModels.AdminSubscriptionOutput{}, fmt.Errorf("%w: plan %q is not configured", ucModels.ErrInvalidInput, planID)
	}

	now := time.Now().UTC()

	// Resolve expiry: explicit ExpiresAt > NoExpiry > DurationDays > plan default (30d for paid).
	var expiresAt *time.Time
	switch {
	case input.ExpiresAt != nil:
		t := input.ExpiresAt.UTC()
		expiresAt = &t
	case input.NoExpiry:
		expiresAt = nil
	case input.DurationDays > 0:
		t := now.AddDate(0, 0, input.DurationDays)
		expiresAt = &t
	case planID == billing.PlanFree:
		expiresAt = nil
	default:
		t := now.AddDate(0, 0, 30)
		expiresAt = &t
	}

	tx, err := uc.repo.DBBeginTransaction(ctx)
	if err != nil {
		return ucModels.AdminSubscriptionOutput{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(context.Background())
		}
	}()

	var adminPtr *int64
	if input.AdminID > 0 {
		adminID := input.AdminID
		adminPtr = &adminID
	}

	if planID == billing.PlanFree {
		if err := uc.repo.DeleteUserSubscription(ctx, tx, input.UserID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ucModels.AdminSubscriptionOutput{}, ucModels.ErrAdminUserNotFound
			}
			return ucModels.AdminSubscriptionOutput{}, err
		}
	} else if err := uc.repo.UpsertUserSubscription(ctx, tx, input.UserID, planID, now, expiresAt, adminPtr); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ucModels.AdminSubscriptionOutput{}, ucModels.ErrAdminUserNotFound
		}
		return ucModels.AdminSubscriptionOutput{}, err
	}

	var coinsGranted int64
	if input.GrantCoins && plan.Coins > 0 {
		if _, err := uc.repo.CreditProfileTokens(ctx, tx, input.UserID, input.AdminID, plan.Coins,
			fmt.Sprintf("subscription:%s", planID)); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ucModels.AdminSubscriptionOutput{}, ucModels.ErrAdminUserNotFound
			}
			return ucModels.AdminSubscriptionOutput{}, err
		}
		coinsGranted = plan.Coins
	}

	if err := tx.Commit(ctx); err != nil {
		return ucModels.AdminSubscriptionOutput{}, err
	}
	committed = true

	user, err := uc.GetAdminUser(ctx, input.UserID)
	if err != nil {
		return ucModels.AdminSubscriptionOutput{}, err
	}

	var startedPtr, expiresPtr *time.Time
	if planID != billing.PlanFree {
		startedPtr = &now
		expiresPtr = expiresAt
	}
	state := uc.buildSubscriptionState(planID, startedPtr, expiresPtr, plans)

	return ucModels.AdminSubscriptionOutput{
		User:         user,
		Subscription: state,
		CoinsGranted: coinsGranted,
	}, nil
}

// AdminClearUserSubscription revokes a subscription (reverts to free).
func (uc *useCase) AdminClearUserSubscription(ctx context.Context, userID int64) (ucModels.AdminSubscriptionOutput, error) {
	if userID <= 0 {
		return ucModels.AdminSubscriptionOutput{}, fmt.Errorf("%w: user_id invalid", ucModels.ErrInvalidInput)
	}
	if err := uc.repo.DeleteUserSubscription(ctx, nil, userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ucModels.AdminSubscriptionOutput{}, ucModels.ErrAdminUserNotFound
		}
		return ucModels.AdminSubscriptionOutput{}, err
	}
	plans, err := uc.loadSubscriptionPlans(ctx)
	if err != nil {
		return ucModels.AdminSubscriptionOutput{}, err
	}
	user, err := uc.GetAdminUser(ctx, userID)
	if err != nil {
		return ucModels.AdminSubscriptionOutput{}, err
	}
	return ucModels.AdminSubscriptionOutput{
		User:         user,
		Subscription: uc.buildSubscriptionState(billing.PlanFree, nil, nil, plans),
	}, nil
}

// buildSubscriptionState applies expiry and derives the display/state fields.
func (uc *useCase) buildSubscriptionState(planID string, started, expires *time.Time, plans []ucModels.SubscriptionPlanItem) ucModels.SubscriptionStateOutput {
	planID = strings.ToLower(strings.TrimSpace(planID))
	if planID == "" {
		planID = billing.PlanFree
	}

	now := time.Now().UTC()
	expired := false
	if expires != nil && !expires.After(now) {
		expired = true
	}

	effectivePlan := planID
	if expired {
		effectivePlan = billing.PlanFree
	}

	state := ucModels.SubscriptionStateOutput{
		PlanID:   effectivePlan,
		PlanRank: billing.PlanRank(effectivePlan),
		IsPaid:   effectivePlan != billing.PlanFree,
		IsActive: effectivePlan != billing.PlanFree,
		Expired:  expired,
	}

	if plan, ok := findPlan(plans, effectivePlan); ok {
		state.PlanName = plan.Name
		state.Coins = plan.Coins
	} else {
		state.PlanName = strings.Title(effectivePlan) //nolint:staticcheck
	}

	if started != nil {
		state.StartedAt = started.UTC().Format(time.RFC3339)
	}

	if effectivePlan != billing.PlanFree && expires != nil {
		state.ExpiresAt = expires.UTC().Format(time.RFC3339)
		remaining := expires.Sub(now)
		if remaining < 0 {
			remaining = 0
		}
		state.HoursLeft = int(math.Ceil(remaining.Hours()))
		state.DaysLeft = int(math.Ceil(remaining.Hours() / 24))
		if state.DaysLeft <= ucModels.ExpiringSoonDays && state.DaysLeft >= 0 {
			state.ExpiringSoon = true
		}
	}

	return state
}

func findPlan(plans []ucModels.SubscriptionPlanItem, planID string) (ucModels.SubscriptionPlanItem, bool) {
	planID = strings.ToLower(strings.TrimSpace(planID))
	for _, p := range plans {
		if strings.ToLower(strings.TrimSpace(p.ID)) == planID {
			return p, true
		}
	}
	return ucModels.SubscriptionPlanItem{}, false
}
