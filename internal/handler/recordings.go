package handler

import (
	"dion-backend/internal/domain"
	"dion-backend/internal/service"
	"dion-backend/internal/utils"

	"log/slog"
	"net/http"
	"strconv"
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
