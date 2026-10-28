package services_test

import (
	"algoforces/internal/domain"
	"algoforces/internal/services"
	"context"
	"errors"
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
		Status:    "registered",
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

func (f *fakeContestRepo) GetAllContests(ctx context.Context) ([]domain.Contest, error) {
	return nil, nil
}

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
		registration    *domain.ContestRegistration
		registrationErr error
		problems        []domain.ContestProblems
		testCasesErr    error
		wantErr         error
	}{
		{
			name:       "contest not found",
			contestErr: gorm.ErrRecordNotFound,
			wantErr:    domain.ErrContestNotFound,
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
			name: "problem not in contest",
			problems: []domain.ContestProblems{
				{ProblemID: "some-other-problem"},
			},
			wantErr: domain.ErrProblemNotInContest,
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
			if tt.contestErr != nil {
				contest = nil
			}

			registration := gateRegisteredRegistration()
			if tt.registration != nil || tt.registrationErr != nil {
				registration = tt.registration
			}
			problems := []domain.ContestProblems{{ProblemID: "problem-1"}}
			if tt.problems != nil {
				problems = tt.problems
			}

			submissionFake := &fakeSubmissionRepo{testCasesErr: tt.testCasesErr}
			contestSvc := services.NewContestService(
				&fakeContestRepo{contest: contest, err: tt.contestErr},
				&fakeUserRepo{},
				&fakeContestRegisterRepo{registration: registration, err: tt.registrationErr},
				&fakeContestProblemsRepo{problems: problems},
			)
			svc := services.NewSubmissionService(submissionFake, contestSvc, "", "")

			_, err := svc.CreateNewSubmission(context.Background(), &domain.CreateSubmissionRequest{
				UserID:    "user-1",
				ContestID: "contest-1",
				ProblemID: "problem-1",
				Code:      "print(1)",
				Language:  "python",
			})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("CreateNewSubmission() error = %v, want %v", err, tt.wantErr)
			}
			if submissionFake.createCalled {
				t.Fatal("CreateNewSubmission() created a submission row despite the gate outcome")
			}
		})
	}
}
