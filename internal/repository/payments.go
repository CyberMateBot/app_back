package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	repoModels "github.com/twelvepills-936/tgapp-/internal/repository/models"
)

// CreatePayment inserts a new pending payment order and returns its id.
func (r *Repository) CreatePayment(ctx context.Context, tx pgx.Tx, p repoModels.Payment) (int64, error) {
	const q = `
INSERT INTO payments(profile_id, provider, idempotence_key, kind, item_id, amount_rub, coins, status)
VALUES($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id`

	provider := p.Provider
	if provider == "" {
		provider = "yookassa"
	}
	status := p.Status
	if status == "" {
		status = repoModels.PaymentStatusPending
	}

	var id int64
	qry := r.getQueryable(tx)
	err := qry.QueryRow(ctx, q,
		p.ProfileID, provider, p.IdempotenceKey, p.Kind, p.ItemID, p.AmountRub, p.Coins, status,
	).Scan(&id)
	return id, err
}

// UpdatePaymentProvider stores the YooKassa payment id and confirmation URL
// once the create-payment API call succeeds.
func (r *Repository) UpdatePaymentProvider(ctx context.Context, tx pgx.Tx, id int64, providerPaymentID, confirmationURL string) error {
	const q = `
UPDATE payments
SET provider_payment_id = $2, confirmation_url = $3, updated_at = NOW()
WHERE id = $1`

	qry := r.getQueryable(tx)
	tag, err := qry.Exec(ctx, q, id, providerPaymentID, confirmationURL)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// GetPaymentByID returns a payment order by its internal id.
func (r *Repository) GetPaymentByID(ctx context.Context, tx pgx.Tx, id int64) (repoModels.Payment, error) {
	const q = `
SELECT id, profile_id, provider, COALESCE(provider_payment_id, ''), idempotence_key,
       kind, item_id, amount_rub, coins, status, COALESCE(confirmation_url, ''), created_at, updated_at
FROM payments WHERE id = $1`

	var p repoModels.Payment
	qry := r.getQueryable(tx)
	err := qry.QueryRow(ctx, q, id).Scan(
		&p.ID, &p.ProfileID, &p.Provider, &p.ProviderPaymentID, &p.IdempotenceKey,
		&p.Kind, &p.ItemID, &p.AmountRub, &p.Coins, &p.Status, &p.ConfirmationURL, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return p, pgx.ErrNoRows
		}
		return p, err
	}
	return p, nil
}

// LockPaymentByProviderID locks (FOR UPDATE) and returns a payment order by its
// YooKassa payment id. Use inside a transaction before mutating payment status
// to avoid double-fulfilling the same webhook notification.
func (r *Repository) LockPaymentByProviderID(ctx context.Context, tx pgx.Tx, providerPaymentID string) (repoModels.Payment, error) {
	const q = `
SELECT id, profile_id, provider, COALESCE(provider_payment_id, ''), idempotence_key,
       kind, item_id, amount_rub, coins, status, COALESCE(confirmation_url, ''), created_at, updated_at
FROM payments WHERE provider_payment_id = $1
FOR UPDATE`

	var p repoModels.Payment
	qry := r.getQueryable(tx)
	err := qry.QueryRow(ctx, q, providerPaymentID).Scan(
		&p.ID, &p.ProfileID, &p.Provider, &p.ProviderPaymentID, &p.IdempotenceKey,
		&p.Kind, &p.ItemID, &p.AmountRub, &p.Coins, &p.Status, &p.ConfirmationURL, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return p, pgx.ErrNoRows
		}
		return p, err
	}
	return p, nil
}

// UpdatePaymentStatus transitions a payment order to a new status.
func (r *Repository) UpdatePaymentStatus(ctx context.Context, tx pgx.Tx, id int64, status string) error {
	const q = `UPDATE payments SET status = $2, updated_at = NOW() WHERE id = $1`
	qry := r.getQueryable(tx)
	tag, err := qry.Exec(ctx, q, id, status)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
