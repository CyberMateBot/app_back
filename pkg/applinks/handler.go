package applinks

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/twelvepills-936/tgapp-/internal"
	ucModels "github.com/twelvepills-936/tgapp-/internal/usecase/models"
	"github.com/twelvepills-936/tgapp-/pkg/config"
	"github.com/twelvepills-936/tgapp-/pkg/tokenguard"
)

// TokenChecker verifies that a telegramId is genuinely owned by the caller
// (via a valid Telegram init-data signature). Satisfied by
// *tokenguard.Guard; declared as a small interface here so this package's
// ownership checks are unit-testable without a real database.
type TokenChecker interface {
	CheckIdentity(ctx context.Context, telegramID, initDataRaw string) error
}

// linksResponse is returned by GET /v1/app/links for the frontend (Support button, referral links).
type linksResponse struct {
	SupportChatURL        string `json:"support_chat_url"`
	BotUsername           string `json:"bot_username,omitempty"`
	ReferralLinkBase      string `json:"referral_link_base,omitempty"`
	MiniAppFullscreenURL  string `json:"mini_app_fullscreen_url,omitempty"`
}

type referralLinkResponse struct {
	ReferralLink string `json:"referral_link"`
}

type referralListItemResponse struct {
	ProfileID           int64  `json:"profile_id"`
	TelegramID          string `json:"telegram_id"`
	Name                string `json:"name"`
	Username            string `json:"username"`
	Avatar              string `json:"avatar"`
	CompletedTasksCount int64  `json:"completed_tasks_count"`
	Earnings            int64  `json:"earnings"`
}

type referralsListResponse struct {
	Referrals      []referralListItemResponse `json:"referrals"`
	TotalCount     int64                      `json:"total_count"`
	TotalEarnings  int64                      `json:"total_earnings"`
	CompletedTasks int64                      `json:"completed_tasks"`
}

