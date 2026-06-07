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
