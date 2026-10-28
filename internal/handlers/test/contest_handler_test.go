package handlers_test

import (
	"algoforces/internal/domain"
	"algoforces/internal/handlers"
	"algoforces/internal/utils"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type fakeContestUseCase struct {
	createErr  error
	updateErr  error
	detailsErr error
	deleteErr  error
	startErr   error
	endErr     error
	allErr     error
	visibleErr error

	detailsResp *domain.CreateContestResponse
	allResp     []domain.CreateContestResponse
	visibleResp []domain.CreateContestResponse
}

func (f *fakeContestUseCase) CreateContest(
	_ context.Context,
	_ *domain.CreateContestRequest,
	_ string,
) (*domain.CreateContestResponse, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	return &domain.CreateContestResponse{Id: uuid.New().String(), Name: "Created"}, nil
}

func (f *fakeContestUseCase) UpdateContest(
	_ context.Context,
	_ *domain.UpdateContestRequest,
) (*domain.UpdateContestResponse, error) {
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	return &domain.UpdateContestResponse{Id: uuid.New().String(), Name: "Updated"}, nil
}

func (f *fakeContestUseCase) GetContestDetails(_ context.Context, _ string) (*domain.CreateContestResponse, error) {
	if f.detailsErr != nil {
		return nil, f.detailsErr
	}
	if f.detailsResp != nil {
		return f.detailsResp, nil
	}
	return &domain.CreateContestResponse{Id: uuid.New().String(), Name: "Details"}, nil
}

func (f *fakeContestUseCase) DeleteContest(_ context.Context, _ *domain.DeleteContestRequest) error {
	return f.deleteErr
}

func (f *fakeContestUseCase) StartContest(_ context.Context, _ string) error {
	return f.startErr
}

func (f *fakeContestUseCase) EndContest(_ context.Context, _ string) error {
	return f.endErr
}

func (f *fakeContestUseCase) GetAllContests(_ context.Context) ([]domain.CreateContestResponse, error) {
	if f.allErr != nil {
		return nil, f.allErr
	}
	if f.allResp != nil {
		return f.allResp, nil
	}
	return []domain.CreateContestResponse{}, nil
}

func (f *fakeContestUseCase) GetVisibleContests(_ context.Context) ([]domain.CreateContestResponse, error) {
	if f.visibleErr != nil {
		return nil, f.visibleErr
	}
	if f.visibleResp != nil {
		return f.visibleResp, nil
	}
	return []domain.CreateContestResponse{}, nil
}

func newContestTestContext(
	t *testing.T,
	method, target, body string,
	params gin.Params,
	withUser bool,
) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, target, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, target, nil)
	}
	c.Request = req
	c.Params = params
	if withUser {
		c.Set("user_id", "user-1")
	}
	return c, w
}

func decodeErrorResponse(t *testing.T, w *httptest.ResponseRecorder) utils.ErrorResponse {
	t.Helper()
	var resp utils.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error response %q: %v", w.Body.String(), err)
	}
	return resp
}

func TestContestHandlerCreateContest(t *testing.T) {
	validBody := `{"name":"X","start_time":"2026-10-29T10:00:00Z","end_time":"2026-10-29T12:00:00Z","duration":120}`

	tests := []struct {
		name       string
		body       string
		withUser   bool
		createErr  error
		wantStatus int
		wantError  string
	}{
		{name: "success", body: validBody, withUser: true, wantStatus: http.StatusCreated},
		{name: "invalid json", body: `{bad`, withUser: true, wantStatus: http.StatusBadRequest},
		{name: "missing user", body: validBody, withUser: false, wantStatus: http.StatusUnauthorized},
		{
			name: "overlap conflict", body: validBody, withUser: true,
			createErr: domain.ErrContestOverlap, wantStatus: http.StatusConflict,
			wantError: domain.ErrContestOverlap.Error(),
		},
		{
			name: "setter not authorized", body: validBody, withUser: true,
			createErr: domain.ErrProblemSetterNotAuthorized, wantStatus: http.StatusBadRequest,
			wantError: domain.ErrProblemSetterNotAuthorized.Error(),
		},
		{
			name: "generic error sanitised", body: validBody, withUser: true,
			createErr: errors.New("pq: connection refused"), wantStatus: http.StatusInternalServerError,
			wantError: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := handlers.NewContestHandler(&fakeContestUseCase{createErr: tt.createErr})
			c, w := newContestTestContext(t, http.MethodPost, "/api/contest/create", tt.body, nil, tt.withUser)
			h.CreateContest(c)
			if w.Code != tt.wantStatus {
				t.Fatalf("CreateContest() status = %d, want %d (body %s)", w.Code, tt.wantStatus, w.Body.String())
			}
			if tt.wantError != "" && decodeErrorResponse(t, w).Error != tt.wantError {
				t.Errorf("CreateContest() error = %q, want %q", decodeErrorResponse(t, w).Error, tt.wantError)
			}
		})
	}
}

