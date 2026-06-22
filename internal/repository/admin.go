package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	repoModels "github.com/twelvepills-936/tgapp-/internal/repository/models"
)

func (r *Repository) GetAdminByEmail(ctx context.Context, tx pgx.Tx, email string) (repoModels.Admin, error) {
	const q = `
SELECT id, email, password_hash, role, created_at
FROM admins WHERE email = $1 LIMIT 1`

	var a repoModels.Admin
	qry := r.getQueryable(tx)
	err := qry.QueryRow(ctx, q, strings.ToLower(strings.TrimSpace(email))).Scan(
		&a.ID, &a.Email, &a.PasswordHash, &a.Role, &a.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return a, pgx.ErrNoRows
		}
		return a, err
	}
	return a, nil
}

func (r *Repository) GetAdminByID(ctx context.Context, tx pgx.Tx, id int64) (repoModels.Admin, error) {
	const q = `
SELECT id, email, password_hash, role, created_at
FROM admins WHERE id = $1 LIMIT 1`

	var a repoModels.Admin
	qry := r.getQueryable(tx)
	err := qry.QueryRow(ctx, q, id).Scan(
		&a.ID, &a.Email, &a.PasswordHash, &a.Role, &a.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return a, pgx.ErrNoRows
		}
		return a, err
	}
	return a, nil
}

func (r *Repository) CreateAdmin(ctx context.Context, tx pgx.Tx, email, passwordHash string) (int64, error) {
	const q = `
INSERT INTO admins(email, password_hash, role)
VALUES($1, $2, 'admin')
RETURNING id`

	var id int64
	qry := r.getQueryable(tx)
	err := qry.QueryRow(ctx, q, strings.ToLower(strings.TrimSpace(email)), passwordHash).Scan(&id)
	return id, err
}

func (r *Repository) CountAdmins(ctx context.Context, tx pgx.Tx) (int64, error) {
	const q = `SELECT COUNT(*) FROM admins`
	var n int64
	qry := r.getQueryable(tx)
	err := qry.QueryRow(ctx, q).Scan(&n)
	return n, err
}

func (r *Repository) GetAdminStats(ctx context.Context, tx pgx.Tx) (repoModels.AdminStats, error) {
	const q = `
SELECT
    (SELECT COUNT(*) FROM profiles),
    (SELECT COUNT(DISTINCT profile_id) FROM prompt_history WHERE created_at >= CURRENT_DATE),
    (SELECT COUNT(*) FROM profiles WHERE created_at >= CURRENT_DATE),
    (SELECT COUNT(*) FROM prompt_history)`

	var s repoModels.AdminStats
	qry := r.getQueryable(tx)
	err := qry.QueryRow(ctx, q).Scan(&s.TotalUsers, &s.ActiveUsersToday, &s.NewUsersToday, &s.TotalMessages)
	return s, err
}

func (r *Repository) ListAdminProfiles(ctx context.Context, tx pgx.Tx, search string, limit, offset int32) ([]repoModels.AdminProfile, int64, error) {
	search = strings.TrimSpace(search)
	qry := r.getQueryable(tx)

	var total int64
	var rows pgx.Rows
	var err error

	if search != "" {
		pattern := "%" + search + "%"
		const countQ = `SELECT COUNT(*) FROM profiles WHERE (
			COALESCE(username, '') ILIKE $1
			OR name ILIKE $1
			OR telegram_id ILIKE $1
		)`
		if err := qry.QueryRow(ctx, countQ, pattern).Scan(&total); err != nil {
			return nil, 0, err
		}
		const listQ = `
SELECT p.id, p.telegram_id, p.name, COALESCE(p.username, ''), p.is_active, COALESCE(w.balance_available, 0), p.created_at,
       COALESCE(us.plan_id, ''), us.started_at, us.expires_at
FROM profiles p
LEFT JOIN wallets w ON w.profile_id = p.id
LEFT JOIN user_subscriptions us ON us.profile_id = p.id
WHERE (
	COALESCE(p.username, '') ILIKE $1
	OR p.name ILIKE $1
	OR p.telegram_id ILIKE $1
)
ORDER BY p.created_at DESC
LIMIT $2 OFFSET $3`
		rows, err = qry.Query(ctx, listQ, pattern, limit, offset)
	} else {
		const countQ = `SELECT COUNT(*) FROM profiles`
		if err := qry.QueryRow(ctx, countQ).Scan(&total); err != nil {
			return nil, 0, err
		}
		const listQ = `
SELECT p.id, p.telegram_id, p.name, COALESCE(p.username, ''), p.is_active, COALESCE(w.balance_available, 0), p.created_at,
       COALESCE(us.plan_id, ''), us.started_at, us.expires_at
FROM profiles p
LEFT JOIN wallets w ON w.profile_id = p.id
LEFT JOIN user_subscriptions us ON us.profile_id = p.id
ORDER BY p.created_at DESC
LIMIT $1 OFFSET $2`
		rows, err = qry.Query(ctx, listQ, limit, offset)
	}
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]repoModels.AdminProfile, 0)
	for rows.Next() {
		var p repoModels.AdminProfile
		if scanErr := rows.Scan(&p.ID, &p.TelegramID, &p.Name, &p.Username, &p.IsActive, &p.Tokens, &p.CreatedAt,
			&p.SubscriptionPlanID, &p.SubscriptionStarted, &p.SubscriptionExpires); scanErr != nil {
			return nil, 0, scanErr
		}
		items = append(items, p)
	}
	return items, total, rows.Err()
}

