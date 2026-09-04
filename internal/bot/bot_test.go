package bot

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TestMiniAppOpenURL(t *testing.T) {
	t.Setenv("TELEGRAM_MINI_APP_URL", "")
	t.Setenv("TELEGRAM_BOT_USERNAME", "CyberMate_bot")

	got := miniAppOpenURL()
	want := "https://t.me/CyberMate_bot?startapp"
	if got != want {
		t.Fatalf("miniAppOpenURL() = %q, want %q", got, want)
	}

	t.Setenv("TELEGRAM_MINI_APP_URL", "https://app.example.com")
	if miniAppOpenURL() != "https://app.example.com" {
		t.Fatalf("custom TELEGRAM_MINI_APP_URL not used")
	}
}

func TestServeWebhook_RequiresSecretToken(t *testing.T) {
	b := &Bot{api: &tgbotapi.BotAPI{}, webhookSecret: "supersecret"}

	// Missing header entirely.
	req := httptest.NewRequest(http.MethodPost, webhookPath, strings.NewReader(`{}`))
	rec := httptest.NewRecorder()
	b.serveWebhook(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 for missing secret token", rec.Code)
	}

	// Wrong secret.
	req = httptest.NewRequest(http.MethodPost, webhookPath, strings.NewReader(`{}`))
	req.Header.Set("X-Telegram-Bot-Api-Secret-Token", "wrong")
	rec = httptest.NewRecorder()
	b.serveWebhook(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 for wrong secret token", rec.Code)
	}

	// Correct secret: update is accepted (empty update, no /start command).
	req = httptest.NewRequest(http.MethodPost, webhookPath, strings.NewReader(`{}`))
	req.Header.Set("X-Telegram-Bot-Api-Secret-Token", "supersecret")
	rec = httptest.NewRecorder()
	b.serveWebhook(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 for correct secret token", rec.Code)
	}
}

func TestServeWebhook_NoSecretConfiguredAllowsRequest(t *testing.T) {
	// Backward-compat: if secret generation ever fails, the webhook must
	// still function (degrading to the old unauthenticated behavior) rather
	// than bricking the bot.
	b := &Bot{api: &tgbotapi.BotAPI{}}

	req := httptest.NewRequest(http.MethodPost, webhookPath, strings.NewReader(`{}`))
	rec := httptest.NewRecorder()
	b.serveWebhook(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 when no secret is configured", rec.Code)
	}
}

func TestGenerateWebhookSecret_Unique(t *testing.T) {
	a, err := generateWebhookSecret()
	if err != nil {
		t.Fatalf("generateWebhookSecret() err = %v", err)
	}
	b, err := generateWebhookSecret()
	if err != nil {
		t.Fatalf("generateWebhookSecret() err = %v", err)
	}
	if a == "" || b == "" {
		t.Fatal("expected non-empty secrets")
	}
	if a == b {
		t.Fatal("expected distinct secrets across calls")
	}
}

func TestStartWelcomeText(t *testing.T) {
	if startWelcomeText == "" {
		t.Fatal("startWelcomeText is empty")
	}
	if strings.Contains(startWelcomeText, "реклам") || strings.Contains(startWelcomeText, "Добро пожаловать") {
		t.Fatalf("old welcome text still present: %q", startWelcomeText)
	}
}