func TestContestHandlerGetContestDetails(t *testing.T) {
	id := uuid.New().String()

	tests := []struct {
		name       string
		params     gin.Params
		detailsErr error
		wantStatus int
		wantError  string
	}{
		{name: "success", params: gin.Params{{Key: "id", Value: id}}, wantStatus: http.StatusOK},
		{name: "missing id", params: nil, wantStatus: http.StatusBadRequest, wantError: "contest ID is required"},
		{
			name: "invalid uuid", params: gin.Params{{Key: "id", Value: "not-a-uuid"}},
			wantStatus: http.StatusBadRequest, wantError: "contest ID must be a valid UUID",
		},
		{
			name: "not found", params: gin.Params{{Key: "id", Value: id}},
			detailsErr: domain.ErrContestNotFound, wantStatus: http.StatusNotFound,
			wantError: domain.ErrContestNotFound.Error(),
		},
		{
			name: "generic error sanitised", params: gin.Params{{Key: "id", Value: id}},
			detailsErr: errors.New("pq: connection refused"), wantStatus: http.StatusInternalServerError,
			wantError: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := handlers.NewContestHandler(&fakeContestUseCase{detailsErr: tt.detailsErr})
			c, w := newContestTestContext(t, http.MethodGet, "/api/contest/"+id, "", tt.params, true)
			h.GetContestDetails(c)
			if w.Code != tt.wantStatus {
				t.Fatalf("GetContestDetails() status = %d, want %d (body %s)", w.Code, tt.wantStatus, w.Body.String())
			}
			if tt.wantError != "" && decodeErrorResponse(t, w).Error != tt.wantError {
				t.Errorf("GetContestDetails() error = %q, want %q", decodeErrorResponse(t, w).Error, tt.wantError)
			}
		})
	}
}

func TestContestHandlerUpdateContest(t *testing.T) {
	validBody := `{"id":"` + uuid.New().String() + `","name":"X","start_time":"2026-10-29T10:00:00Z","end_time":"2026-10-29T12:00:00Z","duration":120}`

	tests := []struct {
		name       string
		body       string
		updateErr  error
		wantStatus int
		wantError  string
	}{
		{name: "success", body: validBody, wantStatus: http.StatusOK},
		{name: "invalid json", body: `{bad`, wantStatus: http.StatusBadRequest},
		{
			name: "not found", body: validBody, updateErr: domain.ErrContestNotFound,
			wantStatus: http.StatusNotFound, wantError: domain.ErrContestNotFound.Error(),
		},
		{
			name: "time locked", body: validBody, updateErr: domain.ErrContestTimeLocked,
			wantStatus: http.StatusConflict, wantError: domain.ErrContestTimeLocked.Error(),
		},
		{
			name: "generic error sanitised", body: validBody, updateErr: errors.New("pq: connection refused"),
			wantStatus: http.StatusInternalServerError, wantError: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := handlers.NewContestHandler(&fakeContestUseCase{updateErr: tt.updateErr})
			c, w := newContestTestContext(t, http.MethodPut, "/api/contest/update", tt.body, nil, true)
			h.UpdateContest(c)
			if w.Code != tt.wantStatus {
				t.Fatalf("UpdateContest() status = %d, want %d (body %s)", w.Code, tt.wantStatus, w.Body.String())
			}
			if tt.wantError != "" && decodeErrorResponse(t, w).Error != tt.wantError {
				t.Errorf("UpdateContest() error = %q, want %q", decodeErrorResponse(t, w).Error, tt.wantError)
			}
		})
	}
}

