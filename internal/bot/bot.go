package bot

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	startWelcomeText = "⚡️ CyberMate\n\n👇 Нажимай на кнопку ниже и начинай создавать!"
	webhookPath      = "/v1/telegram/webhook"
	openAppButton    = "🚀 Открыть CyberMate"
)

type Bot struct {
	api *tgbotapi.BotAPI
	// webhookSecret is sent to Telegram via setWebhook's secret_token and
	// checked against the X-Telegram-Bot-Api-Secret-Token header on every
	// inbound webhook POST. Without it, anyone who finds the webhook URL
	// (it's not otherwise secret) could POST forged Telegram updates.
	webhookSecret string
}

// New creates Telegram bot when TELEGRAM_BOT_ENABLED=true and TELEGRAM_BOT_TOKEN is set.
func New() (*Bot, error) {
	if !botEnabled() {
		slog.Info("telegram bot disabled (TELEGRAM_BOT_ENABLED=false)")
		return &Bot{}, nil
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		slog.Info("telegram bot disabled (TELEGRAM_BOT_TOKEN is empty)")
		return &Bot{}, nil
	}

	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	expected := os.Getenv("TELEGRAM_BOT_USERNAME")
	if expected != "" && api.Self.UserName != expected {
		return nil, fmt.Errorf("bot username mismatch: got @%s, want @%s", api.Self.UserName, expected)
	}

	secret := strings.TrimSpace(os.Getenv("TELEGRAM_WEBHOOK_SECRET"))
	if secret == "" {
		generated, genErr := generateWebhookSecret()
		if genErr != nil {
			slog.Warn("failed to generate telegram webhook secret token; webhook will be unauthenticated", slog.Any("error", genErr))
		} else {
			secret = generated
		}
	}

	slog.Info("telegram bot connected", slog.String("username", "@"+api.Self.UserName))
	return &Bot{api: api, webhookSecret: secret}, nil
}

func generateWebhookSecret() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func (b *Bot) Active() bool {
	return b != nil && b.api != nil
}

func botEnabled() bool {
	v := strings.TrimSpace(os.Getenv("TELEGRAM_BOT_ENABLED"))
	if v == "" {
		return true
	}
	switch strings.ToLower(v) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func botPollingEnabled() bool {
	v := strings.TrimSpace(os.Getenv("TELEGRAM_BOT_POLLING"))
	if v == "" {
		return false
	}
	switch strings.ToLower(v) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

// BotPollingEnabled reports whether long-polling should run (off by default on Railway).
func BotPollingEnabled() bool {
	return botPollingEnabled()
}

// PreparePolling removes an active Telegram webhook so getUpdates can work.
func (b *Bot) PreparePolling(ctx context.Context) error {
	if b.api == nil {
		return nil
	}
	_, err := b.api.Request(tgbotapi.DeleteWebhookConfig{DropPendingUpdates: false})
	if err != nil {
		slog.ErrorContext(ctx, "failed to delete telegram webhook before polling", slog.Any("error", err))
		return err
	}
	slog.InfoContext(ctx, "telegram webhook removed, long polling enabled")
	return nil
}

// SetWebhook registers Telegram webhook URL (e.g. https://api.example.com/v1/telegram/webhook).
func (b *Bot) SetWebhook(ctx context.Context, webhookURL string) error {
	if b.api == nil {
		return nil
	}
	webhookURL = strings.TrimSpace(webhookURL)
	if webhookURL == "" {
		return nil
	}
	if _, err := url.Parse(webhookURL); err != nil {
		return err
	}

	params := tgbotapi.Params{"url": webhookURL}
	if b.webhookSecret != "" {
		params["secret_token"] = b.webhookSecret
	}
	if _, err := b.api.MakeRequest("setWebhook", params); err != nil {
		slog.ErrorContext(ctx, "failed to set telegram webhook", slog.String("url", webhookURL), slog.Any("error", err))
		return err
	}
	slog.InfoContext(ctx, "telegram webhook configured",
		slog.String("url", webhookURL),
		slog.Bool("secret_token_set", b.webhookSecret != ""),
	)
	return nil
}

// HTTPWrap handles POST /v1/telegram/webhook for /start and other bot updates.
func HTTPWrap(next http.Handler, b *Bot) http.Handler {
	if b == nil || !b.Active() {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == webhookPath {
			b.serveWebhook(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (b *Bot) serveWebhook(w http.ResponseWriter, r *http.Request) {
	if b.api == nil {
		http.Error(w, "bot not configured", http.StatusServiceUnavailable)
		return
	}

	if b.webhookSecret != "" {
		got := r.Header.Get("X-Telegram-Bot-Api-Secret-Token")
		if subtle.ConstantTimeCompare([]byte(got), []byte(b.webhookSecret)) != 1 {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	var upd tgbotapi.Update
	if err := json.Unmarshal(body, &upd); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	b.handleUpdate(r.Context(), upd)
	w.WriteHeader(http.StatusOK)
}

// StartPolling starts handling /start command via long polling.
func (b *Bot) StartPolling(ctx context.Context) {
	if b.api == nil {
		return
	}
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30
	updates := b.api.GetUpdatesChan(u)
	for {
		select {
		case <-ctx.Done():
			return
		case upd := <-updates:
			b.handleUpdate(ctx, upd)
		}
	}
}

func (b *Bot) handleUpdate(ctx context.Context, upd tgbotapi.Update) {
	if upd.Message == nil || !upd.Message.IsCommand() || upd.Message.Command() != "start" {
		return
	}
	if err := b.sendStartWelcome(upd.Message.Chat.ID); err != nil {
		slog.ErrorContext(ctx, "failed to send telegram start message", slog.Any("error", err))
	}
}

func (b *Bot) sendStartWelcome(chatID int64) error {
	msg := tgbotapi.NewMessage(chatID, startWelcomeText)
	if btnURL := miniAppOpenURL(); btnURL != "" {
		btn := tgbotapi.NewInlineKeyboardButtonURL(openAppButton, btnURL)
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(btn))
	}
	_, err := b.api.Send(msg)
	return err
}

func miniAppOpenURL() string {
	if u := strings.TrimSpace(os.Getenv("TELEGRAM_MINI_APP_URL")); u != "" {
		return u
	}
	username := strings.TrimPrefix(strings.TrimSpace(os.Getenv("TELEGRAM_BOT_USERNAME")), "@")
	if username == "" {
		return ""
	}
	return "https://t.me/" + username + "?startapp"
}

// SendText sends a plain or formatted message to a Telegram chat.
func (b *Bot) SendText(chatID int64, text, parseMode string) error {
	if b.api == nil {
		return fmt.Errorf("bot not configured")
	}
	msg := tgbotapi.NewMessage(chatID, text)
	if parseMode != "" {
		msg.ParseMode = parseMode
	}
	_, err := b.api.Send(msg)
	return err
}

// SendMessage sends a text with optional inline buttons.
func (b *Bot) SendMessage(telegramID int64, text string, buttons []tgbotapi.InlineKeyboardButton) error {
	if b.api == nil {
		return nil
	}
	msg := tgbotapi.NewMessage(telegramID, text)
	if len(buttons) > 0 {
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(buttons)
	}
	_, err := b.api.Send(msg)
	return err
}
