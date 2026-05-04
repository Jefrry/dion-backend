package repo

import (
	"context"
	"dion-backend/internal/domain"
	"errors"
	"time"

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

func (r *RecordingsDataRepo) GetByID(ctx context.Context, id uint) (domain.Recording, error) {
	var item domain.Recording
	err := r.db.WithContext(ctx).Preload("Artist").Where("id = ?", id).First(&item).Error
	if err != nil {
		return item, err
	}

	attachPendingArtist(&item)
	return item, nil
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

func (r *RecordingsDataRepo) SlugExists(ctx context.Context, slug string, exceptID uint) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).
		Model(&domain.Recording{}).
		Where("slug = ?", slug)

	if exceptID > 0 {
		query = query.Where("id <> ?", exceptID)
	}

	err := query.Count(&count).Error
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

func (r *RecordingsDataRepo) Update(ctx context.Context, id uint, item domain.Recording, artistSlug string) (domain.Recording, error) {
	var updated domain.Recording

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current domain.Recording
		if err := tx.Where("id = ?", id).First(&current).Error; err != nil {
			return err
		}

		var artistID *uint
		if item.Status == domain.StatusApproved {
			artist := domain.Artist{}
			err := tx.Where("name = ?", item.ArtistName).First(&artist).Error
			if err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}

				artist = domain.Artist{
					Name: item.ArtistName,
					Slug: artistSlug,
				}
				if err := tx.Create(&artist).Error; err != nil {
					return err
				}
			}

			artistID = &artist.ID
		}

		now := time.Now()
		externalURL := item.ExternalURL
		updates := map[string]any{
			"title":        item.Title,
			"slug":         item.Slug,
			"description":  item.Description,
			"artist_id":    artistID,
			"artist_name":  item.ArtistName,
			"concert_date": item.ConcertDate,
			"external_url": externalURL,
			"status":       item.Status,
			"moderated_at": &now,
		}

		if err := tx.Model(&current).Updates(updates).Error; err != nil {
			return err
		}

		return tx.Preload("Artist").Where("id = ?", id).First(&updated).Error
	})
	if err != nil {
		return domain.Recording{}, err
	}

	attachPendingArtist(&updated)
	return updated, nil
}

func attachPendingArtist(item *domain.Recording) {
	if item.Status != domain.StatusPending || item.Artist != nil || item.ArtistName == "" {
		return
	}

	item.Artist = &domain.Artist{
		Name: item.ArtistName,
	}
}
