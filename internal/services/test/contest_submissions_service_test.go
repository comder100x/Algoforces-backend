package services_test

import (
	"algoforces/internal/domain"
	"algoforces/internal/services"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

func TestCanViewHiddenTestData(t *testing.T) {
	tests := []struct {
		name string
		role string
		want bool
	}{
		{name: "admin may view hidden test data", role: "admin", want: true},
		{name: "problem-setter may view hidden test data", role: "problem-setter", want: true},
		{name: "regular user may not view hidden test data", role: "user", want: false},
		{name: "empty role may not view hidden test data", role: "", want: false},
		{name: "unknown role may not view hidden test data", role: "superuser", want: false},
		{name: "role matching is case-sensitive", role: "Admin", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := services.CanViewHiddenTestData(tt.role); got != tt.want {
				t.Errorf("CanViewHiddenTestData(%q) = %v, want %v", tt.role, got, tt.want)
			}
		})
	}
}

type fakeContestRepository struct {
	domain.ContestRepository
	err error
}

func (f *fakeContestRepository) GetByID(
	ctx context.Context,
	id string,
) (*domain.Contest, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &domain.Contest{Id: id}, nil
}

// fakeContestSubmissionsRepository mirrors the SQL contract: when
// includeHidden is false it returns only the caller's own submissions,
// preserving the input order. It records the arguments it was called with.
type fakeContestSubmissionsRepository struct {
	contestID     string
	userID        string
	includeHidden bool
	submissions   []domain.Submission
	err           error
}

func (f *fakeContestSubmissionsRepository) GetSubmissionsByContestID(
	ctx context.Context,
	contestID string,
	userID string,
	includeHidden bool,
) ([]domain.Submission, error) {
	f.contestID = contestID
	f.userID = userID
	f.includeHidden = includeHidden
	if f.err != nil {
		return nil, f.err
	}
	if includeHidden {
		return f.submissions, nil
	}
	visible := make([]domain.Submission, 0, len(f.submissions))
	for _, submission := range f.submissions {
		if submission.UserId == userID {
			visible = append(visible, submission)
		}
	}
	return visible, nil
}

type fakeSubmissionRepository struct {
	domain.SubmissionRepository
	details map[string]*domain.Submission
	errIDs  map[string]error
	calls   []string
}

func (f *fakeSubmissionRepository) GetSubmissionDetails(
	ctx context.Context,
	uniqueID string,
) (*domain.Submission, error) {
	f.calls = append(f.calls, uniqueID)
	if err, ok := f.errIDs[uniqueID]; ok {
		return nil, err
	}
	submission, ok := f.details[uniqueID]
	if !ok {
		return nil, errors.New("submission not found")
	}
	return submission, nil
}

func listTestSubmissions() []domain.Submission {
	return []domain.Submission{
		{UniqueID: "sub-newest", UserId: "user-other"},
		{UniqueID: "sub-second", UserId: "user-own"},
		{UniqueID: "sub-third", UserId: "user-other"},
		{UniqueID: "sub-oldest", UserId: "user-own"},
	}
}

func listTestDetails() map[string]*domain.Submission {
	hiddenResult := `{"test_case_id":"tc-hidden","test_number":1,"verdict":"Accepted","is_hidden":true,"test_input":"hidden input","test_expected_output":"hidden output"}`
	visibleResult := `{"test_case_id":"tc-visible","test_number":2,"verdict":"Accepted","is_hidden":false,"test_input":"visible input","test_expected_output":"visible output"}`
	details := make(map[string]*domain.Submission)
	for _, submission := range listTestSubmissions() {
		details[submission.UniqueID] = &domain.Submission{
			UniqueID:        submission.UniqueID,
			UserId:          submission.UserId,
			ContestID:       "contest-1",
			ProblemID:       "problem-1",
			Code:            "print(1)",
			Language:        "python",
			TokenList:       pq.StringArray{"tok-" + submission.UniqueID},
			Verdict:         "Accepted",
			TestCaseResults: pq.StringArray{hiddenResult, visibleResult},
		}
	}
	return details
}

