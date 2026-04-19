package service

import (
	"context"
	"dion-backend/internal/domain"
	"dion-backend/internal/repo"
)

type ArtistsService interface {
	List(ctx context.Context, p domain.Pagination) ([]domain.Artist, error)
	GetBySlug(ctx context.Context, slug string) (domain.Artist, error)
}

type ArtistsDataService struct {
	ar repo.ArtistsRepo
}

func NewArtistsDataService(ar repo.ArtistsRepo) *ArtistsDataService {
	return &ArtistsDataService{ar: ar}
}

func (s *ArtistsDataService) List(ctx context.Context, p domain.Pagination) ([]domain.Artist, error) {
	return s.ar.List(ctx, p)
}

func (s *ArtistsDataService) GetBySlug(ctx context.Context, slug string) (domain.Artist, error) {
	return s.ar.GetBySlug(ctx, slug)
}
