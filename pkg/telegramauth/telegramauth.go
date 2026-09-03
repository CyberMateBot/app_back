package telegramauth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInitDataMissing = errors.New("initDataRaw is required")
	ErrInitDataInvalid = errors.New("invalid telegram init data")
	ErrInitDataExpired = errors.New("telegram init data expired")
)

const maxInitDataAge = 24 * time.Hour

// VerifyInitData validates Telegram WebApp init data and returns the user telegram id.
// initDataRaw may be raw query string or base64-encoded query string (project convention).
func VerifyInitData(initDataRaw, botToken string) (string, error) {
	initDataRaw = strings.TrimSpace(initDataRaw)
	botToken = strings.TrimSpace(botToken)
	if initDataRaw == "" {
		return "", ErrInitDataMissing
	}
	if botToken == "" {
		return "", fmt.Errorf("bot token is not configured")
	}

	raw := initDataRaw
	if decoded, err := base64.StdEncoding.DecodeString(initDataRaw); err == nil && strings.Contains(string(decoded), "=") {
		raw = string(decoded)
	}

	values, err := url.ParseQuery(raw)
	if err != nil {
		return "", ErrInitDataInvalid
	}

	gotHash := strings.ToLower(strings.TrimSpace(values.Get("hash")))
	if gotHash == "" {
		return "", ErrInitDataInvalid
	}
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
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(gotHash)) {
		return "", ErrInitDataInvalid
	}

	if authDate := strings.TrimSpace(values.Get("auth_date")); authDate != "" {
		sec, parseErr := strconv.ParseInt(authDate, 10, 64)
		if parseErr != nil {
			return "", ErrInitDataInvalid
		}
		if time.Since(time.Unix(sec, 0)) > maxInitDataAge {
			return "", ErrInitDataExpired
		}
	}

	userJSON := values.Get("user")
	if userJSON == "" {
		return "", ErrInitDataInvalid
	}
	var user struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal([]byte(userJSON), &user); err != nil || user.ID <= 0 {
		return "", ErrInitDataInvalid
	}
	return strconv.FormatInt(user.ID, 10), nil
}

// ExtractUserID parses telegram user id from init data without signature verification.
// Used for local development when mock init data has no hash.
func ExtractUserID(initDataRaw string) (string, error) {
	initDataRaw = strings.TrimSpace(initDataRaw)
	if initDataRaw == "" {
		return "", ErrInitDataMissing
	}

	raw := initDataRaw
	if decoded, err := base64.StdEncoding.DecodeString(initDataRaw); err == nil && strings.Contains(string(decoded), "=") {
		raw = string(decoded)
	}

	values, err := url.ParseQuery(raw)
	if err != nil {
		return "", ErrInitDataInvalid
	}

	userJSON := values.Get("user")
	if userJSON == "" {
		return "", ErrInitDataInvalid
	}
	var user struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal([]byte(userJSON), &user); err != nil || user.ID <= 0 {
		return "", ErrInitDataInvalid
	}
	return strconv.FormatInt(user.ID, 10), nil
}

// IsDevelopmentEnv reports whether the service is running in a local/dev
// environment, where init data with no Telegram signature is tolerated (the
// browser-based dev mock has no bot secret to sign with).
//
// Defaults to false (strict/production) when ENVIRONMENT is unset: this used
// to default to true, which silently disabled Telegram signature checks
// everywhere (register, wallet, generate, payments, ...) on any deployment
// that never set this specific var — exactly the case in production here,
// where only APP_ENVIRONMENT is configured. Local development must now opt
// in explicitly with ENVIRONMENT=development|dev|local.
func IsDevelopmentEnv() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("ENVIRONMENT"))) {
	case "development", "dev", "local":
		return true
	default:
		return false
	}
}

// InitDataMissingHash reports whether init data has no Telegram signature hash.
func InitDataMissingHash(initDataRaw string) bool {
	initDataRaw = strings.TrimSpace(initDataRaw)
	if initDataRaw == "" {
		return true
	}

	raw := initDataRaw
	if decoded, err := base64.StdEncoding.DecodeString(initDataRaw); err == nil && strings.Contains(string(decoded), "=") {
		raw = string(decoded)
	}

	values, err := url.ParseQuery(raw)
	if err != nil {
		return true
	}
	return strings.TrimSpace(values.Get("hash")) == ""
}
