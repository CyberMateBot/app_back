package telegramauth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
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

func TestVerifyInitData_OK(t *testing.T) {
	botToken := "123456:ABC-DEF"
	values := url.Values{}
	values.Set("auth_date", strconv.FormatInt(time.Now().Unix(), 10))
	values.Set("user", `{"id":777000,"first_name":"Test"}`)
	raw := signInitData(botToken, values)

	got, err := VerifyInitData(raw, botToken)
	if err != nil {
		t.Fatalf("VerifyInitData() err = %v", err)
	}
	if got != "777000" {
		t.Fatalf("telegram id = %q, want 777000", got)
	}
}

func TestExtractUserID_OK(t *testing.T) {
	values := url.Values{}
	values.Set("user", `{"id":777000,"first_name":"Dev"}`)
	raw := values.Encode()

	got, err := ExtractUserID(raw)
	if err != nil {
		t.Fatalf("ExtractUserID() err = %v", err)
	}
	if got != "777000" {
		t.Fatalf("telegram id = %q, want 777000", got)
	}
}

func TestInitDataMissingHash(t *testing.T) {
	values := url.Values{}
	values.Set("user", `{"id":1}`)
	if !InitDataMissingHash(values.Encode()) {
		t.Fatal("expected missing hash")
	}

	values.Set("hash", "abc")
	if InitDataMissingHash(values.Encode()) {
		t.Fatal("expected hash present")
	}
}

func TestVerifyInitData_WrongToken(t *testing.T) {
	values := url.Values{}
	values.Set("auth_date", strconv.FormatInt(time.Now().Unix(), 10))
	values.Set("user", `{"id":1}`)
	raw := signInitData("token-a", values)

	if _, err := VerifyInitData(raw, "token-b"); err != ErrInitDataInvalid {
		t.Fatalf("err = %v, want ErrInitDataInvalid", err)
	}
}

func TestIsDevelopmentEnv(t *testing.T) {
	// Unset must default to strict/production now: a deployment that never
	// sets this specific var (only e.g. APP_ENVIRONMENT) must not silently
	// disable Telegram signature checks.
	t.Setenv("ENVIRONMENT", "")
	if IsDevelopmentEnv() {
		t.Fatal("expected unset ENVIRONMENT to default to production/strict")
	}

	t.Setenv("ENVIRONMENT", "development")
	if !IsDevelopmentEnv() {
		t.Fatal("expected development")
	}

	t.Setenv("ENVIRONMENT", "dev")
	if !IsDevelopmentEnv() {
		t.Fatal("expected dev")
	}

	t.Setenv("ENVIRONMENT", "local")
	if !IsDevelopmentEnv() {
		t.Fatal("expected local")
	}

	t.Setenv("ENVIRONMENT", "production")
	if IsDevelopmentEnv() {
		t.Fatal("expected not development")
	}
}
