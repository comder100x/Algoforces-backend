package services

import (
	"algoforces/internal/domain"
	"context"
	"errors"
)

// CanViewHiddenTestData reports whether a role may see hidden test case
// data. Only admin and problem-setter roles may.
func CanViewHiddenTestData(role string) bool {
	return role == "admin" || role == "problem-setter"
}

// VisibleContestSubmissions returns the submissions a caller may see,
// preserving repository order (newest first).
func VisibleContestSubmissions(submissions []domain.Submission, userID, role string) []domain.Submission {
	if CanViewHiddenTestData(role) {
		return submissions
	}
	visible := make([]domain.Submission, 0, len(submissions))
	for _, submission := range submissions {
		if submission.UserId == userID {
			visible = append(visible, submission)
		}
	}
	return visible
}

// ListContestSubmissions returns the visible submissions, expanded through
// GetSubmissionDetails with the same hidden-test sanitization.
func (s *SubmissionService) ListContestSubmissions(ctx context.Context, contestID, userID, role string) ([]domain.SubmissionDetailsResponse, error) {
	if s.contestSubmissionsRepo == nil {
		return nil, errors.New("contest submissions repository is not configured")
	}
	submissions, err := s.contestSubmissionsRepo.GetSubmissionsByContestID(ctx, contestID)
	if err != nil {
		return nil, err
	}

	visible := VisibleContestSubmissions(submissions, userID, role)
	includeHidden := CanViewHiddenTestData(role)
	results := make([]domain.SubmissionDetailsResponse, 0, len(visible))
	for _, submission := range visible {
		details, err := s.GetSubmissionDetails(ctx, submission.UniqueID, includeHidden)
		if err != nil {
			return nil, err
		}
		results = append(results, *details)
	}
	return results, nil
}
