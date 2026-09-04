package siteapi

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

const bearerPrefix = "bearer "

type envelope[T any] struct {
	Data T `json:"data"`
}

type errorEnvelope struct {
	Error string `json:"error"`
}

type authReq struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName"`
}

type authResp struct {
	AccessToken string `json:"accessToken"`
}

type meResp struct {
	ID          int64  `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
}

type promptCreateReq struct {
	Prompt   string `json:"prompt"`
	Category string `json:"category"`
	Model    string `json:"model"`
}

type promptCreateResp struct {
	ID int64 `json:"id"`
}

type promptItem struct {
	ID       int64  `json:"id"`
	Prompt   string `json:"prompt"`
	Category string `json:"category"`
	Model    string `json:"model"`
}

type promptsListResp struct {
	Items []promptItem `json:"items"`
}

type modelItem struct {
	Slug string `json:"slug"`
	Type string `json:"type"` // text | image
	Name string `json:"name"`
}

type modelsResp struct {
	Items []modelItem `json:"items"`
}

// Wrap adds "site" REST endpoints for the web app.
// All endpoints are under /v1/site/* to not conflict with existing Telegram API.
func Wrap(next http.Handler, uc internal.UseCase, jwtCfg config.ConfigJWT) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/v1/site/") {
			next.ServeHTTP(w, r)
			return
		}

		// Every site-auth route (register/login/me/prompts) either signs or
		// verifies a JWT. The default JWT_SECRET is a well-known placeholder
		// ("your-super-secret-jwt-key-change-this"): RegisterWebAccount and
		// LoginWebAccount only refused an *empty* secret, so a deployment
		// that forgot to set JWT_SECRET happily minted tokens anyone could
		// forge. Refuse in production the same way the admin panel already
		// does, but keep the public model catalog available.
		if r.URL.Path != "/v1/site/models" && config.JWTSecretWeak(jwtCfg.Secret) && config.IsDeployedProduction() {
			writeErr(w, http.StatusServiceUnavailable, "web auth is not configured")
			return
		}

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/site/auth/register":
			var req authReq
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeErr(w, http.StatusBadRequest, "invalid json")
				return
			}
			out, err := uc.RegisterWebAccount(r.Context(), ucModels.RegisterWebAccountInput{
				Email:       req.Email,
				Password:    req.Password,
				DisplayName: req.DisplayName,
			})
			if err != nil {
				writeUsecaseErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, envelope[authResp]{Data: authResp{AccessToken: out.AccessToken}})
			return

		case r.Method == http.MethodPost && r.URL.Path == "/v1/site/auth/login":
			var req authReq
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeErr(w, http.StatusBadRequest, "invalid json")
				return
			}
			out, err := uc.LoginWebAccount(r.Context(), ucModels.LoginWebAccountInput{
				Email:    req.Email,
				Password: req.Password,
			})
			if err != nil {
				writeUsecaseErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, envelope[authResp]{Data: authResp{AccessToken: out.AccessToken}})
			return

		case r.Method == http.MethodGet && r.URL.Path == "/v1/site/auth/me":
			webAccountID, ok := requireAuth(w, r, jwtCfg)
			if !ok {
				return
			}
			out, err := uc.GetWebAccount(r.Context(), webAccountID)
			if err != nil {
				writeUsecaseErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, envelope[meResp]{Data: meResp{
				ID:          out.Data.ID,
				Email:       out.Data.Email,
				DisplayName: out.Data.DisplayName,
			}})
			return

		case r.Method == http.MethodPost && r.URL.Path == "/v1/site/prompts":
			webAccountID, ok := requireAuth(w, r, jwtCfg)
			if !ok {
				return
			}
			var req promptCreateReq
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeErr(w, http.StatusBadRequest, "invalid json")
				return
			}
			out, err := uc.CreateWebPrompt(r.Context(), ucModels.CreateWebPromptInput{
				WebAccountID: webAccountID,
				Prompt:       req.Prompt,
				Category:     req.Category,
				Model:        req.Model,
			})
			if err != nil {
				writeUsecaseErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, envelope[promptCreateResp]{Data: promptCreateResp{ID: out.ID}})
			return

		case r.Method == http.MethodGet && r.URL.Path == "/v1/site/prompts":
			webAccountID, ok := requireAuth(w, r, jwtCfg)
			if !ok {
				return
			}
			limit := parseInt32(r.URL.Query().Get("limit"), 50)
			offset := parseInt32(r.URL.Query().Get("offset"), 0)
			out, err := uc.ListWebPrompts(r.Context(), ucModels.ListWebPromptsInput{
				WebAccountID: webAccountID,
				Limit:        limit,
				Offset:       offset,
			})
			if err != nil {
				writeUsecaseErr(w, err)
				return
			}
			items := make([]promptItem, 0, len(out.Items))
			for _, p := range out.Items {
				items = append(items, promptItem{
					ID:       p.ID,
					Prompt:   p.Prompt,
					Category: p.Category,
					Model:    p.Model,
				})
			}
			writeJSON(w, http.StatusOK, envelope[promptsListResp]{Data: promptsListResp{Items: items}})
			return

		case r.Method == http.MethodGet && r.URL.Path == "/v1/site/models":
			// Backend-driven catalog for the site homepage ("all neural nets from the app").
			writeJSON(w, http.StatusOK, envelope[modelsResp]{Data: modelsResp{Items: []modelItem{
				{Slug: "yandexgpt", Type: "text", Name: "YandexGPT"},
				{Slug: "deepseek-v4-flash", Type: "text", Name: "DeepSeek"},
				{Slug: "gemini-flash", Type: "text", Name: "Gemini Flash"},
				{Slug: "openai", Type: "text", Name: "OpenAI"},
				{Slug: "nano-banana", Type: "image", Name: "Nano Banana (Gemini)"},
				{Slug: "alice-ai-art", Type: "image", Name: "Alice AI ART (Yandex)"},
			}}})
			return
		}

		writeErr(w, http.StatusNotFound, "not found")
	})
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

func requireAuth(w http.ResponseWriter, r *http.Request, jwtCfg config.ConfigJWT) (int64, bool) {
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
	claims, err := jwtutil.ParseAccessToken(jwtCfg.Secret, token)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "invalid token")
		return 0, false
	}
	if claims.WebAccountID <= 0 {
		writeErr(w, http.StatusUnauthorized, "invalid token subject")
		return 0, false
	}
	return claims.WebAccountID, true
}

func writeUsecaseErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ucModels.ErrInvalidInput):
		writeErr(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, ucModels.ErrWebAccountAlreadyExists):
		writeErr(w, http.StatusConflict, "account already exists")
	case errors.Is(err, ucModels.ErrInvalidCredentials):
		writeErr(w, http.StatusUnauthorized, "invalid credentials")
	case errors.Is(err, ucModels.ErrWebAccountNotFound):
		writeErr(w, http.StatusNotFound, "account not found")
	default:
		writeErr(w, http.StatusInternalServerError, "internal error")
	}
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorEnvelope{Error: msg})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

