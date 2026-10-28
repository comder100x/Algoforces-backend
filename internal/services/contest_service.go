package services

import (
	"algoforces/internal/domain"
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type contestService struct {
	contestRepo         domain.ContestRepository
	userRepo            domain.UserRepository
	contestRegisterRepo domain.ContestRegisterRepository
	contestProblemsRepo domain.ContestProblemsRepository
}

func NewContestService(contestRepo domain.ContestRepository, userRepo domain.UserRepository, contestRegisterRepo domain.ContestRegisterRepository, contestProblemsRepo domain.ContestProblemsRepository) domain.ContestUseCase {
	return &contestService{
		contestRepo:         contestRepo,
		userRepo:            userRepo,
		contestRegisterRepo: contestRegisterRepo,
		contestProblemsRepo: contestProblemsRepo,
	}
}

func (s *contestService) CreateContest(ctx context.Context, req *domain.CreateContestRequest, userId string) (*domain.CreateContestResponse, error) {
	contestList, err := s.contestRepo.CheckContestInTimeWindow(ctx, req.StartTime, req.EndTime)
	if err != nil {
		log.Error().Err(err).Str("user_id", userId).Str("contest_name", req.Name).Msg("Failed to check contest time window")
		return nil, fmt.Errorf("failed to check contest time window: %w", err)
	}

	if len(contestList) > 0 {
		return nil, errors.New("already a contest exists in the given time window")
	}

	// Check if correct problem setters are provided
	for _, psUserId := range req.ProblemSetters {
		user, err := s.userRepo.GetByID(ctx, psUserId)
		if err != nil {
			log.Error().Err(err).Str("user_id", userId).Str("problem-setter_id", psUserId).Msg("Failed to get problem setter user")
			return nil, fmt.Errorf("failed to get problem setter user: %w", err)
		}
		if user.Role != "problem-setter" && user.Role != "admin" {
			log.Warn().Str("user_id", userId).Str("problem-setter_id", psUserId).Str("role", user.Role).Msg("Unauthorized user attempted to be added as problem setter")
			return nil, errors.New("user " + psUserId + " is not authorized as problem setter")
		}
	}

	contest := &domain.Contest{
		Id:             uuid.New().String(),
		Name:           req.Name,
		Description:    req.Description,
		StartTime:      req.StartTime,
		EndTime:        req.EndTime,
		Duration:       req.Duration,
		Visible:        req.Visible,
		IsActive:       req.IsActive,
		ProblemSetters: req.ProblemSetters,
		CreatedBy:      userId,
	}

	err = s.contestRepo.CreateContest(ctx, contest)
	if err != nil {
		log.Error().Err(err).Str("user_id", userId).Str("contest_name", req.Name).Msg("Failed to create contest in database")
		return nil, fmt.Errorf("failed to create contest: %w", err)
	}

	response := &domain.CreateContestResponse{
		Id:             contest.Id,
		Name:           contest.Name,
		Description:    contest.Description,
		StartTime:      contest.StartTime,
		EndTime:        contest.EndTime,
		Duration:       contest.Duration,
		Visible:        contest.Visible,
		IsActive:       contest.IsActive,
		ProblemSetters: contest.ProblemSetters,
		CreatedBy:      contest.CreatedBy,
		CreatedAt:      contest.CreatedAt,
		UpdatedAt:      contest.UpdatedAt,
	}

	return response, nil
}

func (s *contestService) UpdateContest(ctx context.Context, req *domain.UpdateContestRequest) (*domain.UpdateContestResponse, error) {
	// Pre Update validations for time window and problem setters
	contestList, err := s.contestRepo.CheckContestInTimeWindow(ctx, req.StartTime, req.EndTime)
	if err != nil {
		return nil, fmt.Errorf("failed to check contest time window: %w", err)
	}

	if len(contestList) > 0 {
		return nil, errors.New("already a contest exists in the given time window")
	}

	// Check if correct problem setters are provided
	for _, psUserId := range req.ProblemSetters {
		user, err := s.userRepo.GetByID(ctx, psUserId)
		if err != nil {
			return nil, fmt.Errorf("failed to get problem setter user: %w", err)
		}
		if user.Role != "problem-setter" && user.Role != "admin" {
			return nil, errors.New("user " + psUserId + " is not authorized as problem setter")
		}
	}

	contest, err := s.contestRepo.GetByID(ctx, req.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to get contest: %w", err)
	}

	contest.Name = req.Name
	contest.Description = req.Description
	contest.StartTime = req.StartTime
	contest.EndTime = req.EndTime
	contest.Duration = req.Duration
	contest.Visible = req.Visible
	contest.IsActive = req.IsActive
	contest.ProblemSetters = req.ProblemSetters

	err = s.contestRepo.UpdateContest(ctx, contest)
	if err != nil {
		return nil, fmt.Errorf("failed to update contest: %w", err)
	}

	response := &domain.UpdateContestResponse{
		Id:             contest.Id,
		Name:           contest.Name,
		Description:    contest.Description,
		StartTime:      contest.StartTime,
		EndTime:        contest.EndTime,
		Duration:       contest.Duration,
		Visible:        contest.Visible,
		IsActive:       contest.IsActive,
		ProblemSetters: contest.ProblemSetters,
		CreatedBy:      contest.CreatedBy,
		CreatedAt:      contest.CreatedAt,
		UpdatedAt:      contest.UpdatedAt,
	}

	return response, nil
}

func (s *contestService) GetContestDetails(ctx context.Context, id string) (*domain.CreateContestResponse, error) {
	contest, err := s.contestRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get contest: %w", err)
	}

	response := &domain.CreateContestResponse{
		Id:             contest.Id,
		Name:           contest.Name,
		Description:    contest.Description,
		StartTime:      contest.StartTime,
		EndTime:        contest.EndTime,
		Duration:       contest.Duration,
		Visible:        contest.Visible,
		IsActive:       contest.IsActive,
		ProblemSetters: contest.ProblemSetters,
		CreatedBy:      contest.CreatedBy,
		CreatedAt:      contest.CreatedAt,
		UpdatedAt:      contest.UpdatedAt,
	}

	return response, nil
}

