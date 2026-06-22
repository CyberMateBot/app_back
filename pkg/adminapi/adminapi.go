package adminapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/twelvepills-936/tgapp-/internal"
	ucModels "github.com/twelvepills-936/tgapp-/internal/usecase/models"
	"github.com/twelvepills-936/tgapp-/pkg/config"
	"github.com/twelvepills-936/tgapp-/pkg/jwtutil"
)

const (
	apiPrefix    = "/api/admin/"
	bearerPrefix = "bearer "
)

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResp struct {
	Token string `json:"token"`
	Admin struct {
		ID    int64  `json:"id"`
		Email string `json:"email"`
	} `json:"admin"`
}

type adminMeResp struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}

type statsResp struct {
	TotalUsers       int64 `json:"total_users"`
	ActiveUsersToday int64 `json:"active_users_today"`
	NewUsersToday    int64 `json:"new_users_today"`
	TotalMessages    int64 `json:"total_messages"`
}

type userResp struct {
	ID                   int64  `json:"id"`
	TelegramID           int64  `json:"telegram_id"`
	Username             string `json:"username"`
	FirstName            string `json:"first_name"`
	LastName             string `json:"last_name"`
	IsActive             bool   `json:"is_active"`
	Tokens               int64  `json:"tokens"`
	CreatedAt            string `json:"created_at"`
	SubscriptionPlanID   string `json:"subscription_plan_id"`
	SubscriptionPlan     string `json:"subscription_plan"`
	SubscriptionExpires  string `json:"subscription_expires,omitempty"`
	SubscriptionDaysLeft int    `json:"subscription_days_left"`
}

type subscriptionStateResp struct {
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

type setSubscriptionReq struct {
	PlanID       string  `json:"plan_id"`
	DurationDays int     `json:"duration_days"`
	ExpiresAt    *string `json:"expires_at"`
	NoExpiry     bool    `json:"no_expiry"`
	GrantCoins   bool    `json:"grant_coins"`
}

type subscriptionChangeResp struct {
	User         userResp              `json:"user"`
	Subscription subscriptionStateResp `json:"subscription"`
	CoinsGranted int64                 `json:"coins_granted"`
}

type tokenChangeReq struct {
	Amount int64  `json:"amount"`
	Reason string `json:"reason"`
}

type tokenChangeResp struct {
	UserID    int64  `json:"user_id"`
	Tokens    int64  `json:"tokens"`
	Delta     int64  `json:"delta"`
	Operation string `json:"operation"`
}

type usersListResp struct {
	Data  []userResp `json:"data"`
	Total int64      `json:"total"`
}

type patchUserReq struct {
	IsActive bool `json:"is_active"`
}

type broadcastReq struct {
	Message   string `json:"message"`
	Target    string `json:"target"`
	ParseMode string `json:"parse_mode"`
}

type broadcastResp struct {
	Sent   int64 `json:"sent"`
	Failed int64 `json:"failed"`
}

type eventItemResp struct {
	ID      string `json:"id"`
	Time    string `json:"time"`
	User    string `json:"user"`
	Action  string `json:"action"`
	Details string `json:"details"`
}

type eventsListResp struct {
	Data []eventItemResp `json:"data"`
}

type transactionItemResp struct {
	ID          int64  `json:"id"`
	User        string `json:"user"`
	Type        string `json:"type"`
	TypeLabel   string `json:"type_label"`
	Amount      int64  `json:"amount"`
	AmountLabel string `json:"amount_label"`
	Method      string `json:"method"`
	MethodLabel string `json:"method_label"`
	CreatedAt   string `json:"created_at"`
	Status      string `json:"status"`
	StatusLabel string `json:"status_label"`
}

type transactionStatsResp struct {
	CreditsMonth    int64 `json:"credits_month"`
	DebitsMonth     int64 `json:"debits_month"`
	OperationsMonth int64 `json:"operations_month"`
	AvgAmount       int64 `json:"avg_amount"`
}

type transactionsListResp struct {
	Stats transactionStatsResp  `json:"stats"`
	Data  []transactionItemResp `json:"data"`
	Total int64                 `json:"total"`
}

type broadcastHistoryItemResp struct {
	ID          int64  `json:"id"`
	Message     string `json:"message"`
	Target      string `json:"target"`
	TargetLabel string `json:"target_label"`
	Sent        int64  `json:"sent"`
	Failed      int64  `json:"failed"`
	CreatedAt   string `json:"created_at"`
	Status      string `json:"status"`
	StatusLabel string `json:"status_label"`
}

type broadcastsListResp struct {
	Data  []broadcastHistoryItemResp `json:"data"`
	Total int64                      `json:"total"`
}

type modelItemResp struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Provider string `json:"provider"`
	Category string `json:"category"`
	Price    int64  `json:"price"`
	Enabled  bool   `json:"enabled"`
}