func (r *Repository) GetAdminProfileByID(ctx context.Context, tx pgx.Tx, id int64) (repoModels.AdminProfile, error) {
	const q = `
SELECT p.id, p.telegram_id, p.name, COALESCE(p.username, ''), p.is_active, COALESCE(w.balance_available, 0), p.created_at,
       COALESCE(us.plan_id, ''), us.started_at, us.expires_at
FROM profiles p
LEFT JOIN wallets w ON w.profile_id = p.id
LEFT JOIN user_subscriptions us ON us.profile_id = p.id
WHERE p.id = $1 LIMIT 1`

	var p repoModels.AdminProfile
	qry := r.getQueryable(tx)
	err := qry.QueryRow(ctx, q, id).Scan(&p.ID, &p.TelegramID, &p.Name, &p.Username, &p.IsActive, &p.Tokens, &p.CreatedAt,
		&p.SubscriptionPlanID, &p.SubscriptionStarted, &p.SubscriptionExpires)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return p, pgx.ErrNoRows
		}
		return p, err
	}
	return p, nil
}

func (r *Repository) UpdateProfileActive(ctx context.Context, tx pgx.Tx, id int64, isActive bool) error {
	const q = `UPDATE profiles SET is_active = $2, updated_at = now() WHERE id = $1`
	qry := r.getQueryable(tx)
	tag, err := qry.Exec(ctx, q, id, isActive)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *Repository) DeleteProfile(ctx context.Context, tx pgx.Tx, id int64) error {
	const q = `DELETE FROM profiles WHERE id = $1`
	qry := r.getQueryable(tx)
	tag, err := qry.Exec(ctx, q, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *Repository) ListBroadcastTelegramIDs(ctx context.Context, tx pgx.Tx, activeOnly bool) ([]string, error) {
	var q string
	if activeOnly {
		q = `
SELECT DISTINCT p.telegram_id
FROM profiles p
JOIN prompt_history h ON h.profile_id = p.id
WHERE p.is_active = TRUE
  AND h.created_at >= now() - interval '7 days'`
	} else {
		q = `SELECT telegram_id FROM profiles WHERE is_active = TRUE`
	}

	qry := r.getQueryable(tx)
	rows, err := qry.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]string, 0)
	for rows.Next() {
		var id string
		if scanErr := rows.Scan(&id); scanErr != nil {
			return nil, scanErr
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *Repository) CreditProfileTokens(ctx context.Context, tx pgx.Tx, profileID, adminID int64, amount int64, reason string) (repoModels.TokenOperationResult, error) {
	var out repoModels.TokenOperationResult
	qry := r.getQueryable(tx)

	var exists bool
	if err := qry.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM profiles WHERE id = $1)`, profileID).Scan(&exists); err != nil {
		return out, err
	}
	if !exists {
		return out, pgx.ErrNoRows
	}

	err := qry.QueryRow(ctx, `
WITH locked AS (
  SELECT id FROM wallets WHERE profile_id = $1 ORDER BY id LIMIT 1 FOR UPDATE
)
UPDATE wallets w
SET balance_available = w.balance_available + $2,
    balance = w.balance + $2
FROM locked
WHERE w.id = locked.id
RETURNING w.balance_available`, profileID, amount).Scan(&out.BalanceAfter)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			if _, insErr := qry.Exec(ctx, `
INSERT INTO wallets(profile_id, balance, total_earned, balance_available)
SELECT $1, 0, 0, 0
WHERE NOT EXISTS (SELECT 1 FROM wallets WHERE profile_id = $1)`, profileID); insErr != nil {
				return out, insErr
			}
			retryErr := qry.QueryRow(ctx, `
WITH locked AS (
  SELECT id FROM wallets WHERE profile_id = $1 ORDER BY id LIMIT 1 FOR UPDATE
)
UPDATE wallets w
SET balance_available = w.balance_available + $2,
    balance = w.balance + $2
FROM locked
WHERE w.id = locked.id
RETURNING w.balance_available`, profileID, amount).Scan(&out.BalanceAfter)
			if retryErr != nil {
				return out, retryErr
			}
		} else {
			return out, err
		}
	}

	_, err = qry.Exec(ctx, `
INSERT INTO token_transactions(profile_id, admin_id, operation, amount, balance_after, reason)
VALUES($1, $2, 'credit', $3, $4, $5)`,
		profileID, adminID, amount, out.BalanceAfter, strings.TrimSpace(reason))
	if err != nil {
		return out, err
	}

	out.ProfileID = profileID
	return out, nil
}

func (r *Repository) DebitProfileTokens(ctx context.Context, tx pgx.Tx, profileID, adminID int64, amount int64, reason string) (repoModels.TokenOperationResult, error) {
	var out repoModels.TokenOperationResult
	qry := r.getQueryable(tx)

	err := qry.QueryRow(ctx, `
WITH locked AS (
  SELECT id FROM wallets WHERE profile_id = $1 ORDER BY id LIMIT 1 FOR UPDATE
)
UPDATE wallets w
SET balance_available = w.balance_available - $2,
    balance = w.balance - $2
FROM locked
WHERE w.id = locked.id
  AND w.balance_available >= $2
RETURNING w.balance_available`, profileID, amount).Scan(&out.BalanceAfter)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			var exists bool
			profileErr := qry.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM profiles WHERE id = $1)`, profileID).Scan(&exists)
			if profileErr != nil {
				return out, profileErr
			}
			if !exists {
				return out, pgx.ErrNoRows
			}
			return out, repoModels.ErrInsufficientTokens
		}
		return out, err
	}

	_, err = qry.Exec(ctx, `
INSERT INTO token_transactions(profile_id, admin_id, operation, amount, balance_after, reason)
VALUES($1, $2, 'debit', $3, $4, $5)`,
		profileID, adminID, amount, out.BalanceAfter, strings.TrimSpace(reason))
	if err != nil {
		return out, err
	}

	out.ProfileID = profileID
	return out, nil
}
