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

// TestCheckSubmissionGate is the PR2 gate matrix. It exercises only the
// pure helper: no DB, no Judge0. The live window in every case is
// [baseNow-1h, baseNow+1h], so the boundary cases use baseNow-1h
// (exactly StartTime) and baseNow+1h (exactly EndTime).
func TestCheckSubmissionGate(t *testing.T) {
	baseNow := time.Date(2026, 10, 27, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name             string
		nilContest       bool
		mutate           func(*domain.Contest)
		registration     *domain.ContestRegistration
		problemInContest bool
		userID           string
		now              time.Time
		wantErr          error
	}{
		{
			name:             "registered, live, visible, problem in contest passes",
			registration:     gateRegisteredRegistration(),
			problemInContest: true,
			userID:           "user-1",
			wantErr:          nil,
		},
		{
			name:             "no registration row is rejected",
			registration:     nil,
			problemInContest: true,
			userID:           "user-1",
			wantErr:          domain.ErrNotRegistered,
		},
		{
			name: "registration with unregistered status is rejected",
			registration: &domain.ContestRegistration{
				ID:        "reg-1",
				UserID:    "user-1",
				ContestID: "contest-1",
				Status:    "unregistered",
			},
			problemInContest: true,
			userID:           "user-1",
			wantErr:          domain.ErrNotRegistered,
		},
		{
			name:             "before start is not live",
			registration:     gateRegisteredRegistration(),
			problemInContest: true,
			userID:           "user-1",
			now:              baseNow.Add(-2 * time.Hour),
			wantErr:          domain.ErrContestNotLive,
		},
		{
			name:             "exactly at start time is live",
			registration:     gateRegisteredRegistration(),
			problemInContest: true,
			userID:           "user-1",
			now:              baseNow.Add(-time.Hour),
			wantErr:          nil,
		},
		{
			name:             "after end is not live",
			registration:     gateRegisteredRegistration(),
			problemInContest: true,
			userID:           "user-1",
			now:              baseNow.Add(2 * time.Hour),
			wantErr:          domain.ErrContestNotLive,
		},
		{
			name:             "exactly at end time is live",
			registration:     gateRegisteredRegistration(),
			problemInContest: true,
			userID:           "user-1",
			now:              baseNow.Add(time.Hour),
			wantErr:          nil,
		},
		{
			name: "inactive contest in its time window is not live",
			mutate: func(c *domain.Contest) {
				c.IsActive = false
			},
			registration:     gateRegisteredRegistration(),
			problemInContest: true,
			userID:           "user-1",
			wantErr:          domain.ErrContestNotLive,
		},
		{
			name: "invisible contest is rejected",
			mutate: func(c *domain.Contest) {
				c.Visible = false
			},
			registration:     gateRegisteredRegistration(),
			problemInContest: true,
			userID:           "user-1",
			wantErr:          domain.ErrContestNotVisible,
		},
		{
			name:             "problem not in contest is rejected",
			registration:     gateRegisteredRegistration(),
			problemInContest: false,
			userID:           "user-1",
			wantErr:          domain.ErrProblemNotInContest,
		},
		{
			name:       "nil contest is not found",
			nilContest: true,
			userID:     "user-1",
			wantErr:    domain.ErrContestNotFound,
		},
		{
			name:             "creator bypasses registration only",
			registration:     nil,
			problemInContest: true,
			userID:           "creator-1",
			wantErr:          nil,
		},
		{
			name:             "problem setter bypasses registration only",
			registration:     nil,
			problemInContest: true,
			userID:           "setter-1",
			wantErr:          nil,
		},
		{
			name: "creator is still rejected for an invisible contest",
			mutate: func(c *domain.Contest) {
				c.Visible = false
			},
			registration:     nil,
			problemInContest: true,
			userID:           "creator-1",
			wantErr:          domain.ErrContestNotVisible,
		},
		{
			name:             "creator is still rejected before start",
			registration:     nil,
			problemInContest: true,
			userID:           "creator-1",
			now:              baseNow.Add(-2 * time.Hour),
			wantErr:          domain.ErrContestNotLive,
		},
		{
			name:             "creator is still rejected for a problem not in the contest",
			registration:     nil,
			problemInContest: false,
			userID:           "creator-1",
			wantErr:          domain.ErrProblemNotInContest,
		},
		{
			name: "invisible beats unregistered",
			mutate: func(c *domain.Contest) {
				c.Visible = false
			},
			registration:     nil,
			problemInContest: true,
			userID:           "user-1",
			wantErr:          domain.ErrContestNotVisible,
		},
		{
			name:             "not live beats unregistered",
			registration:     nil,
			problemInContest: true,
			userID:           "user-1",
			now:              baseNow.Add(-2 * time.Hour),
			wantErr:          domain.ErrContestNotLive,
		},
		{
			name:             "unregistered beats problem not in contest",
			registration:     nil,
			problemInContest: false,
			userID:           "user-1",
			wantErr:          domain.ErrNotRegistered,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := baseNow
			if !tt.now.IsZero() {
				now = tt.now
			}
			var contest *domain.Contest
			if !tt.nilContest {
				contest = gateBaseContest(baseNow)
				if tt.mutate != nil {
					tt.mutate(contest)
				}
			}
			err := services.CheckSubmissionGate(contest, tt.registration, tt.problemInContest, tt.userID, now)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("CheckSubmissionGate() error = %v, want nil", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("CheckSubmissionGate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

// --- Fakes for the service-level gate test ---------------------------------
// These implement the repository interfaces with canned rows so the test
// proves CreateNewSubmission rejects at the gate, before the submission
// row is created and before any Judge0 call, without a DB.

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

// TestCreateNewSubmissionGateRejectsBeforeRowCreation proves the gate
// runs first in CreateNewSubmission: every rejection returns the right
// sentinel and the fake submission repository never sees a row created.
// The gate-pass case uses a stub downstream error to prove execution
// continued past the gate (a real pass would call Judge0, which this
// test must not do).
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
			svc := services.NewSubmissionService(
				submissionFake,
				&fakeContestRepo{contest: contest, err: tt.contestErr},
				&fakeContestRegisterRepo{registration: registration, err: tt.registrationErr},
				&fakeContestProblemsRepo{problems: problems},
				"", "",
			)

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
