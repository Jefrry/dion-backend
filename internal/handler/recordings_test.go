package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path"
	"testing"

	"dion-backend/internal/domain"
	"dion-backend/internal/service"
	"dion-backend/internal/utils"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type recordingServiceMock struct {
	pendingRecordings []domain.Recording
	pendingErr        error
	pendingPagination domain.Pagination
	pendingCalled     bool
	updateRecording   domain.Recording
	updateErr         error
	updateID          uint
	updateInput       service.UpdateRecordingInput
	updateCalled      bool
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

func (m *recordingServiceMock) Update(_ context.Context, id uint, input service.UpdateRecordingInput) (domain.Recording, error) {
	m.updateCalled = true
	m.updateID = id
	m.updateInput = input
	return m.updateRecording, m.updateErr
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

func TestUpdateRecording(t *testing.T) {
	tests := []struct {
		name             string
		target           string
		contentType      string
		body             string
		recording        domain.Recording
		serviceErr       error
		wantStatus       int
		wantUpdateCalled bool
		wantID           uint
		wantStatusInput  domain.RecordingStatus
	}{
		{
			name:        "updates approved recording",
			target:      "/v1/admin/recordings/12",
			contentType: "application/json",
			body:        `{"title":"Updated title","description":"Updated description","concertDate":"2026-05-04","externalURL":"https://example.com/video","artistName":"Existing Artist","status":"approved"}`,
			recording: domain.Recording{
				ID:     12,
				Title:  "Updated title",
				Slug:   "updated-title",
				Status: domain.StatusApproved,
			},
			wantStatus:       http.StatusOK,
			wantUpdateCalled: true,
			wantID:           12,
			wantStatusInput:  domain.StatusApproved,
		},
		{
			name:        "updates rejected recording",
			target:      "/v1/admin/recordings/7",
			contentType: "application/json",
			body:        `{"title":"Rejected title","description":null,"concertDate":null,"externalURL":"https://example.com/video","artistName":"Submitted Artist","status":"rejected"}`,
			recording: domain.Recording{
				ID:     7,
				Title:  "Rejected title",
				Slug:   "rejected-title",
				Status: domain.StatusRejected,
			},
			wantStatus:       http.StatusOK,
			wantUpdateCalled: true,
			wantID:           7,
			wantStatusInput:  domain.StatusRejected,
		},
		{
			name:             "rejects invalid id",
			target:           "/v1/admin/recordings/bad",
			contentType:      "application/json",
			body:             `{"title":"Updated title","externalURL":"https://example.com/video","artistName":"Artist","status":"approved"}`,
			wantStatus:       http.StatusBadRequest,
			wantUpdateCalled: false,
		},
		{
			name:             "rejects invalid status",
			target:           "/v1/admin/recordings/1",
			contentType:      "application/json",
			body:             `{"title":"Updated title","externalURL":"https://example.com/video","artistName":"Artist","status":"pending"}`,
			wantStatus:       http.StatusBadRequest,
			wantUpdateCalled: false,
		},
		{
			name:             "rejects invalid date",
			target:           "/v1/admin/recordings/1",
			contentType:      "application/json",
			body:             `{"title":"Updated title","concertDate":"04-05-2026","externalURL":"https://example.com/video","artistName":"Artist","status":"approved"}`,
			wantStatus:       http.StatusBadRequest,
			wantUpdateCalled: false,
		},
		{
			name:             "rejects wrong content type",
			target:           "/v1/admin/recordings/1",
			contentType:      "text/plain",
			body:             `{"title":"Updated title","externalURL":"https://example.com/video","artistName":"Artist","status":"approved"}`,
			wantStatus:       http.StatusUnsupportedMediaType,
			wantUpdateCalled: false,
		},
		{
			name:             "maps not found",
			target:           "/v1/admin/recordings/1",
			contentType:      "application/json",
			body:             `{"title":"Updated title","externalURL":"https://example.com/video","artistName":"Artist","status":"approved"}`,
			serviceErr:       gorm.ErrRecordNotFound,
			wantStatus:       http.StatusNotFound,
			wantUpdateCalled: true,
			wantID:           1,
			wantStatusInput:  domain.StatusApproved,
		},
		{
			name:             "maps service error",
			target:           "/v1/admin/recordings/1",
			contentType:      "application/json",
			body:             `{"title":"Updated title","externalURL":"https://example.com/video","artistName":"Artist","status":"approved"}`,
			serviceErr:       errors.New("database unavailable"),
			wantStatus:       http.StatusInternalServerError,
			wantUpdateCalled: true,
			wantID:           1,
			wantStatusInput:  domain.StatusApproved,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := &recordingServiceMock{
				updateRecording: tt.recording,
				updateErr:       tt.serviceErr,
			}
			rh := newTestRecordsHandler(rs)

			req := httptest.NewRequest(http.MethodPatch, tt.target, bytes.NewBufferString(tt.body))
			routeCtx := chi.NewRouteContext()
			routeCtx.URLParams.Add("id", path.Base(req.URL.Path))
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			rec := httptest.NewRecorder()

			rh.Update(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d: %s", tt.wantStatus, rec.Code, rec.Body.String())
			}
			if rs.updateCalled != tt.wantUpdateCalled {
				t.Fatalf("expected updateCalled=%v, got %v", tt.wantUpdateCalled, rs.updateCalled)
			}
			if !tt.wantUpdateCalled {
				return
			}
			if rs.updateID != tt.wantID {
				t.Fatalf("expected id %d, got %d", tt.wantID, rs.updateID)
			}
			if rs.updateInput.Status != tt.wantStatusInput {
				t.Fatalf("expected status input %q, got %q", tt.wantStatusInput, rs.updateInput.Status)
			}
		})
	}
}

func newTestRecordsHandler(rs service.RecordingsService) *RecordsHandler {
	return NewRecordingsHandler(slog.New(slog.NewTextHandler(io.Discard, nil)), utils.NewHandlerUtils(), rs)
}
