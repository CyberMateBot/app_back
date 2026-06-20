package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	repoModels "github.com/twelvepills-936/tgapp-/internal/repository/models"
)

func (r *Repository) ListHomeWidgets(ctx context.Context, tx pgx.Tx, activeOnly bool) ([]repoModels.HomeWidget, error) {
	qry := r.getQueryable(tx)
	query := `
SELECT id, sort_order, tag_text, tag_bg, tag_color, title, description,
       background_style, image_url, is_active, created_at, updated_at
FROM home_widgets`
	if activeOnly {
		query += ` WHERE is_active = TRUE`
	}
	query += ` ORDER BY sort_order ASC, id ASC`

	rows, err := qry.Query(ctx, query)
	if err != nil {
		if isMissingRelation(err) {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()

	out := make([]repoModels.HomeWidget, 0)
	for rows.Next() {
		var w repoModels.HomeWidget
		if scanErr := rows.Scan(
			&w.ID, &w.SortOrder, &w.TagText, &w.TagBg, &w.TagColor,
			&w.Title, &w.Description, &w.BackgroundStyle, &w.ImageURL,
			&w.IsActive, &w.CreatedAt, &w.UpdatedAt,
		); scanErr != nil {
			return nil, scanErr
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

func (r *Repository) GetHomeWidgetByID(ctx context.Context, tx pgx.Tx, id int64) (repoModels.HomeWidget, error) {
	qry := r.getQueryable(tx)
	row := qry.QueryRow(ctx, `
SELECT id, sort_order, tag_text, tag_bg, tag_color, title, description,
       background_style, image_url, is_active, created_at, updated_at
FROM home_widgets
WHERE id = $1`, id)

	var w repoModels.HomeWidget
	err := row.Scan(
		&w.ID, &w.SortOrder, &w.TagText, &w.TagBg, &w.TagColor,
		&w.Title, &w.Description, &w.BackgroundStyle, &w.ImageURL,
		&w.IsActive, &w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repoModels.HomeWidget{}, pgx.ErrNoRows
		}
		return repoModels.HomeWidget{}, err
	}
	return w, nil
}

func (r *Repository) CreateHomeWidget(ctx context.Context, tx pgx.Tx, w repoModels.HomeWidget) (int64, error) {
	qry := r.getQueryable(tx)
	var id int64
	err := qry.QueryRow(ctx, `
INSERT INTO home_widgets (
    sort_order, tag_text, tag_bg, tag_color, title, description,
    background_style, image_url, is_active
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id`,
		w.SortOrder, w.TagText, w.TagBg, w.TagColor, w.Title, w.Description,
		w.BackgroundStyle, w.ImageURL, w.IsActive,
	).Scan(&id)
	return id, err
}

func (r *Repository) UpdateHomeWidget(ctx context.Context, tx pgx.Tx, w repoModels.HomeWidget) error {
	qry := r.getQueryable(tx)
	tag, err := qry.Exec(ctx, `
UPDATE home_widgets SET
    sort_order = $2,
    tag_text = $3,
    tag_bg = $4,
    tag_color = $5,
    title = $6,
    description = $7,
    background_style = $8,
    image_url = $9,
    is_active = $10,
    updated_at = NOW()
WHERE id = $1`,
		w.ID, w.SortOrder, w.TagText, w.TagBg, w.TagColor, w.Title, w.Description,
		w.BackgroundStyle, w.ImageURL, w.IsActive,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *Repository) DeleteHomeWidget(ctx context.Context, tx pgx.Tx, id int64) error {
	qry := r.getQueryable(tx)
	tag, err := qry.Exec(ctx, `DELETE FROM home_widgets WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
