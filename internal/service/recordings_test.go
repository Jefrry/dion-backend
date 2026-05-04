package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"dion-backend/internal/domain"

	"gorm.io/gorm"
)

type recordingsRepoMock struct {
	recordings       map[uint]domain.Recording
	recordingSlugs   map[string]bool
	getByIDErr       error
	updateErr        error
	updateCalled     bool
	updateID         uint
	updateItem       domain.Recording
	updateArtistSlug string
}

func (m *recordingsRepoMock) List(context.Context, []domain.RecordingStatus, domain.Pagination) ([]domain.Recording, error) {
	return nil, errors.New("not implemented")
}

func (m *recordingsRepoMock) GetByID(_ context.Context, id uint) (domain.Recording, error) {
	if m.getByIDErr != nil {
		return domain.Recording{}, m.getByIDErr
	}
	recording, ok := m.recordings[id]
	if !ok {
		return domain.Recording{}, gorm.ErrRecordNotFound
	}
	return recording, nil
}

func (m *recordingsRepoMock) GetBySlug(context.Context, string) (domain.Recording, error) {
	return domain.Recording{}, errors.New("not implemented")
}

func (m *recordingsRepoMock) ListByArtistSlug(context.Context, string, domain.Pagination) ([]domain.Recording, error) {
	return nil, errors.New("not implemented")
}

func (m *recordingsRepoMock) SlugExists(_ context.Context, slug string, _ uint) (bool, error) {
	return m.recordingSlugs[slug], nil
}

func (m *recordingsRepoMock) Create(context.Context, domain.Recording) (domain.Recording, error) {
	return domain.Recording{}, errors.New("not implemented")
}

func (m *recordingsRepoMock) Update(_ context.Context, id uint, item domain.Recording, artistSlug string) (domain.Recording, error) {
	m.updateCalled = true
	m.updateID = id
	m.updateItem = item
	m.updateArtistSlug = artistSlug
	if m.updateErr != nil {
		return domain.Recording{}, m.updateErr
	}
	item.ID = id
	return item, nil
}

type artistsRepoMock struct {
	artistSlugs map[string]bool
}

func (m *artistsRepoMock) List(context.Context, domain.Pagination) ([]domain.Artist, error) {
	return nil, errors.New("not implemented")
}

func (m *artistsRepoMock) GetBySlug(context.Context, string) (domain.Artist, error) {
	return domain.Artist{}, errors.New("not implemented")
}

func (m *artistsRepoMock) SlugExists(_ context.Context, slug string) (bool, error) {
	return m.artistSlugs[slug], nil
}

func TestRecordingsDataServiceUpdate(t *testing.T) {
	concertDate := time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		id             uint
		current        domain.Recording
		recordingSlugs map[string]bool
		artistSlugs    map[string]bool
		input          UpdateRecordingInput
		wantErr        error
		wantUpdated    bool
		wantSlug       string
		wantArtistSlug string
		wantStatus     domain.RecordingStatus
	}{
		{
			name: "approved keeps slug when title unchanged",
			id:   1,
			current: domain.Recording{
				ID:     1,
				Title:  "Same title",
				Slug:   "same-title",
				Status: domain.StatusPending,
			},
			input: UpdateRecordingInput{
				Title:       "Same title",
				ConcertDate: &concertDate,
				ExternalURL: " https://example.com/video ",
				ArtistName:  "New Artist",
				Status:      domain.StatusApproved,
			},
			wantUpdated:    true,
			wantSlug:       "same-title",
			wantArtistSlug: "new-artist",
			wantStatus:     domain.StatusApproved,
		},
		{
			name: "approved regenerates unique slug when title changed",
			id:   2,
			current: domain.Recording{
				ID:     2,
				Title:  "Old title",
				Slug:   "old-title",
				Status: domain.StatusPending,
			},
			recordingSlugs: map[string]bool{"new-title": true},
			artistSlugs:    map[string]bool{"new-artist": true},
			input: UpdateRecordingInput{
				Title:       "New title",
				ExternalURL: "https://example.com/video",
				ArtistName:  "New Artist",
				Status:      domain.StatusApproved,
			},
			wantUpdated:    true,
			wantSlug:       "new-title-3",
			wantArtistSlug: "new-artist-3",
			wantStatus:     domain.StatusApproved,
		},
		{
			name: "rejected does not generate artist slug",
			id:   3,
			current: domain.Recording{
				ID:     3,
				Title:  "Old title",
				Slug:   "old-title",
				Status: domain.StatusPending,
			},
			input: UpdateRecordingInput{
				Title:       "Rejected title",
				ExternalURL: "https://example.com/video",
				ArtistName:  "Submitted Artist",
				Status:      domain.StatusRejected,
			},
			wantUpdated:    true,
			wantSlug:       "rejected-title",
			wantArtistSlug: "",
			wantStatus:     domain.StatusRejected,
		},
		{
			name: "returns not found",
			id:   4,
			input: UpdateRecordingInput{
				Title:       "Missing title",
				ExternalURL: "https://example.com/video",
				ArtistName:  "Artist",
				Status:      domain.StatusApproved,
			},
			wantErr: gorm.ErrRecordNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recordings := map[uint]domain.Recording{}
			if tt.current.ID != 0 {
				recordings[tt.id] = tt.current
			}
			rr := &recordingsRepoMock{
				recordings:     recordings,
				recordingSlugs: tt.recordingSlugs,
			}
			ar := &artistsRepoMock{artistSlugs: tt.artistSlugs}
			s := NewRecordingsDataService(rr, ar)

			_, err := s.Update(context.Background(), tt.id, tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected err %v, got %v", tt.wantErr, err)
			}
			if rr.updateCalled != tt.wantUpdated {
				t.Fatalf("expected updateCalled=%v, got %v", tt.wantUpdated, rr.updateCalled)
			}
			if !tt.wantUpdated {
				return
			}
			if rr.updateID != tt.id {
				t.Fatalf("expected update id %d, got %d", tt.id, rr.updateID)
			}
			if rr.updateItem.Slug != tt.wantSlug {
				t.Fatalf("expected slug %q, got %q", tt.wantSlug, rr.updateItem.Slug)
			}
			if rr.updateArtistSlug != tt.wantArtistSlug {
				t.Fatalf("expected artist slug %q, got %q", tt.wantArtistSlug, rr.updateArtistSlug)
			}
			if rr.updateItem.Status != tt.wantStatus {
				t.Fatalf("expected status %q, got %q", tt.wantStatus, rr.updateItem.Status)
			}
		})
	}
}
