package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLimiter_AllowWithinLimit(t *testing.T) {
	l := newLimiter()
	now := time.Now()
	for i := 0; i < 3; i++ {
		if !l.allow("k", 3, time.Minute, now) {
			t.Fatalf("request %d should be allowed", i)
		}
	}
	if l.allow("k", 3, time.Minute, now) {
		t.Fatal("4th request should be rate limited")
	}
}

func TestLimiter_WindowExpires(t *testing.T) {
	l := newLimiter()
	base := time.Now()
	if !l.allow("k", 1, time.Minute, base) {
		t.Fatal("first request should be allowed")
	}
	if l.allow("k", 1, time.Minute, base.Add(30*time.Second)) {
		t.Fatal("second request within window should be blocked")
	}
	if !l.allow("k", 1, time.Minute, base.Add(61*time.Second)) {
		t.Fatal("request after window elapses should be allowed again")
	}
}

func TestLimiter_KeysAreIndependent(t *testing.T) {
	l := newLimiter()
	now := time.Now()
	if !l.allow("a", 1, time.Minute, now) {
		t.Fatal("first key should be allowed")
	}
	if !l.allow("b", 1, time.Minute, now) {
		t.Fatal("distinct key should not share the bucket")
	}
}

func TestWrap_BlocksAfterLimit(t *testing.T) {
	rules := []Rule{{Method: http.MethodPost, Path: "/v1/register", Limit: 2, Window: time.Minute}}
	var calls int
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(http.StatusOK)
	})
	handler := Wrap(next, rules)

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/v1/register", nil)
		req.RemoteAddr = "203.0.113.5:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d, want 200", i, rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/register", nil)
	req.RemoteAddr = "203.0.113.5:1234"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429 after exceeding limit", rec.Code)
	}
	if calls != 2 {
		t.Fatalf("next handler called %d times, want 2", calls)
	}
}

func TestWrap_DifferentIPsHaveSeparateLimits(t *testing.T) {
	rules := []Rule{{Method: http.MethodPost, Path: "/v1/register", Limit: 1, Window: time.Minute}}
	handler := Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), rules)

	req1 := httptest.NewRequest(http.MethodPost, "/v1/register", nil)
	req1.RemoteAddr = "203.0.113.1:1"
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Fatalf("ip1 first request status = %d, want 200", rec1.Code)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/v1/register", nil)
	req2.RemoteAddr = "203.0.113.2:1"
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("ip2 first request status = %d, want 200 (independent bucket)", rec2.Code)
	}
}

func TestWrap_UnmatchedPathsPassThrough(t *testing.T) {
	rules := []Rule{{Method: http.MethodPost, Path: "/v1/register", Limit: 0, Window: time.Minute}}
	handler := Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), rules)

	req := httptest.NewRequest(http.MethodGet, "/v1/other", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 for a path with no matching rule", rec.Code)
	}
}

func TestClientIP_PrefersForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:5555"
	req.Header.Set("X-Forwarded-For", "203.0.113.9, 10.0.0.1")

	if got := ClientIP(req); got != "203.0.113.9" {
		t.Fatalf("ClientIP() = %q, want 203.0.113.9", got)
	}
}

func TestClientIP_FallsBackToRemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "198.51.100.7:4321"

	if got := ClientIP(req); got != "198.51.100.7" {
		t.Fatalf("ClientIP() = %q, want 198.51.100.7", got)
	}
}
