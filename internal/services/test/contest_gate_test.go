package services_test

import (
	"algoforces/internal/domain"
	"algoforces/internal/services"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"
)

func gateBaseContest(now time.Time) *domain.Contest {
	return &domain.Contest{
		Id:             "contest-1",
		StartTime:      now.Add(-time.Hour),
		EndTime:        now.Add(time.Hour),
		Visible:        true,
		IsActive:       true,
		CreatedBy:      "creator-1",
		ProblemSetters: []string{"setter-1"},
	}
}

func gateRegisteredRegistration() *domain.ContestRegistration {
	return &domain.ContestRegistration{
		ID:        "reg-1",
		UserID:    "user-1",
		ContestID: "contest-1",
		Status:    domain.RegistrationStatusRegistered,
	}
}

type fakeSubmissionRepo struct {
	createCalled bool
	testCasesErr error
}

func (f *fakeSubmissionRepo) GetAllTestCasesForProblem(ctx context.Context, problemID string) ([]domain.TestCase, error) {
	return nil, f.testCasesErr
}

func (f *fakeSubmissionRepo) GetTestCaseByID(ctx context.Context, testCaseID string) (*domain.TestCase, error) {
	return nil, nil
}

func (f *fakeSubmissionRepo) GetProblemByID(ctx context.Context, problemID string) (*domain.Problem, error) {
	return nil, nil
}

func (f *fakeSubmissionRepo) GetContestProblemMaxPoints(ctx context.Context, contestID, problemID string) (int, error) {
	return 0, nil
}

func (f *fakeSubmissionRepo) CreateNewSubmission(ctx context.Context, submission *domain.Submission) error {
	f.createCalled = true
	return nil
}

func (f *fakeSubmissionRepo) GetSubmissionDetails(ctx context.Context, uniqueID string) (*domain.Submission, error) {
	return nil, nil
}

func (f *fakeSubmissionRepo) UpdateSubmissionStatus(ctx context.Context, submissionID string, status string) error {
	return nil
}

func (f *fakeSubmissionRepo) UpdateSubmissionResult(ctx context.Context, submissionID string, result *domain.Submission) error {
	return nil
}

func (f *fakeSubmissionRepo) CreateTokenMapping(ctx context.Context, mapping *domain.SubmissionTestCaseMapping) error {
	return nil
}

func (f *fakeSubmissionRepo) GetMappingByToken(ctx context.Context, token string) (*domain.SubmissionTestCaseMapping, error) {
	return nil, nil
}

func (f *fakeSubmissionRepo) UpdateMappingStatus(ctx context.Context, token string, status string) error {
	return nil
}

type fakeContestRepo struct {
	contest *domain.Contest
	err     error
}

func (f *fakeContestRepo) CreateContest(ctx context.Context, contest *domain.Contest) error {
	return nil
}

func (f *fakeContestRepo) GetByID(ctx context.Context, id string) (*domain.Contest, error) {
	return f.contest, f.err
}

func (f *fakeContestRepo) UpdateContest(ctx context.Context, contest *domain.Contest) error {
	return nil
}

func (f *fakeContestRepo) DeleteContest(ctx context.Context, id string) error {
	return nil
}

func (f *fakeContestRepo) CheckContestInTimeWindow(ctx context.Context, startTime, endTime time.Time) ([]domain.Contest, error) {
	return nil, nil
}

// Merged-interface methods from the contest-lifecycle change; unused on this branch.
func (f *fakeContestRepo) CheckContestInTimeWindowExcluding(ctx context.Context, startTime, endTime time.Time, excludeID string) ([]domain.Contest, error) {
	return f.CheckContestInTimeWindow(ctx, startTime, endTime)
}

func (f *fakeContestRepo) GetVisibleContests(ctx context.Context) ([]domain.Contest, error) {
	return nil, nil
}

func (f *fakeContestRepo) GetAllContests(ctx context.Context) ([]domain.Contest, error) {
	return nil, nil
}

type mergedContestRepository interface {
	domain.ContestRepository
	CheckContestInTimeWindowExcluding(ctx context.Context, startTime, endTime time.Time, excludeID string) ([]domain.Contest, error)
	GetVisibleContests(ctx context.Context) ([]domain.Contest, error)
}

var _ mergedContestRepository = (*fakeContestRepo)(nil)

type fakeUserRepo struct{}

func (f *fakeUserRepo) Create(ctx context.Context, user *domain.User) error {
	return nil
}

func (f *fakeUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, nil
}

func (f *fakeUserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return nil, nil
}

func (f *fakeUserRepo) UpdateByID(ctx context.Context, id string, user *domain.User) error {
	return nil
}

type fakeContestRegisterRepo struct {
	registration *domain.ContestRegistration
	err          error
}

func (f *fakeContestRegisterRepo) CreateContestRegistration(ctx context.Context, userID string, contestID string) (*domain.ContestRegistration, error) {
	return nil, nil
}

