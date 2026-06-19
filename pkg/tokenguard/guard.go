package tokenguard

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/twelvepills-936/tgapp-/pkg/telegramauth"
)

var (
	ErrTelegramIDRequired = errors.New("telegramId is required")
	ErrInitDataRequired   = errors.New("initDataRaw is required")
	ErrTelegramIDMismatch = errors.New("telegramId does not match init data")
	ErrProfileNotFound    = errors.New("profile not found")
	ErrProfileInactive    = errors.New("profile is inactive")
	ErrInsufficientTokens = errors.New("insufficient tokens")
)

// Guard blocks AI chat usage when the profile wallet has no available tokens.
type Guard struct {
	db       *pgxpool.Pool
	botToken string
}

func New(db *pgxpool.Pool) *Guard {
	if db == nil {
		return nil
	}
	return &Guard{
		db:       db,
		botToken: strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN")),
	}
}

// CheckAccess validates identity and positive token balance (AI generation).
func (g *Guard) CheckAccess(ctx context.Context, telegramID, initDataRaw string) error {
	if g == nil || g.db == nil {
		return nil
	}
	if err := g.CheckIdentity(ctx, telegramID, initDataRaw); err != nil {
		return err
	}
	return g.checkBalance(ctx, telegramID)
}

// CheckIdentity validates telegram identity without requiring token balance (history read/delete).
func (g *Guard) CheckIdentity(ctx context.Context, telegramID, initDataRaw string) error {
	telegramID = strings.TrimSpace(telegramID)
	if telegramID == "" {
		return ErrTelegramIDRequired
	}
	if g == nil || g.db == nil {
		return nil
	}

	if g.botToken != "" {
		initDataRaw = strings.TrimSpace(initDataRaw)
		if initDataRaw == "" {
			return ErrInitDataRequired
		}
		verifiedID, err := telegramauth.VerifyInitData(initDataRaw, g.botToken)
		if err != nil {
			return err
		}
		if verifiedID != telegramID {
			return ErrTelegramIDMismatch
		}
	}

	var isActive bool
	err := g.db.QueryRow(ctx, `
SELECT COALESCE(p.is_active, TRUE)
FROM profiles p
WHERE p.telegram_id = $1
LIMIT 1`, telegramID).Scan(&isActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrProfileNotFound
		}
		return fmt.Errorf("check profile: %w", err)
	}
	if !isActive {
		return ErrProfileInactive
	}
	return nil
}

func (g *Guard) checkBalance(ctx context.Context, telegramID string) error {
	var balanceAvailable int64
	err := g.db.QueryRow(ctx, `
SELECT COALESCE((SELECT SUM(w.balance_available) FROM wallets w JOIN profiles p ON p.id = w.profile_id WHERE p.telegram_id = $1), 0)`,
		telegramID).Scan(&balanceAvailable)
	if err != nil {
		return fmt.Errorf("check tokens: %w", err)
	}
	if balanceAvailable <= 0 {
		return ErrInsufficientTokens
	}
	return nil
}

// InitDataFromRequest reads init data from JSON body field or X-Telegram-Init-Data header.
func InitDataFromRequest(r *http.Request, bodyValue string) string {
	if v := strings.TrimSpace(bodyValue); v != "" {
		return v
	}
	return strings.TrimSpace(r.Header.Get("X-Telegram-Init-Data"))
}

// WriteHTTPError maps guard errors to HTTP responses. Returns true when a response was written.
func WriteHTTPError(w http.ResponseWriter, r *http.Request, err error) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, ErrTelegramIDRequired):
		writeJSON(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, ErrInitDataRequired):
		writeJSON(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, ErrTelegramIDMismatch):
		writeJSON(w, http.StatusForbidden, err.Error())
	case errors.Is(err, telegramauth.ErrInitDataInvalid), errors.Is(err, telegramauth.ErrInitDataExpired):
		writeJSON(w, http.StatusUnauthorized, "invalid telegram init data")
	case errors.Is(err, ErrProfileNotFound):
		writeJSON(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrProfileInactive):
		writeJSON(w, http.StatusForbidden, err.Error())
	case errors.Is(err, ErrInsufficientTokens):
		writeJSON(w, http.StatusPaymentRequired, err.Error())
	default:
		slog.ErrorContext(r.Context(), "token check failed", slog.Any("error", err))
		writeJSON(w, http.StatusInternalServerError, "failed to verify account tokens")
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
