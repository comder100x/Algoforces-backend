package services_test

import (
	"algoforces/internal/domain"
	"algoforces/internal/services"
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

func participantTestContestWithVisible(isActive, visible bool) *domain.Contest {
	contest := participantTestContest(isActive)
	contest.Visible = visible
	return contest
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
			// Pinned, not changed: Visible is not consulted by this helper.
			// Whether Visible=false must hide problems from registered
			// participants is a held decision; PR #2's gate differs here.
			name:         "registered participant may view during live window even when contest is not visible",
			contest:      participantTestContestWithVisible(true, false),
			userID:       "user-id",
			role:         "user",
			registration: registrationWithStatus("registered"),
			now:          live,
			want:         true,
		},
		{
			name:         "registered participant may view during live window when contest is visible",
			contest:      participantTestContestWithVisible(true, true),
			userID:       "user-id",
			role:         "user",
			registration: registrationWithStatus("registered"),
			now:          live,
			want:         true,
		},
		{
			// Pinned: an empty user ID is not itself a denial; the role and
			// registration checks decide.
			name:         "empty user id with admin role may view before start",
			contest:      participantTestContest(true),
			userID:       "",
			role:         "admin",
			registration: nil,
			now:          beforeStart,
			want:         true,
		},
		{
			name:         "empty user id with registered registration may view during live window",
			contest:      participantTestContest(true),
			userID:       "",
			role:         "user",
			registration: registrationWithStatus("registered"),
			now:          live,
			want:         true,
		},
		{
			name:         "empty user id without registration may not view during live window",
			contest:      participantTestContest(true),
			userID:       "",
			role:         "user",
			registration: nil,
			now:          live,
			want:         false,
		},
		{
			// Pinned: role comparison is case-sensitive.
			name:         "capitalised Admin role may not view without registration",
			contest:      participantTestContest(true),
			userID:       "admin-id",
			role:         "Admin",
			registration: nil,
			now:          live,
			want:         false,
		},
		{
			name:         "upper-case ADMIN role may not view without registration",
			contest:      participantTestContest(true),
			userID:       "admin-id",
			role:         "ADMIN",
			registration: nil,
			now:          live,
			want:         false,
		},
		{
			name:         "capitalised Problem-Setter role may not view without registration",
			contest:      participantTestContest(true),
			userID:       "other-id",
			role:         "Problem-Setter",
			registration: nil,
			now:          live,
			want:         false,
		},
		{
			// Pinned: the listed-setter path is ID-based, so a listed
			// setter keeps access even with a differently-cased role.
			name:         "listed setter keeps access regardless of role casing",
			contest:      participantTestContest(true),
			userID:       "setter-id",
			role:         "Problem-Setter",
			registration: nil,
			now:          live,
			want:         true,
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
