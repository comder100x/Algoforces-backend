package services

import (
	"algoforces/internal/domain"
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type contestRegisterService struct {
	contestRegisterRepo domain.ContestRegisterRepository
	contestRepo         domain.ContestRepository
	userRepo            domain.UserRepository
}

func NewContestRegisterService(contestRegisterRepo domain.ContestRegisterRepository, contestRepo domain.ContestRepository, userRepo domain.UserRepository) domain.ContestRegisterUseCase {
	return &contestRegisterService{
		contestRegisterRepo: contestRegisterRepo,
		contestRepo:         contestRepo,
		userRepo:            userRepo,
	}
}

func (s *contestRegisterService) RegisterContest(ctx context.Context, userID string, req *domain.ContestRegisterRequest) (*domain.ContestRegisterResponse, error) {
	// Verify user exists
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		// User repo returns plain "user not found", not gorm.ErrRecordNotFound.
		if errors.Is(err, gorm.ErrRecordNotFound) || err.Error() == domain.ErrUserNotFound.Error() {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	// Verify contest exists
	contest, err := s.contestRepo.GetByID(ctx, req.ContestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrContestNotFound
		}
		return nil, fmt.Errorf("failed to get contest: %w", err)
	}
	if contest == nil {
		return nil, domain.ErrContestNotFound
	}

	// Registration needs a visible contest that has not ended.
	if !contest.Visible {
		return nil, domain.ErrContestNotVisible
	}
	if time.Now().After(contest.EndTime) {
		return nil, errors.New("contest has ended")
	}

	// Verify contest is open for registration it should be 5 minutes before contest startTime
	contestStartTime := contest.StartTime
	if time.Until(contestStartTime) < 5*time.Minute {
		return nil, errors.New("contest registration is closed")
	}

	// Check if already registered
	existingRegistration, err := s.contestRegisterRepo.GetRegistrationByUserAndContest(ctx, userID, req.ContestID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to get contest registration: %w", err)
	}
	if err == nil && existingRegistration != nil {
		if existingRegistration.Status == domain.RegistrationStatusRegistered {
			return nil, errors.New("already registered for this contest")
		}
		// If previously unregistered, update status to registered
		err = s.contestRegisterRepo.UpdateRegistrationStatus(ctx, userID, req.ContestID, domain.RegistrationStatusRegistered)
		if err != nil {
			return nil, fmt.Errorf("failed to update registration status: %w", err)
		}
		return &domain.ContestRegisterResponse{
			ID:           existingRegistration.ID,
			UserID:       existingRegistration.UserID,
			ContestID:    existingRegistration.ContestID,
			RegisteredAt: existingRegistration.RegisteredAt,
			Status:       domain.RegistrationStatusRegistered,
		}, nil
	}

	// Create new registration
	registration, err := s.contestRegisterRepo.CreateContestRegistration(ctx, userID, req.ContestID)
	if err != nil {
		return nil, fmt.Errorf("failed to create contest registration: %w", err)
	}

	return &domain.ContestRegisterResponse{
		ID:           registration.ID,
		UserID:       registration.UserID,
		ContestID:    registration.ContestID,
		RegisteredAt: registration.RegisteredAt,
		Status:       registration.Status,
	}, nil
}

func (s *contestRegisterService) UnregisterContest(ctx context.Context, userID string, req *domain.ContestUnregisterRequest) error {
	// Check if registered
	registration, err := s.contestRegisterRepo.GetRegistrationByUserAndContest(ctx, userID, req.ContestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrNotRegistered
		}
		return fmt.Errorf("failed to get contest registration: %w", err)
	}
	if registration == nil {
		return domain.ErrNotRegistered
	}

	// Check the time constraint: unregistration allowed only if more than 2 minutes before contest start time

	contest, err := s.contestRepo.GetByID(ctx, req.ContestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrContestNotFound
		}
		return fmt.Errorf("failed to get contest: %w", err)
	}
	if contest == nil {
		return domain.ErrContestNotFound
	}

	contestStartTime := contest.StartTime
	if time.Until(contestStartTime) < 2*time.Minute {
		return errors.New("unregistration period has passed")
	}

	// Update status to unregistered
	err = s.contestRegisterRepo.UpdateRegistrationStatus(ctx, userID, req.ContestID, domain.RegistrationStatusUnregistered)
	if err != nil {
		return fmt.Errorf("failed to update registration status: %w", err)
	}

	return nil
}

func (s *contestRegisterService) GetAllRegistrations(ctx context.Context, userID string) (*domain.AllRegisteredContestForUserResponse, error) {
	registrations, err := s.contestRegisterRepo.GetAllRegistrationsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get registrations: %w", err)
	}

	var responses []domain.ContestRegisterResponse
	for _, reg := range registrations {
		responses = append(responses, domain.ContestRegisterResponse{
			ID:           reg.ID,
			UserID:       reg.UserID,
			ContestID:    reg.ContestID,
			RegisteredAt: reg.RegisteredAt,
			Status:       reg.Status,
		})
	}

	return &domain.AllRegisteredContestForUserResponse{
		Registrations: responses,
	}, nil
}

func (s *contestRegisterService) GetAllRegistrationsForAdmin(ctx context.Context) (*domain.AllRegisteredContestForUserResponse, error) {
	registrations, err := s.contestRegisterRepo.GetAllRegistrationsForAdmin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get registrations: %w", err)
	}

	var responses []domain.ContestRegisterResponse
	for _, reg := range registrations {
		responses = append(responses, domain.ContestRegisterResponse{
			ID:           reg.ID,
			UserID:       reg.UserID,
			ContestID:    reg.ContestID,
			RegisteredAt: reg.RegisteredAt,
			Status:       reg.Status,
		})
	}

	return &domain.AllRegisteredContestForUserResponse{
		Registrations: responses,
	}, nil
}
