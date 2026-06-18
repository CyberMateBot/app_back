package adminapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

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
	ID         int64  `json:"id"`
	TelegramID int64  `json:"telegram_id"`
	Username   string `json:"username"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	IsActive   bool   `json:"is_active"`
	CreatedAt  string `json:"created_at"`
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

		case r.Method == http.MethodPost && path == "auth/logout":
			writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
			return
		}

		adminID, ok := requireAdmin(w, r, jwtCfg)
		if !ok {
			return
		}
		_ = adminID

		switch {
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
			if tail != "" {
				writeErr(w, http.StatusNotFound, "not found")
				return
			}

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

		writeErr(w, http.StatusNotFound, "not found")
	})
}

func parseUserPath(path string) (userID int64, tail string, err error) {
	rest := strings.TrimPrefix(path, "users/")
	parts := strings.SplitN(rest, "/", 2)
	if len(parts) == 0 || parts[0] == "" {
		return 0, "", errors.New("empty")
	}
	userID, err = strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, "", err
	}
	if len(parts) == 2 {
		tail = parts[1]
	}
	return userID, tail, nil
}

func mapUser(u ucModels.AdminUserItem) userResp {
	return userResp{
		ID:         u.ID,
		TelegramID: u.TelegramID,
		Username:   u.Username,
		FirstName:  u.FirstName,
		LastName:   u.LastName,
		IsActive:   u.IsActive,
		CreatedAt:  u.CreatedAt,
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

func writeUsecaseErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ucModels.ErrInvalidInput):
		writeErr(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, ucModels.ErrInvalidCredentials):
		writeErr(w, http.StatusUnauthorized, "invalid credentials")
	case errors.Is(err, ucModels.ErrAdminNotFound):
		writeErr(w, http.StatusNotFound, "admin not found")
	case errors.Is(err, ucModels.ErrAdminUserNotFound):
		writeErr(w, http.StatusNotFound, "user not found")
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
