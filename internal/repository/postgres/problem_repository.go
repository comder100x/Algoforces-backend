package postgres

import (
	"algoforces/internal/domain"
	"context"

	"gorm.io/gorm"
)

type problemRepository struct {
	db *gorm.DB
}

func NewProblemRepository(db *gorm.DB) domain.ProblemRepository {
	return &problemRepository{
		db: db,
	}
}

func (r *problemRepository) CreateProblem(ctx context.Context, problem *domain.Problem) error {
	return r.db.WithContext(ctx).Create(problem).Error
}

func (r *problemRepository) GetProblemByID(ctx context.Context, id string) (*domain.Problem, error) {
	var problem domain.Problem
	err := r.db.WithContext(ctx).Where("unique_id = ?", id).First(&problem).Error
	if err != nil {
		return nil, err
	}
	return &problem, nil
}

func (r *problemRepository) UpdateProblem(ctx context.Context, problem *domain.Problem) error {
	return r.db.WithContext(ctx).Save(problem).Error
}

func (r *problemRepository) DeleteProblem(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("unique_id = ?", id).Delete(&domain.Problem{}).Error
}

func (r *problemRepository) GetAllProblems(ctx context.Context) ([]domain.Problem, error) {
	var problems []domain.Problem
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&problems).Error
	if err != nil {
		return nil, err
	}
	return problems, nil
}
