package services

import (
	"algoforces/internal/domain"
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// CanViewHiddenTestData reports whether a role may see hidden test case
// data. Only admin and problem-setter roles may.
func CanViewHiddenTestData(role string) bool {
	return role == "admin" || role == "problem-setter"
}

// ListContestSubmissions loads the contest first, then returns the
// caller's visible submissions, newest first. No visibility or liveness
// rule is applied beyond existence. Each row is re-queried through
// GetSubmissionDetails so hidden-test sanitisation stays in one place.
func (s *SubmissionService) ListContestSubmissions(
	ctx context.Context,
	contestID string,
	userID string,
	role string,
) ([]domain.SubmissionDetailsResponse, error) {
	if s.contestRepo == nil ||
		s.contestSubmissionsRepo == nil ||
		s.submissionRepo == nil {
		return nil, errors.New("contest submissions dependencies are not configured")
	}
	if _, err := s.contestRepo.GetByID(ctx, contestID); err != nil {
		if errors.Is(err, domain.ErrContestNotFound) ||
			errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf(
				"contest %s: %w",
				contestID,
				domain.ErrContestNotFound,
			)
		}
		return nil, fmt.Errorf("get contest %s: %w", contestID, err)
	}

	includeHidden := CanViewHiddenTestData(role)
	submissions, err := s.contestSubmissionsRepo.GetSubmissionsByContestID(
		ctx,
		contestID,
		userID,
		includeHidden,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"get submissions for contest %s: %w",
			contestID,
			err,
		)
	}

	results := make([]domain.SubmissionDetailsResponse, 0, len(submissions))
	for _, submission := range submissions {
		details, err := s.GetSubmissionDetails(
			ctx,
			submission.UniqueID,
			includeHidden,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"list contest submissions for contest %s: %w",
				contestID,
				err,
			)
		}
		results = append(results, *details)
	}
	return results, nil
}
