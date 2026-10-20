package postgres

import (
	"algoforces/internal/domain"
	"context"

	"gorm.io/gorm"
)

type contestProblemsRepository struct {
	db *gorm.DB
}

// NewContestProblemsRepository creates a new contest problems repository
func NewContestProblemsRepository(db *gorm.DB) domain.ContestProblemsRepository {
	return &contestProblemsRepository{
		db: db,
	}
}

func (r *contestProblemsRepository) CreateContestProblem(ctx context.Context, contestProblem *domain.ContestProblems) error {
	return r.db.WithContext(ctx).Create(contestProblem).Error
}

func (r *contestProblemsRepository) GetContestProblemByID(ctx context.Context, uniqueID string) (*domain.ContestProblems, error) {
	var contestProblem domain.ContestProblems
	err := r.db.WithContext(ctx).Where("unique_id = ?", uniqueID).First(&contestProblem).Error
	if err != nil {
		return nil, err
	}
	return &contestProblem, nil
}

func (r *contestProblemsRepository) GetContestProblemsByContestID(ctx context.Context, contestID string) ([]domain.ContestProblems, error) {
	var contestProblems []domain.ContestProblems
	err := r.db.WithContext(ctx).
		Where("contest_id = ?", contestID).
		Order("order_position ASC").
		Find(&contestProblems).Error
	if err != nil {
		return nil, err
	}
	return contestProblems, nil
}

func (r *contestProblemsRepository) UpdateContestProblem(ctx context.Context, contestProblem *domain.ContestProblems) error {
	return r.db.WithContext(ctx).Save(contestProblem).Error
}

func (r *contestProblemsRepository) DeleteContestProblem(ctx context.Context, uniqueID string) error {
	return r.db.WithContext(ctx).Where("unique_id = ?", uniqueID).Delete(&domain.ContestProblems{}).Error
}

func (r *contestProblemsRepository) GetContestProblemsByContestIDWithDetails(ctx context.Context, contestID string) ([]domain.ContestProblemDetail, error) {
	var contestProblems []domain.ContestProblems
	err := r.db.WithContext(ctx).
		Where("contest_id = ?", contestID).
		Order("order_position ASC").
		Find(&contestProblems).Error
	if err != nil {
		return nil, err
	}

	var details []domain.ContestProblemDetail
	for _, cp := range contestProblems {
		var problem domain.Problem
		err := r.db.WithContext(ctx).Where("unique_id = ?", cp.ProblemID).First(&problem).Error

		detail := domain.ContestProblemDetail{
			UniqueID:      cp.UniqueID,
			ContestID:     cp.ContestID,
			ProblemID:     cp.ProblemID,
			OrderPosition: cp.OrderPosition,
			MaxPoints:     cp.MaxPoints,
			CreatedAt:     cp.CreatedAt,
			UpdatedAt:     cp.UpdatedAt,
		}

		if err == nil {
			detail.Problem = &problem
		}

		details = append(details, detail)
	}

	return details, nil
}
