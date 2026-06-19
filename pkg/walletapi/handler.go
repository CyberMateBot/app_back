package walletapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const walletPathPrefix = "/v1/wallet/telegram/"

type walletDTO struct {
	Balance          int64 `json:"balance"`
	BalanceAvailable int64 `json:"balance_available"`
	Tokens           int64 `json:"tokens"`
	TotalEarned      int64 `json:"total_earned"`
}

type walletResponse struct {
	Wallet       walletDTO `json:"wallet"`
	Transactions []any     `json:"transactions"`
}

// Wrap serves GET /v1/wallet/telegram/{telegramId} with token balances from wallets table.
func Wrap(next http.Handler, db *pgxpool.Pool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || !strings.HasPrefix(r.URL.Path, walletPathPrefix) {
			next.ServeHTTP(w, r)
			return
		}

		telegramID := strings.Trim(strings.TrimPrefix(r.URL.Path, walletPathPrefix), "/")
		if telegramID == "" {
			http.Error(w, "telegram_id required", http.StatusBadRequest)
			return
		}

		if db == nil {
			http.Error(w, "wallet api is not configured", http.StatusServiceUnavailable)
			return
		}

		payload, err := loadWallet(r.Context(), db, telegramID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				http.Error(w, "profile not found", http.StatusNotFound)
				return
			}
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(payload)
	})
}

func loadWallet(ctx context.Context, db *pgxpool.Pool, telegramID string) (walletResponse, error) {
	var balance int64
	var balanceAvailable int64
	var totalEarned int64

	err := db.QueryRow(ctx, `
SELECT
	COALESCE(w.balance, 0),
	COALESCE(w.balance_available, 0),
	COALESCE(w.total_earned, 0)
FROM profiles p
LEFT JOIN wallets w ON w.profile_id = p.id
WHERE p.telegram_id = $1
ORDER BY w.id NULLS LAST
LIMIT 1`, telegramID).Scan(&balance, &balanceAvailable, &totalEarned)
	if err != nil {
		return walletResponse{}, err
	}

	return walletResponse{
		Wallet: walletDTO{
			Balance:          balance,
			BalanceAvailable: balanceAvailable,
			Tokens:           balanceAvailable,
			TotalEarned:      totalEarned,
		},
		Transactions: []any{},
	}, nil
}
