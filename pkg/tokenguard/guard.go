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
	"github.com/twelvepills-936/tgapp-/pkg/billing"
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
// Billing is skipped only when BILLING_DISABLED=true.
func (g *Guard) CheckAccess(ctx context.Context, telegramID, initDataRaw string) error {
	return g.CheckAccessForModel(ctx, telegramID, initDataRaw, "", "")
}

// CheckAccessForModel validates identity and sufficient balance for a model operation.
func (g *Guard) CheckAccessForModel(ctx context.Context, telegramID, initDataRaw, modelID, category string) error {
	if g == nil || g.db == nil {
		return nil
	}
	if err := g.CheckIdentity(ctx, telegramID, initDataRaw); err != nil {
		return err
	}
	if isBillingDisabled() {
		return nil
	}

	price := int64(g.resolveModelPrice(ctx, modelID, category))
	if price <= 0 {
		return g.checkBalance(ctx, telegramID)
	}

	return g.checkBalanceAtLeast(ctx, telegramID, price)
}

// ChargeForGeneration debits CyberCoins after a successful generation.
func (g *Guard) ChargeForGeneration(ctx context.Context, telegramID, modelID, category string) error {
	if g == nil || g.db == nil || isBillingDisabled() {
		return nil
	}

	telegramID = strings.TrimSpace(telegramID)
	if telegramID == "" {
		return ErrTelegramIDRequired
	}

	price := int64(g.resolveModelPrice(ctx, modelID, category))
	if price <= 0 {
		return nil
	}

	return g.debitProfileByTelegram(ctx, telegramID, price, fmt.Sprintf("generation:%s", strings.TrimSpace(modelID)))
}

func (g *Guard) resolveModelPrice(ctx context.Context, modelID, category string) int {
	modelID = strings.TrimSpace(modelID)
	category = strings.TrimSpace(category)
	if modelID == "" {
		return billing.DefaultModelPrice(modelID, category)
	}

	var price int
	err := g.db.QueryRow(ctx, `
SELECT price_coins
FROM model_configs
WHERE model_id = $1
LIMIT 1`, modelID).Scan(&price)
	if err == nil && price > 0 {
		return price
	}

	return billing.DefaultModelPrice(modelID, category)
}

func (g *Guard) checkBalanceAtLeast(ctx context.Context, telegramID string, amount int64) error {
	var balanceAvailable int64
	err := g.db.QueryRow(ctx, `
SELECT COALESCE((SELECT SUM(w.balance_available) FROM wallets w JOIN profiles p ON p.id = w.profile_id WHERE p.telegram_id = $1), 0)`,
		telegramID).Scan(&balanceAvailable)
	if err != nil {
		return fmt.Errorf("check tokens: %w", err)
	}
	if balanceAvailable < amount {
		return ErrInsufficientTokens
	}
	return nil
}

func (g *Guard) debitProfileByTelegram(ctx context.Context, telegramID string, amount int64, reason string) error {
	if amount <= 0 {
		return nil
	}

	tx, err := g.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin debit tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var profileID int64
	err = tx.QueryRow(ctx, `SELECT id FROM profiles WHERE telegram_id = $1 LIMIT 1`, telegramID).Scan(&profileID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrProfileNotFound
		}
		return fmt.Errorf("resolve profile: %w", err)
	}

	var balanceAfter int64
	err = tx.QueryRow(ctx, `
WITH locked AS (
  SELECT id FROM wallets WHERE profile_id = $1 ORDER BY id LIMIT 1 FOR UPDATE
)
UPDATE wallets w
SET balance_available = w.balance_available - $2,
    balance = w.balance - $2
FROM locked
WHERE w.id = locked.id
  AND w.balance_available >= $2
RETURNING w.balance_available`, profileID, amount).Scan(&balanceAfter)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInsufficientTokens
		}
		return fmt.Errorf("debit wallet: %w", err)
	}

	_, err = tx.Exec(ctx, `
INSERT INTO token_transactions(profile_id, admin_id, operation, amount, balance_after, reason)
VALUES($1, NULL, 'debit', $2, $3, $4)`,
		profileID, amount, balanceAfter, strings.TrimSpace(reason))
	if err != nil {
		return fmt.Errorf("log token transaction: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit debit tx: %w", err)
	}

	return nil
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
		if err != nil && isDevelopmentEnv() && telegramauth.InitDataMissingHash(initDataRaw) {
			verifiedID, err = telegramauth.ExtractUserID(initDataRaw)
		}
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

func isDevelopmentEnv() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("ENVIRONMENT"))) {
	case "", "development", "dev", "local":
		return true
	default:
		return false
	}
}

func isBillingDisabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("BILLING_DISABLED"))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func writeJSON(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
