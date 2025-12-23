package postgres

import (
	"algoforces/internal/domain"
	"context"
	"fmt"

	"gorm.io/gorm"
)

type adminRepository struct {
	*userRepository // Embed userRepository to inherit all its methods
}

func NewAdminRepository(db *gorm.DB) domain.AdminRepository {
	return &adminRepository{
		userRepository: &userRepository{db: db},
	}
}

func (r *adminRepository) UpdateByEmail(ctx context.Context, email string, user *domain.User) error {
	fmt.Printf("DEBUG: Repository UpdateByEmail called for email: %s\n", email)
	err := r.db.WithContext(ctx).Model(&domain.User{}).Where("email = ?", email).Updates(user).Error
	if err != nil {
		fmt.Printf("DEBUG: Repository UpdateByEmail failed: %v\n", err)
	} else {
		fmt.Println("DEBUG: Repository UpdateByEmail successful")
	}
	return err
}

func (r *adminRepository) GetAllUsers(ctx context.Context) ([]domain.User, error) {
	var users []domain.User
	err := r.db.WithContext(ctx).Find(&users).Error
	return users, err
}

func (r *adminRepository) GetUsersByRole(ctx context.Context, role string) ([]domain.User, error) {
	var users []domain.User
	err := r.db.WithContext(ctx).Where("role = ?", role).Find(&users).Error
	return users, err
}
