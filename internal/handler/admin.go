package handler

import (
	"dion-backend/internal/config"
	"dion-backend/internal/utils"
	"log/slog"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const adminTokenTTL = 24 * time.Hour

type adminLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type adminLoginResponse struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expiresAt"`
}

type AdminHandler struct {
	l   *slog.Logger
	u   utils.HandlerUtils
	cfg config.AdminConfig
}

func NewAdminHandler(l *slog.Logger, u utils.HandlerUtils, cfg config.AdminConfig) *AdminHandler {
	return &AdminHandler{
		l:   l,
		u:   u,
		cfg: cfg,
	}
}

// Login godoc
// @Summary     Admin login
// @Tags        admin
// @Accept      json
// @Produce     json
// @Param       request  body      adminLoginRequest   true  "Admin credentials"
// @Success     200      {object}  adminLoginResponse
// @Failure     400      {string}  string  "bad request"
// @Failure     401      {string}  string  "unauthorized"
// @Failure     415      {string}  string  "unsupported media type"
// @Failure     500      {string}  string  "internal server error"
// @Router      /admin/login [post]
func (ah *AdminHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req adminLoginRequest
	if ok := ah.u.ReadJSON(w, r, &req); !ok {
		return
	}

	if req.Username != ah.cfg.Username {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(ah.cfg.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	expiresAt := time.Now().Add(adminTokenTTL)
	tokenString, err := ah.signToken(expiresAt)
	if err != nil {
		ah.l.Error("failed to sign admin JWT", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	ah.u.WriteJSON(w, http.StatusOK, adminLoginResponse{
		Token:     tokenString,
		ExpiresAt: expiresAt.Format(time.RFC3339),
	})
}

func (ah *AdminHandler) signToken(expiresAt time.Time) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   ah.cfg.Username,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(ah.cfg.JWTSecret))
}
