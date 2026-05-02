package middlewares

import (
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	adminLoginMaxAttempts = 5
	adminLoginWindow      = time.Minute
)

type loginRateLimiter struct {
	mu          sync.Mutex
	attempts    map[string]loginAttempts
	maxAttempts int
	window      time.Duration
	now         func() time.Time
}

type loginAttempts struct {
	count   int
	resetAt time.Time
}

func NewAdminLoginRateLimiter() func(http.Handler) http.Handler {
	return newLoginRateLimiter(adminLoginMaxAttempts, adminLoginWindow, time.Now)
}

func newLoginRateLimiter(maxAttempts int, window time.Duration, now func() time.Time) func(http.Handler) http.Handler {
	limiter := &loginRateLimiter{
		attempts:    make(map[string]loginAttempts),
		maxAttempts: maxAttempts,
		window:      window,
		now:         now,
	}

	return limiter.middleware
}

func (l *loginRateLimiter) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowed, retryAfter := l.allow(clientIP(r.RemoteAddr))
		if !allowed {
			w.Header().Set("Retry-After", strconv.Itoa(int(retryAfter.Seconds())))
			http.Error(w, "too many login attempts", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (l *loginRateLimiter) allow(key string) (bool, time.Duration) {
	now := l.now()

	l.mu.Lock()
	defer l.mu.Unlock()

	attempts := l.attempts[key]
	if attempts.resetAt.IsZero() || !now.Before(attempts.resetAt) {
		l.attempts[key] = loginAttempts{
			count:   1,
			resetAt: now.Add(l.window),
		}
		return true, 0
	}

	if attempts.count >= l.maxAttempts {
		return false, attempts.resetAt.Sub(now)
	}

	attempts.count++
	l.attempts[key] = attempts

	return true, 0
}

func clientIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err == nil && host != "" {
		return host
	}

	if remoteAddr == "" {
		return "unknown"
	}

	return remoteAddr
}
