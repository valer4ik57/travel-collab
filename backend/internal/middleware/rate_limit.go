package middleware

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type rateLimitEntry struct {
	Count     int
	ResetTime time.Time
}

type RateLimiter struct {
	limit  int
	window time.Duration
	mu     sync.Mutex
	items  map[string]rateLimitEntry
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	if limit < 1 {
		limit = 1
	}
	if window <= 0 {
		window = time.Minute
	}
	return &RateLimiter{
		limit:  limit,
		window: window,
		items:  make(map[string]rateLimitEntry),
	}
}

func (l *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := clientIP(r) + " " + r.Method + " " + r.URL.Path
		now := time.Now()

		l.mu.Lock()
		entry := l.items[key]
		if entry.ResetTime.IsZero() || now.After(entry.ResetTime) {
			entry = rateLimitEntry{Count: 0, ResetTime: now.Add(l.window)}
		}
		entry.Count++
		l.items[key] = entry
		allowed := entry.Count <= l.limit
		retryAfter := time.Until(entry.ResetTime)
		l.cleanupLocked(now)
		l.mu.Unlock()

		if !allowed {
			if retryAfter < 0 {
				retryAfter = l.window
			}
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", strconv.FormatInt(int64(retryAfter.Round(time.Second).Seconds()), 10))
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "too many requests, try again later"})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (l *RateLimiter) cleanupLocked(now time.Time) {
	if len(l.items) < 1024 {
		return
	}
	for key, entry := range l.items {
		if now.After(entry.ResetTime) {
			delete(l.items, key)
		}
	}
}

func clientIP(r *http.Request) string {
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		if ip := strings.TrimSpace(parts[0]); ip != "" {
			return ip
		}
	}
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}
