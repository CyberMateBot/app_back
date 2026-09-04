package tokenguard

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/twelvepills-936/tgapp-/pkg/telegramauth"
)

func signInitData(botToken string, values url.Values) string {
	values.Del("hash")
	pairs := make([]string, 0, len(values))
	for key := range values {
		pairs = append(pairs, key+"="+values.Get(key))
	}
	sort.Strings(pairs)
	dataCheckString := strings.Join(pairs, "\n")

	secretMAC := hmac.New(sha256.New, []byte("WebAppData"))
	secretMAC.Write([]byte(botToken))
	secretKey := secretMAC.Sum(nil)

	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(dataCheckString))
	values.Set("hash", hex.EncodeToString(mac.Sum(nil)))
	return values.Encode()
}

func TestDevelopmentMockInitData(t *testing.T) {
	t.Setenv("ENVIRONMENT", "development")

	raw := `user={"id":777000,"first_name":"Dev_User"}`
	if !telegramauth.InitDataMissingHash(raw) {
		t.Fatal("mock init data should not include hash")
	}

	id, err := telegramauth.ExtractUserID(raw)
	if err != nil {
		t.Fatalf("ExtractUserID() err = %v", err)
	}
	if id != "777000" {
		t.Fatalf("telegram id = %q, want 777000", id)
	}
}

func TestCheckAccessSkipsBalanceWhenBillingDisabled(t *testing.T) {
	t.Setenv("BILLING_DISABLED", "true")
	g := &Guard{db: nil}
	if err := g.CheckAccess(t.Context(), "777000", ""); err != nil {
		t.Fatalf("CheckAccess() err = %v, want nil when db is nil", err)
	}
}

func TestVerifyTelegramOwnership_EmptyBotTokenProductionDenies(t *testing.T) {
	t.Setenv("ENVIRONMENT", "production")
	err := verifyTelegramOwnership("", "777000", "")
	if !errors.Is(err, ErrBotTokenNotConfigured) {
		t.Fatalf("err = %v, want ErrBotTokenNotConfigured", err)
	}

	// Even with initData supplied, a deployment missing its bot token must
	// still refuse the request rather than trust the client's claim.
	err = verifyTelegramOwnership("", "777000", "user=%7B%22id%22%3A777000%7D")
	if !errors.Is(err, ErrBotTokenNotConfigured) {
		t.Fatalf("err = %v, want ErrBotTokenNotConfigured", err)
	}
}

func TestVerifyTelegramOwnership_EmptyBotTokenDevAllowsNoInitData(t *testing.T) {
	t.Setenv("ENVIRONMENT", "development")
	if err := verifyTelegramOwnership("", "777000", ""); err != nil {
		t.Fatalf("err = %v, want nil for dev tooling without init data", err)
	}
}

func TestVerifyTelegramOwnership_EmptyBotTokenDevChecksMismatch(t *testing.T) {
	t.Setenv("ENVIRONMENT", "development")
	raw := "user=" + url.QueryEscape(`{"id":777000}`)

	if err := verifyTelegramOwnership("", "777000", raw); err != nil {
		t.Fatalf("err = %v, want nil for matching mock id", err)
	}
	if err := verifyTelegramOwnership("", "1", raw); !errors.Is(err, ErrTelegramIDMismatch) {
		t.Fatalf("err = %v, want ErrTelegramIDMismatch", err)
	}
}

func TestVerifyTelegramOwnership_SignedInitDataRequiredInProduction(t *testing.T) {
	t.Setenv("ENVIRONMENT", "production")
	botToken := "123456:ABC-DEF"

	if err := verifyTelegramOwnership(botToken, "777000", ""); !errors.Is(err, ErrInitDataRequired) {
		t.Fatalf("err = %v, want ErrInitDataRequired", err)
	}

	values := url.Values{}
	values.Set("auth_date", strconv.FormatInt(time.Now().Unix(), 10))
	values.Set("user", `{"id":777000,"first_name":"Test"}`)
	raw := signInitData(botToken, values)

	if err := verifyTelegramOwnership(botToken, "777000", raw); err != nil {
		t.Fatalf("err = %v, want nil for correctly signed init data", err)
	}
	if err := verifyTelegramOwnership(botToken, "1", raw); !errors.Is(err, ErrTelegramIDMismatch) {
		t.Fatalf("err = %v, want ErrTelegramIDMismatch for telegramId not matching signed identity", err)
	}

	// A forged/unsigned initData claiming an arbitrary id must be rejected
	// once a bot token is configured, even in production.
	forged := url.Values{}
	forged.Set("user", `{"id":42}`)
	if err := verifyTelegramOwnership(botToken, "42", forged.Encode()); err == nil {
		t.Fatal("expected error for unsigned init data when bot token is configured")
	}
}

func TestIsBillingDisabled(t *testing.T) {
	t.Setenv("BILLING_DISABLED", "")
	if isBillingDisabled() {
		t.Fatal("expected billing enabled by default")
	}

	t.Setenv("BILLING_DISABLED", "true")
	if !isBillingDisabled() {
		t.Fatal("expected billing disabled")
	}
}
