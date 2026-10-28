package domain_test

import (
	"algoforces/internal/domain"
	"testing"
	"time"
)

func TestContestStatus(t *testing.T) {
	// Fixed anchor: now is noon; the window runs 10:00-14:00 the same day.
	start := time.Date(2026, 10, 27, 10, 0, 0, 0, time.UTC)
	end := time.Date(2026, 10, 27, 14, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		contest  domain.Contest
		now      time.Time
		expected string
	}{
		{
			name:     "upcoming before start, inactive",
			contest:  domain.Contest{StartTime: start, EndTime: end, IsActive: false},
			now:      time.Date(2026, 10, 27, 9, 0, 0, 0, time.UTC),
			expected: "upcoming",
		},
		{
			name:     "upcoming before start, activated early",
			contest:  domain.Contest{StartTime: start, EndTime: end, IsActive: true},
			now:      time.Date(2026, 10, 27, 9, 59, 59, 0, time.UTC),
			expected: "upcoming",
		},
		{
			name:     "live mid-window, active",
			contest:  domain.Contest{StartTime: start, EndTime: end, IsActive: true},
			now:      time.Date(2026, 10, 27, 12, 0, 0, 0, time.UTC),
			expected: "live",
		},
		{
			name:     "live at start boundary, active",
			contest:  domain.Contest{StartTime: start, EndTime: end, IsActive: true},
			now:      start,
			expected: "live",
		},
		{
			name:     "live at end boundary, active",
			contest:  domain.Contest{StartTime: start, EndTime: end, IsActive: true},
			now:      end,
			expected: "live",
		},
		{
			name:     "ended after end, active",
			contest:  domain.Contest{StartTime: start, EndTime: end, IsActive: true},
			now:      time.Date(2026, 10, 27, 14, 0, 1, 0, time.UTC),
			expected: "ended",
		},
		{
			name:     "ended after end, inactive",
			contest:  domain.Contest{StartTime: start, EndTime: end, IsActive: false},
			now:      time.Date(2026, 10, 27, 15, 0, 0, 0, time.UTC),
			expected: "ended",
		},
		{
			name:     "ended started but inactive mid-window",
			contest:  domain.Contest{StartTime: start, EndTime: end, IsActive: false},
			now:      time.Date(2026, 10, 27, 12, 0, 0, 0, time.UTC),
			expected: "ended",
		},
		{
			name:     "ended at start boundary, never started",
			contest:  domain.Contest{StartTime: start, EndTime: end, IsActive: false},
			now:      start,
			expected: "ended",
		},
		{
			name:     "ended at end boundary, inactive",
			contest:  domain.Contest{StartTime: start, EndTime: end, IsActive: false},
			now:      end,
			expected: "ended",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := domain.ContestStatus(tt.contest, tt.now); got != tt.expected {
				t.Errorf("ContestStatus() = %q, want %q", got, tt.expected)
			}
		})
	}
}
