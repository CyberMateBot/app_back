package repository

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	repoModels "github.com/twelvepills-936/tgapp-/internal/repository/models"
)

// CreateProfile inserts a new profile and returns its id.
func (r *Repository) CreateProfile(ctx context.Context, tx pgx.Tx, p repoModels.Profile) (int64, error) {
	const q = `
INSERT INTO profiles(name, telegram_id, avatar, location, role, description, telegram_init_data, username, verified)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)
RETURNING id`

	var id int64
	qry := r.getQueryable(tx)
	err := qry.QueryRow(ctx, q, p.Name, p.TelegramID, p.Avatar, p.Location, p.Role, p.Description, p.TelegramInitData, p.Username, p.Verified).Scan(&id)
	return id, err
}

// GetProfileByTelegramID returns profile by telegram id or pgx.ErrNoRows.
func (r *Repository) GetProfileByTelegramID(ctx context.Context, tx pgx.Tx, telegramID string) (repoModels.Profile, error) {
	const q = `
SELECT id, name, telegram_id,
       COALESCE(avatar, ''), COALESCE(location, ''), COALESCE(role, ''),
       COALESCE(description, ''), COALESCE(telegram_init_data, ''), COALESCE(username, ''),
       verified, ui_theme, created_at, updated_at
FROM profiles WHERE telegram_id = $1 LIMIT 1`

	var p repoModels.Profile
	qry := r.getQueryable(tx)
	err := qry.QueryRow(ctx, q, telegramID).Scan(
		&p.ID, &p.Name, &p.TelegramID, &p.Avatar, &p.Location, &p.Role, &p.Description,
		&p.TelegramInitData, &p.Username, &p.Verified, &p.UITheme, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return p, pgx.ErrNoRows
		}
		slog.ErrorContext(ctx, "failed to get profile", slog.Any("error", err), slog.String("telegram_id", telegramID))
		return p, err
	}
	return p, nil
}

// UpdateProfileTheme sets ui_theme for profile identified by telegram_id.
func (r *Repository) UpdateProfileTheme(ctx context.Context, tx pgx.Tx, telegramID, theme string) error {
	const q = `
UPDATE profiles SET ui_theme = $2, updated_at = now()
WHERE telegram_id = $1`

	qry := r.getQueryable(tx)
	tag, err := qry.Exec(ctx, q, telegramID, theme)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update profile theme", slog.Any("error", err), slog.String("telegram_id", telegramID))
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// CreateWalletForUser creates wallet for profile.
func (r *Repository) CreateWalletForUser(ctx context.Context, tx pgx.Tx, profileID int64) (int64, error) {
	const q = `INSERT INTO wallets(profile_id, balance, total_earned, balance_available) VALUES($1,0,0,0) RETURNING id`
	var id int64
	qry := r.getQueryable(tx)
	err := qry.QueryRow(ctx, q, profileID).Scan(&id)
	return id, err
}

// AddReferral links referrer to referee with zeroed stats.
func (r *Repository) AddReferral(ctx context.Context, tx pgx.Tx, referrerProfileID int64, refereeProfileID int64) error {
	const q = `
INSERT INTO referrals(referrer_profile_id, referee_profile_id, completed_tasks_count, earnings)
VALUES($1,$2,0,0)
ON CONFLICT (referrer_profile_id, referee_profile_id) DO NOTHING`
	qry := r.getQueryable(tx)
	_, err := qry.Exec(ctx, q, referrerProfileID, refereeProfileID)
	return err
}
