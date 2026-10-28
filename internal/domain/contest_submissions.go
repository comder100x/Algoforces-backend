package domain

import (
	"context"
	"errors"
)

// ErrContestNotFound is returned when no contest exists with the requested ID.
var ErrContestNotFound = errors.New("contest not found")

// ContestSubmissionsRepository lists submissions belonging to a contest.
// It is separate from SubmissionRepository so existing fakes keep compiling.
type ContestSubmissionsRepository interface {
	// GetSubmissionsByContestID returns the submissions the caller may
	// see, newest first (SubmittedAt descending). When includeHidden is
	// false only the caller's own submissions are returned; the
	// ownership filter lives in SQL, not in the caller.
	GetSubmissionsByContestID(
		ctx context.Context,
		contestID string,
		userID string,
		includeHidden bool,
	) ([]Submission, error)
}
