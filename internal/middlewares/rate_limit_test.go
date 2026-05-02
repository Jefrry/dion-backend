package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAdminLoginRateLimiter(t *testing.T) {
	now := time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		requests   []rateLimitRequest
		wantStatus []int
		wantCalled int
	}{
		{
			name: "limits sixth attempt in window",
			requests: []rateLimitRequest{
				{remoteAddr: "192.0.2.1:1234", at: now},
				{remoteAddr: "192.0.2.1:1234", at: now.Add(time.Second)},
				{remoteAddr: "192.0.2.1:1234", at: now.Add(2 * time.Second)},
				{remoteAddr: "192.0.2.1:1234", at: now.Add(3 * time.Second)},
				{remoteAddr: "192.0.2.1:1234", at: now.Add(4 * time.Second)},
				{remoteAddr: "192.0.2.1:1234", at: now.Add(5 * time.Second)},
			},
			wantStatus: []int{
				http.StatusNoContent,
				http.StatusNoContent,
				http.StatusNoContent,
				http.StatusNoContent,
				http.StatusNoContent,
				http.StatusTooManyRequests,
			},
			wantCalled: 5,
		},
		{
			name: "uses independent client IP buckets",
			requests: []rateLimitRequest{
				{remoteAddr: "192.0.2.1:1234", at: now},
				{remoteAddr: "192.0.2.1:1234", at: now.Add(time.Second)},
				{remoteAddr: "192.0.2.1:1234", at: now.Add(2 * time.Second)},
				{remoteAddr: "192.0.2.1:1234", at: now.Add(3 * time.Second)},
				{remoteAddr: "192.0.2.1:1234", at: now.Add(4 * time.Second)},
				{remoteAddr: "198.51.100.10:1234", at: now.Add(5 * time.Second)},
			},
			wantStatus: []int{
				http.StatusNoContent,
				http.StatusNoContent,
				http.StatusNoContent,
				http.StatusNoContent,
				http.StatusNoContent,
				http.StatusNoContent,
			},
			wantCalled: 6,
		},
		{
			name: "resets after window",
			requests: []rateLimitRequest{
				{remoteAddr: "192.0.2.1:1234", at: now},
				{remoteAddr: "192.0.2.1:1234", at: now.Add(time.Second)},
				{remoteAddr: "192.0.2.1:1234", at: now.Add(2 * time.Second)},
				{remoteAddr: "192.0.2.1:1234", at: now.Add(3 * time.Second)},
				{remoteAddr: "192.0.2.1:1234", at: now.Add(4 * time.Second)},
				{remoteAddr: "192.0.2.1:1234", at: now.Add(time.Minute)},
			},
			wantStatus: []int{
				http.StatusNoContent,
				http.StatusNoContent,
				http.StatusNoContent,
				http.StatusNoContent,
				http.StatusNoContent,
				http.StatusNoContent,
			},
			wantCalled: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var currentTime time.Time
			called := 0
			handler := newLoginRateLimiter(5, time.Minute, func() time.Time {
				return currentTime
			})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called++
				w.WriteHeader(http.StatusNoContent)
			}))

			for i, request := range tt.requests {
				currentTime = request.at
				req := httptest.NewRequest(http.MethodPost, "/v1/admin/login", nil)
				req.RemoteAddr = request.remoteAddr
				rec := httptest.NewRecorder()

				handler.ServeHTTP(rec, req)

				if rec.Code != tt.wantStatus[i] {
					t.Fatalf("request %d: expected status %d, got %d", i, tt.wantStatus[i], rec.Code)
				}

				if rec.Code == http.StatusTooManyRequests && rec.Header().Get("Retry-After") == "" {
					t.Fatalf("request %d: expected Retry-After header", i)
				}
			}

			if called != tt.wantCalled {
				t.Fatalf("expected called=%d, got %d", tt.wantCalled, called)
			}
		})
	}
}

func TestClientIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		want       string
	}{
		{name: "host port", remoteAddr: "192.0.2.1:1234", want: "192.0.2.1"},
		{name: "host only", remoteAddr: "192.0.2.1", want: "192.0.2.1"},
		{name: "empty", remoteAddr: "", want: "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := clientIP(tt.remoteAddr); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

type rateLimitRequest struct {
	remoteAddr string
	at         time.Time
}