func (f *fakeContestRegisterRepo) GetRegistrationByUserAndContest(ctx context.Context, userID string, contestID string) (*domain.ContestRegistration, error) {
	return f.registration, f.err
}

func (f *fakeContestRegisterRepo) UpdateRegistrationStatus(ctx context.Context, userID string, contestID string, status string) error {
	return nil
}

func (f *fakeContestRegisterRepo) GetAllRegistrationsByUserID(ctx context.Context, userID string) ([]domain.ContestRegistration, error) {
	return nil, nil
}

func (f *fakeContestRegisterRepo) GetAllRegistrationsForAdmin(ctx context.Context) ([]domain.ContestRegistration, error) {
	return nil, nil
}

type fakeContestProblemsRepo struct {
	problems []domain.ContestProblems
	err      error
}

func (f *fakeContestProblemsRepo) CreateContestProblem(ctx context.Context, contestProblem *domain.ContestProblems) error {
	return nil
}

func (f *fakeContestProblemsRepo) GetContestProblemByID(ctx context.Context, uniqueID string) (*domain.ContestProblems, error) {
	return nil, nil
}

func (f *fakeContestProblemsRepo) GetContestProblemsByContestID(ctx context.Context, contestID string) ([]domain.ContestProblems, error) {
	return f.problems, f.err
}

func (f *fakeContestProblemsRepo) UpdateContestProblem(ctx context.Context, contestProblem *domain.ContestProblems) error {
	return nil
}

func (f *fakeContestProblemsRepo) DeleteContestProblem(ctx context.Context, uniqueID string) error {
	return nil
}

func (f *fakeContestProblemsRepo) GetContestProblemsByContestIDWithDetails(ctx context.Context, contestID string) ([]domain.ContestProblemDetail, error) {
	return nil, nil
}

func TestCreateNewSubmissionGateRejectsBeforeRowCreation(t *testing.T) {
	downstreamErr := errors.New("downstream stub failure")

	tests := []struct {
		name            string
		mutate          func(*domain.Contest)
		contestErr      error
		nilContest      bool
		registration    *domain.ContestRegistration
		registrationErr error
		noRegistration  bool
		problems        []domain.ContestProblems
		problemsErr     error
		testCasesErr    error
		userID          string
		wantErr         error
		wantErrContains string
	}{
		{
			name:       "contest not found",
			contestErr: gorm.ErrRecordNotFound,
			wantErr:    domain.ErrContestNotFound,
		},
		{
			name:       "nil contest without error is not found",
			nilContest: true,
			wantErr:    domain.ErrContestNotFound,
		},
		{
			name:            "contest lookup error is wrapped, not a sentinel",
			contestErr:      errors.New("db down"),
			wantErrContains: "failed to get contest: db down",
		},
		{
			name:            "contest not found beats registration error",
			contestErr:      errors.New("db down"),
			registrationErr: errors.New("db down"),
			wantErrContains: "failed to get contest: db down",
		},
		{
			name: "invisible contest",
			mutate: func(c *domain.Contest) {
				c.Visible = false
			},
			wantErr: domain.ErrContestNotVisible,
		},
		{
			name: "contest not live",
			mutate: func(c *domain.Contest) {
				c.IsActive = false
			},
			wantErr: domain.ErrContestNotLive,
		},
		{
			name:            "unregistered user",
			registrationErr: gorm.ErrRecordNotFound,
			wantErr:         domain.ErrNotRegistered,
		},
		{
			name:           "nil registration without error is not registered",
			noRegistration: true,
			wantErr:        domain.ErrNotRegistered,
		},
		{
			name:            "registration lookup error is wrapped, not a sentinel",
			registrationErr: errors.New("db down"),
			wantErrContains: "failed to get contest registration: db down",
		},
		{
			name: "not live beats unregistered",
			mutate: func(c *domain.Contest) {
				c.IsActive = false
			},
			registrationErr: gorm.ErrRecordNotFound,
			wantErr:         domain.ErrContestNotLive,
		},
		{
			name:            "contest problems lookup error is wrapped, not a sentinel",
			problemsErr:     errors.New("db down"),
			wantErrContains: "failed to get contest problems: db down",
		},
		{
			name: "problem not in contest",
			problems: []domain.ContestProblems{
				{ProblemID: "some-other-problem"},
			},
			wantErr: domain.ErrProblemNotInContest,
		},
		{
			name:   "creator bypasses an unregistered-status row",
			userID: "creator-1",
			registration: &domain.ContestRegistration{
				ID:        "reg-1",
				UserID:    "creator-1",
				ContestID: "contest-1",
				Status:    domain.RegistrationStatusUnregistered,
			},
			testCasesErr: downstreamErr,
			wantErr:      downstreamErr,
		},
		{
			name:   "problem setter bypasses an unregistered-status row",
			userID: "setter-1",
			registration: &domain.ContestRegistration{
				ID:        "reg-1",
				UserID:    "setter-1",
				ContestID: "contest-1",
				Status:    domain.RegistrationStatusUnregistered,
			},
			testCasesErr: downstreamErr,
			wantErr:      downstreamErr,
		},
		{
			name:   "creator is still rejected for an inactive in-window contest",
			userID: "creator-1",
			mutate: func(c *domain.Contest) {
				c.IsActive = false
			},
			wantErr: domain.ErrContestNotLive,
		},
		{
			name:         "gate passes and execution continues past the gate",
			testCasesErr: downstreamErr,
			wantErr:      downstreamErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Now()
			contest := gateBaseContest(now)
			if tt.mutate != nil {
				tt.mutate(contest)
			}
			if tt.contestErr != nil || tt.nilContest {
				contest = nil
			}

			registration := gateRegisteredRegistration()
			if tt.registration != nil || tt.registrationErr != nil {
				registration = tt.registration
			}
			if tt.noRegistration {
				registration = nil
			}
			problems := []domain.ContestProblems{{ProblemID: "problem-1"}}
			if tt.problems != nil {
				problems = tt.problems
			}
			userID := tt.userID
			if userID == "" {
				userID = "user-1"
			}

			submissionFake := &fakeSubmissionRepo{testCasesErr: tt.testCasesErr}
			contestSvc := services.NewContestService(
				&fakeContestRepo{contest: contest, err: tt.contestErr},
				&fakeUserRepo{},
				&fakeContestRegisterRepo{registration: registration, err: tt.registrationErr},
				&fakeContestProblemsRepo{problems: problems, err: tt.problemsErr},
			)
			svc := services.NewSubmissionService(submissionFake, contestSvc, "", "")

			_, err := svc.CreateNewSubmission(context.Background(), &domain.CreateSubmissionRequest{
				UserID:    userID,
				ContestID: "contest-1",
				ProblemID: "problem-1",
				Code:      "print(1)",
				Language:  "python",
			})
			if tt.wantErrContains != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErrContains) {
					t.Fatalf("CreateNewSubmission() error = %v, want error containing %q", err, tt.wantErrContains)
				}
			} else if !errors.Is(err, tt.wantErr) {
				t.Fatalf("CreateNewSubmission() error = %v, want %v", err, tt.wantErr)
			}
			if submissionFake.createCalled {
				t.Fatal("CreateNewSubmission() created a submission row despite the gate outcome")
			}
		})
	}
}