type modelsListResp struct {
	Data []modelItemResp `json:"data"`
}

type subscriptionPlanItemResp struct {
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
	Enabled    bool     `json:"enabled"`
	SortOrder  int32    `json:"sort_order"`
}

type subscriptionPlansListResp struct {
	Data []subscriptionPlanItemResp `json:"data"`
}

type updateSubscriptionPlansReq struct {
	Data []subscriptionPlanItemResp `json:"data"`
}

type coinPackItemResp struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Coins     int64  `json:"coins"`
	PriceRub  int64  `json:"price_rub"`
	Badge     string `json:"badge,omitempty"`
	Enabled   bool   `json:"enabled"`
	SortOrder int32  `json:"sort_order"`
}

type coinPacksListResp struct {
	Data []coinPackItemResp `json:"data"`
}

type updateCoinPacksReq struct {
	Data []coinPackItemResp `json:"data"`
}

type patchModelReq struct {
	Price   *int64 `json:"price"`
	Enabled *bool  `json:"enabled"`
}

type homeWidgetItemResp struct {
	ID              int64  `json:"id"`
	SortOrder       int32  `json:"sort_order"`
	TagText         string `json:"tag_text"`
	TagBg           string `json:"tag_bg"`
	TagColor        string `json:"tag_color"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	BackgroundStyle string `json:"background_style"`
	ImageURL        string `json:"image_url"`
	IsActive        bool   `json:"is_active"`
}

type homeWidgetsListResp struct {
	Data []homeWidgetItemResp `json:"data"`
}

type createHomeWidgetReq struct {
	SortOrder       int32  `json:"sort_order"`
	TagText         string `json:"tag_text"`
	TagBg           string `json:"tag_bg"`
	TagColor        string `json:"tag_color"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	BackgroundStyle string `json:"background_style"`
	ImageURL        string `json:"image_url"`
	IsActive        bool   `json:"is_active"`
}

type patchHomeWidgetReq struct {
	SortOrder       *int32  `json:"sort_order"`
	TagText         *string `json:"tag_text"`
	TagBg           *string `json:"tag_bg"`
	TagColor        *string `json:"tag_color"`
	Title           *string `json:"title"`
	Description     *string `json:"description"`
	BackgroundStyle *string `json:"background_style"`
	ImageURL        *string `json:"image_url"`
	IsActive        *bool   `json:"is_active"`
}

type errorResp struct {
	Error string `json:"error"`
}

// Messenger sends Telegram broadcast messages.
type Messenger interface {
	Active() bool
	SendText(chatID int64, text, parseMode string) error
}

