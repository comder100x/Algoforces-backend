package domain_test

import (
	"algoforces/internal/domain"
	"encoding/json"
	"testing"
	"time"
)

func TestDeriveContestStatus(t *testing.T) {
	// Fixed anchor: now is noon; the window runs 10:00-14:00 the same day.
	start := time.Date(2026, 10, 27, 10, 0, 0, 0, time.UTC)
	end := time.Date(2026, 10, 27, 14, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		contest  domain.Contest
		now      time.Time
		expected domain.ContestStatus
	}{
		{
			name:     "upcoming before start, inactive",
			contest:  domain.Contest{StartTime: start, EndTime: end, IsActive: false},
			now:      time.Date(2026, 10, 27, 9, 0, 0, 0, time.UTC),
			expected: domain.ContestStatusUpcoming,
		},
		{
			name:     "upcoming before start, activated early",
			contest:  domain.Contest{StartTime: start, EndTime: end, IsActive: true},
			now:      time.Date(2026, 10, 27, 9, 59, 59, 0, time.UTC),
			expected: domain.ContestStatusUpcoming,
		},
		{
			name:     "live mid-window, active",
			contest:  domain.Contest{StartTime: start, EndTime: end, IsActive: true},
			now:      time.Date(2026, 10, 27, 12, 0, 0, 0, time.UTC),
			expected: domain.ContestStatusLive,
		},
		{
			name:     "live at start boundary, active",
			contest:  domain.Contest{StartTime: start, EndTime: end, IsActive: true},
			now:      start,
			expected: domain.ContestStatusLive,
		},
		{
			name:     "live at end boundary, active",
			contest:  domain.Contest{StartTime: start, EndTime: end, IsActive: true},
			now:      end,
			expected: domain.ContestStatusLive,
		},
		{
			name:     "ended after end, active",
			contest:  domain.Contest{StartTime: start, EndTime: end, IsActive: true},
			now:      time.Date(2026, 10, 27, 14, 0, 1, 0, time.UTC),
			expected: domain.ContestStatusEnded,
		},
		{
			name:     "ended after end, inactive",
			contest:  domain.Contest{StartTime: start, EndTime: end, IsActive: false},
			now:      time.Date(2026, 10, 27, 15, 0, 0, 0, time.UTC),
			expected: domain.ContestStatusEnded,
		},
		{
			name:     "ended started but inactive mid-window",
			contest:  domain.Contest{StartTime: start, EndTime: end, IsActive: false},
			now:      time.Date(2026, 10, 27, 12, 0, 0, 0, time.UTC),
			expected: domain.ContestStatusEnded,
		},
		{
			name:     "ended at start boundary, never started",
			contest:  domain.Contest{StartTime: start, EndTime: end, IsActive: false},
			now:      start,
			expected: domain.ContestStatusEnded,
		},
		{
			name:     "ended at end boundary, inactive",
			contest:  domain.Contest{StartTime: start, EndTime: end, IsActive: false},
			now:      end,
			expected: domain.ContestStatusEnded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := domain.DeriveContestStatus(tt.contest, tt.now); got != tt.expected {
				t.Errorf("DeriveContestStatus() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestContestStatusEnumValues(t *testing.T) {
	if string(domain.ContestStatusUpcoming) != "upcoming" || string(domain.ContestStatusLive) != "live" || string(domain.ContestStatusEnded) != "ended" {
		t.Errorf("enum values = %q, %q, %q, want \"upcoming\", \"live\", \"ended\"",
			domain.ContestStatusUpcoming, domain.ContestStatusLive, domain.ContestStatusEnded)
	}

	// The enum is string-based, so the API JSON for a response status is
	// unchanged: it still renders as the plain string.
	resp := domain.CreateContestResponse{Id: "c1", Status: domain.ContestStatusLive}
	b, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if decoded["status"] != "live" {
		t.Errorf("JSON status = %v, want %q", decoded["status"], "live")
	}
}
