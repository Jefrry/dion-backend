package repo

import (
	"context"
	"dion-backend/internal/domain"
)

type RecordingsRepo interface {
	List(ctx context.Context, statuses []domain.RecordingStatus, p domain.Pagination) ([]domain.Recording, error)
	GetByID(ctx context.Context, id uint) (domain.Recording, error)
	GetBySlug(ctx context.Context, slug string) (domain.Recording, error)
	ListByArtistSlug(ctx context.Context, artistSlug string, p domain.Pagination) ([]domain.Recording, error)
	SlugExists(ctx context.Context, slug string, exceptID uint) (bool, error)
	Create(ctx context.Context, item domain.Recording) (domain.Recording, error)
	Update(ctx context.Context, id uint, item domain.Recording, artistSlug string) (domain.Recording, error)
}

type ArtistsRepo interface {
	List(ctx context.Context, p domain.Pagination) ([]domain.Artist, error)
	GetBySlug(ctx context.Context, slug string) (domain.Artist, error)
	SlugExists(ctx context.Context, slug string) (bool, error)
}
