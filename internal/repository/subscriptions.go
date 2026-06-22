package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	repoModels "github.com/twelvepills-936/tgapp-/internal/repository/models"
)

// GetUserSubscription returns the stored subscription for a profile, or pgx.ErrNoRows.
func (r *Repository) GetUserSubscription(ctx context.Context, tx pgx.Tx, profileID int64) (repoModels.UserSubscription, error) {
	const q = `
SELECT profile_id, plan_id, started_at, expires_at, granted_by, updated_at
FROM user_subscriptions
WHERE profile_id = $1
LIMIT 1`

	var s repoModels.UserSubscription
	qry := r.getQueryable(tx)
	err := qry.QueryRow(ctx, q, profileID).Scan(
		&s.ProfileID, &s.PlanID, &s.StartedAt, &s.ExpiresAt, &s.GrantedBy, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return s, pgx.ErrNoRows
		}
		return s, err
	}
	return s, nil
}

// GetUserSubscriptionByTelegramID resolves a profile by telegram id and returns its subscription.
func (r *Repository) GetUserSubscriptionByTelegramID(ctx context.Context, tx pgx.Tx, telegramID string) (repoModels.UserSubscription, error) {
	const q = `
SELECT us.profile_id, us.plan_id, us.started_at, us.expires_at, us.granted_by, us.updated_at
FROM user_subscriptions us
JOIN profiles p ON p.id = us.profile_id
WHERE p.telegram_id = $1
LIMIT 1`

	var s repoModels.UserSubscription
	qry := r.getQueryable(tx)
	err := qry.QueryRow(ctx, q, telegramID).Scan(
		&s.ProfileID, &s.PlanID, &s.StartedAt, &s.ExpiresAt, &s.GrantedBy, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return s, pgx.ErrNoRows
		}
		return s, err
	}
	return s, nil
}

// UpsertUserSubscription creates or updates a profile's subscription.
func (r *Repository) UpsertUserSubscription(ctx context.Context, tx pgx.Tx, profileID int64, planID string, startedAt time.Time, expiresAt *time.Time, grantedBy *int64) error {
	const q = `
INSERT INTO user_subscriptions(profile_id, plan_id, started_at, expires_at, granted_by, updated_at)
VALUES($1, $2, $3, $4, $5, NOW())
ON CONFLICT (profile_id) DO UPDATE
SET plan_id = EXCLUDED.plan_id,
    started_at = EXCLUDED.started_at,
    expires_at = EXCLUDED.expires_at,
    granted_by = EXCLUDED.granted_by,
    updated_at = NOW()`

	qry := r.getQueryable(tx)
	var exists bool
	if err := qry.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM profiles WHERE id = $1)`, profileID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return pgx.ErrNoRows
	}
	_, err := qry.Exec(ctx, q, profileID, planID, startedAt, expiresAt, grantedBy)
	return err
}

// DeleteUserSubscription removes a profile's subscription row (reverts to free).
func (r *Repository) DeleteUserSubscription(ctx context.Context, tx pgx.Tx, profileID int64) error {
	qry := r.getQueryable(tx)
	var exists bool
	if err := qry.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM profiles WHERE id = $1)`, profileID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return pgx.ErrNoRows
	}
	_, err := qry.Exec(ctx, `DELETE FROM user_subscriptions WHERE profile_id = $1`, profileID)
	return err
}