func TestContestHandlerDeleteContest(t *testing.T) {
	id := uuid.New().String()

	tests := []struct {
		name       string
		params     gin.Params
		deleteErr  error
		wantStatus int
		wantError  string
	}{
		{name: "success", params: gin.Params{{Key: "id", Value: id}}, wantStatus: http.StatusOK},
		{name: "missing id", params: nil, wantStatus: http.StatusBadRequest, wantError: "contest ID is required"},
		{
			name: "invalid uuid", params: gin.Params{{Key: "id", Value: "not-a-uuid"}},
			wantStatus: http.StatusBadRequest, wantError: "contest ID must be a valid UUID",
		},
		{
			name: "live conflict", params: gin.Params{{Key: "id", Value: id}},
			deleteErr: domain.ErrContestLive, wantStatus: http.StatusConflict,
			wantError: domain.ErrContestLive.Error(),
		},
		{
			name: "not found", params: gin.Params{{Key: "id", Value: id}},
			deleteErr: domain.ErrContestNotFound, wantStatus: http.StatusNotFound,
			wantError: domain.ErrContestNotFound.Error(),
		},
		{
			name: "generic error sanitised", params: gin.Params{{Key: "id", Value: id}},
			deleteErr: errors.New("pq: connection refused"), wantStatus: http.StatusInternalServerError,
			wantError: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := handlers.NewContestHandler(&fakeContestUseCase{deleteErr: tt.deleteErr})
			c, w := newContestTestContext(t, http.MethodDelete, "/api/contest/"+id, "", tt.params, true)
			h.DeleteContest(c)
			if w.Code != tt.wantStatus {
				t.Fatalf("DeleteContest() status = %d, want %d (body %s)", w.Code, tt.wantStatus, w.Body.String())
			}
			if tt.wantError != "" && decodeErrorResponse(t, w).Error != tt.wantError {
				t.Errorf("DeleteContest() error = %q, want %q", decodeErrorResponse(t, w).Error, tt.wantError)
			}
		})
	}
}

func TestContestHandlerStartContest(t *testing.T) {
	id := uuid.New().String()

	tests := []struct {
		name       string
		params     gin.Params
		withUser   bool
		startErr   error
		wantStatus int
		wantError  string
	}{
		{name: "success", params: gin.Params{{Key: "id", Value: id}}, withUser: true, wantStatus: http.StatusOK},
		{
			name: "missing id", params: nil, withUser: true,
			wantStatus: http.StatusBadRequest, wantError: "contest ID is required",
		},
		{
			name: "invalid uuid", params: gin.Params{{Key: "id", Value: "not-a-uuid"}}, withUser: true,
			wantStatus: http.StatusBadRequest, wantError: "contest ID must be a valid UUID",
		},
		{
			name: "missing user", params: gin.Params{{Key: "id", Value: id}}, withUser: false,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "ended conflict", params: gin.Params{{Key: "id", Value: id}}, withUser: true,
			startErr: domain.ErrContestEnded, wantStatus: http.StatusConflict,
			wantError: domain.ErrContestEnded.Error(),
		},
		{
			name: "no problems", params: gin.Params{{Key: "id", Value: id}}, withUser: true,
			startErr: domain.ErrContestHasNoProblems, wantStatus: http.StatusBadRequest,
			wantError: domain.ErrContestHasNoProblems.Error(),
		},
		{
			name: "generic error sanitised", params: gin.Params{{Key: "id", Value: id}}, withUser: true,
			startErr: errors.New("pq: connection refused"), wantStatus: http.StatusInternalServerError,
			wantError: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := handlers.NewContestHandler(&fakeContestUseCase{startErr: tt.startErr})
			c, w := newContestTestContext(t, http.MethodPost, "/api/contest/"+id+"/start", "", tt.params, tt.withUser)
			h.StartContest(c)
			if w.Code != tt.wantStatus {
				t.Fatalf("StartContest() status = %d, want %d (body %s)", w.Code, tt.wantStatus, w.Body.String())
			}
			if tt.wantError != "" && decodeErrorResponse(t, w).Error != tt.wantError {
				t.Errorf("StartContest() error = %q, want %q", decodeErrorResponse(t, w).Error, tt.wantError)
			}
		})
	}
}

