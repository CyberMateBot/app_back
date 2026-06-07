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
