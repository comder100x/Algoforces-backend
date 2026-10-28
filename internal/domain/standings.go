package domain

import (
	"context"
	"errors"
	"time"
)

// Contest standings status values, computed from a contest's
// StartTime, EndTime and IsActive fields. See the standings service
// for the exact derivation.
const (
	ContestStatusUpcoming = "upcoming"
	ContestStatusLive     = "live"
	ContestStatusEnded    = "ended"
)

// ErrContestNotFound is returned when the standings are requested for
// a contest that does not exist.
var ErrContestNotFound = errors.New("contest not found")

// ErrStandingsForbidden is returned when a participant requests the
// standings before the contest has started. Admins and problem
// setters may still view them.
var ErrStandingsForbidden = errors.New("standings are not available before the contest starts")

// StandingsProblemScore is one user's best result on one problem.
type StandingsProblemScore struct {
	ProblemID       string     `json:"problem_id"`
	BestScore       int        `json:"best_score"`
	FirstAcceptedAt *time.Time `json:"first_accepted_at"`
	Attempts        int        `json:"attempts"`
}

// StandingsEntry is one user's row on the contest leaderboard.
type StandingsEntry struct {
	UserID         string                  `json:"user_id"`
	Username       string                  `json:"username"`
	TotalScore     int                     `json:"total_score"`
	SolvedCount    int                     `json:"solved_count"`
	LastAcceptedAt *time.Time              `json:"last_accepted_at"`
	ProblemScores  []StandingsProblemScore `json:"problem_scores"`
}

// StandingsResponse is the leaderboard for one contest.
type StandingsResponse struct {
	ContestID   string           `json:"contest_id"`
	Status      string           `json:"status"`
	Entries     []StandingsEntry `json:"entries"`
	GeneratedAt time.Time        `json:"generated_at"`
}

// SubmissionCountKey identifies one user's submissions on one
// problem. It is the key of the per-problem attempt counts returned
// by StandingsRepository.GetSubmissionCountsForContest.
type SubmissionCountKey struct {
	UserID    string
	ProblemID string
}

// StandingsRepository reads the submission data standings are computed
// from. It is separate from SubmissionRepository so its fakes keep compiling.
type StandingsRepository interface {
	// GetAcceptedSubmissionsForContest returns every submission in
	// the contest whose verdict is Accepted.
	GetAcceptedSubmissionsForContest(ctx context.Context, contestID string) ([]Submission, error)
	// GetSubmissionCountsForContest returns, for every user/problem
	// pair with at least one submission in the contest, how many
	// submissions that user made on that problem (all verdicts).
	GetSubmissionCountsForContest(ctx context.Context, contestID string) (map[SubmissionCountKey]int, error)
}

// StandingsUseCase computes contest standings.
type StandingsUseCase interface {
	GetStandings(ctx context.Context, contestID string, userID string, role string) (*StandingsResponse, error)
}
