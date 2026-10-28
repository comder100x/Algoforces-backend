package handlers_test

import (
	"algoforces/internal/domain"
	"algoforces/internal/handlers"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type participantCall struct {
	contestID string
	userID    string
	role      string
}

// fakeContestProblemsUseCase answers the participant method from configured
// fields and records every call. Its legacy GetContestProblems method must
// never be reached by the handlers under test: both routes go through the
// gated participant method.
type fakeContestProblemsUseCase struct {
	participantResp  *domain.GetContestProblemsResponse
	participantErr   error
	participantCalls []participantCall
	legacyCalled     bool
}

func (f *fakeContestProblemsUseCase) CreateContestProblem(
	ctx context.Context,
	req *domain.CreateContestProblemRequest,
) (*domain.CreateContestProblemResponse, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeContestProblemsUseCase) BulkCreateContestProblems(
	ctx context.Context,
	req *domain.BulkCreateContestProblemRequest,
) (*domain.BulkCreateContestProblemResponse, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeContestProblemsUseCase) GetContestProblem(
	ctx context.Context,
	req *domain.GetContestProblemRequest,
) (*domain.GetContestProblemResponse, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeContestProblemsUseCase) GetContestProblems(
	ctx context.Context,
	req *domain.GetContestProblemsRequest,
) (*domain.GetContestProblemsResponse, error) {
	f.legacyCalled = true
	return nil, errors.New("legacy GetContestProblems must not be called by handlers")
}

func (f *fakeContestProblemsUseCase) GetContestProblemsForParticipant(
	ctx context.Context,
	contestID, userID, role string,
) (*domain.GetContestProblemsResponse, error) {
	f.participantCalls = append(f.participantCalls, participantCall{
		contestID: contestID,
		userID:    userID,
		role:      role,
	})
	return f.participantResp, f.participantErr
}

func (f *fakeContestProblemsUseCase) UpdateContestProblem(
	ctx context.Context,
	req *domain.UpdateContestProblemRequest,
) (*domain.UpdateContestProblemResponse, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeContestProblemsUseCase) DeleteContestProblem(
	ctx context.Context,
	req *domain.DeleteContestProblemRequest,
) (*domain.DeleteContestProblemResponse, error) {
	return nil, errors.New("not implemented")
}

func callContestProblemsHandler(
	t *testing.T,
	fn func(*gin.Context),
	params gin.Params,
	hasUser bool,
	userID string,
	hasRole bool,
	role string,
) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	if params != nil {
		c.Params = params
	}
	if hasUser {
		c.Set("user_id", userID)
	}
	if hasRole {
		c.Set("user_role", role)
	}
	fn(c)
	return w
}

func successParticipantResponse() *domain.GetContestProblemsResponse {
	return &domain.GetContestProblemsResponse{
		Problems: []domain.ContestProblemDetail{
			{
				UniqueID:      "contest-problem-id",
				ContestID:     "contest-id",
				ProblemID:     "problem-id",
				OrderPosition: 1,
				MaxPoints:     100,
			},
		},
	}
}

type contestProblemsHandlerCase struct {
	name              string
	params            gin.Params
	hasUser           bool
	userID            string
	hasRole           bool
	role              string
	fakeErr           error
	wantStatus        int
	wantErrorContains string
	wantUseCaseCall   bool
}

func runContestProblemsHandlerCases(
	t *testing.T,
	cases []contestProblemsHandlerCase,
	fn func(*handlers.ContestProblemsHandler) func(*gin.Context),
) {
	t.Helper()

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeContestProblemsUseCase{
				participantResp: successParticipantResponse(),
				participantErr:  tt.fakeErr,
			}
			handler := handlers.NewContestProblemsHandler(fake)

			w := callContestProblemsHandler(
				t,
				fn(handler),
				tt.params,
				tt.hasUser,
				tt.userID,
				tt.hasRole,
				tt.role,
			)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
			if fake.legacyCalled {
				t.Error("legacy GetContestProblems was called; both routes must use the gated participant method")
			}
			if tt.wantUseCaseCall && len(fake.participantCalls) != 1 {
				t.Fatalf("participant calls = %d, want 1", len(fake.participantCalls))
			}
			if !tt.wantUseCaseCall && len(fake.participantCalls) != 0 {
				t.Fatalf("participant calls = %d, want 0", len(fake.participantCalls))
			}

			var body map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}
			if tt.wantErrorContains != "" {
				errField, _ := body["error"].(string)
				if !strings.Contains(errField, tt.wantErrorContains) {
					t.Errorf("error = %q, want it to contain %q", errField, tt.wantErrorContains)
				}
			}
			if tt.wantStatus == http.StatusOK {
				var success struct {
					Data domain.GetContestProblemsResponse `json:"data"`
				}
				if err := json.Unmarshal(w.Body.Bytes(), &success); err != nil {
					t.Fatalf("json.Unmarshal() error = %v", err)
				}
				if len(success.Data.Problems) != 1 {
					t.Errorf("got %d problems, want 1", len(success.Data.Problems))
				}
			}
		})
	}
}

