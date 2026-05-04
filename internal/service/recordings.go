package service

import (
	"context"
	"fmt"
	"strings"

	"dion-backend/internal/domain"
	"dion-backend/internal/lib/slug"
	"dion-backend/internal/repo"
)

type RecordingsDataService struct {
	rr repo.RecordingsRepo
	ar repo.ArtistsRepo
}

func NewRecordingsDataService(rr repo.RecordingsRepo, ar repo.ArtistsRepo) *RecordingsDataService {
	return &RecordingsDataService{
		rr: rr,
		ar: ar,
	}
}

func (s *RecordingsDataService) List(ctx context.Context, p domain.Pagination) ([]domain.Recording, error) {
	return s.rr.List(ctx, []domain.RecordingStatus{domain.StatusApproved, domain.StatusPending, domain.StatusRejected}, p)
}

func (s *RecordingsDataService) GetBySlug(ctx context.Context, slug string) (domain.Recording, error) {
	return s.rr.GetBySlug(ctx, slug)
}

func (s *RecordingsDataService) ListByArtistSlug(ctx context.Context, artistSlug string, p domain.Pagination) ([]domain.Recording, error) {
	return s.rr.ListByArtistSlug(ctx, artistSlug, p)
}

func (s *RecordingsDataService) Create(ctx context.Context, input CreateRecordingInput) (domain.Recording, error) {
	recordingSlug, err := s.uniqueRecordingSlug(ctx, input.Title, 0)
	if err != nil {
		return domain.Recording{}, err
	}

	externalURL := strings.TrimSpace(input.ExternalURL)
	item := domain.Recording{
		Title:       strings.TrimSpace(input.Title),
		Slug:        recordingSlug,
		Description: input.Description,
		ArtistName:  strings.TrimSpace(input.ArtistName),
		ConcertDate: input.ConcertDate,
		ExternalURL: &externalURL,
		Status:      domain.StatusPending,
	}

	return s.rr.Create(ctx, item)
}

func (s *RecordingsDataService) Update(ctx context.Context, id uint, input UpdateRecordingInput) (domain.Recording, error) {
	current, err := s.rr.GetByID(ctx, id)
	if err != nil {
		return domain.Recording{}, err
	}

	title := strings.TrimSpace(input.Title)
	recordingSlug := current.Slug
	if title != current.Title {
		recordingSlug, err = s.uniqueRecordingSlug(ctx, title, id)
		if err != nil {
			return domain.Recording{}, err
		}
	}

	artistName := strings.TrimSpace(input.ArtistName)
	artistSlug := ""
	if input.Status == domain.StatusApproved {
		artistSlug, err = s.uniqueArtistSlug(ctx, artistName)
		if err != nil {
			return domain.Recording{}, err
		}
	}

	externalURL := strings.TrimSpace(input.ExternalURL)
	item := domain.Recording{
		Title:       title,
		Slug:        recordingSlug,
		Description: input.Description,
		ArtistName:  artistName,
		ConcertDate: input.ConcertDate,
		ExternalURL: &externalURL,
		Status:      input.Status,
	}

	return s.rr.Update(ctx, id, item, artistSlug)
}

func (s *RecordingsDataService) PendingList(ctx context.Context, _ StatusPending, p domain.Pagination) ([]domain.Recording, error) {
	return s.rr.List(ctx, []domain.RecordingStatus{domain.StatusPending}, p)
}

func (s *RecordingsDataService) ApprovedList(ctx context.Context, _ StatusApproved, p domain.Pagination) ([]domain.Recording, error) {
	return s.rr.List(ctx, []domain.RecordingStatus{domain.StatusApproved}, p)
}

func (s *RecordingsDataService) RejectedList(ctx context.Context, _ StatusRejected, p domain.Pagination) ([]domain.Recording, error) {
	return s.rr.List(ctx, []domain.RecordingStatus{domain.StatusRejected}, p)
}

func (s *RecordingsDataService) uniqueRecordingSlug(ctx context.Context, title string, exceptID uint) (string, error) {
	baseSlug := slug.Slugify(title)
	if baseSlug == "" {
		baseSlug = "recording"
	}

	recordingSlug := baseSlug
	for i := 3; ; i++ {
		var (
			exists bool
			err    error
		)
		exists, err = s.rr.SlugExists(ctx, recordingSlug, exceptID)
		if err != nil {
			return "", err
		}
		if !exists {
			return recordingSlug, nil
		}

		recordingSlug = fmt.Sprintf("%s-%d", baseSlug, i)
	}
}

func (s *RecordingsDataService) uniqueArtistSlug(ctx context.Context, artistName string) (string, error) {
	baseSlug := slug.Slugify(artistName)
	if baseSlug == "" {
		baseSlug = "artist"
	}

	artistSlug := baseSlug
	for i := 3; ; i++ {
		exists, err := s.ar.SlugExists(ctx, artistSlug)
		if err != nil {
			return "", err
		}
		if !exists {
			return artistSlug, nil
		}

		artistSlug = fmt.Sprintf("%s-%d", baseSlug, i)
	}
}
