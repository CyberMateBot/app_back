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
SELECT id, telegram_id, name, COALESCE(username, ''), is_active, created_at
FROM profiles
WHERE (
	COALESCE(username, '') ILIKE $1
	OR name ILIKE $1
	OR telegram_id ILIKE $1
)
ORDER BY created_at DESC
LIMIT $2 OFFSET $3`
		rows, err = qry.Query(ctx, listQ, pattern, limit, offset)
	} else {
		const countQ = `SELECT COUNT(*) FROM profiles`
		if err := qry.QueryRow(ctx, countQ).Scan(&total); err != nil {
			return nil, 0, err
		}
		const listQ = `
SELECT id, telegram_id, name, COALESCE(username, ''), is_active, created_at
FROM profiles
ORDER BY created_at DESC
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
		if scanErr := rows.Scan(&p.ID, &p.TelegramID, &p.Name, &p.Username, &p.IsActive, &p.CreatedAt); scanErr != nil {
			return nil, 0, scanErr
		}
		items = append(items, p)
	}
	return items, total, rows.Err()
}

func (r *Repository) GetAdminProfileByID(ctx context.Context, tx pgx.Tx, id int64) (repoModels.AdminProfile, error) {
	const q = `
SELECT id, telegram_id, name, COALESCE(username, ''), is_active, created_at
FROM profiles WHERE id = $1 LIMIT 1`

	var p repoModels.AdminProfile
	qry := r.getQueryable(tx)
	err := qry.QueryRow(ctx, q, id).Scan(&p.ID, &p.TelegramID, &p.Name, &p.Username, &p.IsActive, &p.CreatedAt)
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
