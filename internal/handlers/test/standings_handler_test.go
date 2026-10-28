package handlers_test

import (
	"algoforces/internal/domain"
	"algoforces/internal/handlers"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeStandingsUseCase struct {
	resp  *domain.StandingsResponse
	err   error
	calls int
}

func (f *fakeStandingsUseCase) GetStandings(ctx context.Context, contestID string, userID string, role string) (*domain.StandingsResponse, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	return f.resp, nil
}

func standingsHandlerContext(w *httptest.ResponseRecorder, withID bool, userID, role *string) *gin.Context {
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/contest/contest-1/standings", nil)
	if withID {
		c.Params = gin.Params{{Key: "id", Value: "contest-1"}}
	}
	if userID != nil {
		c.Set("user_id", *userID)
	}
	if role != nil {
		c.Set("user_role", *role)
	}
	return c
}

func decodeStandingsBody(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	return body
}

func TestStandingsHandlerGetStandings(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := "user-1"
	role := "user"

	tests := []struct {
		name       string
		withID     bool
		userID     *string
		role       *string
		fake       *fakeStandingsUseCase
		wantStatus int
		wantCalls  int
	}{
		{
			name:       "missing id returns 400 without calling the use case",
			withID:     false,
			fake:       &fakeStandingsUseCase{},
			wantStatus: http.StatusBadRequest,
			wantCalls:  0,
		},
		{
			name:       "missing user returns 401",
			withID:     true,
			fake:       &fakeStandingsUseCase{},
			wantStatus: http.StatusUnauthorized,
			wantCalls:  0,
		},
		{
			name:       "missing role returns 403",
			withID:     true,
			userID:     &userID,
			fake:       &fakeStandingsUseCase{},
			wantStatus: http.StatusForbidden,
			wantCalls:  0,
		},
		{
			name:       "contest not found returns 404",
			withID:     true,
			userID:     &userID,
			role:       &role,
			fake:       &fakeStandingsUseCase{err: domain.ErrContestNotFound},
			wantStatus: http.StatusNotFound,
			wantCalls:  1,
		},
		{
			name:       "standings forbidden returns 403",
			withID:     true,
			userID:     &userID,
			role:       &role,
			fake:       &fakeStandingsUseCase{err: domain.ErrStandingsForbidden},
			wantStatus: http.StatusForbidden,
			wantCalls:  1,
		},
		{
			name:       "internal error returns 500",
			withID:     true,
			userID:     &userID,
			role:       &role,
			fake:       &fakeStandingsUseCase{err: errors.New("db exploded")},
			wantStatus: http.StatusInternalServerError,
			wantCalls:  1,
		},
		{
			name:   "success returns standings body",
			withID: true,
			userID: &userID,
			role:   &role,
			fake: &fakeStandingsUseCase{resp: &domain.StandingsResponse{
				ContestID:   "contest-1",
				Status:      domain.ContestStatusLive,
				Entries:     []domain.StandingsEntry{{UserID: "user-1", Username: "alice", TotalScore: 100}},
				GeneratedAt: time.Now(),
			}},
			wantStatus: http.StatusOK,
			wantCalls:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c := standingsHandlerContext(w, tt.withID, tt.userID, tt.role)
			handler := handlers.NewStandingsHandler(tt.fake)

			handler.GetStandings(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, tt.wantCalls, tt.fake.calls)
			body := decodeStandingsBody(t, w)
			assert.Equal(t, float64(tt.wantStatus), body["status_code"])
			if tt.wantStatus == http.StatusOK {
				data, ok := body["data"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "contest-1", data["contest_id"])
				assert.Equal(t, domain.ContestStatusLive, data["status"])
			}
		})
	}
}

func TestStandingsHandlerGetStandingsSanitisesInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := "user-1"
	role := "user"
	w := httptest.NewRecorder()
	c := standingsHandlerContext(w, true, &userID, &role)
	handler := handlers.NewStandingsHandler(&fakeStandingsUseCase{err: errors.New("db exploded: connection refused")})

	handler.GetStandings(c)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	body := decodeStandingsBody(t, w)
	assert.Equal(t, "internal server error", body["error"])
	assert.NotContains(t, w.Body.String(), "db exploded")
}

func TestStandingsHandlerGetStandingsNilUseCase(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := "user-1"
	role := "user"
	w := httptest.NewRecorder()
	c := standingsHandlerContext(w, true, &userID, &role)
	handler := handlers.NewStandingsHandler(nil)

	handler.GetStandings(c)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	body := decodeStandingsBody(t, w)
	assert.Equal(t, "internal server error", body["error"])
}
