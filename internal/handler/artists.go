package handler

import (
	"dion-backend/internal/domain"
	"dion-backend/internal/service"
	"dion-backend/internal/utils"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type ArtistsHandler struct {
	l  *slog.Logger
	u  utils.HandlerUtils
	as service.ArtistsService
}

func NewArtistsHandler(l *slog.Logger, u utils.HandlerUtils, as service.ArtistsService) *ArtistsHandler {
	return &ArtistsHandler{
		l:  l,
		u:  u,
		as: as,
	}
}

func (ah *ArtistsHandler) GetList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	limit, err := strconv.Atoi(q.Get("limit"))
	if err != nil || limit <= 0 {
		limit = 20
	}

	offset, err := strconv.Atoi(q.Get("offset"))
	if err != nil || offset < 0 {
		offset = 0
	}

	artists, err := ah.as.List(r.Context(), domain.Pagination{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		ah.l.Error("ArtistsService.List failed", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	ah.u.WriteJSON(w, http.StatusOK, artists)
}

func (ah *ArtistsHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	artist, err := ah.as.GetBySlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		ah.l.Error("ArtistsService.GetBySlug failed", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	ah.u.WriteJSON(w, http.StatusOK, artist)
}
