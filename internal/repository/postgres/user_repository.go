package postgres

import (
	"algoforces/internal/domain"
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) UpdateByID(ctx context.Context, id string, user *domain.User) error {
	fmt.Printf("DEBUG: Repository UpdateByID called for user ID: %s\n", id)
	err := r.db.WithContext(ctx).Model(&domain.User{}).Where("id = ?", id).Updates(user).Error
	if err != nil {
		fmt.Printf("DEBUG: Repository UpdateByID failed: %v\n", err)
	} else {
		fmt.Println("DEBUG: Repository UpdateByID successful")
	}
	return err
}
