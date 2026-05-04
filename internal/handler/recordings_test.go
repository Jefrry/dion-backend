package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"dion-backend/internal/domain"
	"dion-backend/internal/service"
	"dion-backend/internal/utils"
)

type recordingServiceMock struct {
	pendingRecordings []domain.Recording
	pendingErr        error
	pendingPagination domain.Pagination
	pendingCalled     bool
}

func (m *recordingServiceMock) List(context.Context, domain.Pagination) ([]domain.Recording, error) {
	return nil, errors.New("not implemented")
}

func (m *recordingServiceMock) GetBySlug(context.Context, string) (domain.Recording, error) {
	return domain.Recording{}, errors.New("not implemented")
}

func (m *recordingServiceMock) ListByArtistSlug(context.Context, string, domain.Pagination) ([]domain.Recording, error) {
	return nil, errors.New("not implemented")
}

func (m *recordingServiceMock) Create(context.Context, service.CreateRecordingInput) (domain.Recording, error) {
	return domain.Recording{}, errors.New("not implemented")
}

func (m *recordingServiceMock) PendingList(_ context.Context, _ service.StatusPending, p domain.Pagination) ([]domain.Recording, error) {
	m.pendingCalled = true
	m.pendingPagination = p
	return m.pendingRecordings, m.pendingErr
}

func (m *recordingServiceMock) ApprovedList(context.Context, service.StatusApproved, domain.Pagination) ([]domain.Recording, error) {
	return nil, errors.New("not implemented")
}

func (m *recordingServiceMock) RejectedList(context.Context, service.StatusRejected, domain.Pagination) ([]domain.Recording, error) {
	return nil, errors.New("not implemented")
}

func TestGetPendingList(t *testing.T) {
	tests := []struct {
		name              string
		target            string
		recordings        []domain.Recording
		serviceErr        error
		wantStatus        int
		wantPagination    domain.Pagination
		wantPendingCalled bool
		wantRecordings    int
	}{
		{
			name:   "uses requested pagination",
			target: "/v1/admin/recordings/pending?limit=10&offset=5",
			recordings: []domain.Recording{
				{ID: 1, Title: "Pending recording", Slug: "pending-recording", Status: domain.StatusPending},
			},
			wantStatus:        http.StatusOK,
			wantPagination:    domain.Pagination{Limit: 10, Offset: 5},
			wantPendingCalled: true,
			wantRecordings:    1,
		},
		{
			name:              "uses pagination defaults",
			target:            "/v1/admin/recordings/pending?limit=bad&offset=-1",
			wantStatus:        http.StatusOK,
			wantPagination:    domain.Pagination{Limit: 20, Offset: 0},
			wantPendingCalled: true,
			wantRecordings:    0,
		},
		{
			name:              "handles service error",
			target:            "/v1/admin/recordings/pending",
			serviceErr:        errors.New("database unavailable"),
			wantStatus:        http.StatusInternalServerError,
			wantPagination:    domain.Pagination{Limit: 20, Offset: 0},
			wantPendingCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := &recordingServiceMock{
				pendingRecordings: tt.recordings,
				pendingErr:        tt.serviceErr,
			}
			rh := newTestRecordsHandler(rs)

			req := httptest.NewRequest(http.MethodGet, tt.target, nil)
			rec := httptest.NewRecorder()

			rh.GetPendingList(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d: %s", tt.wantStatus, rec.Code, rec.Body.String())
			}
			if rs.pendingCalled != tt.wantPendingCalled {
				t.Fatalf("expected pendingCalled=%v, got %v", tt.wantPendingCalled, rs.pendingCalled)
			}
			if rs.pendingPagination != tt.wantPagination {
				t.Fatalf("expected pagination %+v, got %+v", tt.wantPagination, rs.pendingPagination)
			}
			if tt.wantStatus != http.StatusOK {
				return
			}

			var got []domain.Recording
			if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}
			if len(got) != tt.wantRecordings {
				t.Fatalf("expected %d recordings, got %+v", tt.wantRecordings, got)
			}
			for _, recording := range got {
				if recording.Status != domain.StatusPending {
					t.Fatalf("expected pending recording, got %+v", recording)
				}
			}
		})
	}
}

func newTestRecordsHandler(rs service.RecordingsService) *RecordsHandler {
	return NewRecordingsHandler(slog.New(slog.NewTextHandler(io.Discard, nil)), utils.NewHandlerUtils(), rs)
}
