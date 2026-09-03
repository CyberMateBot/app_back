package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	repoModels "github.com/twelvepills-936/tgapp-/internal/repository/models"
	ucModels "github.com/twelvepills-936/tgapp-/internal/usecase/models"
	"github.com/twelvepills-936/tgapp-/pkg/billing"
	"github.com/twelvepills-936/tgapp-/pkg/yookassa"
)

const subscriptionDurationDays = 30

// StartCheckout resolves the requested catalog item (coin pack or subscription
// plan), records a pending payment order, and starts a YooKassa redirect
// checkout for it.
func (uc *useCase) StartCheckout(ctx context.Context, input ucModels.StartCheckoutInput) (ucModels.StartCheckoutOutput, error) {
	if err := input.Validate(); err != nil {
		return ucModels.StartCheckoutOutput{}, err
	}
	if uc.yookassa == nil || !uc.yookassa.Enabled() {
		return ucModels.StartCheckoutOutput{}, ucModels.ErrPaymentsDisabled
	}

	settings, err := uc.GetAdminSettings(ctx)
	if err != nil {
		return ucModels.StartCheckoutOutput{}, err
	}
	if !settings.YookassaEnabled {
		return ucModels.StartCheckoutOutput{}, ucModels.ErrPaymentsDisabled
	}

	telegramID := strings.TrimSpace(input.TelegramID)
	profile, err := uc.repo.GetProfileByTelegramID(ctx, nil, telegramID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ucModels.StartCheckoutOutput{}, ucModels.ErrProfileNotFound
		}
		return ucModels.StartCheckoutOutput{}, err
	}

	kind := strings.TrimSpace(input.Kind)
	itemID := strings.ToLower(strings.TrimSpace(input.ItemID))

	var (
		amountRub   float64
		coins       int64
		description string
	)

	switch kind {
	case ucModels.PaymentKindCoinPack:
		packs, err := uc.loadCoinPacks(ctx)
		if err != nil {
			return ucModels.StartCheckoutOutput{}, err
		}
		pack, ok := findCoinPack(packs, itemID)
		if !ok || !pack.Enabled || pack.PriceRub <= 0 {
			return ucModels.StartCheckoutOutput{}, ucModels.ErrItemNotFound
		}
		amountRub = float64(pack.PriceRub)
		coins = pack.Coins
		description = fmt.Sprintf("CyberMate: %s", pack.Name)

	case ucModels.PaymentKindSubscription:
		if itemID == billing.PlanFree {
			return ucModels.StartCheckoutOutput{}, ucModels.ErrItemNotFound
		}
		plans, err := uc.loadSubscriptionPlans(ctx)
		if err != nil {
			return ucModels.StartCheckoutOutput{}, err
		}
		plan, ok := findPlan(plans, itemID)
		if !ok || !plan.Enabled || plan.PriceRub <= 0 {
			return ucModels.StartCheckoutOutput{}, ucModels.ErrItemNotFound
		}
		amountRub = float64(plan.PriceRub)
		coins = plan.Coins
		description = fmt.Sprintf("CyberMate: подписка %s (1 месяц)", plan.Name)

	default:
		return ucModels.StartCheckoutOutput{}, fmt.Errorf("%w: unknown kind %q", ucModels.ErrInvalidInput, kind)
	}

	idempotenceKey, err := randomPaymentToken()
	if err != nil {
		return ucModels.StartCheckoutOutput{}, err
	}

	localID, err := uc.repo.CreatePayment(ctx, nil, repoModels.Payment{
		ProfileID:      profile.ID,
		Provider:       "yookassa",
		IdempotenceKey: idempotenceKey,
		Kind:           kind,
		ItemID:         itemID,
		AmountRub:      amountRub,
		Coins:          coins,
		Status:         repoModels.PaymentStatusPending,
	})
	if err != nil {
		return ucModels.StartCheckoutOutput{}, err
	}

	remote, err := uc.yookassa.CreatePayment(ctx, yookassa.CreatePaymentRequest{
		Amount:         yookassa.RubAmount(amountRub),
		Description:    description,
		ReturnURL:      uc.yookassaReturnURL,
		IdempotenceKey: idempotenceKey,
		Metadata: map[string]string{
			"payment_id":  strconv.FormatInt(localID, 10),
			"telegram_id": telegramID,
			"kind":        kind,
			"item_id":     itemID,
		},
	})
	if err != nil {
		slog.ErrorContext(ctx, "yookassa create payment failed", slog.Any("error", err), slog.Int64("payment_id", localID))
		return ucModels.StartCheckoutOutput{}, fmt.Errorf("failed to start checkout: %w", err)
	}

	if err := uc.repo.UpdatePaymentProvider(ctx, nil, localID, remote.ID, remote.Confirmation.ConfirmationURL); err != nil {
		return ucModels.StartCheckoutOutput{}, err
	}

	return ucModels.StartCheckoutOutput{
		PaymentID:       localID,
		ConfirmationURL: remote.Confirmation.ConfirmationURL,
		AmountRub:       amountRub,
		Coins:           coins,
	}, nil
}

