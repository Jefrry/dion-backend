package repo

import (
	"context"
	"dion-backend/internal/domain"

	"gorm.io/gorm"
)

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

func (r *ArtistsDataRepo) GetBySlug(ctx context.Context, slug string) (domain.Artist, error) {
	var item domain.Artist
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&item).Error
	return item, err
}

func (r *ArtistsDataRepo) SlugExists(ctx context.Context, slug string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Artist{}).
		Where("slug = ?", slug).
		Count(&count).Error

	return count > 0, err
}