// Wrap registers Admin Panel REST API under /api/admin/*.
func Wrap(next http.Handler, uc internal.UseCase, jwtCfg config.ConfigJWT, messenger Messenger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, apiPrefix) {
			next.ServeHTTP(w, r)
			return
		}

		path := strings.TrimPrefix(r.URL.Path, apiPrefix)

		switch {
		case r.Method == http.MethodPost && path == "auth/login":
			var req loginReq
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeErr(w, http.StatusBadRequest, "invalid json")
				return
			}
			out, err := uc.AdminLogin(r.Context(), ucModels.AdminLoginInput{
				Email:    req.Email,
				Password: req.Password,
			})
			if err != nil {
				writeUsecaseErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, loginResp{
				Token: out.Token,
				Admin: struct {
					ID    int64  `json:"id"`
					Email string `json:"email"`
				}{ID: out.Admin.ID, Email: out.Admin.Email},
			})
			return

		}

		if config.JWTSecretWeak(jwtCfg.Secret) && config.IsDeployedProduction() {
			writeErr(w, http.StatusServiceUnavailable, "jwt is not configured")
			return
		}

		adminID, ok := requireAdmin(w, r, jwtCfg)
		if !ok {
			return
		}

		switch {
		case r.Method == http.MethodPost && path == "auth/logout":
			w.WriteHeader(http.StatusNoContent)
			return

		case r.Method == http.MethodGet && path == "auth/me":
			out, err := uc.GetAdmin(r.Context(), adminID)
			if err != nil {
				writeUsecaseErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, adminMeResp{ID: out.ID, Email: out.Email})
			return

		case r.Method == http.MethodGet && path == "stats":
			out, err := uc.GetAdminStats(r.Context())
			if err != nil {
				writeUsecaseErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, statsResp{
				TotalUsers:       out.TotalUsers,
				ActiveUsersToday: out.ActiveUsersToday,
				NewUsersToday:    out.NewUsersToday,
				TotalMessages:    out.TotalMessages,
			})
			return

		case r.Method == http.MethodGet && path == "events":
			limit := parseInt32(r.URL.Query().Get("limit"), 20)
			out, err := uc.ListAdminEvents(r.Context(), limit)
			if err != nil {
				writeUsecaseErr(w, err)
				return
			}
			items := make([]eventItemResp, 0, len(out.Data))
			for _, item := range out.Data {
				items = append(items, eventItemResp(item))
			}
			writeJSON(w, http.StatusOK, eventsListResp{Data: items})
			return

		case r.Method == http.MethodGet && path == "transactions":
			page := parseInt32(r.URL.Query().Get("page"), 1)
			perPage := parseInt32(r.URL.Query().Get("per_page"), 20)
			out, err := uc.ListAdminTransactions(r.Context(), ucModels.AdminListTransactionsInput{
				Page:      page,
				PerPage:   perPage,
				Operation: r.URL.Query().Get("operation"),
			})
			if err != nil {
				writeUsecaseErr(w, err)
				return
			}
			items := make([]transactionItemResp, 0, len(out.Data))
			for _, item := range out.Data {
				items = append(items, transactionItemResp(item))
			}
			writeJSON(w, http.StatusOK, transactionsListResp{
				Stats: transactionStatsResp(out.Stats),
				Data:  items,
				Total: out.Total,
			})
			return

		case r.Method == http.MethodGet && path == "broadcasts":
			page := parseInt32(r.URL.Query().Get("page"), 1)
			perPage := parseInt32(r.URL.Query().Get("per_page"), 20)
			out, err := uc.ListAdminBroadcasts(r.Context(), ucModels.AdminListBroadcastsInput{
				Page:    page,
				PerPage: perPage,
			})
			if err != nil {
				writeUsecaseErr(w, err)
				return
			}
			items := make([]broadcastHistoryItemResp, 0, len(out.Data))
			for _, item := range out.Data {
				items = append(items, broadcastHistoryItemResp(item))
			}
			writeJSON(w, http.StatusOK, broadcastsListResp{Data: items, Total: out.Total})
			return

		case r.Method == http.MethodGet && path == "settings":
			out, err := uc.GetAdminSettings(r.Context())
			if err != nil {
				writeUsecaseErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, out)
			return

		case r.Method == http.MethodPut && path == "settings":
			var req ucModels.AdminUpdateSettingsInput
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeErr(w, http.StatusBadRequest, "invalid json")
				return
			}
			out, err := uc.UpdateAdminSettings(r.Context(), req)
			if err != nil {
				writeUsecaseErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, out)
			return

		case r.Method == http.MethodGet && path == "billing/subscription-plans":
			out, err := uc.ListAdminSubscriptionPlans(r.Context())
			if err != nil {
				writeUsecaseErr(w, err)
				return
			}
			items := make([]subscriptionPlanItemResp, 0, len(out.Data))
			for _, item := range out.Data {
				items = append(items, subscriptionPlanItemResp(item))
			}
			writeJSON(w, http.StatusOK, subscriptionPlansListResp{Data: items})
			return

		case r.Method == http.MethodPut && path == "billing/subscription-plans":
			var req updateSubscriptionPlansReq
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeErr(w, http.StatusBadRequest, "invalid json")
				return
			}
			plans := make([]ucModels.SubscriptionPlanItem, 0, len(req.Data))
			for _, item := range req.Data {
				plans = append(plans, ucModels.SubscriptionPlanItem(item))
			}
			out, err := uc.UpdateAdminSubscriptionPlans(r.Context(), ucModels.AdminUpdateSubscriptionPlansInput{Data: plans})
			if err != nil {
				writeUsecaseErr(w, err)
				return
			}
			items := make([]subscriptionPlanItemResp, 0, len(out.Data))
			for _, item := range out.Data {
				items = append(items, subscriptionPlanItemResp(item))
			}
			writeJSON(w, http.StatusOK, subscriptionPlansListResp{Data: items})
			return

		case r.Method == http.MethodPost && path == "billing/subscription-plans/reset-defaults":
			out, err := uc.ResetAdminSubscriptionPlans(r.Context())
			if err != nil {
				writeUsecaseErr(w, err)
				return
			}
			items := make([]subscriptionPlanItemResp, 0, len(out.Data))
			for _, item := range out.Data {
				items = append(items, subscriptionPlanItemResp(item))
			}
			writeJSON(w, http.StatusOK, subscriptionPlansListResp{Data: items})
			return

		case r.Method == http.MethodGet && path == "billing/coin-packs":
			out, err := uc.ListAdminCoinPacks(r.Context())
			if err != nil {
				writeUsecaseErr(w, err)
				return
			}
			items := make([]coinPackItemResp, 0, len(out.Data))
			for _, item := range out.Data {
				items = append(items, coinPackItemResp(item))
			}
			writeJSON(w, http.StatusOK, coinPacksListResp{Data: items})
			return

		case r.Method == http.MethodPut && path == "billing/coin-packs":
			var req updateCoinPacksReq
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeErr(w, http.StatusBadRequest, "invalid json")
				return
			}
			packs := make([]ucModels.CoinPackItem, 0, len(req.Data))
			for _, item := range req.Data {
				packs = append(packs, ucModels.CoinPackItem(item))
			}
			out, err := uc.UpdateAdminCoinPacks(r.Context(), ucModels.AdminUpdateCoinPacksInput{Data: packs})
			if err != nil {
				writeUsecaseErr(w, err)
				return
			}
			items := make([]coinPackItemResp, 0, len(out.Data))
			for _, item := range out.Data {
				items = append(items, coinPackItemResp(item))
			}
			writeJSON(w, http.StatusOK, coinPacksListResp{Data: items})
			return

		case r.Method == http.MethodPost && path == "billing/coin-packs/reset-defaults":
			out, err := uc.ResetAdminCoinPacks(r.Context())
			if err != nil {
				writeUsecaseErr(w, err)
				return
			}
			items := make([]coinPackItemResp, 0, len(out.Data))
			for _, item := range out.Data {
				items = append(items, coinPackItemResp(item))
			}
			writeJSON(w, http.StatusOK, coinPacksListResp{Data: items})
			return

		case r.Method == http.MethodGet && path == "home-widgets":
			out, err := uc.ListAdminHomeWidgets(r.Context())
			if err != nil {
				writeUsecaseErr(w, err)
				return
			}
			items := make([]homeWidgetItemResp, 0, len(out.Data))
			for _, item := range out.Data {
				items = append(items, mapHomeWidgetItem(item))
			}
			writeJSON(w, http.StatusOK, homeWidgetsListResp{Data: items})
			return

		case r.Method == http.MethodPost && path == "home-widgets":
			var req createHomeWidgetReq
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeErr(w, http.StatusBadRequest, "invalid json")
				return
			}
			out, err := uc.CreateAdminHomeWidget(r.Context(), ucModels.AdminCreateHomeWidgetInput{
				SortOrder:       req.SortOrder,
				TagText:         req.TagText,
				TagBg:           req.TagBg,
				TagColor:        req.TagColor,
				Title:           req.Title,
				Description:     req.Description,
				BackgroundStyle: req.BackgroundStyle,
				ImageURL:        req.ImageURL,
				IsActive:        req.IsActive,
			})
			if err != nil {
				writeUsecaseErr(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, mapHomeWidgetItem(out.HomeWidgetItem))
			return

		case strings.HasPrefix(path, "home-widgets/"):
			idStr := strings.TrimPrefix(path, "home-widgets/")
			if idStr == "" || strings.Contains(idStr, "/") {
				writeErr(w, http.StatusNotFound, "not found")
				return
			}
			widgetID, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil || widgetID < 1 {
				writeErr(w, http.StatusBadRequest, "invalid widget id")
				return
			}
			switch r.Method {
			case http.MethodPatch:
				var req patchHomeWidgetReq
				if decodeErr := json.NewDecoder(r.Body).Decode(&req); decodeErr != nil {
					writeErr(w, http.StatusBadRequest, "invalid json")
					return
				}
				out, patchErr := uc.UpdateAdminHomeWidget(r.Context(), ucModels.AdminUpdateHomeWidgetInput{
					ID:              widgetID,
					SortOrder:       req.SortOrder,
					TagText:         req.TagText,
					TagBg:           req.TagBg,
					TagColor:        req.TagColor,
					Title:           req.Title,
					Description:     req.Description,
					BackgroundStyle: req.BackgroundStyle,
					ImageURL:        req.ImageURL,
					IsActive:        req.IsActive,
				})
				if patchErr != nil {
					writeUsecaseErr(w, patchErr)
					return
				}
				writeJSON(w, http.StatusOK, mapHomeWidgetItem(out.HomeWidgetItem))
				return
			case http.MethodDelete:
				if delErr := uc.DeleteAdminHomeWidget(r.Context(), widgetID); delErr != nil {
					writeUsecaseErr(w, delErr)
					return
				}
				w.WriteHeader(http.StatusNoContent)
				return
			default:
				writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}

		case r.Method == http.MethodGet && path == "models":
			out, err := uc.ListAdminModels(r.Context())
			if err != nil {
				writeUsecaseErr(w, err)
				return
			}
			items := make([]modelItemResp, 0, len(out.Data))
			for _, item := range out.Data {
				items = append(items, modelItemResp(item))
			}
			writeJSON(w, http.StatusOK, modelsListResp{Data: items})
			return

		case strings.HasPrefix(path, "models/"):
			modelID := strings.TrimPrefix(path, "models/")
			if modelID == "" || strings.Contains(modelID, "/") {
				writeErr(w, http.StatusNotFound, "not found")
				return
			}
			if r.Method != http.MethodPatch {
				writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
			var req patchModelReq
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeErr(w, http.StatusBadRequest, "invalid json")
				return
			}
			out, err := uc.UpdateAdminModel(r.Context(), ucModels.AdminUpdateModelInput{
				ModelID: modelID,
				Price:   req.Price,
				Enabled: req.Enabled,
			})
			if err != nil {
				writeUsecaseErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, modelItemResp(out))
			return

		case r.Method == http.MethodGet && path == "users":
			page := parseInt32(r.URL.Query().Get("page"), 1)
			perPage := parseInt32(r.URL.Query().Get("per_page"), 20)
			out, err := uc.ListAdminUsers(r.Context(), ucModels.AdminListUsersInput{
				Page:    page,
				PerPage: perPage,
				Search:  r.URL.Query().Get("search"),
			})
			if err != nil {
				writeUsecaseErr(w, err)
				return
			}
			items := make([]userResp, 0, len(out.Data))
			for _, u := range out.Data {
				items = append(items, mapUser(u))
			}
			writeJSON(w, http.StatusOK, usersListResp{Data: items, Total: out.Total})
			return

		case r.Method == http.MethodPost && path == "broadcast":
			var req broadcastReq
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeErr(w, http.StatusBadRequest, "invalid json")
				return
			}
			out, err := uc.AdminBroadcast(r.Context(), ucModels.AdminBroadcastInput{
				AdminID:   adminID,
				Message:   req.Message,
				Target:    req.Target,
				ParseMode: req.ParseMode,
			}, messenger)
			if err != nil {
				writeUsecaseErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, broadcastResp{Sent: out.Sent, Failed: out.Failed})
			return

		case strings.HasPrefix(path, "users/"):
			userID, tail, err := parseUserPath(path)
			if err != nil {
				writeErr(w, http.StatusBadRequest, "invalid user id")
				return
			}

			switch tail {
			case "tokens/credit":
				if r.Method != http.MethodPost {
					writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
					return
				}
				var req tokenChangeReq
				if decErr := json.NewDecoder(r.Body).Decode(&req); decErr != nil {
					writeErr(w, http.StatusBadRequest, "invalid json")
					return
				}
				if msg := validateTokenChangeReq(req.Amount, req.Reason); msg != "" {
					writeErr(w, http.StatusUnprocessableEntity, msg)
					return
				}
				out, creditErr := uc.AdminCreditTokens(r.Context(), ucModels.AdminTokenChangeInput{
					UserID:  userID,
					AdminID: adminID,
					Amount:  req.Amount,
					Reason:  req.Reason,
				})
				if creditErr != nil {
					writeUsecaseErr(w, creditErr)
					return
				}
				writeJSON(w, http.StatusOK, mapTokenChange(out))
				return

			case "tokens/debit":
				if r.Method != http.MethodPost {
					writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
					return
				}
				var req tokenChangeReq
				if decErr := json.NewDecoder(r.Body).Decode(&req); decErr != nil {
					writeErr(w, http.StatusBadRequest, "invalid json")
					return
				}
				if msg := validateTokenChangeReq(req.Amount, req.Reason); msg != "" {
					writeErr(w, http.StatusUnprocessableEntity, msg)
					return
				}
				out, debitErr := uc.AdminDebitTokens(r.Context(), ucModels.AdminTokenChangeInput{
					UserID:  userID,
					AdminID: adminID,
					Amount:  req.Amount,
					Reason:  req.Reason,
				})
				if debitErr != nil {
					writeUsecaseErr(w, debitErr)
					return
				}
				writeJSON(w, http.StatusOK, mapTokenChange(out))
				return

			case "subscription":
				switch r.Method {
				case http.MethodPost:
					var req setSubscriptionReq
					if decErr := json.NewDecoder(r.Body).Decode(&req); decErr != nil {
						writeErr(w, http.StatusBadRequest, "invalid json")
						return
					}
					input := ucModels.AdminSetSubscriptionInput{
						UserID:       userID,
						AdminID:      adminID,
						PlanID:       req.PlanID,
						DurationDays: req.DurationDays,
						NoExpiry:     req.NoExpiry,
						GrantCoins:   req.GrantCoins,
					}
					if req.ExpiresAt != nil && strings.TrimSpace(*req.ExpiresAt) != "" {
						parsed, parseErr := time.Parse(time.RFC3339, strings.TrimSpace(*req.ExpiresAt))
						if parseErr != nil {
							writeErr(w, http.StatusUnprocessableEntity, "expires_at must be RFC3339")
							return
						}
						input.ExpiresAt = &parsed
					}
					out, setErr := uc.AdminSetUserSubscription(r.Context(), input)
					if setErr != nil {
						writeUsecaseErr(w, setErr)
						return
					}
					writeJSON(w, http.StatusOK, subscriptionChangeResp{
						User:         mapUser(out.User),
						Subscription: mapSubscriptionState(out.Subscription),
						CoinsGranted: out.CoinsGranted,
					})
					return
				case http.MethodDelete:
					out, clearErr := uc.AdminClearUserSubscription(r.Context(), userID)
					if clearErr != nil {
						writeUsecaseErr(w, clearErr)
						return
					}
					writeJSON(w, http.StatusOK, subscriptionChangeResp{
						User:         mapUser(out.User),
						Subscription: mapSubscriptionState(out.Subscription),
					})
					return
				default:
					writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
					return
				}

			case "":
				switch r.Method {
				case http.MethodGet:
					out, getErr := uc.GetAdminUser(r.Context(), userID)
					if getErr != nil {
						writeUsecaseErr(w, getErr)
						return
					}
					writeJSON(w, http.StatusOK, mapUser(out))
					return

				case http.MethodPatch:
					var req patchUserReq
					if decErr := json.NewDecoder(r.Body).Decode(&req); decErr != nil {
						writeErr(w, http.StatusBadRequest, "invalid json")
						return
					}
					out, patchErr := uc.UpdateAdminUserActive(r.Context(), ucModels.AdminUpdateUserInput{
						UserID:   userID,
						IsActive: req.IsActive,
					})
					if patchErr != nil {
						writeUsecaseErr(w, patchErr)
						return
					}
					writeJSON(w, http.StatusOK, mapUser(out))
					return

				case http.MethodDelete:
					if delErr := uc.DeleteAdminUser(r.Context(), userID); delErr != nil {
						writeUsecaseErr(w, delErr)
						return
					}
					w.WriteHeader(http.StatusNoContent)
					return
				}
			}
		}

		writeErr(w, http.StatusNotFound, "not found")
	})
}

