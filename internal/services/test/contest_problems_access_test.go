package services_test

import (
	"algoforces/internal/domain"
	"algoforces/internal/services"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/lib/pq"
)

func participantTestContest(isActive bool) *domain.Contest {
	return &domain.Contest{
		Id:             "contest-id",
		Name:           "Weekly Contest",
		StartTime:      time.Date(2026, 10, 27, 9, 0, 0, 0, time.UTC),
		EndTime:        time.Date(2026, 10, 27, 11, 0, 0, 0, time.UTC),
		IsActive:       isActive,
		CreatedBy:      "creator-id",
		ProblemSetters: pq.StringArray{"setter-id"},
	}
}

func registrationWithStatus(status string) *domain.ContestRegistration {
	return &domain.ContestRegistration{
		ID:        "registration-id",
		UserID:    "user-id",
		ContestID: "contest-id",
		Status:    status,
	}
}

// TestCanViewContestProblemsForParticipant is the gate matrix for the
// participant-facing contest-problems view. It exercises only the pure
// access helper with plain structs: no database, no mocks, no HTTP.
func TestCanViewContestProblemsForParticipant(t *testing.T) {
	beforeStart := time.Date(2026, 10, 27, 8, 0, 0, 0, time.UTC)
	atStart := time.Date(2026, 10, 27, 9, 0, 0, 0, time.UTC)
	live := time.Date(2026, 10, 27, 10, 0, 0, 0, time.UTC)
	atEnd := time.Date(2026, 10, 27, 11, 0, 0, 0, time.UTC)
	afterEnd := time.Date(2026, 10, 27, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		contest      *domain.Contest
		userID       string
		role         string
		registration *domain.ContestRegistration
		now          time.Time
		want         bool
	}{
		{
			name:         "admin may view before start",
			contest:      participantTestContest(true),
			userID:       "admin-id",
			role:         "admin",
			registration: nil,
			now:          beforeStart,
			want:         true,
		},
		{
			name:         "admin may view after end even when contest is inactive",
			contest:      participantTestContest(false),
			userID:       "admin-id",
			role:         "admin",
			registration: nil,
			now:          afterEnd,
			want:         true,
		},
		{
			name:         "problem-setter role may view before start",
			contest:      participantTestContest(true),
			userID:       "setter-id",
			role:         "problem-setter",
			registration: nil,
			now:          beforeStart,
			want:         true,
		},
		{
			name:         "problem-setter role may view after end even when contest is inactive",
			contest:      participantTestContest(false),
			userID:       "other-setter-id",
			role:         "problem-setter",
			registration: nil,
			now:          afterEnd,
			want:         true,
		},
		{
			name:         "contest creator may view before start",
			contest:      participantTestContest(true),
			userID:       "creator-id",
			role:         "user",
			registration: nil,
			now:          beforeStart,
			want:         true,
		},
		{
			name:         "contest creator may view after end even when contest is inactive",
			contest:      participantTestContest(false),
			userID:       "creator-id",
			role:         "user",
			registration: nil,
			now:          afterEnd,
			want:         true,
		},
		{
			name:         "listed problem-setter may view before start",
			contest:      participantTestContest(true),
			userID:       "setter-id",
			role:         "user",
			registration: nil,
			now:          beforeStart,
			want:         true,
		},
		{
			name:         "listed problem-setter may view after end",
			contest:      participantTestContest(true),
			userID:       "setter-id",
			role:         "user",
			registration: nil,
			now:          afterEnd,
			want:         true,
		},
		{
			name:         "registered participant may view during live window",
			contest:      participantTestContest(true),
			userID:       "user-id",
			role:         "user",
			registration: registrationWithStatus("registered"),
			now:          live,
			want:         true,
		},
		{
			name:         "registered participant may view at exactly start time",
			contest:      participantTestContest(true),
			userID:       "user-id",
			role:         "user",
			registration: registrationWithStatus("registered"),
			now:          atStart,
			want:         true,
		},
		{
			name:         "registered participant may view at exactly end time",
			contest:      participantTestContest(true),
			userID:       "user-id",
			role:         "user",
			registration: registrationWithStatus("registered"),
			now:          atEnd,
			want:         true,
		},
		{
			name:         "registered participant may not view before start",
			contest:      participantTestContest(true),
			userID:       "user-id",
			role:         "user",
			registration: registrationWithStatus("registered"),
			now:          beforeStart,
			want:         false,
		},
		{
			name:         "registered participant may not view after end",
			contest:      participantTestContest(true),
			userID:       "user-id",
			role:         "user",
			registration: registrationWithStatus("registered"),
			now:          afterEnd,
			want:         false,
		},
		{
			name:         "registered participant may not view when contest is inactive",
			contest:      participantTestContest(false),
			userID:       "user-id",
			role:         "user",
			registration: registrationWithStatus("registered"),
			now:          live,
			want:         false,
		},
		{
			name:         "unregistered user may not view during live window",
			contest:      participantTestContest(true),
			userID:       "user-id",
			role:         "user",
			registration: nil,
			now:          live,
			want:         false,
		},
		{
			name:         "unregistered-status registration may not view during live window",
			contest:      participantTestContest(true),
			userID:       "user-id",
			role:         "user",
			registration: registrationWithStatus("unregistered"),
			now:          live,
			want:         false,
		},
		{
			name:         "stranger may not view before start",
			contest:      participantTestContest(true),
			userID:       "stranger-id",
			role:         "user",
			registration: nil,
			now:          beforeStart,
			want:         false,
		},
		{
			name:         "nil contest allows no one, not even an admin",
			contest:      nil,
			userID:       "admin-id",
			role:         "admin",
			registration: nil,
			now:          live,
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := services.CanViewContestProblemsForParticipant(tt.contest, tt.userID, tt.role, tt.registration, tt.now)
			if got != tt.want {
				t.Errorf("CanViewContestProblemsForParticipant() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGetContestProblemsResponseContainsNoTestCaseData verifies that the
// participant-facing response type serializes the problem statement and
// carries no test-case input/output fields anywhere in its JSON shape.
func TestGetContestProblemsResponseContainsNoTestCaseData(t *testing.T) {
	response := &domain.GetContestProblemsResponse{
		Problems: []domain.ContestProblemDetail{
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
	}

	payload, err := json.Marshal(response)
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
