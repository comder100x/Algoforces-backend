package domain

import (
	"context"
)

// ContestSubmissionsRepository lists submissions belonging to a contest.
// It is separate from SubmissionRepository so existing fakes keep compiling.
type ContestSubmissionsRepository interface {
	// GetSubmissionsByContestID returns every submission for a contest,
	// newest first (SubmittedAt descending). It applies no ownership or
	// role filtering; callers must filter before returning results.
	GetSubmissionsByContestID(ctx context.Context, contestID string) ([]Submission, error)
}
