package postgres

import (
	"algoforces/internal/domain"
	"context"
	"strconv"

	"github.com/rs/zerolog/log"
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

func (r *problemRepository) GetAllProblems(ctx context.Context, pageOffset int, limit int) (*domain.AllProblemListResponse, error) {
	log.Info().Msg("Getting all problems for page offset: " + strconv.Itoa(pageOffset) + " and limit: " + strconv.Itoa(limit))
	var allProblemListResponse domain.AllProblemListResponse

	// Get total count
	var total int64
	if err := r.db.WithContext(ctx).Model(&domain.Problem{}).Count(&total).Error; err != nil {
		return nil, err
	}

	// Get paginated results
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Offset(pageOffset * limit).
		Limit(limit).
		Find(&allProblemListResponse.Problems).Error
	if err != nil {
		log.Error().Msg("Error getting all problems: " + err.Error())
		return nil, err
	}

	log.Info().Msg("Total problems: " + strconv.Itoa(int(total)))
	allProblemListResponse.Total = int(total)
	return &allProblemListResponse, nil
}
