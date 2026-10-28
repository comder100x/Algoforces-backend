package postgres

import (
	"algoforces/internal/domain"
	"context"
	"fmt"

	"gorm.io/gorm"
)

// NewContestSubmissionsRepository returns a ContestSubmissionsRepository
// backed by the same submissionRepository struct.
func NewContestSubmissionsRepository(db *gorm.DB) domain.ContestSubmissionsRepository {
	return &submissionRepository{
		db: db,
	}
}

// GetSubmissionsByContestID returns the submissions the caller may see,
// newest first. Non-privileged callers get only their own submissions;
// privileged callers get every submission for the contest.
func (r *submissionRepository) GetSubmissionsByContestID(
	ctx context.Context,
	contestID string,
	userID string,
	includeHidden bool,
) ([]domain.Submission, error) {
	query := r.db.WithContext(ctx).Where("contest_id = ?", contestID)
	if !includeHidden {
		query = query.Where("user_id = ?", userID)
	}
	var submissions []domain.Submission
	err := query.Order("submitted_at DESC").Find(&submissions).Error
	if err != nil {
		return nil, fmt.Errorf(
			"get submissions for contest %s: %w",
			contestID,
			err,
		)
	}
	return submissions, nil
}
