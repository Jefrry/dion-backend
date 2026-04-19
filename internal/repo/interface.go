package repo

import (
	"context"
	"dion-backend/internal/domain"
)

type RecordingsRepo interface {
	List(ctx context.Context, statuses []domain.RecordingStatus, p domain.Pagination) ([]domain.Recording, error)
}
