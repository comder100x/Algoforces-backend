package postgres

import (
	"algoforces/internal/domain"
	"context"

	"gorm.io/gorm"
)

// NewContestSubmissionsRepository returns a ContestSubmissionsRepository
// backed by the same submissionRepository struct.
func NewContestSubmissionsRepository(db *gorm.DB) domain.ContestSubmissionsRepository {
	return &submissionRepository{
		db: db,
	}
}

// GetSubmissionsByContestID returns every submission for a contest, newest
// first. It applies no ownership or role filtering; the service layer
// filters the result before returning it to a caller.
func (r *submissionRepository) GetSubmissionsByContestID(ctx context.Context, contestID string) ([]domain.Submission, error) {
	var submissions []domain.Submission
	err := r.db.WithContext(ctx).Where("contest_id = ?", contestID).Order("submitted_at DESC").Find(&submissions).Error
	if err != nil {
		return nil, err
	}
	return submissions, nil
}