type homeWidgetItemResponse struct {
	ID              int64  `json:"id"`
	SortOrder       int32  `json:"sort_order"`
	TagText         string `json:"tag_text"`
	TagBg           string `json:"tag_bg"`
	TagColor        string `json:"tag_color"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	BackgroundStyle string `json:"background_style"`
	ImageURL        string `json:"image_url"`
}

type homeWidgetsResponse struct {
	Data []homeWidgetItemResponse `json:"data"`
}

type billingPlanResponse struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Badge      string   `json:"badge"`
	BadgeClass string   `json:"badge_class"`
	PriceRub   int64    `json:"price_rub"`
	PriceSub   string   `json:"price_sub"`
	Coins      int64    `json:"coins"`
	Features   []string `json:"features"`
	Locked     []string `json:"locked,omitempty"`
	Popular    bool     `json:"popular"`
	SortOrder  int32    `json:"sort_order"`
}

type billingCoinPackResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Coins     int64  `json:"coins"`
	PriceRub  int64  `json:"price_rub"`
	Badge     string `json:"badge,omitempty"`
	SortOrder int32  `json:"sort_order"`
}

type billingCatalogResponse struct {
	CoinRateRub float64                   `json:"coin_rate_rub"`
	Plans       []billingPlanResponse     `json:"plans"`
	CoinPacks   []billingCoinPackResponse `json:"coin_packs"`
}

type subscriptionStateResponse struct {
	PlanID       string `json:"plan_id"`
	PlanName     string `json:"plan_name"`
	PlanRank     int    `json:"plan_rank"`
	IsPaid       bool   `json:"is_paid"`
	Coins        int64  `json:"coins"`
	StartedAt    string `json:"started_at,omitempty"`
	ExpiresAt    string `json:"expires_at,omitempty"`
	DaysLeft     int    `json:"days_left"`
	HoursLeft    int    `json:"hours_left"`
	IsActive     bool   `json:"is_active"`
	ExpiringSoon bool   `json:"expiring_soon"`
	Expired      bool   `json:"expired"`
}

const referralLinkPathPrefix = "/v1/users/telegram/"
const referralLinkPathSuffix = "/referral-link"
const referralsPathSuffix = "/referrals"
const subscriptionPathSuffix = "/subscription"
const themePathSuffix = "/theme"

// knownTelegramSubPaths lists the recognized suffixes under
// /v1/users/telegram/{id}/... so the bare profile path (no suffix) can be
// distinguished from them.
var knownTelegramSubPaths = []string{
	referralLinkPathSuffix,
	referralsPathSuffix,
	subscriptionPathSuffix,
	themePathSuffix,
}

func isBareTelegramProfilePath(rest string) bool {
	if rest == "" || strings.Contains(rest, "/") {
		return false
	}
	for _, suffix := range knownTelegramSubPaths {
		if strings.HasSuffix(rest, suffix) {
			return false
		}
	}
	return true
}

// authorizeTelegramOwner verifies the caller's Telegram init data proves
// ownership of telegramID before any profile/subscription/referral data for
// that id is served. These routes used to be fully unauthenticated, letting
// anyone enumerate any user's profile, plan, or referral earnings by
// telegram id (IDOR). tokens may be nil in tests, in which case the check is
// skipped (mirrors tokenguard.Guard's own nil-safety).
func authorizeTelegramOwner(w http.ResponseWriter, r *http.Request, tokens TokenChecker, telegramID string) bool {
	telegramID = strings.TrimSpace(telegramID)
	if telegramID == "" {
		http.Error(w, "telegram_id required", http.StatusBadRequest)
		return false
	}
	if tokens == nil {
		return true
	}
	err := tokens.CheckIdentity(r.Context(), telegramID, tokenguard.InitDataFromRequest(r, ""))
	return !tokenguard.WriteHTTPError(w, r, err)
}

// Wrap adds app link endpoints with Telegram deep links from config.
func Wrap(next http.Handler, app config.ConfigApp, uc internal.UseCase, tokens TokenChecker) http.Handler {
	botUsername := NormalizeBotUsername(app.TelegramBotUsername)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The profile GET and theme PATCH routes are handled further down the
		// chain by the gRPC-gateway mux, not in this switch. They still need
		// an ownership check here (before forwarding) because the generated
		// gateway handlers have no notion of Telegram init data.
		if strings.HasPrefix(r.URL.Path, referralLinkPathPrefix) {
			rest := strings.TrimPrefix(r.URL.Path, referralLinkPathPrefix)
			switch {
			case r.Method == http.MethodPatch && strings.HasSuffix(rest, themePathSuffix):
				telegramID := strings.TrimSuffix(rest, themePathSuffix)
				if !authorizeTelegramOwner(w, r, tokens, telegramID) {
					return
				}
				next.ServeHTTP(w, r)
				return
			case r.Method == http.MethodGet && isBareTelegramProfilePath(rest):
				if !authorizeTelegramOwner(w, r, tokens, rest) {
					return
				}
				next.ServeHTTP(w, r)
				return
			}
		}

		if r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}

		switch {
		case r.URL.Path == "/v1/app/links":
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			_ = json.NewEncoder(w).Encode(linksResponse{
				SupportChatURL:       app.SupportTelegramInviteURL,
				BotUsername:          botUsername,
				ReferralLinkBase:     ReferralLinkBase(botUsername, app.TelegramReferralParamPrefix),
				MiniAppFullscreenURL: MainMiniAppFullscreenURL(botUsername),
			})
			return

		case r.URL.Path == "/v1/billing/catalog":
			if uc == nil {
				http.Error(w, "billing catalog is not configured", http.StatusServiceUnavailable)
				return
			}
			out, err := uc.GetPublicBillingCatalog(r.Context())
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			plans := make([]billingPlanResponse, 0, len(out.Plans))
			for _, item := range out.Plans {
				plans = append(plans, billingPlanResponse{
					ID: item.ID, Name: item.Name, Badge: item.Badge, BadgeClass: item.BadgeClass,
					PriceRub: item.PriceRub, PriceSub: item.PriceSub, Coins: item.Coins,
					Features: item.Features, Locked: item.Locked, Popular: item.Popular, SortOrder: item.SortOrder,
				})
			}
			packs := make([]billingCoinPackResponse, 0, len(out.CoinPacks))
			for _, item := range out.CoinPacks {
				packs = append(packs, billingCoinPackResponse{
					ID: item.ID, Name: item.Name, Coins: item.Coins,
					PriceRub: item.PriceRub, Badge: item.Badge, SortOrder: item.SortOrder,
				})
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			_ = json.NewEncoder(w).Encode(billingCatalogResponse{
				CoinRateRub: out.CoinRateRub,
				Plans:       plans,
				CoinPacks:   packs,
			})
			return

		case r.URL.Path == "/v1/home/widgets":
			if uc == nil {
				http.Error(w, "home widgets are not configured", http.StatusServiceUnavailable)
				return
			}
			out, err := uc.ListHomeWidgets(r.Context())
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			items := make([]homeWidgetItemResponse, 0, len(out.Data))
			for _, item := range out.Data {
				items = append(items, homeWidgetItemResponse{
					ID:              item.ID,
					SortOrder:       item.SortOrder,
					TagText:         item.TagText,
					TagBg:           item.TagBg,
					TagColor:        item.TagColor,
					Title:           item.Title,
					Description:     item.Description,
					BackgroundStyle: item.BackgroundStyle,
					ImageURL:        item.ImageURL,
				})
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			_ = json.NewEncoder(w).Encode(homeWidgetsResponse{Data: items})
			return

		case strings.HasPrefix(r.URL.Path, referralLinkPathPrefix) &&
			strings.HasSuffix(r.URL.Path, referralLinkPathSuffix):
			telegramID := strings.TrimSuffix(
				strings.TrimPrefix(r.URL.Path, referralLinkPathPrefix),
				referralLinkPathSuffix,
			)
			if telegramID == "" {
				http.Error(w, "telegram_id required", http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			_ = json.NewEncoder(w).Encode(referralLinkResponse{
				ReferralLink: ReferralLink(botUsername, telegramID, app.TelegramReferralParamPrefix),
			})
			return

		case strings.HasPrefix(r.URL.Path, referralLinkPathPrefix) &&
			strings.HasSuffix(r.URL.Path, subscriptionPathSuffix):
			telegramID := strings.TrimSuffix(
				strings.TrimPrefix(r.URL.Path, referralLinkPathPrefix),
				subscriptionPathSuffix,
			)
			if telegramID == "" {
				http.Error(w, "telegram_id required", http.StatusBadRequest)
				return
			}
			if !authorizeTelegramOwner(w, r, tokens, telegramID) {
				return
			}
			if uc == nil {
				http.Error(w, "subscriptions are not configured", http.StatusServiceUnavailable)
				return
			}
			out, err := uc.GetSubscriptionByTelegramID(r.Context(), telegramID)
			if err != nil {
				switch {
				case errors.Is(err, ucModels.ErrProfileNotFound):
					http.Error(w, "profile not found", http.StatusNotFound)
				case errors.Is(err, ucModels.ErrInvalidInput):
					http.Error(w, err.Error(), http.StatusBadRequest)
				default:
					http.Error(w, "internal error", http.StatusInternalServerError)
				}
				return
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			_ = json.NewEncoder(w).Encode(subscriptionStateResponse{
				PlanID:       out.PlanID,
				PlanName:     out.PlanName,
				PlanRank:     out.PlanRank,
				IsPaid:       out.IsPaid,
				Coins:        out.Coins,
				StartedAt:    out.StartedAt,
				ExpiresAt:    out.ExpiresAt,
				DaysLeft:     out.DaysLeft,
				HoursLeft:    out.HoursLeft,
				IsActive:     out.IsActive,
				ExpiringSoon: out.ExpiringSoon,
				Expired:      out.Expired,
			})
			return

		case strings.HasPrefix(r.URL.Path, referralLinkPathPrefix) &&
			strings.HasSuffix(r.URL.Path, referralsPathSuffix):
			telegramID := strings.TrimSuffix(
				strings.TrimPrefix(r.URL.Path, referralLinkPathPrefix),
				referralsPathSuffix,
			)
			if telegramID == "" {
				http.Error(w, "telegram_id required", http.StatusBadRequest)
				return
			}
			if !authorizeTelegramOwner(w, r, tokens, telegramID) {
				return
			}
			if uc == nil {
				http.Error(w, "referrals are not configured", http.StatusServiceUnavailable)
				return
			}
			out, err := uc.ListReferralsByTelegramID(r.Context(), telegramID)
			if err != nil {
				switch {
				case errors.Is(err, ucModels.ErrProfileNotFound):
					http.Error(w, "profile not found", http.StatusNotFound)
				case errors.Is(err, ucModels.ErrInvalidInput):
					http.Error(w, err.Error(), http.StatusBadRequest)
				default:
					http.Error(w, "internal error", http.StatusInternalServerError)
				}
				return
			}
			items := make([]referralListItemResponse, 0, len(out.Referrals))
			for _, item := range out.Referrals {
				items = append(items, referralListItemResponse{
					ProfileID:           item.ProfileID,
					TelegramID:          item.TelegramID,
					Name:                item.Name,
					Username:            item.Username,
					Avatar:              item.Avatar,
					CompletedTasksCount: item.CompletedTasksCount,
					Earnings:            item.Earnings,
				})
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			_ = json.NewEncoder(w).Encode(referralsListResponse{
				Referrals:      items,
				TotalCount:     out.TotalCount,
				TotalEarnings:  out.TotalEarnings,
				CompletedTasks: out.CompletedTasks,
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}
