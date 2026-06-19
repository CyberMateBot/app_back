package cors

import (
	"net/http"
	"strings"

	"github.com/twelvepills-936/tgapp-/pkg/config"
)

// Wrap adds CORS headers for browser clients (Telegram Mini App / Vite dev server).
func Wrap(next http.Handler, cfg config.ConfigCORS) http.Handler {
	allowed := normalizeOrigins(cfg.AllowedOrigins)
	allowAll := config.CORSAllowsAll(cfg.AllowedOrigins) || config.CORSAllowsAll(allowed)
	methods := strings.Join(cfg.AllowedMethods, ", ")
	headers := strings.Join(cfg.AllowedHeaders, ", ")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if allowAll {
			w.Header().Set("Access-Control-Allow-Origin", pickAllowOrigin(origin, true))
			w.Header().Add("Vary", "Origin")
		} else if origin != "" && originAllowed(origin, allowed) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Add("Vary", "Origin")
		}

		if allowAll || origin != "" {
			w.Header().Set("Access-Control-Allow-Methods", methods)
			w.Header().Set("Access-Control-Allow-Headers", headers)
			w.Header().Set("Access-Control-Max-Age", "86400")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func normalizeOrigin(origin string) string {
	return strings.TrimSuffix(strings.TrimSpace(origin), "/")
}

func normalizeOrigins(origins []string) []string {
	out := make([]string, 0, len(origins))
	for _, o := range origins {
		o = normalizeOrigin(o)
		if o != "" && o != "*" {
			out = append(out, o)
		}
	}
	if len(out) == 0 {
		return []string{"*"}
	}
	return out
}

func originAllowed(origin string, allowed []string) bool {
	origin = normalizeOrigin(origin)
	for _, a := range allowed {
		if normalizeOrigin(a) == origin {
			return true
		}
	}
	return false
}

func pickAllowOrigin(requestOrigin string, allowAll bool) string {
	if allowAll {
		if requestOrigin != "" && requestOrigin != "null" {
			return requestOrigin
		}
		return "*"
	}
	return requestOrigin
}
