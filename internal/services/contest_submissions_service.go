package services

import (
	"algoforces/internal/domain"
	"context"
	"errors"
)

// CanViewHiddenTestData reports whether a role may see hidden test case
// data (input, expected output, and test case IDs). Only admin and
// problem-setter roles may; every other role, including an empty or
// unknown role, may not. This is the single decision point for the
// includeHidden flag passed to GetSubmissionDetails from the contest
// submissions list, so hidden test data can never leak to regular users.
func CanViewHiddenTestData(role string) bool {
	return role == "admin" || role == "problem-setter"
}

// VisibleContestSubmissions returns the submissions a caller with the
// given userID and role may see, preserving the repository order (newest
// first). Admin and problem-setter callers see every submission; any
// other caller sees only their own submissions.
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

// ListContestSubmissions returns the contest submissions visible to the
// caller, newest first. Each visible submission is expanded through the
// existing GetSubmissionDetails path, which applies the same hidden-test
// sanitization as the single-submission endpoint: includeHidden is true
// only for admin/problem-setter callers, and regular users additionally
// only ever see their own submissions (see VisibleContestSubmissions).
// It does not change submission creation or judging.
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
