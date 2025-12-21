package services

import (
	"algoforces/internal/domain"
	"algoforces/internal/utils"
	"context"
	"errors"

	"github.com/google/uuid"
)

type authService struct {
	userRepo domain.UserRepository
}

func NewAuthService(userRepo domain.UserRepository) domain.UserUseCase {
	return &authService{
		userRepo: userRepo,
	}
}

func (s *authService) Signup(ctx context.Context, req *domain.SignupRequest) (*domain.AuthResponse, error) {
	// Check if user already exists
	existingUser, _ := s.userRepo.GetByEmail(ctx, req.Email) // Ignore error as it's expected when user doesn't exist
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	hashedPassword, err := utils.HashPassword(req.Password)

	if err != nil {
		return nil, err
	}

	newUser := &domain.User{
		Id:       uuid.New().String(),
		Email:    req.Email,
		Password: hashedPassword,
		Username: req.Username,
		Role:     "user",
	}

	err = s.userRepo.Create(ctx, newUser)

	if err != nil {
		return nil, err
	}

	jwtToken, err := utils.GenerateToken(newUser.Id, newUser.Role, newUser.Email)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		AccessToken: jwtToken,
		User:        *newUser,
	}, nil
}

func (s *authService) Login(ctx context.Context, req *domain.LoginRequest) (*domain.AuthResponse, error) {

	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	if !utils.VerifyPassword(req.Password, user.Password) {
		return nil, errors.New("Password is incorrect")
	}

	jwtToken, err := utils.GenerateToken(user.Id, user.Role, user.Email)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		AccessToken: jwtToken,
		User:        *user,
	}, nil
}

func (s *authService) GetUserProfile(ctx context.Context, userID string) (*domain.UserProfileResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	return &domain.UserProfileResponse{
		Username:  user.Email, // Using email as username since User struct doesn't have username field
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}, nil
}

func (s *authService) UpdateUserProfile(ctx context.Context, userID string, req *domain.UpdateUserProfileRequest) (*domain.UserProfileResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	user.Email = req.Email
	user.Username = req.Username

	err = s.userRepo.UpdateByID(ctx, userID, user)
	if err != nil {
		return nil, err
	}

	return &domain.UserProfileResponse{
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}, nil
}
