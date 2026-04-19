package service

import (
	"context"
	"dion-backend/internal/domain"
)

type StatusPending struct{}
type StatusApproved struct{}
type StatusRejected struct{}

type RecordingsService interface {
	List(ctx context.Context, p domain.Pagination) ([]domain.Recording, error)
	PendingList(ctx context.Context, status StatusPending, p domain.Pagination) ([]domain.Recording, error)
	ApprovedList(ctx context.Context, status StatusApproved, p domain.Pagination) ([]domain.Recording, error)
	RejectedList(ctx context.Context, status StatusRejected, p domain.Pagination) ([]domain.Recording, error)
}