func TestContestHandlerEndContest(t *testing.T) {
	id := uuid.New().String()

	tests := []struct {
		name       string
		params     gin.Params
		withUser   bool
		endErr     error
		wantStatus int
		wantError  string
	}{
		{name: "success", params: gin.Params{{Key: "id", Value: id}}, withUser: true, wantStatus: http.StatusOK},
		{
			name: "invalid uuid", params: gin.Params{{Key: "id", Value: "not-a-uuid"}}, withUser: true,
			wantStatus: http.StatusBadRequest, wantError: "contest ID must be a valid UUID",
		},
		{
			name: "missing user", params: gin.Params{{Key: "id", Value: id}}, withUser: false,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "not found", params: gin.Params{{Key: "id", Value: id}}, withUser: true,
			endErr: domain.ErrContestNotFound, wantStatus: http.StatusNotFound,
			wantError: domain.ErrContestNotFound.Error(),
		},
		{
			name: "generic error sanitised", params: gin.Params{{Key: "id", Value: id}}, withUser: true,
			endErr: errors.New("pq: connection refused"), wantStatus: http.StatusInternalServerError,
			wantError: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := handlers.NewContestHandler(&fakeContestUseCase{endErr: tt.endErr})
			c, w := newContestTestContext(t, http.MethodPost, "/api/contest/"+id+"/end", "", tt.params, tt.withUser)
			h.EndContest(c)
			if w.Code != tt.wantStatus {
				t.Fatalf("EndContest() status = %d, want %d (body %s)", w.Code, tt.wantStatus, w.Body.String())
			}
			if tt.wantError != "" && decodeErrorResponse(t, w).Error != tt.wantError {
				t.Errorf("EndContest() error = %q, want %q", decodeErrorResponse(t, w).Error, tt.wantError)
			}
		})
	}
}

func TestContestHandlerListContests(t *testing.T) {
	tests := []struct {
		name       string
		handler    func(*handlers.ContestHandler, *gin.Context)
		fake       *fakeContestUseCase
		wantStatus int
		wantError  string
		wantEmpty  bool
	}{
		{
			name:       "get all success empty list",
			handler:    func(h *handlers.ContestHandler, c *gin.Context) { h.GetAllContests(c) },
			fake:       &fakeContestUseCase{allResp: []domain.CreateContestResponse{}},
			wantStatus: http.StatusOK, wantEmpty: true,
		},
		{
			name:       "get all error sanitised",
			handler:    func(h *handlers.ContestHandler, c *gin.Context) { h.GetAllContests(c) },
			fake:       &fakeContestUseCase{allErr: errors.New("pq: connection refused")},
			wantStatus: http.StatusInternalServerError, wantError: "internal server error",
		},
		{
			name:       "get visible success empty list",
			handler:    func(h *handlers.ContestHandler, c *gin.Context) { h.GetVisibleContests(c) },
			fake:       &fakeContestUseCase{visibleResp: []domain.CreateContestResponse{}},
			wantStatus: http.StatusOK, wantEmpty: true,
		},
		{
			name:       "get visible error sanitised",
			handler:    func(h *handlers.ContestHandler, c *gin.Context) { h.GetVisibleContests(c) },
			fake:       &fakeContestUseCase{visibleErr: errors.New("pq: connection refused")},
			wantStatus: http.StatusInternalServerError, wantError: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := handlers.NewContestHandler(tt.fake)
			c, w := newContestTestContext(t, http.MethodGet, "/api/contest/all", "", nil, true)
			tt.handler(h, c)
			if w.Code != tt.wantStatus {
				t.Fatalf("list status = %d, want %d (body %s)", w.Code, tt.wantStatus, w.Body.String())
			}
			if tt.wantError != "" && decodeErrorResponse(t, w).Error != tt.wantError {
				t.Errorf("list error = %q, want %q", decodeErrorResponse(t, w).Error, tt.wantError)
			}
			if tt.wantEmpty && !strings.Contains(w.Body.String(), `"data":[]`) {
				t.Errorf("list body = %s, want empty array data ([] not null)", w.Body.String())
			}
		})
	}
}