const maxTokenAmount int64 = 1_000_000_000

func validateTokenChangeReq(amount int64, reason string) string {
	if amount <= 0 {
		return "amount must be greater than 0"
	}
	if amount > maxTokenAmount {
		return "amount is too large"
	}
	if len(strings.TrimSpace(reason)) > 255 {
		return "reason is too long"
	}
	return ""
}

func parseUserPath(path string) (userID int64, tail string, err error) {
	rest := strings.TrimPrefix(path, "users/")
	parts := strings.SplitN(rest, "/", 2)
	if len(parts) == 0 || parts[0] == "" {
		return 0, "", errors.New("empty")
	}
	userID, err = strconv.ParseInt(parts[0], 10, 64)
	if err != nil || userID <= 0 {
		return 0, "", errors.New("invalid")
	}
	if len(parts) == 2 {
		tail = parts[1]
	}
	return userID, tail, nil
}

func mapUser(u ucModels.AdminUserItem) userResp {
	return userResp{
		ID:                   u.ID,
		TelegramID:           u.TelegramID,
		Username:             u.Username,
		FirstName:            u.FirstName,
		LastName:             u.LastName,
		IsActive:             u.IsActive,
		Tokens:               u.Tokens,
		CreatedAt:            u.CreatedAt,
		SubscriptionPlanID:   u.SubscriptionPlanID,
		SubscriptionPlan:     u.SubscriptionPlan,
		SubscriptionExpires:  u.SubscriptionExpires,
		SubscriptionDaysLeft: u.SubscriptionDaysLeft,
	}
}

