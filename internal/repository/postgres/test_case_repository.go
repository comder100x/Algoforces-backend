package postgres

import (
	"algoforces/internal/domain"
	"context"

	"gorm.io/gorm"
)

type testCaseRepository struct {
	db *gorm.DB
}

func NewTestCaseRepository(db *gorm.DB) *testCaseRepository {
	return &testCaseRepository{
		db: db,
	}
}

func (r *testCaseRepository) CreateTestCase(ctx context.Context, testCase *domain.TestCase) error {
	return r.db.WithContext(ctx).Create(testCase).Error
}

func (r *testCaseRepository) UpdateTestCase(ctx context.Context, testCase *domain.TestCase) error {
	return r.db.WithContext(ctx).Save(testCase).Error
}

func (r *testCaseRepository) DeleteTestCase(ctx context.Context, uniqueID string) error {
	return r.db.WithContext(ctx).Delete(&domain.TestCase{}, "unique_id = ?", uniqueID).Error
}

func (r *testCaseRepository) GetTestCasesByProblemID(ctx context.Context, problemID string) ([]*domain.TestCase, error) {
	var testCases []*domain.TestCase
	err := r.db.WithContext(ctx).Where("problem_id = ?", problemID).Order("order_position ASC").Find(&testCases).Error
	if err != nil {
		return nil, err
	}
	return testCases, nil
}

func (r *testCaseRepository) GetTestCaseByUniqueID(ctx context.Context, uniqueID string) (*domain.TestCase, error) {
	var testCase domain.TestCase
	err := r.db.WithContext(ctx).Where("unique_id = ?", uniqueID).First(&testCase).Error
	if err != nil {
		return nil, err
	}
	return &testCase, nil
}
