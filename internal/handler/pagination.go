package handler

import (
	"dion-backend/internal/domain"
	"net/http"
	"strconv"
)

func parsePagination(r *http.Request) domain.Pagination {
	q := r.URL.Query()

	limit, err := strconv.Atoi(q.Get("limit"))
	if err != nil || limit <= 0 {
		limit = 20
	}

	offset, err := strconv.Atoi(q.Get("offset"))
	if err != nil || offset < 0 {
		offset = 0
	}

	return domain.Pagination{
		Limit:  limit,
		Offset: offset,
	}
}
