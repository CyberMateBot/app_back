package walletapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/twelvepills-936/tgapp-/pkg/tokenguard"
)

const walletPathPrefix = "/v1/wallet/telegram/"

type walletDTO struct {
	Balance          int64 `json:"balance"`
	BalanceAvailable int64 `json:"balance_available"`
	Tokens           int64 `json:"tokens"`
	TotalEarned      int64 `json:"total_earned"`
}

type walletTransactionDTO struct {
	ID          int64  `json:"id"`
	Type        string `json:"type"`
	Operation   string `json:"operation"`
	Amount      int64  `json:"amount"`
	Description string `json:"description"`
	Reason      string `json:"reason,omitempty"`
	CreatedAt   string `json:"created_at"`
}

type walletResponse struct {
	Wallet       walletDTO              `json:"wallet"`
	Transactions []walletTransactionDTO `json:"transactions"`
}

// Wrap serves GET /v1/wallet/telegram/{telegramId} with token balances from wallets table.
// tokens gates access so only the owner of telegramId (proven via a valid
// Telegram init-data signature) can read that wallet's balance and
// transaction history — this endpoint used to be fully unauthenticated, so
// anyone could enumerate any user's finances by telegram id.
func Wrap(next http.Handler, db *pgxpool.Pool, tokens *tokenguard.Guard) http.Handler {
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

		if tokens != nil {
			identityErr := tokens.CheckIdentity(r.Context(), telegramID, tokenguard.InitDataFromRequest(r, ""))
			if tokenguard.WriteHTTPError(w, r, identityErr) {
				return
			}
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
	var profileID int64
	var balance int64
	var balanceAvailable int64
	var totalEarned int64

	err := db.QueryRow(ctx, `
SELECT
	p.id,
	COALESCE(w.balance, 0),
	COALESCE(w.balance_available, 0),
	COALESCE(w.total_earned, 0)
FROM profiles p
LEFT JOIN wallets w ON w.profile_id = p.id
WHERE p.telegram_id = $1
ORDER BY w.id NULLS LAST
LIMIT 1`, telegramID).Scan(&profileID, &balance, &balanceAvailable, &totalEarned)
	if err != nil {
		return walletResponse{}, err
	}

	transactions, err := loadWalletTransactions(ctx, db, profileID)
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
		Transactions: transactions,
	}, nil
}

func loadWalletTransactions(ctx context.Context, db *pgxpool.Pool, profileID int64) ([]walletTransactionDTO, error) {
	rows, err := db.Query(ctx, `
SELECT id, operation, amount, COALESCE(reason, ''), created_at
FROM token_transactions
WHERE profile_id = $1
ORDER BY created_at DESC
LIMIT 50`, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]walletTransactionDTO, 0)
	for rows.Next() {
		var item walletTransactionDTO
		var operation string
		var amount int64
		var reason string
		var createdAt any

		if scanErr := rows.Scan(&item.ID, &operation, &amount, &reason, &createdAt); scanErr != nil {
			return nil, scanErr
		}

		item.Operation = operation
		item.Type = operation
		item.Reason = strings.TrimSpace(reason)
		item.Description = item.Reason
		if createdAt != nil {
			item.CreatedAt = formatTimestamp(createdAt)
		}

		switch strings.ToLower(strings.TrimSpace(operation)) {
		case "debit":
			item.Amount = -amount
		default:
			item.Amount = amount
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if items == nil {
		items = []walletTransactionDTO{}
	}

	return items, nil
}

func formatTimestamp(value any) string {
	switch v := value.(type) {
	case time.Time:
		return v.UTC().Format(time.RFC3339)
	case *time.Time:
		if v != nil {
			return v.UTC().Format(time.RFC3339)
		}
	case string:
		return v
	}
	return ""
}
