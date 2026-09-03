package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	repoModels "github.com/twelvepills-936/tgapp-/internal/repository/models"
)

// ListReferralsByReferrerProfileID returns referrals invited by the given profile.
func (r *Repository) ListReferralsByReferrerProfileID(ctx context.Context, tx pgx.Tx, referrerProfileID int64) ([]repoModels.Referral, error) {
	const q = `
SELECT r.id, r.referee_profile_id, p.telegram_id, p.name,
       COALESCE(p.username, ''), COALESCE(p.avatar, ''),
       r.completed_tasks_count, r.earnings
FROM referrals r
JOIN profiles p ON p.id = r.referee_profile_id
WHERE r.referrer_profile_id = $1
ORDER BY r.id DESC`

	qry := r.getQueryable(tx)
	rows, err := qry.Query(ctx, q, referrerProfileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]repoModels.Referral, 0)
	for rows.Next() {
		var item repoModels.Referral
		if scanErr := rows.Scan(
			&item.ID,
			&item.RefereeProfileID,
			&item.RefereeTelegramID,
			&item.RefereeName,
			&item.RefereeUsername,
			&item.RefereeAvatar,
			&item.CompletedTasksCount,
			&item.Earnings,
		); scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// LockPendingReferralByReferee looks up the referral row that credits this
// referee, IF the referrer hasn't already been paid out for them. The row is
// locked FOR UPDATE so a concurrent second call (e.g. two overlapping
// subscription webhooks) can't double-pay the referrer. `pgx.ErrNoRows` is
// returned when the referee has no unpaid referral (either no inviter or the
// bonus was already granted).
func (r *Repository) LockPendingReferralByReferee(ctx context.Context, tx pgx.Tx, refereeProfileID int64) (int64, int64, error) {
	const q = `
SELECT id, referrer_profile_id
FROM referrals
WHERE referee_profile_id = $1 AND completed_tasks_count = 0
LIMIT 1
FOR UPDATE`

	qry := r.getQueryable(tx)
	var referralID, referrerProfileID int64
	if err := qry.QueryRow(ctx, q, refereeProfileID).Scan(&referralID, &referrerProfileID); err != nil {
		return 0, 0, err
	}
	return referralID, referrerProfileID, nil
}

// MarkReferralCompleted flips a referral's completed flag and records how
// much the referrer earned from it. Idempotent by design of the callsite —
// LockPendingReferralByReferee only returns rows that are still pending.
func (r *Repository) MarkReferralCompleted(ctx context.Context, tx pgx.Tx, referralID int64, earnings int64) error {
	const q = `UPDATE referrals SET completed_tasks_count = 1, earnings = $2 WHERE id = $1`
	qry := r.getQueryable(tx)
	_, err := qry.Exec(ctx, q, referralID, earnings)
	return err
}
