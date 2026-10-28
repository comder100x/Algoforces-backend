package services_test

import (
	"algoforces/internal/domain"
	"algoforces/internal/services"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

// Hand fakes for the repositories behind contestProblemsService. Each fake
// answers from configured fields; unconfigured methods fail loudly so a
// test that reaches them by mistake cannot pass silently.

type fakeContestRepository struct {
	contest *domain.Contest
	err     error
}

func (f *fakeContestRepository) CreateContest(
	ctx context.Context,
	contest *domain.Contest,
) error {
	return errors.New("fakeContestRepository.CreateContest not configured")
}

func (f *fakeContestRepository) GetByID(
	ctx context.Context,
	id string,
) (*domain.Contest, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.contest, nil
}

func (f *fakeContestRepository) UpdateContest(
	ctx context.Context,
	contest *domain.Contest,
) error {
	return errors.New("fakeContestRepository.UpdateContest not configured")
}

func (f *fakeContestRepository) DeleteContest(ctx context.Context, id string) error {
	return errors.New("fakeContestRepository.DeleteContest not configured")
}

func (f *fakeContestRepository) CheckContestInTimeWindow(
	ctx context.Context,
	startTime, endTime time.Time,
) ([]domain.Contest, error) {
	return nil, errors.New("fakeContestRepository.CheckContestInTimeWindow not configured")
}

func (f *fakeContestRepository) GetAllContests(
	ctx context.Context,
) ([]domain.Contest, error) {
	return nil, errors.New("fakeContestRepository.GetAllContests not configured")
}

type fakeProblemRepository struct {
	problem *domain.Problem
	err     error
}

func (f *fakeProblemRepository) CreateProblem(
	ctx context.Context,
	problem *domain.Problem,
) error {
	return errors.New("fakeProblemRepository.CreateProblem not configured")
}

func (f *fakeProblemRepository) GetProblemByID(
	ctx context.Context,
	id string,
) (*domain.Problem, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.problem, nil
}

func (f *fakeProblemRepository) UpdateProblem(
	ctx context.Context,
	problem *domain.Problem,
) error {
	return errors.New("fakeProblemRepository.UpdateProblem not configured")
}

func (f *fakeProblemRepository) DeleteProblem(ctx context.Context, id string) error {
	return errors.New("fakeProblemRepository.DeleteProblem not configured")
}

func (f *fakeProblemRepository) GetAllProblems(
	ctx context.Context,
	pageOffset int,
	limit int,
) (*domain.AllProblemListResponse, error) {
	return nil, errors.New("fakeProblemRepository.GetAllProblems not configured")
}

type fakeContestProblemsRepository struct {
	problems []domain.ContestProblemDetail
	err      error
}

func (f *fakeContestProblemsRepository) CreateContestProblem(
	ctx context.Context,
	contestProblem *domain.ContestProblems,
) error {
	return errors.New("fakeContestProblemsRepository.CreateContestProblem not configured")
}

func (f *fakeContestProblemsRepository) GetContestProblemByID(
	ctx context.Context,
	uniqueID string,
) (*domain.ContestProblems, error) {
	return nil, errors.New("fakeContestProblemsRepository.GetContestProblemByID not configured")
}

func (f *fakeContestProblemsRepository) GetContestProblemsByContestID(
	ctx context.Context,
	contestID string,
) ([]domain.ContestProblems, error) {
	return nil, errors.New("fakeContestProblemsRepository.GetContestProblemsByContestID not configured")
}

func (f *fakeContestProblemsRepository) UpdateContestProblem(
	ctx context.Context,
	contestProblem *domain.ContestProblems,
) error {
	return errors.New("fakeContestProblemsRepository.UpdateContestProblem not configured")
}

func (f *fakeContestProblemsRepository) DeleteContestProblem(
	ctx context.Context,
	uniqueID string,
) error {
	return errors.New("fakeContestProblemsRepository.DeleteContestProblem not configured")
}

func (f *fakeContestProblemsRepository) GetContestProblemsByContestIDWithDetails(
	ctx context.Context,
	contestID string,
) ([]domain.ContestProblemDetail, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.problems, nil
}

type fakeContestRegisterRepository struct {
	registration *domain.ContestRegistration
	err          error
}

func (f *fakeContestRegisterRepository) CreateContestRegistration(
	ctx context.Context,
	userID string,
	contestID string,
) (*domain.ContestRegistration, error) {
	return nil, errors.New("fakeContestRegisterRepository.CreateContestRegistration not configured")
}

func (f *fakeContestRegisterRepository) GetRegistrationByUserAndContest(
	ctx context.Context,
	userID string,
	contestID string,
) (*domain.ContestRegistration, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.registration, nil
}

func (f *fakeContestRegisterRepository) UpdateRegistrationStatus(
	ctx context.Context,
	userID string,
	contestID string,
	status string,
) error {
	return errors.New("fakeContestRegisterRepository.UpdateRegistrationStatus not configured")
}

func (f *fakeContestRegisterRepository) GetAllRegistrationsByUserID(
	ctx context.Context,
	userID string,
) ([]domain.ContestRegistration, error) {
	return nil, errors.New("fakeContestRegisterRepository.GetAllRegistrationsByUserID not configured")
}

func (f *fakeContestRegisterRepository) GetAllRegistrationsForAdmin(
	ctx context.Context,
) ([]domain.ContestRegistration, error) {
	return nil, errors.New("fakeContestRegisterRepository.GetAllRegistrationsForAdmin not configured")
}

func liveServiceTestContest(now time.Time) *domain.Contest {
	return &domain.Contest{
		Id:             "contest-id",
		Name:           "Weekly Contest",
		StartTime:      now.Add(-time.Hour),
		EndTime:        now.Add(time.Hour),
		IsActive:       true,
		CreatedBy:      "creator-id",
		ProblemSetters: pq.StringArray{"setter-id"},
	}
}

func TestGetContestProblemsForParticipant(t *testing.T) {
	now := time.Now()
	detail := func(order int) domain.ContestProblemDetail {
		return domain.ContestProblemDetail{
			UniqueID:      "contest-problem-id",
			ContestID:     "contest-id",
			ProblemID:     "problem-id",
			OrderPosition: order,
			MaxPoints:     100,
		}
	}

	tests := []struct {
		name            string
		contest         *domain.Contest
		contestErr      error
		registration    *domain.ContestRegistration
		registrationErr error
		problems        []domain.ContestProblemDetail
		problemsErr     error
		userID          string
		role            string
		wantErrIs       error
		wantErrContains string
		wantOrder       []int
	}{
		{
			name:       "contest missing maps to ErrContestNotFound",
			contestErr: gorm.ErrRecordNotFound,
			userID:     "admin-id",
			role:       "admin",
			wantErrIs:  domain.ErrContestNotFound,
		},
		{
			name:            "contest lookup failure other than missing row is wrapped, not not-found",
			contestErr:      errors.New("db down"),
			userID:          "admin-id",
			role:            "admin",
			wantErrContains: "db down",
		},
		{
			name:            "missing registration row means unregistered, not an error",
			contest:         liveServiceTestContest(now),
			registrationErr: gorm.ErrRecordNotFound,
			userID:          "user-id",
			role:            "user",
			wantErrIs:       domain.ErrContestProblemsForbidden,
		},
		{
			name:            "missing registration row does not block an admin",
			contest:         liveServiceTestContest(now),
			registrationErr: gorm.ErrRecordNotFound,
			problems:        []domain.ContestProblemDetail{detail(1)},
			userID:          "admin-id",
			role:            "admin",
			wantOrder:       []int{1},
		},
		{
			name:            "registration lookup failure other than missing row propagates",
			contest:         liveServiceTestContest(now),
			registrationErr: errors.New("db down"),
			userID:          "user-id",
			role:            "user",
			wantErrContains: "db down",
		},
		{
			name:            "problems repo error propagates wrapped",
			contest:         liveServiceTestContest(now),
			registration:    registrationWithStatus("registered"),
			problemsErr:     errors.New("db down"),
			userID:          "user-id",
			role:            "user",
			wantErrContains: "db down",
		},
		{
			name:         "unregistered caller is forbidden during the live window",
			contest:      liveServiceTestContest(now),
			registration: nil,
			userID:       "user-id",
			role:         "user",
			wantErrIs:    domain.ErrContestProblemsForbidden,
		},
		{
			name:         "registered participant gets problems in repo order",
			contest:      liveServiceTestContest(now),
			registration: registrationWithStatus("registered"),
			problems: []domain.ContestProblemDetail{
				detail(2),
				detail(1),
			},
			userID:    "user-id",
			role:      "user",
			wantOrder: []int{2, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := services.NewContestProblemsService(
				&fakeContestProblemsRepository{problems: tt.problems, err: tt.problemsErr},
				&fakeContestRepository{contest: tt.contest, err: tt.contestErr},
				&fakeProblemRepository{},
				&fakeContestRegisterRepository{
					registration: tt.registration,
					err:          tt.registrationErr,
				},
			)

			resp, err := svc.GetContestProblemsForParticipant(
				context.Background(),
				"contest-id",
				tt.userID,
				tt.role,
			)

			if tt.wantErrIs != nil {
				if !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("error = %v, want errors.Is %v", err, tt.wantErrIs)
				}
				if resp != nil {
					t.Errorf("response = %v, want nil on error", resp)
				}
				return
			}
			if tt.wantErrContains != "" {
				if err == nil {
					t.Fatalf("error = nil, want error containing %q", tt.wantErrContains)
				}
				if !strings.Contains(err.Error(), tt.wantErrContains) {
					t.Errorf("error = %v, want it to contain %q", err, tt.wantErrContains)
				}
				if errors.Is(err, domain.ErrContestNotFound) {
					t.Errorf("error = %v, must not be ErrContestNotFound", err)
				}
				if errors.Is(err, domain.ErrContestProblemsForbidden) {
					t.Errorf("error = %v, must not be ErrContestProblemsForbidden", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("error = %v, want nil", err)
			}
			if resp == nil {
				t.Fatal("response = nil, want non-nil")
			}
			if len(tt.wantOrder) != len(resp.Problems) {
				t.Fatalf("got %d problems, want %d", len(resp.Problems), len(tt.wantOrder))
			}
			for i, want := range tt.wantOrder {
				if resp.Problems[i].OrderPosition != want {
					t.Errorf(
						"problems[%d].OrderPosition = %d, want %d (order must pass through unchanged)",
						i,
						resp.Problems[i].OrderPosition,
						want,
					)
				}
			}
		})
	}
}

// TestGetContestProblemsForParticipantResponseContainsNoTestCaseData
// marshals the ACTUAL service return, built from fakes, and asserts the
// participant-facing JSON carries the problem statement and no test-case
// input/output anywhere.
func TestGetContestProblemsForParticipantResponseContainsNoTestCaseData(t *testing.T) {
	now := time.Now()
	svc := services.NewContestProblemsService(
		&fakeContestProblemsRepository{
			problems: []domain.ContestProblemDetail{
				{
					UniqueID:      "contest-problem-id",
					ContestID:     "contest-id",
					ProblemID:     "problem-id",
					OrderPosition: 1,
					MaxPoints:     100,
					Problem: &domain.Problem{
						UniqueID:   "problem-id",
						Title:      "Two Sum",
						Statement:  "Find two numbers that add up to the target.",
						Difficulty: "easy",
					},
				},
			},
		},
		&fakeContestRepository{contest: liveServiceTestContest(now)},
		&fakeProblemRepository{},
		&fakeContestRegisterRepository{},
	)

	resp, err := svc.GetContestProblemsForParticipant(context.Background(), "contest-id", "admin-id", "admin")
	if err != nil {
		t.Fatalf("GetContestProblemsForParticipant() error = %v", err)
	}

	payload, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	body := string(payload)

	if !strings.Contains(body, `"statement":"Find two numbers that add up to the target."`) {
		t.Errorf("response JSON should include the problem statement, got: %s", body)
	}
	for _, forbidden := range []string{"test_case", "testcase", "TestCase", "expected_output", `"input"`, `"output"`} {
		if strings.Contains(body, forbidden) {
			t.Errorf("response JSON must not contain %q, got: %s", forbidden, body)
		}
	}
}

// TestGetContestProblemsForParticipantEmptyProblemsSerialiseAsArray pins
// that a contest with no problems returns [] in JSON, never null.
func TestGetContestProblemsForParticipantEmptyProblemsSerialiseAsArray(t *testing.T) {
	now := time.Now()
	svc := services.NewContestProblemsService(
		&fakeContestProblemsRepository{problems: nil},
		&fakeContestRepository{contest: liveServiceTestContest(now)},
		&fakeProblemRepository{},
		&fakeContestRegisterRepository{},
	)

	resp, err := svc.GetContestProblemsForParticipant(context.Background(), "contest-id", "admin-id", "admin")
	if err != nil {
		t.Fatalf("GetContestProblemsForParticipant() error = %v", err)
	}
	if resp.Problems == nil {
		t.Fatal("response.Problems is nil, want an empty non-nil slice")
	}

	payload, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if !strings.Contains(string(payload), `"problems":[]`) {
		t.Errorf("response JSON should contain \"problems\":[], got: %s", payload)
	}
}

// TestGetContestProblemsSentinelAndEmptySlice covers the legacy list
// method kept on the use-case interface: a missing contest maps to the
// not-found sentinel and an empty list still serialises as [].
func TestGetContestProblemsSentinelAndEmptySlice(t *testing.T) {
	svc := services.NewContestProblemsService(
		&fakeContestProblemsRepository{problems: nil},
		&fakeContestRepository{err: gorm.ErrRecordNotFound},
		&fakeProblemRepository{},
		&fakeContestRegisterRepository{},
	)

	_, err := svc.GetContestProblems(
		context.Background(),
		&domain.GetContestProblemsRequest{ContestID: "contest-id"},
	)
	if !errors.Is(err, domain.ErrContestNotFound) {
		t.Fatalf("error = %v, want errors.Is ErrContestNotFound", err)
	}
}

// TestContestProblemsServiceNilRequests pins that nil request pointers are
// rejected with an error instead of panicking on a nil dereference.
func TestContestProblemsServiceNilRequests(t *testing.T) {
	svc := services.NewContestProblemsService(
		&fakeContestProblemsRepository{},
		&fakeContestRepository{},
		&fakeProblemRepository{},
		&fakeContestRegisterRepository{},
	)
	ctx := context.Background()

	if _, err := svc.CreateContestProblem(ctx, nil); err == nil {
		t.Error("CreateContestProblem(nil) error = nil, want non-nil")
	}
	if _, err := svc.BulkCreateContestProblems(ctx, nil); err == nil {
		t.Error("BulkCreateContestProblems(nil) error = nil, want non-nil")
	}
	if _, err := svc.GetContestProblem(ctx, nil); err == nil {
		t.Error("GetContestProblem(nil) error = nil, want non-nil")
	}
	if _, err := svc.GetContestProblems(ctx, nil); err == nil {
		t.Error("GetContestProblems(nil) error = nil, want non-nil")
	}
	if _, err := svc.UpdateContestProblem(ctx, nil); err == nil {
		t.Error("UpdateContestProblem(nil) error = nil, want non-nil")
	}
	if _, err := svc.DeleteContestProblem(ctx, nil); err == nil {
		t.Error("DeleteContestProblem(nil) error = nil, want non-nil")
	}
}
