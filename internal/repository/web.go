package repository

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	repoModels "github.com/twelvepills-936/tgapp-/internal/repository/models"
)

func (r *Repository) CreateWebAccount(ctx context.Context, tx pgx.Tx, a repoModels.WebAccount) (int64, error) {
	const q = `
INSERT INTO web_accounts(email, password_hash, display_name)
VALUES($1,$2,$3)
RETURNING id`
	var id int64
	qry := r.getQueryable(tx)
	if err := qry.QueryRow(ctx, q, a.Email, a.PasswordHash, a.DisplayName).Scan(&id); err != nil {
		slog.ErrorContext(ctx, "failed to create web account", slog.Any("error", err), slog.String("email", a.Email))
		return 0, err
	}
	return id, nil
}

func (r *Repository) GetWebAccountByEmail(ctx context.Context, tx pgx.Tx, email string) (repoModels.WebAccount, error) {
	const q = `
SELECT id, email, password_hash, display_name, created_at, updated_at
FROM web_accounts
WHERE email = $1
LIMIT 1`
	var a repoModels.WebAccount
	qry := r.getQueryable(tx)
	err := qry.QueryRow(ctx, q, email).Scan(&a.ID, &a.Email, &a.PasswordHash, &a.DisplayName, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return a, pgx.ErrNoRows
		}
		slog.ErrorContext(ctx, "failed to get web account by email", slog.Any("error", err), slog.String("email", email))
		return a, err
	}
	return a, nil
}

func (r *Repository) GetWebAccountByID(ctx context.Context, tx pgx.Tx, id int64) (repoModels.WebAccount, error) {
	const q = `
SELECT id, email, password_hash, display_name, created_at, updated_at
FROM web_accounts
WHERE id = $1
LIMIT 1`
	var a repoModels.WebAccount
	qry := r.getQueryable(tx)
	err := qry.QueryRow(ctx, q, id).Scan(&a.ID, &a.Email, &a.PasswordHash, &a.DisplayName, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return a, pgx.ErrNoRows
		}
		slog.ErrorContext(ctx, "failed to get web account by id", slog.Any("error", err), slog.Int64("web_account_id", id))
		return a, err
	}
	return a, nil
}

func (r *Repository) CreateWebPrompt(ctx context.Context, tx pgx.Tx, p repoModels.WebPrompt) (int64, error) {
	const q = `
INSERT INTO web_prompt_history(web_account_id, prompt, category, model)
VALUES($1,$2,$3,$4)
RETURNING id`
	var id int64
	qry := r.getQueryable(tx)
	if err := qry.QueryRow(ctx, q, p.WebAccountID, p.Prompt, p.Category, p.Model).Scan(&id); err != nil {
		slog.ErrorContext(ctx, "failed to create web prompt", slog.Any("error", err), slog.Int64("web_account_id", p.WebAccountID))
		return 0, err
	}
	return id, nil
}

func (r *Repository) ListWebPrompts(ctx context.Context, tx pgx.Tx, webAccountID int64, limit int32, offset int32) ([]repoModels.WebPrompt, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	const q = `
SELECT id, web_account_id, prompt, category, model, created_at
FROM web_prompt_history
WHERE web_account_id = $1
ORDER BY created_at DESC, id DESC
LIMIT $2 OFFSET $3`

	qry := r.getQueryable(tx)
	rows, err := qry.Query(ctx, q, webAccountID, limit, offset)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list web prompts", slog.Any("error", err), slog.Int64("web_account_id", webAccountID))
		return nil, err
	}
	defer rows.Close()

	out := make([]repoModels.WebPrompt, 0, limit)
	for rows.Next() {
		var p repoModels.WebPrompt
		if scanErr := rows.Scan(&p.ID, &p.WebAccountID, &p.Prompt, &p.Category, &p.Model, &p.CreatedAt); scanErr != nil {
			return nil, scanErr
		}
		out = append(out, p)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return out, nil
}

