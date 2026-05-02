package repo

import (
	"context"
	"dion-backend/internal/domain"

	"gorm.io/gorm"
)

type RecordingsDataRepo struct {
	db *gorm.DB
}

func NewRecordingsRepo(db *gorm.DB) *RecordingsDataRepo {
	return &RecordingsDataRepo{db: db}
}

func (r *RecordingsDataRepo) List(ctx context.Context, statuses []domain.RecordingStatus, p domain.Pagination) ([]domain.Recording, error) {
	var items []domain.Recording

	err := r.db.WithContext(ctx).
		Preload("Artist").
		Where("status IN ?", statuses).
		Limit(p.Limit).
		Offset(p.Offset).
		Find(&items).Error
	if err != nil {
		return nil, err
	}

	// Pending recordings may have only artist name,
	// artist as an entity is gonna be created when status changes
	for i := range items {
		attachPendingArtist(&items[i])
	}

	return items, nil
}

func (r *RecordingsDataRepo) GetBySlug(ctx context.Context, slug string) (domain.Recording, error) {
	var item domain.Recording
	err := r.db.WithContext(ctx).Preload("Artist").Where("slug = ?", slug).First(&item).Error
	if err != nil {
		return item, err
	}

	attachPendingArtist(&item)
	return item, err
}

func (r *RecordingsDataRepo) ListByArtistSlug(ctx context.Context, artistSlug string, p domain.Pagination) ([]domain.Recording, error) {
	var items []domain.Recording
	err := r.db.WithContext(ctx).
		Preload("Artist").
		Joins("JOIN artists ON artists.id = recordings.artist_id").
		Where("artists.slug = ?", artistSlug).
		Limit(p.Limit).
		Offset(p.Offset).
		Find(&items).Error
	if err != nil {
		return nil, err
	}

	for i := range items {
		attachPendingArtist(&items[i])
	}
	return items, nil
}

func (r *RecordingsDataRepo) SlugExists(ctx context.Context, slug string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Recording{}).
		Where("slug = ?", slug).
		Count(&count).Error

	return count > 0, err
}

func (r *RecordingsDataRepo) Create(ctx context.Context, item domain.Recording) (domain.Recording, error) {
	err := r.db.WithContext(ctx).Create(&item).Error
	if err != nil {
		return item, err
	}

	attachPendingArtist(&item)
	return item, err
}

func attachPendingArtist(item *domain.Recording) {
	if item.Artist != nil || item.ArtistName == "" {
		return
	}

	item.Artist = &domain.Artist{
		Name: item.ArtistName,
	}
}
