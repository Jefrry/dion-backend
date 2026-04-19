package repo

import (
	"context"
	"dion-backend/internal/domain"

	"gorm.io/gorm"
)

type ArtistsRepo interface {
	List(ctx context.Context, p domain.Pagination) ([]domain.Artist, error)
}

type ArtistsDataRepo struct {
	db *gorm.DB
}

func NewArtistsRepo(db *gorm.DB) *ArtistsDataRepo {
	return &ArtistsDataRepo{db: db}
}

func (r *ArtistsDataRepo) List(ctx context.Context, p domain.Pagination) ([]domain.Artist, error) {
	var items []domain.Artist

	err := r.db.WithContext(ctx).
		Limit(p.Limit).
		Offset(p.Offset).
		Find(&items).Error

	return items, err
}
