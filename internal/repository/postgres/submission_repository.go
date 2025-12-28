package postgres

import (
	"algoforces/internal/domain"
	"context"

	"gorm.io/gorm"
)

type submissionRepository struct {
	db *gorm.DB
}

// NewSubmissionRepository returns the interface type
func NewSubmissionRepository(db *gorm.DB) domain.SubmissionRepository {
	return &submissionRepository{
		db: db,
	}
}

func (r *submissionRepository) CreateNewSubmission(ctx context.Context, submission *domain.Submission) error {
	return r.db.WithContext(ctx).Create(submission).Error
}

// GetSubmissionDetails retrieves submission by unique ID
func (r *submissionRepository) GetSubmissionDetails(ctx context.Context, uniqueID string) (*domain.Submission, error) {
	var submission domain.Submission
	err := r.db.WithContext(ctx).Where("unique_id = ?", uniqueID).First(&submission).Error
	if err != nil {
		return nil, err
	}
	return &submission, nil
}

// UpdateSubmissionStatus updates the submission record
func (r *submissionRepository) UpdateSubmissionStatus(ctx context.Context, submissionID, status string) error {
	return r.db.WithContext(ctx).Model(&domain.Submission{}).Where("unique_id = ?", submissionID).Update("verdict", status).Error
}

func (r *submissionRepository) GetAllTestCasesForProblem(ctx context.Context, problemID string) ([]domain.TestCase, error) {
	var testCases []domain.TestCase
	err := r.db.WithContext(ctx).Where("problem_id = ?", problemID).Find(&testCases).Error
	if err != nil {
		return nil, err
	}
	return testCases, nil
}

func (r *submissionRepository) UpdateSubmissionResult(ctx context.Context, submissionID string, result *domain.Submission) error {
	return r.db.WithContext(ctx).Model(&domain.Submission{}).Where("unique_id = ?", submissionID).Updates(result).Error
}
