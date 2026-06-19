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

func TestVerifyInitData_WrongToken(t *testing.T) {
	values := url.Values{}
	values.Set("auth_date", strconv.FormatInt(time.Now().Unix(), 10))
	values.Set("user", `{"id":1}`)
	raw := signInitData("token-a", values)

	if _, err := VerifyInitData(raw, "token-b"); err != ErrInitDataInvalid {
		t.Fatalf("err = %v, want ErrInitDataInvalid", err)
	}
}