func TestListContestSubmissions(t *testing.T) {
	tests := []struct {
		name              string
		role              string
		userID            string
		submissions       []domain.Submission
		wantIDs           []string
		wantIncludeHidden bool
		wantHiddenResults bool
	}{
		{
			name:              "admin sees all submissions with hidden data",
			role:              "admin",
			userID:            "user-own",
			submissions:       listTestSubmissions(),
			wantIDs:           []string{"sub-newest", "sub-second", "sub-third", "sub-oldest"},
			wantIncludeHidden: true,
			wantHiddenResults: true,
		},
		{
			name:              "problem-setter sees all submissions with hidden data",
			role:              "problem-setter",
			userID:            "user-own",
			submissions:       listTestSubmissions(),
			wantIDs:           []string{"sub-newest", "sub-second", "sub-third", "sub-oldest"},
			wantIncludeHidden: true,
			wantHiddenResults: true,
		},
		{
			name:              "regular user sees only own submissions without hidden data",
			role:              "user",
			userID:            "user-own",
			submissions:       listTestSubmissions(),
			wantIDs:           []string{"sub-second", "sub-oldest"},
			wantIncludeHidden: false,
			wantHiddenResults: false,
		},
		{
			name:              "newest-first order survives filtering for unknown roles",
			role:              "superuser",
			userID:            "user-own",
			submissions:       listTestSubmissions(),
			wantIDs:           []string{"sub-second", "sub-oldest"},
			wantIncludeHidden: false,
			wantHiddenResults: false,
		},
		{
			name:              "regular user with no submissions sees none",
			role:              "user",
			userID:            "user-none",
			submissions:       listTestSubmissions(),
			wantIDs:           []string{},
			wantIncludeHidden: false,
			wantHiddenResults: false,
		},
		{
			name:              "empty contest returns an empty list",
			role:              "admin",
			userID:            "user-own",
			submissions:       []domain.Submission{},
			wantIDs:           []string{},
			wantIncludeHidden: true,
			wantHiddenResults: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subsRepo := &fakeContestSubmissionsRepository{submissions: tt.submissions}
			svc := services.NewSubmissionService(
				&fakeSubmissionRepository{details: listTestDetails()},
				subsRepo,
				&fakeContestRepository{},
				"",
				"",
			)

			got, err := svc.ListContestSubmissions(
				context.Background(),
				"contest-1",
				tt.userID,
				tt.role,
			)
			if err != nil {
				t.Fatalf("ListContestSubmissions() error = %v", err)
			}
			if got == nil {
				t.Fatalf("ListContestSubmissions() returned nil, want empty list")
			}
			if len(got) != len(tt.wantIDs) {
				t.Fatalf("ListContestSubmissions() returned %d results, want %d", len(got), len(tt.wantIDs))
			}
			for i, wantID := range tt.wantIDs {
				if got[i].UniqueID != wantID {
					t.Errorf("result[%d].UniqueID = %q, want %q", i, got[i].UniqueID, wantID)
				}
				if tt.wantHiddenResults {
					if len(got[i].TestCaseResults) != 2 {
						t.Errorf("result[%d] has %d test case results, want 2", i, len(got[i].TestCaseResults))
					}
					if len(got[i].TokenList) != 1 {
						t.Errorf("result[%d] has %d tokens, want 1", i, len(got[i].TokenList))
					}
				} else {
					if len(got[i].TestCaseResults) != 1 {
						t.Errorf("result[%d] has %d test case results, want 1 visible result", i, len(got[i].TestCaseResults))
					}
					if len(got[i].TestCaseResults) == 1 {
						result := got[i].TestCaseResults[0]
						if result.IsHidden {
							t.Errorf("result[%d] exposes a hidden test case result", i)
						}
						if result.TestCaseID != "" || result.TestInput != "" {
							t.Errorf("result[%d] visible result is not sanitised: %+v", i, result)
						}
					}
					if len(got[i].TokenList) != 0 {
						t.Errorf("result[%d] exposes %d tokens, want 0", i, len(got[i].TokenList))
					}
				}
			}
			if subsRepo.contestID != "contest-1" {
				t.Errorf("repository got contestID %q, want %q", subsRepo.contestID, "contest-1")
			}
			if subsRepo.userID != tt.userID {
				t.Errorf("repository got userID %q, want %q", subsRepo.userID, tt.userID)
			}
			if subsRepo.includeHidden != tt.wantIncludeHidden {
				t.Errorf("repository got includeHidden %v, want %v", subsRepo.includeHidden, tt.wantIncludeHidden)
			}
		})
	}
}

