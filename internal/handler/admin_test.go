package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"dion-backend/internal/config"
	"dion-backend/internal/utils"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func TestAdminLoginSuccess(t *testing.T) {
	tests := []struct {
		name     string
		password string
		body     string
	}{
		{
			name:     "valid credentials",
			password: "strong-password",
			body:     `{"username":"admin","password":"strong-password"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adminHandler := newTestAdminHandler(t, tt.password)

			req := httptest.NewRequest(http.MethodPost, "/v1/admin/login", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			adminHandler.Login(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
			}

			var res adminLoginResponse
			if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if res.Token == "" {
				t.Fatal("expected token in response")
			}

			expiresAt, err := time.Parse(time.RFC3339, res.ExpiresAt)
			if err != nil {
				t.Fatalf("expected RFC3339 expiresAt, got %q: %v", res.ExpiresAt, err)
			}
			if time.Until(expiresAt) < 23*time.Hour || time.Until(expiresAt) > 24*time.Hour {
				t.Fatalf("expected expiresAt about 24h from now, got %s", expiresAt)
			}

			claims := &jwt.RegisteredClaims{}
			token, err := jwt.ParseWithClaims(res.Token, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(adminHandler.cfg.JWTSecret), nil
			}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired())
			if err != nil {
				t.Fatalf("failed to parse token: %v", err)
			}
			if !token.Valid {
				t.Fatal("expected valid token")
			}
			if claims.Subject != adminHandler.cfg.Username {
				t.Fatalf("expected subject %q, got %q", adminHandler.cfg.Username, claims.Subject)
			}
		})
	}
}

func TestAdminLoginInvalidCredentials(t *testing.T) {
	adminHandler := newTestAdminHandler(t, "strong-password")

	tests := []struct {
		name string
		body string
	}{
		{name: "wrong username", body: `{"username":"root","password":"strong-password"}`},
		{name: "wrong password", body: `{"username":"admin","password":"wrong-password"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/v1/admin/login", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			adminHandler.Login(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
			}
		})
	}
}

func TestAdminLoginRequestErrors(t *testing.T) {
	adminHandler := newTestAdminHandler(t, "strong-password")

	tests := []struct {
		name        string
		contentType string
		body        string
		wantStatus  int
	}{
		{name: "malformed json", contentType: "application/json", body: `{`, wantStatus: http.StatusBadRequest},
		{name: "wrong content type", contentType: "text/plain", body: `{"username":"admin","password":"strong-password"}`, wantStatus: http.StatusUnsupportedMediaType},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/v1/admin/login", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", tt.contentType)
			rec := httptest.NewRecorder()

			adminHandler.Login(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}
		})
	}
}

func newTestAdminHandler(t *testing.T, password string) *AdminHandler {
	t.Helper()

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("failed to generate bcrypt hash: %v", err)
	}

	return NewAdminHandler(slog.New(slog.NewTextHandler(io.Discard, nil)), utils.NewHandlerUtils(), config.AdminConfig{
		Username:     "admin",
		PasswordHash: string(passwordHash),
		JWTSecret:    "0123456789012345678901234567890123456789012345678901234567890123",
	})
}
