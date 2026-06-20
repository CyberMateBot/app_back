package applinks

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/twelvepills-936/tgapp-/internal"
	ucModels "github.com/twelvepills-936/tgapp-/internal/usecase/models"
	"github.com/twelvepills-936/tgapp-/pkg/config"
)

// linksResponse is returned by GET /v1/app/links for the frontend (Support button, referral links).
type linksResponse struct {
	SupportChatURL   string `json:"support_chat_url"`
	BotUsername      string `json:"bot_username,omitempty"`
	ReferralLinkBase string `json:"referral_link_base,omitempty"`
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

const referralLinkPathPrefix = "/v1/users/telegram/"
const referralLinkPathSuffix = "/referral-link"
const referralsPathSuffix = "/referrals"

// Wrap adds app link endpoints with Telegram deep links from config.
func Wrap(next http.Handler, app config.ConfigApp, uc internal.UseCase) http.Handler {
	botUsername := NormalizeBotUsername(app.TelegramBotUsername)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}

		switch {
		case r.URL.Path == "/v1/app/links":
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			_ = json.NewEncoder(w).Encode(linksResponse{
				SupportChatURL:   app.SupportTelegramInviteURL,
				BotUsername:      botUsername,
				ReferralLinkBase: ReferralLinkBase(botUsername, app.TelegramReferralParamPrefix),
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
			strings.HasSuffix(r.URL.Path, referralsPathSuffix):
			telegramID := strings.TrimSuffix(
				strings.TrimPrefix(r.URL.Path, referralLinkPathPrefix),
				referralsPathSuffix,
			)
			if telegramID == "" {
				http.Error(w, "telegram_id required", http.StatusBadRequest)
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
