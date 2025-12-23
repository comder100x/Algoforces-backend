package services

import (
	"algoforces/internal/domain"
	"context"
	"errors"
)

type adminService struct {
	adminRepo domain.AdminRepository
}

func NewAdminService(adminRepo domain.AdminRepository) domain.AdminUseCase {
	return &adminService{
		adminRepo: adminRepo,
	}
}
func (s *adminService) AddRole(ctx context.Context, req *domain.AddRoleRequest) (*domain.AddRoleResponse, error) {
	user, err := s.adminRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("Error fetching user by email: " + err.Error())
	}

	if user == nil {
		return nil, errors.New("User not found")
	}

	user.Role = req.Role

	err = s.adminRepo.UpdateByEmail(ctx, req.Email, user)
	if err != nil {
		return nil, errors.New("Error updating user role: " + err.Error())
	}

	return &domain.AddRoleResponse{
		Email:       user.Email,
		Role:        req.Role,
		CurrentRole: user.Role,
	}, nil

}
func (s *adminService) RemoveRole(ctx context.Context, req *domain.RemoveRoleRequest) (*domain.RemoveRoleResponse, error) {
	user, err := s.adminRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("Error fetching user by email: " + err.Error())
	}

	if user == nil {
		return nil, errors.New("User not found")
	}

	user.Role = "user" // Default role after removal

	err = s.adminRepo.UpdateByEmail(ctx, req.Email, user)
	if err != nil {
		return nil, errors.New("Error updating user role: " + err.Error())
	}

	return &domain.RemoveRoleResponse{
		Email:       user.Email,
		Role:        req.Role,
		CurrentRole: user.Role,
	}, nil
}

func (s *adminService) GetAllUsers(ctx context.Context) (*domain.UserListResponse, error) {
	users, err := s.adminRepo.GetAllUsers(ctx)
	if err != nil {
		return nil, errors.New("Error fetching all users: " + err.Error())
	}

	return s.toUserListResponse(users), nil
}

func (s *adminService) GetUsersByRole(ctx context.Context, role string) (*domain.UserListResponse, error) {
	users, err := s.adminRepo.GetUsersByRole(ctx, role)
	if err != nil {
		return nil, errors.New("Error fetching users by role: " + err.Error())
	}

	return s.toUserListResponse(users), nil
}

func (s *adminService) toUserListResponse(users []domain.User) *domain.UserListResponse {
	userItems := make([]domain.UserListItem, len(users))
	for i, user := range users {
		userItems[i] = domain.UserListItem{
			ID:        user.Id,
			Username:  user.Username,
			Email:     user.Email,
			Role:      user.Role,
			CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return &domain.UserListResponse{
		Users: userItems,
		Total: len(userItems),
	}
}