func (s *contestService) DeleteContest(ctx context.Context, req *domain.DeleteContestRequest) error {
	if err := s.contestRepo.DeleteContest(ctx, req.Id); err != nil {
		return fmt.Errorf("failed to delete contest: %w", err)
	}
	return nil
}

func (s *contestService) StartContest(ctx context.Context, id string) error {
	contest, err := s.contestRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get contest: %w", err)
	}

	contest.IsActive = true

	if err := s.contestRepo.UpdateContest(ctx, contest); err != nil {
		return fmt.Errorf("failed to update contest: %w", err)
	}
	return nil
}

func (s *contestService) GetAllContests(ctx context.Context) ([]domain.CreateContestResponse, error) {
	contests, err := s.contestRepo.GetAllContests(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get contests: %w", err)
	}

	var contestResponses []domain.CreateContestResponse
	for _, contest := range contests {
		contestResponses = append(contestResponses, domain.CreateContestResponse{
			Id:             contest.Id,
			Name:           contest.Name,
			Description:    contest.Description,
			StartTime:      contest.StartTime,
			EndTime:        contest.EndTime,
			Duration:       contest.Duration,
			Visible:        contest.Visible,
			IsActive:       contest.IsActive,
			ProblemSetters: contest.ProblemSetters,
			CreatedBy:      contest.CreatedBy,
			CreatedAt:      contest.CreatedAt,
			UpdatedAt:      contest.UpdatedAt,
		})
	}

	return contestResponses, nil
}

// Gate checks in order: exists, visible, live, registered or privileged, problem in contest.
func (s *contestService) CheckSubmissionGate(ctx context.Context, userID, contestID, problemID string) error {
	if s.contestRepo == nil || s.contestRegisterRepo == nil || s.contestProblemsRepo == nil {
		return errors.New("contest gate repositories are not configured")
	}
	contest, err := s.contestRepo.GetByID(ctx, contestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrContestNotFound
		}
		return fmt.Errorf("failed to get contest: %w", err)
	}
	if contest == nil {
		return domain.ErrContestNotFound
	}
	registration, err := s.contestRegisterRepo.GetRegistrationByUserAndContest(ctx, userID, contestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			registration = nil
		} else {
			return fmt.Errorf("failed to get contest registration: %w", err)
		}
	}
	contestProblems, err := s.contestProblemsRepo.GetContestProblemsByContestID(ctx, contestID)
	if err != nil {
		return fmt.Errorf("failed to get contest problems: %w", err)
	}
	isProblemInContest := false
	for _, contestProblem := range contestProblems {
		if contestProblem.ProblemID == problemID {
			isProblemInContest = true
			break
		}
	}
	return evaluateSubmissionGate(contest, registration, isProblemInContest, userID, time.Now())
}

// Creators and problem setters bypass only the registration check.
func evaluateSubmissionGate(contest *domain.Contest, registration *domain.ContestRegistration, isProblemInContest bool, userID string, now time.Time) error {
	if contest == nil {
		return domain.ErrContestNotFound
	}
	if !contest.Visible {
		return domain.ErrContestNotVisible
	}
	isLive := contest.IsActive && !now.Before(contest.StartTime) && !now.After(contest.EndTime)
	if !isLive {
		return domain.ErrContestNotLive
	}
	privileged := contest.CreatedBy == userID || slices.Contains(contest.ProblemSetters, userID)
	if !privileged && (registration == nil || registration.Status != domain.RegistrationStatusRegistered) {
		return domain.ErrNotRegistered
	}
	if !isProblemInContest {
		return domain.ErrProblemNotInContest
	}
	return nil
}