func mapSubscriptionState(s ucModels.SubscriptionStateOutput) subscriptionStateResp {
	return subscriptionStateResp{
		PlanID:       s.PlanID,
		PlanName:     s.PlanName,
		PlanRank:     s.PlanRank,
		IsPaid:       s.IsPaid,
		Coins:        s.Coins,
		StartedAt:    s.StartedAt,
		ExpiresAt:    s.ExpiresAt,
		DaysLeft:     s.DaysLeft,
		HoursLeft:    s.HoursLeft,
		IsActive:     s.IsActive,
		ExpiringSoon: s.ExpiringSoon,
		Expired:      s.Expired,
	}
}

func mapTokenChange(o ucModels.AdminTokenChangeOutput) tokenChangeResp {
	return tokenChangeResp{
		UserID:    o.UserID,
		Tokens:    o.Tokens,
		Delta:     o.Delta,
		Operation: o.Operation,
	}
}

func requireAdmin(w http.ResponseWriter, r *http.Request, jwtCfg config.ConfigJWT) (int64, bool) {
	h := strings.TrimSpace(r.Header.Get("Authorization"))
	if h == "" {
		writeErr(w, http.StatusUnauthorized, "missing authorization header")
		return 0, false
	}
	lh := strings.ToLower(h)
	if !strings.HasPrefix(lh, bearerPrefix) {
		writeErr(w, http.StatusUnauthorized, "invalid authorization header")
		return 0, false
	}
	token := strings.TrimSpace(h[len(bearerPrefix):])
	if token == "" {
		writeErr(w, http.StatusUnauthorized, "missing token")
		return 0, false
	}
	claims, err := jwtutil.ParseAdminToken(jwtCfg.Secret, token)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "invalid token")
		return 0, false
	}
	return claims.AdminID, true
}

