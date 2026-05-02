package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"dion-backend/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

func TestAdminAuthAllowsValidToken(t *testing.T) {
	cfg := testAdminConfig()

	tests := []struct {
		name       string
		path       string
		header     string
		wantStatus int
		wantCalled bool
	}{
		{
			name:       "valid token",
			path:       "/v1/admin/recordings",
			header:     "Bearer " + signTestToken(t, cfg, cfg.Username, time.Now().Add(time.Hour), jwt.SigningMethodHS256),
			wantStatus: http.StatusNoContent,
			wantCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			handler := AdminAuth(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusNoContent)
			}))

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			req.Header.Set("Authorization", tt.header)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}
			if called != tt.wantCalled {
				t.Fatalf("expected called=%v, got %v", tt.wantCalled, called)
			}
		})
	}
}

func TestAdminAuthRejectsInvalidRequests(t *testing.T) {
	cfg := testAdminConfig()

	tests := []struct {
		name   string
		header string
	}{
		{name: "missing header"},
		{name: "not bearer", header: "Basic abc"},
		{name: "empty bearer", header: "Bearer "},
		{name: "invalid token", header: "Bearer invalid-token"},
		{name: "expired token", header: "Bearer " + signTestToken(t, cfg, cfg.Username, time.Now().Add(-time.Hour), jwt.SigningMethodHS256)},
		{name: "wrong subject", header: "Bearer " + signTestToken(t, cfg, "other-admin", time.Now().Add(time.Hour), jwt.SigningMethodHS256)},
		{name: "wrong method", header: "Bearer " + signTestToken(t, cfg, cfg.Username, time.Now().Add(time.Hour), jwt.SigningMethodHS384)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			handler := AdminAuth(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusNoContent)
			}))

			req := httptest.NewRequest(http.MethodGet, "/v1/admin/recordings", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
			}
			if called {
				t.Fatal("expected protected handler not to be called")
			}
		})
	}
}

func TestAdminAuthSkipsLogin(t *testing.T) {
	cfg := testAdminConfig()

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantCalled bool
	}{
		{
			name:       "login without token",
			method:     http.MethodPost,
			path:       "/v1/admin/login",
			wantStatus: http.StatusNoContent,
			wantCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			handler := AdminAuth(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusNoContent)
			}))

			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}
			if called != tt.wantCalled {
				t.Fatalf("expected called=%v, got %v", tt.wantCalled, called)
			}
		})
	}
}

func testAdminConfig() config.AdminConfig {
	return config.AdminConfig{
		Username:     "admin",
		PasswordHash: "$2a$04$unusedunusedunusedunuseduOgivM6gnQsoJUHg18BDuEfd4g.n24vO",
		JWTSecret:    "0123456789012345678901234567890123456789012345678901234567890123",
	}
}

func signTestToken(t *testing.T, cfg config.AdminConfig, subject string, expiresAt time.Time, method jwt.SigningMethod) string {
	t.Helper()

	claims := jwt.RegisteredClaims{
		Subject:   subject,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	}
	token := jwt.NewWithClaims(method, claims)
	tokenString, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	return tokenString
}
