package applinks

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/twelvepills-936/tgapp-/pkg/config"
	"github.com/twelvepills-936/tgapp-/pkg/tokenguard"
)

func TestWrap_AppLinks(t *testing.T) {
	t.Parallel()

	mux := Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}), config.ConfigApp{
		TelegramBotUsername:      "CyberMate_bot",
		SupportTelegramInviteURL: "https://t.me/+test",
	}, nil, nil)

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

	mux := Wrap(http.NotFoundHandler(), config.ConfigApp{TelegramBotUsername: "CyberMate_bot"}, nil, nil)

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

	mux := Wrap(http.NotFoundHandler(), config.ConfigApp{TelegramBotUsername: "CyberMate_bot"}, &fakeReferralsUC{}, nil)

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

// fakeTokenChecker lets tests simulate the IDOR-prevention ownership check
// without a real database/tokenguard.Guard.
type fakeTokenChecker struct {
	err error
}

func (f *fakeTokenChecker) CheckIdentity(_ context.Context, telegramID, _ string) error {
	return f.err
}

func TestWrap_ReferralsList_RequiresOwnership(t *testing.T) {
	t.Parallel()

	mux := Wrap(http.NotFoundHandler(), config.ConfigApp{TelegramBotUsername: "CyberMate_bot"}, &fakeReferralsUC{},
		&fakeTokenChecker{err: tokenguard.ErrTelegramIDMismatch})

	req := httptest.NewRequest(http.MethodGet, "/v1/users/telegram/42/referrals", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 when caller does not own telegramId", rec.Code)
	}
}

func TestWrap_Subscription_RequiresOwnership(t *testing.T) {
	t.Parallel()

	mux := Wrap(http.NotFoundHandler(), config.ConfigApp{TelegramBotUsername: "CyberMate_bot"}, &fakeReferralsUC{},
		&fakeTokenChecker{err: tokenguard.ErrTelegramIDMismatch})

	req := httptest.NewRequest(http.MethodGet, "/v1/users/telegram/42/subscription", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 when caller does not own telegramId", rec.Code)
	}
}

func TestWrap_ProfileAndTheme_ForwardOnlyWhenOwned(t *testing.T) {
	t.Parallel()

	var forwarded bool
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		forwarded = true
		w.WriteHeader(http.StatusOK)
	})

	deny := Wrap(next, config.ConfigApp{TelegramBotUsername: "CyberMate_bot"}, &fakeReferralsUC{},
		&fakeTokenChecker{err: tokenguard.ErrTelegramIDMismatch})

	forwarded = false
	rec := httptest.NewRecorder()
	deny.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/users/telegram/42", nil))
	if forwarded {
		t.Fatal("profile GET must not be forwarded when ownership check fails")
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}

	forwarded = false
	rec = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/v1/users/telegram/42/theme", nil)
	deny.ServeHTTP(rec, req)
	if forwarded {
		t.Fatal("theme PATCH must not be forwarded when ownership check fails")
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}

	allow := Wrap(next, config.ConfigApp{TelegramBotUsername: "CyberMate_bot"}, &fakeReferralsUC{},
		&fakeTokenChecker{err: nil})

	forwarded = false
	rec = httptest.NewRecorder()
	allow.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/users/telegram/42", nil))
	if !forwarded {
		t.Fatal("profile GET must be forwarded when ownership check passes")
	}

	forwarded = false
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPatch, "/v1/users/telegram/42/theme", nil)
	allow.ServeHTTP(rec, req)
	if !forwarded {
		t.Fatal("theme PATCH must be forwarded when ownership check passes")
	}
}
