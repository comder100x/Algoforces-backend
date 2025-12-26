package postgres

import (
	"algoforces/internal/domain"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type contestRegisterRepository struct {
	db *gorm.DB
}

func NewContestRegisterRepository(db *gorm.DB) domain.ContestRegisterRepository {
	return &contestRegisterRepository{
		db: db,
	}
}

func (r *contestRegisterRepository) CreateContestRegistration(ctx context.Context, userID string, contestID string) (*domain.ContestRegistration, error) {
	registration := &domain.ContestRegistration{
		ID:        uuid.New().String(),
		UserID:    userID,
		ContestID: contestID,
		Status:    "registered",
	}
	err := r.db.WithContext(ctx).Create(registration).Error
	if err != nil {
		return nil, err
	}
	return registration, nil
}

func (r *contestRegisterRepository) GetRegistrationByUserAndContest(ctx context.Context, userID string, contestID string) (*domain.ContestRegistration, error) {
	var registration domain.ContestRegistration
	err := r.db.WithContext(ctx).Where("user_id = ? AND contest_id = ?", userID, contestID).First(&registration).Error
	if err != nil {
		return nil, err
	}
	return &registration, nil
}

func (r *contestRegisterRepository) UpdateRegistrationStatus(ctx context.Context, userID string, contestID string, status string) error {
	err := r.db.WithContext(ctx).Model(&domain.ContestRegistration{}).Where("user_id = ? AND contest_id = ?", userID, contestID).Update("status", status).Error
	return err
}

func (r *contestRegisterRepository) GetAllRegistrationsByUserID(ctx context.Context, userID string) ([]domain.ContestRegistration, error) {
	var registrations []domain.ContestRegistration
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&registrations).Error
	return registrations, err
}

func (r *contestRegisterRepository) GetAllRegistrationsForAdmin(ctx context.Context) ([]domain.ContestRegistration, error) {
	var registrations []domain.ContestRegistration
	err := r.db.WithContext(ctx).Order("registered_at DESC").Find(&registrations).Error
	if err != nil {
		return nil, err
	}
	return registrations, nil
}
