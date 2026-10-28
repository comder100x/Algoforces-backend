package handlers_test

import (
	"algoforces/internal/domain"
	"algoforces/internal/handlers"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakeSubmissionUseCase struct {
	domain.SubmissionUseCase
	called    bool
	contestID string
	userID    string
	role      string
	results   []domain.SubmissionDetailsResponse
	err       error
}

func (f *fakeSubmissionUseCase) ListContestSubmissions(
	ctx context.Context,
	contestID string,
	userID string,
	role string,
) ([]domain.SubmissionDetailsResponse, error) {
	f.called = true
	f.contestID = contestID
	f.userID = userID
	f.role = role
	return f.results, f.err
}

func TestListContestSubmissionsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		contestID    string
		setUser      bool
		setRole      bool
		nilUseCase   bool
		fakeErr      error
		fakeResults  []domain.SubmissionDetailsResponse
		wantStatus   int
		wantContains []string
		wantAbsent   []string
	}{
		{
			name:         "missing contest id returns 400 without calling the use case",
			contestID:    "",
			setUser:      true,
			setRole:      true,
			wantStatus:   http.StatusBadRequest,
			wantContains: []string{"Contest ID is required"},
		},
		{
			name:         "missing user returns 401",
			contestID:    "contest-1",
			setUser:      false,
			setRole:      true,
			wantStatus:   http.StatusUnauthorized,
			wantContains: []string{"User ID not found in token"},
		},
		{
			name:         "missing role returns 403",
			contestID:    "contest-1",
			setUser:      true,
			setRole:      false,
			wantStatus:   http.StatusForbidden,
			wantContains: []string{"User role not found"},
		},
		{
			name:         "unconfigured use case returns a generic 500",
			contestID:    "contest-1",
			setUser:      true,
			setRole:      true,
			nilUseCase:   true,
			wantStatus:   http.StatusInternalServerError,
			wantContains: []string{"Failed to list contest submissions"},
		},
		{
			name:         "unknown contest returns 404",
			contestID:    "contest-1",
			setUser:      true,
			setRole:      true,
			fakeErr:      fmt.Errorf("contest contest-1: %w", domain.ErrContestNotFound),
			wantStatus:   http.StatusNotFound,
			wantContains: []string{"Contest not found"},
		},
		{
			name:         "wrapped sentinel returns 404",
			contestID:    "contest-1",
			setUser:      true,
			setRole:      true,
			fakeErr:      errors.Join(errors.New("load failed"), domain.ErrContestNotFound),
			wantStatus:   http.StatusNotFound,
			wantContains: []string{"Contest not found"},
		},
		{
			name:         "use case failure returns a sanitised 500",
			contestID:    "contest-1",
			setUser:      true,
			setRole:      true,
			fakeErr:      errors.New("db exploded: connection refused"),
			wantStatus:   http.StatusInternalServerError,
			wantContains: []string{"Failed to list contest submissions"},
			wantAbsent:   []string{"db exploded"},
		},
		{
			name:        "success returns the submissions body",
			contestID:   "contest-1",
			setUser:     true,
			setRole:     true,
			fakeResults: []domain.SubmissionDetailsResponse{{UniqueID: "sub-1", UserID: "user-1"}},
			wantStatus:  http.StatusOK,
			wantContains: []string{
				"Contest submissions retrieved successfully",
				"sub-1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(
				http.MethodGet,
				"/api/contest/"+tt.contestID+"/submissions",
				nil,
			)
			if tt.contestID != "" {
				c.Params = gin.Params{{Key: "id", Value: tt.contestID}}
			}
			if tt.setUser {
				c.Set("user_id", "user-1")
			}
			if tt.setRole {
				c.Set("user_role", "user")
			}

			fake := &fakeSubmissionUseCase{
				results: tt.fakeResults,
				err:     tt.fakeErr,
			}
			var handler *handlers.SubmissionHandler
			if tt.nilUseCase {
				handler = handlers.NewSubmissionHandler(nil)
			} else {
				handler = handlers.NewSubmissionHandler(fake)
			}

			handler.ListContestSubmissions(c)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body %s)", w.Code, tt.wantStatus, w.Body.String())
			}
			body := w.Body.String()
			for _, want := range tt.wantContains {
				if !strings.Contains(body, want) {
					t.Errorf("body %q does not contain %q", body, want)
				}
			}
			for _, absent := range tt.wantAbsent {
				if strings.Contains(body, absent) {
					t.Errorf("body %q contains %q, want it sanitised", body, absent)
				}
			}
			if tt.wantStatus == http.StatusOK {
				if !fake.called {
					t.Errorf("use case was not called")
				}
				if fake.contestID != tt.contestID || fake.userID != "user-1" || fake.role != "user" {
					t.Errorf("use case called with (%q, %q, %q), want (%q, %q, %q)",
						fake.contestID, fake.userID, fake.role, tt.contestID, "user-1", "user")
				}
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				data, ok := response["data"].([]interface{})
				if !ok || len(data) != 1 {
					t.Fatalf("data = %v, want one submission", response["data"])
				}
				first, ok := data[0].(map[string]interface{})
				if !ok || first["unique_id"] != "sub-1" {
					t.Errorf("data[0] = %v, want unique_id sub-1", data[0])
				}
			}
			if tt.contestID == "" && fake.called {
				t.Errorf("use case was called for a missing contest id")
			}
		})
	}
}