func TestCheckSubmissionGateNilRepositories(t *testing.T) {
	tests := []struct {
		name                string
		contestRepo         domain.ContestRepository
		contestRegisterRepo domain.ContestRegisterRepository
		contestProblemsRepo domain.ContestProblemsRepository
	}{
		{
			name:                "nil contest repository",
			contestRegisterRepo: &fakeContestRegisterRepo{},
			contestProblemsRepo: &fakeContestProblemsRepo{},
		},
		{
			name:                "nil registration repository",
			contestRepo:         &fakeContestRepo{},
			contestProblemsRepo: &fakeContestProblemsRepo{},
		},
		{
			name:                "nil contest problems repository",
			contestRepo:         &fakeContestRepo{},
			contestRegisterRepo: &fakeContestRegisterRepo{},
		},
		{
			name: "all gate repositories nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contestSvc := services.NewContestService(
				tt.contestRepo,
				&fakeUserRepo{},
				tt.contestRegisterRepo,
				tt.contestProblemsRepo,
			)
			err := contestSvc.CheckSubmissionGate(context.Background(), "user-1", "contest-1", "problem-1")
			if err == nil || !strings.Contains(err.Error(), "contest gate repositories are not configured") {
				t.Fatalf("CheckSubmissionGate() error = %v, want repositories-not-configured error", err)
			}
		})
	}
}

func TestCreateNewSubmissionNilGuards(t *testing.T) {
	svc := services.NewSubmissionService(&fakeSubmissionRepo{}, nil, "", "")
	if _, err := svc.CreateNewSubmission(context.Background(), &domain.CreateSubmissionRequest{}); err == nil || !strings.Contains(err.Error(), "contest use case is not configured") {
		t.Fatalf("CreateNewSubmission() error = %v, want contest-use-case-not-configured error", err)
	}

	contestSvc := services.NewContestService(
		&fakeContestRepo{contest: gateBaseContest(time.Now())},
		&fakeUserRepo{},
		&fakeContestRegisterRepo{registration: gateRegisteredRegistration()},
		&fakeContestProblemsRepo{problems: []domain.ContestProblems{{ProblemID: "problem-1"}}},
	)
	svc = services.NewSubmissionService(&fakeSubmissionRepo{}, contestSvc, "", "")
	if _, err := svc.CreateNewSubmission(context.Background(), nil); err == nil || !strings.Contains(err.Error(), "create submission request is nil") {
		t.Fatalf("CreateNewSubmission(nil) error = %v, want nil-request error", err)
	}
}
