package service

import (
	"context"
	"dion-backend/internal/domain"
	"time"
)

type StatusPending struct{}
type StatusApproved struct{}
type StatusRejected struct{}

type CreateRecordingInput struct {
	Title       string
	Description *string
	ConcertDate *time.Time
	ExternalURL string
	ArtistName  string
}

type RecordingsService interface {
	List(ctx context.Context, p domain.Pagination) ([]domain.Recording, error)
	GetBySlug(ctx context.Context, slug string) (domain.Recording, error)
	ListByArtistSlug(ctx context.Context, artistSlug string, p domain.Pagination) ([]domain.Recording, error)
	Create(ctx context.Context, input CreateRecordingInput) (domain.Recording, error)
	PendingList(ctx context.Context, status StatusPending, p domain.Pagination) ([]domain.Recording, error)
	ApprovedList(ctx context.Context, status StatusApproved, p domain.Pagination) ([]domain.Recording, error)
	RejectedList(ctx context.Context, status StatusRejected, p domain.Pagination) ([]domain.Recording, error)
}
