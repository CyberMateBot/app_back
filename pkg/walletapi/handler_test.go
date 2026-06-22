package walletapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWrap_PassesThroughOtherPaths(t *testing.T) {
	t.Parallel()

	mux := Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}), nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/users/telegram/1", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusTeapot {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestWrap_MissingTelegramID(t *testing.T) {
	t.Parallel()

	mux := Wrap(http.NotFoundHandler(), nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/wallet/telegram/", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestWalletResponseJSON(t *testing.T) {
	t.Parallel()

	body, err := json.Marshal(walletResponse{
		Wallet: walletDTO{
			Balance:          100,
			BalanceAvailable: 100,
			Tokens:           100,
			TotalEarned:      0,
		},
        Transactions: []walletTransactionDTO{},
	})
	if err != nil {
		t.Fatal(err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatal(err)
	}

	wallet, ok := decoded["wallet"].(map[string]any)
	if !ok {
		t.Fatalf("wallet = %#v", decoded["wallet"])
	}
	if wallet["tokens"] != float64(100) {
		t.Fatalf("tokens = %v", wallet["tokens"])
	}
}
