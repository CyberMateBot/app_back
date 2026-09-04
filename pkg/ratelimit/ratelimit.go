// Package ratelimit provides a minimal in-memory, per-client-IP sliding
// window rate limiter for a small set of high-value, otherwise
// unauthenticated or brute-forceable HTTP endpoints (admin login, account
// registration, unauthenticated media-prepare, payment webhooks). It is
// intentionally simple (no Redis dependency) since a single-instance
// in-memory limiter is enough to blunt casual abuse and brute-forcing; if
// the service is ever scaled horizontally this should move to a shared
// store.
package ratelimit

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Rule rate-limits one exact (method, path) pair to `Limit` requests per
// `Window`, per client IP.
type Rule struct {
	Method string
	Path   string
	Limit  int
	Window time.Duration
}

type bucket struct {
	hits []time.Time
}

// Limiter tracks per-key sliding windows of recent request timestamps.
type Limiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
}

func newLimiter() *Limiter {
	return &Limiter{buckets: make(map[string]*bucket)}
}

// allow reports whether another request for key is permitted, given at most
// `limit` requests are allowed per `window`. It also prunes timestamps older
// than window so buckets don't grow unbounded for a single busy key.
func (l *Limiter) allow(key string, limit int, window time.Duration, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	b, ok := l.buckets[key]
	if !ok {
		b = &bucket{}
		l.buckets[key] = b
	}

	cutoff := now.Add(-window)
	kept := b.hits[:0]
	for _, t := range b.hits {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	b.hits = kept

	if len(b.hits) >= limit {
		return false
	}
	b.hits = append(b.hits, now)
	return true
}

// sweep removes buckets that have had no activity for a while, bounding
// memory usage from the ever-growing set of distinct client IPs.
func (l *Limiter) sweep(maxAge time.Duration, now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for key, b := range l.buckets {
		if len(b.hits) == 0 || now.Sub(b.hits[len(b.hits)-1]) > maxAge {
			delete(l.buckets, key)
		}
	}
}

func (l *Limiter) startSweeper() {
	ticker := time.NewTicker(10 * time.Minute)
	go func() {
		for now := range ticker.C {
			l.sweep(30*time.Minute, now)
		}
	}()
}

// Wrap enforces `rules` on top of `next`. Only requests matching a rule's
// exact method+path are rate-limited; everything else passes straight
// through. At most one rule applies per request (first match wins).
func Wrap(next http.Handler, rules []Rule) http.Handler {
	if len(rules) == 0 {
		return next
	}

	limiter := newLimiter()
	limiter.startSweeper()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, rule := range rules {
			if r.Method != rule.Method || r.URL.Path != rule.Path {
				continue
			}
			key := rule.Method + " " + rule.Path + "|" + ClientIP(r)
			if !limiter.allow(key, rule.Limit, rule.Window, time.Now()) {
				w.Header().Set("Retry-After", strconv.Itoa(int(rule.Window.Seconds())))
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(`{"error":"too many requests, please slow down"}`))
				return
			}
			break
		}
		next.ServeHTTP(w, r)
	})
}

// ClientIP best-effort resolves the caller's IP, preferring the first hop in
// X-Forwarded-For (set by Railway's edge proxy in front of this service)
// over RemoteAddr, which would otherwise always be the proxy's own address.
func ClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if first := strings.TrimSpace(parts[0]); first != "" {
			return first
		}
	}
	if xrip := strings.TrimSpace(r.Header.Get("X-Real-IP")); xrip != "" {
		return xrip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
