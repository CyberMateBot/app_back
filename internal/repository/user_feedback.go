package repository

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5"
	repoModels "github.com/twelvepills-936/tgapp-/internal/repository/models"
)

func (r *Repository) CreateUserFeedback(ctx context.Context, tx pgx.Tx, profileID int64, kind, message string) (int64, error) {
	const q = `
INSERT INTO user_feedback (profile_id, kind, message)
VALUES ($1, $2, $3)
RETURNING id`

	var id int64
	qry := r.getQueryable(tx)
	err := qry.QueryRow(ctx, q, profileID, kind, message).Scan(&id)
	return id, err
}

func (r *Repository) ListAdminUserFeedback(ctx context.Context, tx pgx.Tx, kind string, limit, offset int32) ([]repoModels.UserFeedback, int64, error) {
	qry := r.getQueryable(tx)
	kind = strings.TrimSpace(strings.ToLower(kind))

	var total int64
	var rows pgx.Rows
	var err error

	if kind == repoModels.UserFeedbackKindSuggestion || kind == repoModels.UserFeedbackKindBug {
		if err := qry.QueryRow(ctx, `SELECT COUNT(*) FROM user_feedback WHERE kind = $1`, kind).Scan(&total); err != nil {
			if isMissingRelation(err) {
				return nil, 0, nil
			}
			return nil, 0, err
		}
		rows, err = qry.Query(ctx, `
SELECT uf.id, uf.profile_id, uf.kind, uf.message, uf.created_at,
       COALESCE(NULLIF(p.name, ''), NULLIF(p.username, ''), p.telegram_id)
FROM user_feedback uf
JOIN profiles p ON p.id = uf.profile_id
WHERE uf.kind = $1
ORDER BY uf.created_at DESC
LIMIT $2 OFFSET $3`, kind, limit, offset)
	} else {
		if err := qry.QueryRow(ctx, `SELECT COUNT(*) FROM user_feedback`).Scan(&total); err != nil {
			if isMissingRelation(err) {
				return nil, 0, nil
			}
			return nil, 0, err
		}
		rows, err = qry.Query(ctx, `
SELECT uf.id, uf.profile_id, uf.kind, uf.message, uf.created_at,
       COALESCE(NULLIF(p.name, ''), NULLIF(p.username, ''), p.telegram_id)
FROM user_feedback uf
JOIN profiles p ON p.id = uf.profile_id
ORDER BY uf.created_at DESC
LIMIT $1 OFFSET $2`, limit, offset)
	}
	if err != nil {
		if isMissingRelation(err) {
			return nil, 0, nil
		}
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]repoModels.UserFeedback, 0)
	for rows.Next() {
		var item repoModels.UserFeedback
		if scanErr := rows.Scan(&item.ID, &item.ProfileID, &item.Kind, &item.Message, &item.CreatedAt, &item.UserName); scanErr != nil {
			return nil, 0, scanErr
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *Repository) DeleteAdminUserFeedback(ctx context.Context, tx pgx.Tx, id int64) error {
	qry := r.getQueryable(tx)
	_, err := qry.Exec(ctx, `DELETE FROM user_feedback WHERE id = $1`, id)
	return err
}
