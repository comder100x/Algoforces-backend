package handlers_test

import (
	"algoforces/internal/domain"
	"algoforces/internal/handlers"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakeSubmissionUseCase struct {
	result       *domain.CreateSubmissionResponse
	err          error
	createCalled bool
}

func (f *fakeSubmissionUseCase) CreateNewSubmission(ctx context.Context, req *domain.CreateSubmissionRequest) (*domain.CreateSubmissionResponse, error) {
	f.createCalled = true
	return f.result, f.err
}

func (f *fakeSubmissionUseCase) GetSubmissionDetails(ctx context.Context, uniqueID string, includeHidden bool) (*domain.SubmissionDetailsResponse, error) {
	return nil, nil
}

func (f *fakeSubmissionUseCase) UpdateSubmissionStatus(ctx context.Context, submissionID string, status string) error {
	return nil
}

func (f *fakeSubmissionUseCase) UpdateSubmissionResult(ctx context.Context, submissionID string, req *domain.UpdateSubmissionResultRequest) (*domain.UpdateSubmissionResultResponse, error) {
	return nil, nil
}

func (f *fakeSubmissionUseCase) JudgeSubmissionCallback(ctx context.Context, req *domain.JudgeSubmissionCallbackRequest) error {
	return nil
}

func TestCreateSubmissionHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	validBody := `{"contest_id":"11111111-1111-1111-1111-111111111111","problem_id":"22222222-2222-2222-2222-222222222222","code":"print(1)","language":"python"}`

	tests := []struct {
		name             string
		body             string
		setUser          bool
		useCaseErr       error
		wantStatus       int
		wantMessage      string
		wantError        string
		wantCreateCalled bool
	}{
		{
			name:        "invalid request body",
			body:        `{bad json`,
			setUser:     true,
			wantStatus:  http.StatusBadRequest,
			wantMessage: "Invalid Request Body",
		},
		{
			name:        "missing user id",
			body:        validBody,
			wantStatus:  http.StatusUnauthorized,
			wantMessage: "User ID not found in token",
		},
		{
			name:             "contest not found",
			body:             validBody,
			setUser:          true,
			useCaseErr:       domain.ErrContestNotFound,
			wantStatus:       http.StatusNotFound,
			wantMessage:      "Contest not found",
			wantError:        "contest not found",
			wantCreateCalled: true,
		},
		{
			name:             "contest not visible",
			body:             validBody,
			setUser:          true,
			useCaseErr:       domain.ErrContestNotVisible,
			wantStatus:       http.StatusForbidden,
			wantMessage:      "Contest is not visible",
			wantError:        "contest is not visible",
			wantCreateCalled: true,
		},
		{
			name:             "contest not live",
			body:             validBody,
			setUser:          true,
			useCaseErr:       domain.ErrContestNotLive,
			wantStatus:       http.StatusForbidden,
			wantMessage:      "Contest is not live",
			wantError:        "contest is not live",
			wantCreateCalled: true,
		},
		{
			name:             "not registered",
			body:             validBody,
			setUser:          true,
			useCaseErr:       domain.ErrNotRegistered,
			wantStatus:       http.StatusForbidden,
			wantMessage:      "You are not registered for this contest",
			wantError:        "user is not registered for this contest",
			wantCreateCalled: true,
		},
		{
			name:             "wrapped sentinel still maps to its status",
			body:             validBody,
			setUser:          true,
			useCaseErr:       fmt.Errorf("gate check failed: %w", domain.ErrNotRegistered),
			wantStatus:       http.StatusForbidden,
			wantMessage:      "You are not registered for this contest",
			wantError:        "gate check failed: user is not registered for this contest",
			wantCreateCalled: true,
		},
		{
			name:             "problem not in contest",
			body:             validBody,
			setUser:          true,
			useCaseErr:       domain.ErrProblemNotInContest,
			wantStatus:       http.StatusConflict,
			wantMessage:      "Problem is not part of this contest",
			wantError:        "problem is not part of this contest",
			wantCreateCalled: true,
		},
		{
			name:             "unexpected error is a sanitised 500",
			body:             validBody,
			setUser:          true,
			useCaseErr:       errors.New("pq: connection refused"),
			wantStatus:       http.StatusInternalServerError,
			wantMessage:      "Failed to create submission",
			wantError:        "internal server error",
			wantCreateCalled: true,
		},
		{
			name:             "success",
			body:             validBody,
			setUser:          true,
			wantStatus:       http.StatusCreated,
			wantMessage:      "Submission created successfully",
			wantCreateCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeSubmissionUseCase{
				result: &domain.CreateSubmissionResponse{UniqueID: "sub-1"},
				err:    tt.useCaseErr,
			}
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/submission/create", bytes.NewBufferString(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")
			if tt.setUser {
				c.Set("user_id", "user-1")
			}

			handlers.NewSubmissionHandler(fake).CreateSubmission(c)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body = %s", w.Code, tt.wantStatus, w.Body.String())
			}
			if fake.createCalled != tt.wantCreateCalled {
				t.Fatalf("createCalled = %v, want %v", fake.createCalled, tt.wantCreateCalled)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}
			if response["message"] != tt.wantMessage {
				t.Errorf("message = %v, want %q", response["message"], tt.wantMessage)
			}
			if tt.wantError != "" && response["error"] != tt.wantError {
				t.Errorf("error = %v, want %q", response["error"], tt.wantError)
			}
			if tt.wantStatus == http.StatusCreated {
				data, ok := response["data"].(map[string]interface{})
				if !ok || data["unique_id"] != "sub-1" {
					t.Errorf("data = %v, want unique_id sub-1", response["data"])
				}
			}
		})
	}
}
