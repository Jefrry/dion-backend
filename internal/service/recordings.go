package service

import (
	"context"

	"dion-backend/internal/domain"
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

func (s *RecordingsDataService) PendingList(ctx context.Context, _ StatusPending, p domain.Pagination) ([]domain.Recording, error) {
	return s.rr.List(ctx, []domain.RecordingStatus{domain.StatusPending}, p)
}

func (s *RecordingsDataService) ApprovedList(ctx context.Context, _ StatusApproved, p domain.Pagination) ([]domain.Recording, error) {
	return s.rr.List(ctx, []domain.RecordingStatus{domain.StatusApproved}, p)
}

func (s *RecordingsDataService) RejectedList(ctx context.Context, _ StatusRejected, p domain.Pagination) ([]domain.Recording, error) {
	return s.rr.List(ctx, []domain.RecordingStatus{domain.StatusRejected}, p)
}
