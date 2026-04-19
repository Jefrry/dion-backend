package repo

import (
	"dion-backend/internal/domain"

	"context"

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
		Where("status IN ?", statuses).
		Limit(p.Limit).
		Offset(p.Offset).
		Find(&items).Error

	return items, err
}

func (r *RecordingsDataRepo) GetBySlug(ctx context.Context, slug string) (domain.Recording, error) {
	var item domain.Recording
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&item).Error
	return item, err
}