func TestListContestSubmissionsErrors(t *testing.T) {
	tests := []struct {
		name              string
		contestErr        error
		subsErr           error
		detailErrIDs      map[string]error
		nilContestRepo    bool
		nilSubsRepo       bool
		nilSubmissionRepo bool
		wantErrContains   string
		wantNotFound      bool
		wantCalls         []string
	}{
		{
			name:            "missing contest returns ErrContestNotFound",
			contestErr:      domain.ErrContestNotFound,
			wantErrContains: "contest-1",
			wantNotFound:    true,
		},
		{
			name:            "record-not-found from the repository becomes ErrContestNotFound",
			contestErr:      gorm.ErrRecordNotFound,
			wantErrContains: "contest-1",
			wantNotFound:    true,
		},
		{
			name:            "contest lookup failure propagates",
			contestErr:      errors.New("db unavailable"),
			wantErrContains: "get contest contest-1",
		},
		{
			name:            "submissions lookup failure propagates",
			subsErr:         errors.New("db unavailable"),
			wantErrContains: "get submissions for contest contest-1",
		},
		{
			name:            "mid-loop detail failure aborts the list",
			detailErrIDs:    map[string]error{"sub-second": errors.New("detail store unavailable")},
			wantErrContains: "sub-second",
			wantCalls:       []string{"sub-newest", "sub-second"},
		},
		{
			name:            "nil contest repository errors",
			nilContestRepo:  true,
			wantErrContains: "not configured",
		},
		{
			name:            "nil submissions repository errors",
			nilSubsRepo:     true,
			wantErrContains: "not configured",
		},
		{
			name:              "nil submission repository errors",
			nilSubmissionRepo: true,
			wantErrContains:   "not configured",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var contestRepo domain.ContestRepository = &fakeContestRepository{err: tt.contestErr}
			if tt.nilContestRepo {
				contestRepo = nil
			}
			var subsRepo domain.ContestSubmissionsRepository = &fakeContestSubmissionsRepository{
				submissions: listTestSubmissions(),
				err:         tt.subsErr,
			}
			if tt.nilSubsRepo {
				subsRepo = nil
			}
			submissionRepo := &fakeSubmissionRepository{
				details: listTestDetails(),
				errIDs:  tt.detailErrIDs,
			}
			var submissionRepoIface domain.SubmissionRepository = submissionRepo
			if tt.nilSubmissionRepo {
				submissionRepoIface = nil
			}
			svc := services.NewSubmissionService(
				submissionRepoIface,
				subsRepo,
				contestRepo,
				"",
				"",
			)

			_, err := svc.ListContestSubmissions(
				context.Background(),
				"contest-1",
				"user-own",
				"admin",
			)
			if err == nil {
				t.Fatalf("ListContestSubmissions() error = nil, want error containing %q", tt.wantErrContains)
			}
			if !strings.Contains(err.Error(), tt.wantErrContains) {
				t.Errorf("error %q does not contain %q", err.Error(), tt.wantErrContains)
			}
			if tt.wantNotFound && !errors.Is(err, domain.ErrContestNotFound) {
				t.Errorf("error %q does not match domain.ErrContestNotFound", err.Error())
			}
			if tt.wantCalls != nil {
				if len(submissionRepo.calls) != len(tt.wantCalls) {
					t.Fatalf("detail lookups = %v, want %v", submissionRepo.calls, tt.wantCalls)
				}
				for i, want := range tt.wantCalls {
					if submissionRepo.calls[i] != want {
						t.Errorf("detail lookup[%d] = %q, want %q", i, submissionRepo.calls[i], want)
					}
				}
			}
		})
	}
}