func mapHomeWidgetItem(item ucModels.HomeWidgetItem) homeWidgetItemResp {
	return homeWidgetItemResp{
		ID:              item.ID,
		SortOrder:       item.SortOrder,
		TagText:         item.TagText,
		TagBg:           item.TagBg,
		TagColor:        item.TagColor,
		Title:           item.Title,
		Description:     item.Description,
		BackgroundStyle: item.BackgroundStyle,
		ImageURL:        item.ImageURL,
		IsActive:        item.IsActive,
	}
}

func writeUsecaseErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ucModels.ErrInvalidInput):
		writeErr(w, http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, ucModels.ErrInvalidCredentials):
		writeErr(w, http.StatusUnauthorized, "invalid credentials")
	case errors.Is(err, ucModels.ErrAdminNotFound):
		writeErr(w, http.StatusNotFound, "admin not found")
	case errors.Is(err, ucModels.ErrAdminUserNotFound):
		writeErr(w, http.StatusNotFound, "user not found")
	case errors.Is(err, ucModels.ErrHomeWidgetNotFound):
		writeErr(w, http.StatusNotFound, "widget not found")
	case errors.Is(err, ucModels.ErrInsufficientTokens):
		writeErr(w, http.StatusUnprocessableEntity, "Insufficient tokens")
	case errors.Is(err, ucModels.ErrBroadcastNotReady):
		writeErr(w, http.StatusServiceUnavailable, "telegram bot is not configured")
	default:
		writeErr(w, http.StatusInternalServerError, "internal error")
	}
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorResp{Error: msg})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func parseInt32(s string, def int32) int32 {
	if s == "" {
		return def
	}
	v, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return def
	}
	return int32(v)
}
