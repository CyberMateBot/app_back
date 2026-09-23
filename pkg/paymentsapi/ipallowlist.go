package paymentsapi

import (
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/twelvepills-936/tgapp-/pkg/ratelimit"
)

// yookassaCIDRs is the official list of YooKassa server IP ranges from which
// webhook notifications are sent. Requests originating from any other address
// are rejected before the handler even reads the body.
//
// Source: https://yookassa.ru/developers/using-api/webhooks#ip
var yookassaCIDRs = func() []*net.IPNet {
	prefixes := []string{
		"185.71.76.0/27",
		"185.71.77.0/27",
		"77.75.153.0/25",
		"77.75.154.128/25",
		"77.75.156.11/32",
		"77.75.156.35/32",
	}
	nets := make([]*net.IPNet, 0, len(prefixes))
	for _, p := range prefixes {
		_, cidr, err := net.ParseCIDR(p)
		if err != nil {
			panic("paymentsapi: invalid YooKassa CIDR " + p + ": " + err.Error())
		}
		nets = append(nets, cidr)
	}
	return nets
}()

// isYooKassaIP reports whether ip belongs to one of the known YooKassa
// notification server ranges.
func isYooKassaIP(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	for _, cidr := range yookassaCIDRs {
		if cidr.Contains(parsed) {
			return true
		}
	}
	return false
}

// webhookIPCheckDisabled returns true when YOOKASSA_WEBHOOK_SKIP_IP_CHECK=true,
// which allows curl/local testing without spoofing a YooKassa IP.
// Must NEVER be set in production.
func webhookIPCheckDisabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("YOOKASSA_WEBHOOK_SKIP_IP_CHECK"))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

// webhookAllowlistMiddleware wraps h and rejects requests to the YooKassa
// webhook path that do not originate from a known YooKassa IP range.
func webhookAllowlistMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == pathYooKassaWebhook {
			if !webhookIPCheckDisabled() {
				clientIP := ratelimit.ClientIP(r)
				if !isYooKassaIP(clientIP) {
					slog.Warn("yookassa webhook: request rejected from non-YooKassa IP",
						slog.String("ip", clientIP))
					writeJSON(w, http.StatusForbidden, map[string]string{
						"error": "forbidden",
					})
					return
				}
			}
		}
		h.ServeHTTP(w, r)
	})
}
