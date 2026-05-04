package handler

import (
	"dion-backend/internal/service"
	"dion-backend/internal/utils"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type createRecordingRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description"`
	ConcertDate *string `json:"concertDate"`
	ExternalURL string  `json:"externalURL"`
	ArtistName  string  `json:"artistName"`
}
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

// Create godoc
// @Summary     Create recording submission
// @Tags        recordings
// @Accept      json
// @Produce     json
// @Param       request  body      createRecordingRequest  true  "Recording submission"
// @Success     201      {object}  domain.Recording
// @Failure     400      {string}  string  "bad request"
// @Failure     415      {string}  string  "unsupported media type"
// @Failure     500      {string}  string  "internal server error"
// @Router      /recordings [post]
func (rh *RecordsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createRecordingRequest
	if ok := rh.u.ReadJSON(w, r, &req); !ok {
		return
	}

	input, ok := validateCreateRequest(w, req)
	if !ok {
		return
	}

	recording, err := rh.rs.Create(r.Context(), input)
	if err != nil {
		rh.l.Error("RecordingsService.Create failed", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	rh.u.WriteJSON(w, http.StatusCreated, recording)
}

func validateCreateRequest(w http.ResponseWriter, req createRecordingRequest) (service.CreateRecordingInput, bool) {
	rules := []utils.LengthRule{
		{Field: "title", Value: req.Title, Min: 3, Max: 280, Required: true},
		{Field: "externalURL", Value: req.ExternalURL, Min: 5, Max: 2048, Required: true},
		{Field: "artistName", Value: req.ArtistName, Min: 2, Max: 255, Required: true},
	}

	if req.Description != nil {
		rules = append(rules, utils.LengthRule{
			Field: "description", Value: *req.Description, Min: 3, Max: 1000, Required: false,
		})
	}

	for _, rule := range rules {
		if err := utils.ValidateStringLength(rule); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return service.CreateRecordingInput{}, false
		}
	}

	var description *string
	if req.Description != nil {
		value := strings.TrimSpace(*req.Description)
		if value != "" {
			description = &value
		}
	}

	var concertDate *time.Time
	if req.ConcertDate != nil && strings.TrimSpace(*req.ConcertDate) != "" {
		parsedDate, err := time.Parse("2006-01-02", strings.TrimSpace(*req.ConcertDate))
		if err != nil {
			http.Error(w, "concertDate must use YYYY-MM-DD format", http.StatusBadRequest)
			return service.CreateRecordingInput{}, false
		}
		concertDate = &parsedDate
	}

	return service.CreateRecordingInput{
		Title:       strings.TrimSpace(req.Title),
		Description: description,
		ConcertDate: concertDate,
		ExternalURL: strings.TrimSpace(req.ExternalURL),
		ArtistName:  strings.TrimSpace(req.ArtistName),
	}, true
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
	pagination := parsePagination(r)

	recordings, err := rh.rs.ApprovedList(r.Context(), service.StatusApproved{}, pagination)
	if err != nil {
		rh.l.Error("ApprovedList failed", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	rh.u.WriteJSON(w, http.StatusOK, recordings)
}

// GetPendingList godoc
// @Summary     List pending recordings
// @Tags        admin
// @Produce     json
// @Param       limit   query    int  false  "Page size"   default(20)
// @Param       offset  query    int  false  "Page offset" default(0)
// @Success     200  {array}   domain.Recording
// @Failure     401  {string}  string  "unauthorized"
// @Failure     500  {string}  string  "internal server error"
// @Security    BearerAuth
// @Router      /admin/recordings/pending [get]
func (rh *RecordsHandler) GetPendingList(w http.ResponseWriter, r *http.Request) {
	pagination := parsePagination(r)

	recordings, err := rh.rs.PendingList(r.Context(), service.StatusPending{}, pagination)
	if err != nil {
		rh.l.Error("PendingList failed", "err", err)
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
	pagination := parsePagination(r)

	recordings, err := rh.rs.ListByArtistSlug(r.Context(), artistSlug, pagination)
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
