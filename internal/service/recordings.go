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
}

func NewRecordingsDataService(rr repo.RecordingsRepo) *RecordingsDataService {
	return &RecordingsDataService{
		rr: rr,
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
	baseSlug := slug.Slugify(input.Title)
	if baseSlug == "" {
		baseSlug = "recording"
	}

	recordingSlug := baseSlug

	for i := 3; ; i++ {
		exists, err := s.rr.SlugExists(ctx, recordingSlug)
		if err != nil {
			return domain.Recording{}, err
		}
		if !exists {
			break
		}

		recordingSlug = fmt.Sprintf("%s-%d", baseSlug, i)
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

func (s *RecordingsDataService) PendingList(ctx context.Context, _ StatusPending, p domain.Pagination) ([]domain.Recording, error) {
	return s.rr.List(ctx, []domain.RecordingStatus{domain.StatusPending}, p)
}

func (s *RecordingsDataService) ApprovedList(ctx context.Context, _ StatusApproved, p domain.Pagination) ([]domain.Recording, error) {
	return s.rr.List(ctx, []domain.RecordingStatus{domain.StatusApproved}, p)
}

func (s *RecordingsDataService) RejectedList(ctx context.Context, _ StatusRejected, p domain.Pagination) ([]domain.Recording, error) {
	return s.rr.List(ctx, []domain.RecordingStatus{domain.StatusRejected}, p)
}
