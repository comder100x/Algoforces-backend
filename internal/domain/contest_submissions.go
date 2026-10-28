package domain

import (
	"context"
)

// ContestSubmissionsRepository lists submissions belonging to a contest.
//
// It is deliberately separate from SubmissionRepository so existing
// SubmissionRepository fakes keep compiling unchanged: only callers that
// need the contest-wide list depend on this interface.
type ContestSubmissionsRepository interface {
	// GetSubmissionsByContestID returns every submission for a contest,
	// newest first (SubmittedAt descending). It applies no ownership or
	// role filtering; callers must filter before returning results.
	GetSubmissionsByContestID(ctx context.Context, contestID string) ([]Submission, error)
}
