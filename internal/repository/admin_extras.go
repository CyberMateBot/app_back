package repository

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/jackc/pgx/v5"
	repoModels "github.com/twelvepills-936/tgapp-/internal/repository/models"
)

func (r *Repository) ListAdminEvents(ctx context.Context, tx pgx.Tx, limit int32) ([]repoModels.AdminEvent, error) {
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	perSource := limit
	if perSource < 10 {
		perSource = 10
	}

	qry := r.getQueryable(tx)
	events := make([]repoModels.AdminEvent, 0, limit)

	tokenRows, err := qry.Query(ctx, `
SELECT
    'token_' || tt.id::text,
    tt.created_at,
    COALESCE(NULLIF(p.name, ''), NULLIF(p.username, ''), p.telegram_id),
    CASE tt.operation WHEN 'credit' THEN 'Начисление токенов' ELSE 'Списание токенов' END,
    (CASE tt.operation WHEN 'credit' THEN '+' ELSE '-' END) || tt.amount::text || ' монет'
        || COALESCE(' · ' || NULLIF(tt.reason, ''), '')
FROM token_transactions tt
JOIN profiles p ON p.id = tt.profile_id
ORDER BY tt.created_at DESC
LIMIT $1`, perSource)
	if err != nil {
		return nil, err
	}
	for tokenRows.Next() {
		var e repoModels.AdminEvent
		if scanErr := tokenRows.Scan(&e.ID, &e.Time, &e.User, &e.Action, &e.Details); scanErr != nil {
			tokenRows.Close()
			return nil, scanErr
		}
		events = append(events, e)
	}
	tokenRows.Close()
	if tokenRows.Err() != nil {
		return nil, tokenRows.Err()
	}

	userRows, err := qry.Query(ctx, `
SELECT
    'user_' || p.id::text,
    p.created_at,
    COALESCE(NULLIF(p.name, ''), NULLIF(p.username, ''), p.telegram_id),
    'Новый пользователь',
    'Telegram ID ' || p.telegram_id
FROM profiles p
ORDER BY p.created_at DESC
LIMIT $1`, perSource)
	if err != nil {
		return nil, err
	}
	for userRows.Next() {
		var e repoModels.AdminEvent
		if scanErr := userRows.Scan(&e.ID, &e.Time, &e.User, &e.Action, &e.Details); scanErr != nil {
			userRows.Close()
			return nil, scanErr
		}
		events = append(events, e)
	}
	userRows.Close()
	if userRows.Err() != nil {
		return nil, userRows.Err()
	}

	sortEventsDesc(events)
	if int32(len(events)) > limit {
		events = events[:limit]
	}
	return events, nil
}

func sortEventsDesc(events []repoModels.AdminEvent) {
	for i := 0; i < len(events); i++ {
		for j := i + 1; j < len(events); j++ {
			if events[j].Time.After(events[i].Time) {
				events[i], events[j] = events[j], events[i]
			}
		}
	}
}

func (r *Repository) GetAdminTransactionStats(ctx context.Context, tx pgx.Tx) (repoModels.AdminTransactionStats, error) {
	const q = `
SELECT
    COALESCE(SUM(amount) FILTER (WHERE operation = 'credit' AND created_at >= date_trunc('month', now())), 0),
    COALESCE(SUM(amount) FILTER (WHERE operation = 'debit' AND created_at >= date_trunc('month', now())), 0),
    COUNT(*) FILTER (WHERE created_at >= date_trunc('month', now())),
    COALESCE(ROUND(AVG(amount) FILTER (WHERE created_at >= date_trunc('month', now()))), 0)
FROM token_transactions`

	var s repoModels.AdminTransactionStats
	qry := r.getQueryable(tx)
	err := qry.QueryRow(ctx, q).Scan(&s.CreditsMonth, &s.DebitsMonth, &s.OperationsMonth, &s.AvgAmount)
	return s, err
}