// HandleYooKassaWebhookNotification is called whenever YooKassa pings our
// webhook endpoint. It never trusts the notification body: it re-fetches the
// payment from the API (the only way to reliably verify a notification, since
// YooKassa does not sign the payload) and fulfills the order accordingly.
// Unknown payment ids and already-fulfilled orders are treated as a no-op so
// retried notifications stay idempotent.
func (uc *useCase) HandleYooKassaWebhookNotification(ctx context.Context, providerPaymentID string) error {
	providerPaymentID = strings.TrimSpace(providerPaymentID)
	if providerPaymentID == "" {
		return fmt.Errorf("%w: payment id is required", ucModels.ErrInvalidInput)
	}
	if uc.yookassa == nil || !uc.yookassa.Enabled() {
		return ucModels.ErrPaymentsDisabled
	}

	remote, err := uc.yookassa.GetPayment(ctx, providerPaymentID)
	if err != nil {
		return fmt.Errorf("failed to verify payment with yookassa: %w", err)
	}

	tx, err := uc.repo.DBBeginTransaction(ctx)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(context.Background())
		}
	}()

	payment, err := uc.repo.LockPaymentByProviderID(ctx, tx, providerPaymentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.WarnContext(ctx, "yookassa webhook: unknown payment id", slog.String("provider_payment_id", providerPaymentID))
			return nil
		}
		return err
	}

	if payment.Status != repoModels.PaymentStatusPending {
		// Already fulfilled/canceled by an earlier notification — idempotent no-op.
		return nil
	}

	switch remote.Status {
	case "succeeded":
		if err := uc.repo.UpdatePaymentStatus(ctx, tx, payment.ID, repoModels.PaymentStatusSucceeded); err != nil {
			return err
		}

		switch payment.Kind {
		case ucModels.PaymentKindCoinPack:
			if payment.Coins > 0 {
				if _, err := uc.repo.CreditProfileTokens(ctx, tx, payment.ProfileID, 0, payment.Coins,
					fmt.Sprintf("yookassa:pack:%s", payment.ItemID)); err != nil {
					return err
				}
			}
		case ucModels.PaymentKindSubscription:
			now := time.Now().UTC()
			expiresAt := now.AddDate(0, 0, subscriptionDurationDays)
			if err := uc.repo.UpsertUserSubscription(ctx, tx, payment.ProfileID, payment.ItemID, now, &expiresAt, nil); err != nil {
				return err
			}
			if payment.Coins > 0 {
				if _, err := uc.repo.CreditProfileTokens(ctx, tx, payment.ProfileID, 0, payment.Coins,
					fmt.Sprintf("yookassa:subscription:%s", payment.ItemID)); err != nil {
					return err
				}
			}
			// A paid subscription is the only event that unlocks the
			// referrer bonus — see creditReferralBonusOnSubscription
			// for the "why this trigger" rationale. Any error here must
			// abort the whole webhook so the payment isn't marked
			// succeeded without the bonus (retried webhook will try
			// again idempotently).
			if err := uc.creditReferralBonusOnSubscription(ctx, tx, payment.ProfileID); err != nil {
				return err
			}
		}

	case "canceled":
		if err := uc.repo.UpdatePaymentStatus(ctx, tx, payment.ID, repoModels.PaymentStatusCanceled); err != nil {
			return err
		}

	default:
		// "pending" / "waiting_for_capture" — nothing to fulfill yet, wait for the next notification.
		return nil
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	committed = true
	return nil
}

func findCoinPack(packs []ucModels.CoinPackItem, packID string) (ucModels.CoinPackItem, bool) {
	packID = strings.ToLower(strings.TrimSpace(packID))
	for _, p := range packs {
		if strings.ToLower(strings.TrimSpace(p.ID)) == packID {
			return p, true
		}
	}
	return ucModels.CoinPackItem{}, false
}

func randomPaymentToken() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
