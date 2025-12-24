package postgres

import (
	"algoforces/internal/domain"
	"context"
	"time"

	"gorm.io/gorm"
)

type contestRepository struct {
	db *gorm.DB
}

// NewContestRepository creates a new contest repository
func NewContestRepository(db *gorm.DB) *contestRepository {
	return &contestRepository{
		db: db,
	}
}

func (r *contestRepository) CreateContest(ctx context.Context, contest *domain.Contest) error {
	return r.db.WithContext(ctx).Create(contest).Error
}

func (r *contestRepository) GetByID(ctx context.Context, id string) (*domain.Contest, error) {
	var contest domain.Contest
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&contest).Error
	if err != nil {
		return nil, err
	}
	return &contest, nil
}

func (r *contestRepository) UpdateContest(ctx context.Context, contest *domain.Contest) error {
	return r.db.WithContext(ctx).Save(contest).Error
}

func (r *contestRepository) DeleteContest(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.Contest{}).Error
}

func (r *contestRepository) CheckContestInTimeWindow(ctx context.Context, startTime, endTime time.Time) ([]domain.Contest, error) {
	var contests []domain.Contest
	err := r.db.WithContext(ctx).
		Where("start_time < ? AND end_time > ?", endTime, startTime).
		Find(&contests).Error
	if err != nil {
		return nil, err
	}
	return contests, nil
}
