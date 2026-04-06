package handler

import (
	"dion-backend/internal/utils"

	"log/slog"
	"net/http"
)

type RecordsHandler struct {
	l *slog.Logger
	u utils.HandlerUtils
}

func NewRecordsHandler(l *slog.Logger, u utils.HandlerUtils) *RecordsHandler {
	return &RecordsHandler{
		l: l,
		u: u,
	}
}

func (rh *RecordsHandler) Get(w http.ResponseWriter, r *http.Request) {
	rh.u.WritePlain(w, http.StatusOK, "result")
}
