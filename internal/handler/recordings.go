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

type RecordsHandler struct {
	l  *slog.Logger
	u  utils.HandlerUtils
	rs service.RecordingsService
}

func NewRecordingsHandler(l *slog.Logger, u utils.HandlerUtils, rs service.RecordingsService) *RecordsHandler {
	return &RecordsHandler{
		l:  l,
		u:  u,
		rs: rs,
	}
}

func (rh *RecordsHandler) Get(w http.ResponseWriter, r *http.Request) {
	rh.u.WritePlain(w, http.StatusOK, "result")
}

// GetApprovedList godoc
// @Summary     List approved recordings
// @Tags        recordings
// @Produce     json
// @Param       limit   query    int  false  "Page size"   default(20)
// @Param       offset  query    int  false  "Page offset" default(0)
// @Success     200  {array}   domain.Recording
// @Failure     500  {string}  string  "internal server error"
// @Router      /recordings [get]
func (rh *RecordsHandler) GetApprovedList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	limit, err := strconv.Atoi(q.Get("limit"))
	if err != nil || limit <= 0 {
		limit = 20
	}

	offset, err := strconv.Atoi(q.Get("offset"))
	if err != nil || offset < 0 {
		offset = 0
	}

	recordings, err := rh.rs.ApprovedList(r.Context(), service.StatusApproved{}, domain.Pagination{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		rh.l.Error("ApprovedList failed", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	rh.u.WriteJSON(w, http.StatusOK, recordings)
}

// GetListByArtistSlug godoc
// @Summary     List recordings by artist slug
// @Tags        recordings
// @Produce     json
// @Param       slug    path     string  true   "Artist slug"
// @Param       limit   query    int     false  "Page size"   default(20)
// @Param       offset  query    int     false  "Page offset" default(0)
// @Success     200  {array}   domain.Recording
// @Failure     500  {string}  string  "internal server error"
// @Router      /artists/{slug}/recordings [get]
func (rh *RecordsHandler) GetListByArtistSlug(w http.ResponseWriter, r *http.Request) {
	artistSlug := chi.URLParam(r, "slug")
	q := r.URL.Query()

	limit, err := strconv.Atoi(q.Get("limit"))
	if err != nil || limit <= 0 {
		limit = 20
	}

	offset, err := strconv.Atoi(q.Get("offset"))
	if err != nil || offset < 0 {
		offset = 0
	}

	recordings, err := rh.rs.ListByArtistSlug(r.Context(), artistSlug, domain.Pagination{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		rh.l.Error("RecordingsService.ListByArtistSlug failed", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	rh.u.WriteJSON(w, http.StatusOK, recordings)
}

// GetBySlug godoc
// @Summary     Get recording by slug
// @Tags        recordings
// @Produce     json
// @Param       slug  path      string  true  "Recording slug"
// @Success     200   {object}  domain.Recording
// @Failure     404   {string}  string  "not found"
// @Failure     500   {string}  string  "internal server error"
// @Router      /recordings/{slug} [get]
func (rh *RecordsHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	recording, err := rh.rs.GetBySlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		rh.l.Error("RecordingsService.GetBySlug failed", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	rh.u.WriteJSON(w, http.StatusOK, recording)
}
