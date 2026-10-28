package services

import (
	"algoforces/internal/domain"
	"errors"
	"testing"
	"time"
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

func TestEvaluateSubmissionGate(t *testing.T) {
	baseNow := time.Date(2026, 10, 27, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name               string
		nilContest         bool
		mutate             func(*domain.Contest)
		registration       *domain.ContestRegistration
		isProblemInContest bool
		userID             string
		now                time.Time
		wantErr            error
	}{
		{
			name:               "registered, live, visible, problem in contest passes",
			registration:       gateRegisteredRegistration(),
			isProblemInContest: true,
			userID:             "user-1",
			wantErr:            nil,
		},
		{
			name:               "no registration row is rejected",
			registration:       nil,
			isProblemInContest: true,
			userID:             "user-1",
			wantErr:            domain.ErrNotRegistered,
		},
		{
			name: "registration with unregistered status is rejected",
			registration: &domain.ContestRegistration{
				ID:        "reg-1",
				UserID:    "user-1",
				ContestID: "contest-1",
				Status:    domain.RegistrationStatusUnregistered,
			},
			isProblemInContest: true,
			userID:             "user-1",
			wantErr:            domain.ErrNotRegistered,
		},
		{
			name:               "before start is not live",
			registration:       gateRegisteredRegistration(),
			isProblemInContest: true,
			userID:             "user-1",
			now:                baseNow.Add(-2 * time.Hour),
			wantErr:            domain.ErrContestNotLive,
		},
		{
			name:               "exactly at start time is live",
			registration:       gateRegisteredRegistration(),
			isProblemInContest: true,
			userID:             "user-1",
			now:                baseNow.Add(-time.Hour),
			wantErr:            nil,
		},
		{
			name:               "after end is not live",
			registration:       gateRegisteredRegistration(),
			isProblemInContest: true,
			userID:             "user-1",
			now:                baseNow.Add(2 * time.Hour),
			wantErr:            domain.ErrContestNotLive,
		},
		{
			name:               "exactly at end time is live",
			registration:       gateRegisteredRegistration(),
			isProblemInContest: true,
			userID:             "user-1",
			now:                baseNow.Add(time.Hour),
			wantErr:            nil,
		},
		{
			name: "inactive contest in its time window is not live",
			mutate: func(c *domain.Contest) {
				c.IsActive = false
			},
			registration:       gateRegisteredRegistration(),
			isProblemInContest: true,
			userID:             "user-1",
			wantErr:            domain.ErrContestNotLive,
		},
		{
			name: "invisible contest is rejected",
			mutate: func(c *domain.Contest) {
				c.Visible = false
			},
			registration:       gateRegisteredRegistration(),
			isProblemInContest: true,
			userID:             "user-1",
			wantErr:            domain.ErrContestNotVisible,
		},
		{
			name:               "problem not in contest is rejected",
			registration:       gateRegisteredRegistration(),
			isProblemInContest: false,
			userID:             "user-1",
			wantErr:            domain.ErrProblemNotInContest,
		},
		{
			name:       "nil contest is not found",
			nilContest: true,
			userID:     "user-1",
			wantErr:    domain.ErrContestNotFound,
		},
		{
			name:               "creator bypasses registration only",
			registration:       nil,
			isProblemInContest: true,
			userID:             "creator-1",
			wantErr:            nil,
		},
		{
			name:               "problem setter bypasses registration only",
			registration:       nil,
			isProblemInContest: true,
			userID:             "setter-1",
			wantErr:            nil,
		},
		{
			name: "creator is still rejected for an invisible contest",
			mutate: func(c *domain.Contest) {
				c.Visible = false
			},
			registration:       nil,
			isProblemInContest: true,
			userID:             "creator-1",
			wantErr:            domain.ErrContestNotVisible,
		},
		{
			name:               "creator is still rejected before start",
			registration:       nil,
			isProblemInContest: true,
			userID:             "creator-1",
			now:                baseNow.Add(-2 * time.Hour),
			wantErr:            domain.ErrContestNotLive,
		},
		{
			name:               "creator is still rejected for a problem not in the contest",
			registration:       nil,
			isProblemInContest: false,
			userID:             "creator-1",
			wantErr:            domain.ErrProblemNotInContest,
		},
		{
			name: "invisible beats unregistered",
			mutate: func(c *domain.Contest) {
				c.Visible = false
			},
			registration:       nil,
			isProblemInContest: true,
			userID:             "user-1",
			wantErr:            domain.ErrContestNotVisible,
		},
		{
			name:               "not live beats unregistered",
			registration:       nil,
			isProblemInContest: true,
			userID:             "user-1",
			now:                baseNow.Add(-2 * time.Hour),
			wantErr:            domain.ErrContestNotLive,
		},
		{
			name:               "unregistered beats problem not in contest",
			registration:       nil,
			isProblemInContest: false,
			userID:             "user-1",
			wantErr:            domain.ErrNotRegistered,
		},
		{
			name: "creator bypasses an unregistered-status row",
			registration: &domain.ContestRegistration{
				ID:        "reg-1",
				UserID:    "creator-1",
				ContestID: "contest-1",
				Status:    domain.RegistrationStatusUnregistered,
			},
			isProblemInContest: true,
			userID:             "creator-1",
			wantErr:            nil,
		},
		{
			name: "problem setter bypasses an unregistered-status row",
			registration: &domain.ContestRegistration{
				ID:        "reg-1",
				UserID:    "setter-1",
				ContestID: "contest-1",
				Status:    domain.RegistrationStatusUnregistered,
			},
			isProblemInContest: true,
			userID:             "setter-1",
			wantErr:            nil,
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
			err := evaluateSubmissionGate(contest, tt.registration, tt.isProblemInContest, tt.userID, now)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("evaluateSubmissionGate() error = %v, want nil", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("evaluateSubmissionGate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
