package router

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"dion-backend/internal/config"
	"dion-backend/internal/domain"
	"dion-backend/internal/handler"
	"dion-backend/internal/service"
	"dion-backend/internal/utils"

	"github.com/golang-jwt/jwt/v5"
)

type routerRecordingServiceMock struct {
	pendingCalled bool
	updateCalled  bool
}

func (m *routerRecordingServiceMock) List(context.Context, domain.Pagination) ([]domain.Recording, error) {
	return nil, errors.New("not implemented")
}

func (m *routerRecordingServiceMock) GetBySlug(context.Context, string) (domain.Recording, error) {
	return domain.Recording{}, errors.New("not implemented")
}

func (m *routerRecordingServiceMock) ListByArtistSlug(context.Context, string, domain.Pagination) ([]domain.Recording, error) {
	return nil, errors.New("not implemented")
}

func (m *routerRecordingServiceMock) Create(context.Context, service.CreateRecordingInput) (domain.Recording, error) {
	return domain.Recording{}, errors.New("not implemented")
}

func (m *routerRecordingServiceMock) Update(context.Context, uint, service.UpdateRecordingInput) (domain.Recording, error) {
	m.updateCalled = true
	return domain.Recording{ID: 1, Title: "Updated", Slug: "updated", Status: domain.StatusApproved}, nil
}

func (m *routerRecordingServiceMock) PendingList(context.Context, service.StatusPending, domain.Pagination) ([]domain.Recording, error) {
	m.pendingCalled = true
	return []domain.Recording{}, nil
}

func (m *routerRecordingServiceMock) ApprovedList(context.Context, service.StatusApproved, domain.Pagination) ([]domain.Recording, error) {
	return nil, errors.New("not implemented")
}

func (m *routerRecordingServiceMock) RejectedList(context.Context, service.StatusRejected, domain.Pagination) ([]domain.Recording, error) {
	return nil, errors.New("not implemented")
}

func TestAdminPendingRecordingsRouteAuth(t *testing.T) {
	tests := []struct {
		name              string
		withToken         bool
		wantStatus        int
		wantPendingCalled bool
	}{
		{
			name:              "requires jwt",
			wantStatus:        http.StatusUnauthorized,
			wantPendingCalled: false,
		},
		{
			name:              "allows valid jwt",
			withToken:         true,
			wantStatus:        http.StatusOK,
			wantPendingCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := &routerRecordingServiceMock{}
			cfg := routerTestAdminConfig()
			rh := handler.NewRecordingsHandler(slog.New(slog.NewTextHandler(io.Discard, nil)), utils.NewHandlerUtils(), rs)
			r := NewRouter(rh, nil, nil, cfg).MustRun()

			req := httptest.NewRequest(http.MethodGet, "/v1/admin/recordings/pending", nil)
			if tt.withToken {
				req.Header.Set("Authorization", "Bearer "+signRouterTestToken(t, cfg))
			}
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d: %s", tt.wantStatus, rec.Code, rec.Body.String())
			}
			if rs.pendingCalled != tt.wantPendingCalled {
				t.Fatalf("expected pendingCalled=%v, got %v", tt.wantPendingCalled, rs.pendingCalled)
			}
		})
	}
}

func TestAdminUpdateRecordingRouteAuth(t *testing.T) {
	tests := []struct {
		name             string
		withToken        bool
		wantStatus       int
		wantUpdateCalled bool
	}{
		{
			name:             "requires jwt",
			wantStatus:       http.StatusUnauthorized,
			wantUpdateCalled: false,
		},
		{
			name:             "allows valid jwt",
			withToken:        true,
			wantStatus:       http.StatusOK,
			wantUpdateCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := &routerRecordingServiceMock{}
			cfg := routerTestAdminConfig()
			rh := handler.NewRecordingsHandler(slog.New(slog.NewTextHandler(io.Discard, nil)), utils.NewHandlerUtils(), rs)
			r := NewRouter(rh, nil, nil, cfg).MustRun()

			body := `{"title":"Updated title","externalURL":"https://example.com/video","artistName":"Artist","status":"approved"}`
			req := httptest.NewRequest(http.MethodPatch, "/v1/admin/recordings/1", bytes.NewBufferString(body))
			req.Header.Set("Content-Type", "application/json")
			if tt.withToken {
				req.Header.Set("Authorization", "Bearer "+signRouterTestToken(t, cfg))
			}
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d: %s", tt.wantStatus, rec.Code, rec.Body.String())
			}
			if rs.updateCalled != tt.wantUpdateCalled {
				t.Fatalf("expected updateCalled=%v, got %v", tt.wantUpdateCalled, rs.updateCalled)
			}
		})
	}
}

func routerTestAdminConfig() config.AdminConfig {
	return config.AdminConfig{
		Username:     "admin",
		PasswordHash: "$2a$04$unusedunusedunusedunuseduOgivM6gnQsoJUHg18BDuEfd4g.n24vO",
		JWTSecret:    "0123456789012345678901234567890123456789012345678901234567890123",
	}
}

func signRouterTestToken(t *testing.T, cfg config.AdminConfig) string {
	t.Helper()

	claims := jwt.RegisteredClaims{
		Subject:   cfg.Username,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	return tokenString
}
