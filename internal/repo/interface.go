package repo

import (
	"context"
	"dion-backend/internal/domain"
)

type RecordingsRepo interface {
	List(ctx context.Context, statuses []domain.RecordingStatus, p domain.Pagination) ([]domain.Recording, error)
	GetBySlug(ctx context.Context, slug string) (domain.Recording, error)
	ListByArtistSlug(ctx context.Context, artistSlug string, p domain.Pagination) ([]domain.Recording, error)
	SlugExists(ctx context.Context, slug string) (bool, error)
	Create(ctx context.Context, item domain.Recording) (domain.Recording, error)
}
