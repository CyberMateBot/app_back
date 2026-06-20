package prompthistory

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrProfileNotFound = errors.New("profile not found")

// Store persists prompt history in PostgreSQL (profile_id FK to profiles).
type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

func (s *Store) resolveProfileID(ctx context.Context, telegramID string) (int64, error) {
	telegramID = strings.TrimSpace(telegramID)
	if telegramID == "" {
		return 0, ErrProfileNotFound
	}

	var profileID int64
	err := s.db.QueryRow(ctx, `SELECT id FROM profiles WHERE telegram_id = $1 LIMIT 1`, telegramID).Scan(&profileID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrProfileNotFound
		}
		return 0, fmt.Errorf("resolve profile: %w", err)
	}
	return profileID, nil
}

func (s *Store) Insert(ctx context.Context, in saveRequest) (Item, error) {
	profileID, err := s.resolveProfileID(ctx, in.TelegramID)
	if err != nil {
		return Item{}, err
	}

	category := strings.TrimSpace(in.Category)
	if category == "" {
		category = "general"
	}

	const q = `
INSERT INTO prompt_history (profile_id, prompt, response, category, model, session_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, prompt, COALESCE(response, ''), category,
          COALESCE(model, ''), COALESCE(session_id, ''), created_at`

	var item Item
	err = s.db.QueryRow(ctx, q,
		profileID,
		strings.TrimSpace(in.Prompt),
		strings.TrimSpace(in.Response),
		category,
		strings.TrimSpace(in.Model),
		strings.TrimSpace(in.SessionID),
	).Scan(
		&item.ID, &item.Prompt, &item.Response,
		&item.Category, &item.Model, &item.SessionID, &item.CreatedAt,
	)
	if err != nil {
		return Item{}, fmt.Errorf("insert prompt history: %w", err)
	}
	item.TelegramID = strings.TrimSpace(in.TelegramID)
	item.CreatedAt = item.CreatedAt.UTC()
	return item, nil
}

func (s *Store) ListByTelegram(ctx context.Context, telegramID string, limit int) ([]Item, error) {
	if limit <= 0 {
		limit = 200
	}

	const q = `
SELECT h.id, p.telegram_id, h.prompt, COALESCE(h.response, ''), h.category,
       COALESCE(h.model, ''), COALESCE(h.session_id, ''), h.created_at
FROM prompt_history h
JOIN profiles p ON p.id = h.profile_id
WHERE p.telegram_id = $1
ORDER BY h.created_at DESC
LIMIT $2`

	rows, err := s.db.Query(ctx, q, strings.TrimSpace(telegramID), limit)
	if err != nil {
		return nil, fmt.Errorf("list prompt history: %w", err)
	}
	defer rows.Close()

	items := make([]Item, 0)
	for rows.Next() {
		var item Item
		if err := rows.Scan(
			&item.ID, &item.TelegramID, &item.Prompt, &item.Response,
			&item.Category, &item.Model, &item.SessionID, &item.CreatedAt,
		); err != nil {
			return nil, err
		}
		item.CreatedAt = item.CreatedAt.UTC()
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) DeleteByTelegram(ctx context.Context, telegramID string) error {
	profileID, err := s.resolveProfileID(ctx, telegramID)
	if err != nil {
		if errors.Is(err, ErrProfileNotFound) {
			return nil
		}
		return err
	}

	_, err = s.db.Exec(ctx, `DELETE FROM prompt_history WHERE profile_id = $1`, profileID)
	if err != nil {
		return fmt.Errorf("delete prompt history: %w", err)
	}
	return nil
}

func (s *Store) DeleteTopic(ctx context.Context, telegramID, sessionKey string, ids []int64) error {
	profileID, err := s.resolveProfileID(ctx, telegramID)
	if err != nil {
		return err
	}

	sessionKey = strings.TrimSpace(sessionKey)
	if sessionKey != "" {
		if strings.HasPrefix(sessionKey, "legacy-") {
			rawID := strings.TrimPrefix(sessionKey, "legacy-")
			var itemID int64
			if _, parseErr := fmt.Sscan(rawID, &itemID); parseErr == nil && itemID > 0 {
				_, err = s.db.Exec(ctx, `DELETE FROM prompt_history WHERE profile_id = $1 AND id = $2`, profileID, itemID)
				if err != nil {
					return fmt.Errorf("delete legacy prompt history item: %w", err)
				}
				return nil
			}
		}

		_, err = s.db.Exec(ctx, `DELETE FROM prompt_history WHERE profile_id = $1 AND session_id = $2`, profileID, sessionKey)
		if err != nil {
			return fmt.Errorf("delete prompt history session: %w", err)
		}
		return nil
	}

	if len(ids) == 0 {
		return nil
	}

	_, err = s.db.Exec(ctx, `DELETE FROM prompt_history WHERE profile_id = $1 AND id = ANY($2::bigint[])`, profileID, ids)
	if err != nil {
		return fmt.Errorf("delete prompt history items: %w", err)
	}
	return nil
}

// SaveAfterGenerate stores a generation result when telegramId is present.
func (s *Store) SaveAfterGenerate(ctx context.Context, telegramID, prompt, response, category, model, sessionID string) (*Item, error) {
	telegramID = strings.TrimSpace(telegramID)
	prompt = strings.TrimSpace(prompt)
	if telegramID == "" || prompt == "" {
		return nil, nil
	}

	item, err := s.Insert(ctx, saveRequest{
		TelegramID: telegramID,
		Prompt:     prompt,
		Response:   response,
		Category:   category,
		Model:      model,
		SessionID:  sessionID,
	})
	if err != nil {
		if errors.Is(err, ErrProfileNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}
