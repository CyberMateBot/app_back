package applinks

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/twelvepills-936/tgapp-/pkg/config"
)

func TestWrap_AppLinks(t *testing.T) {
	t.Parallel()

	mux := Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}), config.ConfigApp{
		TelegramBotUsername:      "CyberMate_bot",
		SupportTelegramInviteURL: "https://t.me/+test",
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/app/links", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}

	var body linksResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.BotUsername != "CyberMate_bot" {
		t.Fatalf("bot_username = %q", body.BotUsername)
	}
	if body.ReferralLinkBase != "https://t.me/CyberMate_bot?startapp=ref_" {
		t.Fatalf("referral_link_base = %q", body.ReferralLinkBase)
	}
	if body.MiniAppFullscreenURL != "https://t.me/CyberMate_bot?startapp&mode=fullscreen" {
		t.Fatalf("mini_app_fullscreen_url = %q", body.MiniAppFullscreenURL)
	}
}

func TestWrap_ReferralLink(t *testing.T) {
	t.Parallel()

	mux := Wrap(http.NotFoundHandler(), config.ConfigApp{TelegramBotUsername: "CyberMate_bot"}, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/users/telegram/42/referral-link", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	var body referralLinkResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	want := "https://t.me/CyberMate_bot?startapp=ref_42"
	if body.ReferralLink != want {
		t.Fatalf("referral_link = %q, want %q", body.ReferralLink, want)
	}
}

func TestWrap_ReferralsList(t *testing.T) {
	t.Parallel()

	mux := Wrap(http.NotFoundHandler(), config.ConfigApp{TelegramBotUsername: "CyberMate_bot"}, &fakeReferralsUC{})

	req := httptest.NewRequest(http.MethodGet, "/v1/users/telegram/42/referrals", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}

	var body referralsListResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.TotalCount != 1 || len(body.Referrals) != 1 {
		t.Fatalf("unexpected body: %+v", body)
	}
	if body.Referrals[0].TelegramID != "999" {
		t.Fatalf("telegram_id = %q", body.Referrals[0].TelegramID)
	}
}