func TestGetContestProblemsForParticipantHandler(t *testing.T) {
	runContestProblemsHandlerCases(
		t,
		[]contestProblemsHandlerCase{
			{
				name:       "missing contest id returns 400 without calling the use case",
				params:     nil,
				hasUser:    true,
				userID:     "user-id",
				hasRole:    true,
				role:       "user",
				wantStatus: http.StatusBadRequest,
			},
			{
				name: "missing user id returns 401 without calling the use case",
				params: gin.Params{
					{Key: "id", Value: "contest-id"},
				},
				hasUser:    false,
				hasRole:    true,
				role:       "user",
				wantStatus: http.StatusUnauthorized,
			},
			{
				name: "missing role returns 403 without calling the use case",
				params: gin.Params{
					{Key: "id", Value: "contest-id"},
				},
				hasUser:    true,
				userID:     "user-id",
				hasRole:    false,
				wantStatus: http.StatusForbidden,
			},
			{
				name: "forbidden sentinel returns 403",
				params: gin.Params{
					{Key: "id", Value: "contest-id"},
				},
				hasUser:           true,
				userID:            "user-id",
				hasRole:           true,
				role:              "user",
				fakeErr:           domain.ErrContestProblemsForbidden,
				wantStatus:        http.StatusForbidden,
				wantErrorContains: "not visible",
				wantUseCaseCall:   true,
			},
			{
				name: "not-found sentinel returns 404",
				params: gin.Params{
					{Key: "id", Value: "contest-id"},
				},
				hasUser:           true,
				userID:            "user-id",
				hasRole:           true,
				role:              "user",
				fakeErr:           domain.ErrContestNotFound,
				wantStatus:        http.StatusNotFound,
				wantErrorContains: "contest not found",
				wantUseCaseCall:   true,
			},
			{
				name: "any other error returns a sanitised 500",
				params: gin.Params{
					{Key: "id", Value: "contest-id"},
				},
				hasUser:           true,
				userID:            "user-id",
				hasRole:           true,
				role:              "user",
				fakeErr:           errors.New("db down"),
				wantStatus:        http.StatusInternalServerError,
				wantErrorContains: "internal server error",
				wantUseCaseCall:   true,
			},
			{
				name: "success returns 200 with the problems body",
				params: gin.Params{
					{Key: "id", Value: "contest-id"},
				},
				hasUser:         true,
				userID:          "user-id",
				hasRole:         true,
				role:            "user",
				wantStatus:      http.StatusOK,
				wantUseCaseCall: true,
			},
		},
		func(h *handlers.ContestProblemsHandler) func(*gin.Context) {
			return h.GetContestProblemsForParticipant
		},
	)
}

func TestGetContestProblemsHandlerIsGated(t *testing.T) {
	runContestProblemsHandlerCases(
		t,
		[]contestProblemsHandlerCase{
			{
				name:       "missing contest id returns 400 without calling the use case",
				params:     nil,
				hasUser:    true,
				userID:     "user-id",
				hasRole:    true,
				role:       "user",
				wantStatus: http.StatusBadRequest,
			},
			{
				name: "missing user id returns 401 without calling the use case",
				params: gin.Params{
					{Key: "contestId", Value: "contest-id"},
				},
				hasUser:    false,
				hasRole:    true,
				role:       "user",
				wantStatus: http.StatusUnauthorized,
			},
			{
				name: "missing role returns 403 without calling the use case",
				params: gin.Params{
					{Key: "contestId", Value: "contest-id"},
				},
				hasUser:    true,
				userID:     "user-id",
				hasRole:    false,
				wantStatus: http.StatusForbidden,
			},
			{
				name: "forbidden sentinel returns 403",
				params: gin.Params{
					{Key: "contestId", Value: "contest-id"},
				},
				hasUser:           true,
				userID:            "user-id",
				hasRole:           true,
				role:              "user",
				fakeErr:           domain.ErrContestProblemsForbidden,
				wantStatus:        http.StatusForbidden,
				wantErrorContains: "not visible",
				wantUseCaseCall:   true,
			},
			{
				name: "not-found sentinel returns 404",
				params: gin.Params{
					{Key: "contestId", Value: "contest-id"},
				},
				hasUser:           true,
				userID:            "user-id",
				hasRole:           true,
				role:              "user",
				fakeErr:           domain.ErrContestNotFound,
				wantStatus:        http.StatusNotFound,
				wantErrorContains: "contest not found",
				wantUseCaseCall:   true,
			},
			{
				name: "any other error returns a sanitised 500",
				params: gin.Params{
					{Key: "contestId", Value: "contest-id"},
				},
				hasUser:           true,
				userID:            "user-id",
				hasRole:           true,
				role:              "user",
				fakeErr:           errors.New("db down"),
				wantStatus:        http.StatusInternalServerError,
				wantErrorContains: "internal server error",
				wantUseCaseCall:   true,
			},
			{
				name: "success returns 200 with the problems body",
				params: gin.Params{
					{Key: "contestId", Value: "contest-id"},
				},
				hasUser:         true,
				userID:          "user-id",
				hasRole:         true,
				role:            "user",
				wantStatus:      http.StatusOK,
				wantUseCaseCall: true,
			},
		},
		func(h *handlers.ContestProblemsHandler) func(*gin.Context) {
			return h.GetContestProblems
		},
	)
}

// TestGetContestProblemsHandlerForwardsIdentity pins that the legacy route
// forwards the exact contest, user, and role to the gated participant
// method, so the same gate decides for both routes.
func TestGetContestProblemsHandlerForwardsIdentity(t *testing.T) {
	fake := &fakeContestProblemsUseCase{participantResp: successParticipantResponse()}
	handler := handlers.NewContestProblemsHandler(fake)

	w := callContestProblemsHandler(
		t,
		handler.GetContestProblems,
		gin.Params{{Key: "contestId", Value: "contest-id"}},
		true,
		"user-id",
		true,
		"admin",
	)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
	}
	if fake.legacyCalled {
		t.Fatal("legacy GetContestProblems was called")
	}
	if len(fake.participantCalls) != 1 {
		t.Fatalf("participant calls = %d, want 1", len(fake.participantCalls))
	}
	got := fake.participantCalls[0]
	want := participantCall{contestID: "contest-id", userID: "user-id", role: "admin"}
	if got != want {
		t.Errorf("participant call = %+v, want %+v", got, want)
	}
}