func (r *Repository) ListAdminTokenTransactions(ctx context.Context, tx pgx.Tx, operation string, limit, offset int32) ([]repoModels.AdminTokenTransaction, int64, error) {
	qry := r.getQueryable(tx)
	operation = strings.TrimSpace(strings.ToLower(operation))

	var total int64
	var rows pgx.Rows
	var err error

	if operation == "credit" || operation == "debit" {
		if err := qry.QueryRow(ctx, `
SELECT COUNT(*) FROM token_transactions WHERE operation = $1`, operation).Scan(&total); err != nil {
			return nil, 0, err
		}
		rows, err = qry.Query(ctx, `
SELECT tt.id,
       COALESCE(NULLIF(p.name, ''), NULLIF(p.username, ''), p.telegram_id),
       tt.operation, tt.amount, COALESCE(tt.reason, ''), tt.created_at
FROM token_transactions tt
JOIN profiles p ON p.id = tt.profile_id
WHERE tt.operation = $1
ORDER BY tt.created_at DESC
LIMIT $2 OFFSET $3`, operation, limit, offset)
	} else {
		if err := qry.QueryRow(ctx, `SELECT COUNT(*) FROM token_transactions`).Scan(&total); err != nil {
			return nil, 0, err
		}
		rows, err = qry.Query(ctx, `
SELECT tt.id,
       COALESCE(NULLIF(p.name, ''), NULLIF(p.username, ''), p.telegram_id),
       tt.operation, tt.amount, COALESCE(tt.reason, ''), tt.created_at
FROM token_transactions tt
JOIN profiles p ON p.id = tt.profile_id
ORDER BY tt.created_at DESC
LIMIT $1 OFFSET $2`, limit, offset)
	}
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]repoModels.AdminTokenTransaction, 0)
	for rows.Next() {
		var item repoModels.AdminTokenTransaction
		if scanErr := rows.Scan(&item.ID, &item.UserName, &item.Operation, &item.Amount, &item.Reason, &item.CreatedAt); scanErr != nil {
			return nil, 0, scanErr
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *Repository) CreateAdminBroadcast(ctx context.Context, tx pgx.Tx, adminID int64, message, target, parseMode string, sent, failed int64) (int64, error) {
	const q = `
INSERT INTO admin_broadcasts(admin_id, message, target, parse_mode, sent_count, failed_count)
VALUES($1, $2, $3, $4, $5, $6)
RETURNING id`

	var id int64
	qry := r.getQueryable(tx)
	err := qry.QueryRow(ctx, q, adminID, message, target, parseMode, sent, failed).Scan(&id)
	return id, err
}

func (r *Repository) ListAdminBroadcasts(ctx context.Context, tx pgx.Tx, limit, offset int32) ([]repoModels.AdminBroadcastRecord, int64, error) {
	qry := r.getQueryable(tx)

	var total int64
	if err := qry.QueryRow(ctx, `SELECT COUNT(*) FROM admin_broadcasts`).Scan(&total); err != nil {
		if isMissingRelation(err) {
			return nil, 0, nil
		}
		return nil, 0, err
	}

	rows, err := qry.Query(ctx, `
SELECT id, message, target, sent_count, failed_count, created_at
FROM admin_broadcasts
ORDER BY created_at DESC
LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]repoModels.AdminBroadcastRecord, 0)
	for rows.Next() {
		var item repoModels.AdminBroadcastRecord
		if scanErr := rows.Scan(&item.ID, &item.Message, &item.Target, &item.SentCount, &item.FailedCount, &item.CreatedAt); scanErr != nil {
			return nil, 0, scanErr
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *Repository) GetAdminSettings(ctx context.Context, tx pgx.Tx) (map[string]json.RawMessage, error) {
	qry := r.getQueryable(tx)
	rows, err := qry.Query(ctx, `SELECT key, value FROM admin_settings`)
	if err != nil {
		if isMissingRelation(err) {
			return map[string]json.RawMessage{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]json.RawMessage)
	for rows.Next() {
		var key string
		var value json.RawMessage
		if scanErr := rows.Scan(&key, &value); scanErr != nil {
			return nil, scanErr
		}
		out[key] = value
	}
	return out, rows.Err()
}

func (r *Repository) UpsertAdminSetting(ctx context.Context, tx pgx.Tx, key string, value any) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	qry := r.getQueryable(tx)
	_, err = qry.Exec(ctx, `
INSERT INTO admin_settings(key, value, updated_at)
VALUES($1, $2::jsonb, now())
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = now()`, key, string(raw))
	return err
}

func (r *Repository) ListModelConfigs(ctx context.Context, tx pgx.Tx) (map[string]repoModels.ModelConfig, error) {
	qry := r.getQueryable(tx)
	rows, err := qry.Query(ctx, `
SELECT model_id, category, name, provider, price_coins, enabled
FROM model_configs`)
	if err != nil {
		if isMissingRelation(err) {
			return map[string]repoModels.ModelConfig{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]repoModels.ModelConfig)
	for rows.Next() {
		var item repoModels.ModelConfig
		if scanErr := rows.Scan(&item.ModelID, &item.Category, &item.Name, &item.Provider, &item.PriceCoins, &item.Enabled); scanErr != nil {
			return nil, scanErr
		}
		out[item.ModelID] = item
	}
	return out, rows.Err()
}

func (r *Repository) UpsertModelConfig(ctx context.Context, tx pgx.Tx, cfg repoModels.ModelConfig) error {
	qry := r.getQueryable(tx)
	_, err := qry.Exec(ctx, `
INSERT INTO model_configs(model_id, category, name, provider, price_coins, enabled, updated_at)
VALUES($1, $2, $3, $4, $5, $6, now())
ON CONFLICT (model_id) DO UPDATE SET
    category = EXCLUDED.category,
    name = EXCLUDED.name,
    provider = EXCLUDED.provider,
    price_coins = EXCLUDED.price_coins,
    enabled = EXCLUDED.enabled,
    updated_at = now()`,
		cfg.ModelID, cfg.Category, cfg.Name, cfg.Provider, cfg.PriceCoins, cfg.Enabled)
	return err
}
