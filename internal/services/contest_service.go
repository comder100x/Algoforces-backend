package services

import (
	"algoforces/internal/domain"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type contestService struct {
	contestRepo         domain.ContestRepository
	userRepo            domain.UserRepository
	contestProblemsRepo domain.ContestProblemsRepository
}

func NewContestService(
	contestRepo domain.ContestRepository,
	userRepo domain.UserRepository,
	contestProblemsRepo domain.ContestProblemsRepository,
) domain.ContestUseCase {
	return &contestService{
		contestRepo:         contestRepo,
		userRepo:            userRepo,
		contestProblemsRepo: contestProblemsRepo,
	}
}

func (s *contestService) getContest(ctx context.Context, id string) (*domain.Contest, error) {
	contest, err := s.contestRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrContestNotFound
		}
		return nil, fmt.Errorf("failed to get contest: %w", err)
	}
	if contest == nil {
		return nil, domain.ErrContestNotFound
	}
	return contest, nil
}

func (s *contestService) checkProblemSetters(
	ctx context.Context,
	setterIDs []string,
) error {
	for _, psUserId := range setterIDs {
		user, err := s.userRepo.GetByID(ctx, psUserId)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: %s", domain.ErrProblemSetterNotFound, psUserId)
			}
			return fmt.Errorf("failed to get problem setter user %s: %w", psUserId, err)
		}
		if user == nil || (user.Role != "problem-setter" && user.Role != "admin") {
			return fmt.Errorf("%w: %s", domain.ErrProblemSetterNotAuthorized, psUserId)
		}
	}
	return nil
}

func (s *contestService) CreateContest(
	ctx context.Context,
	req *domain.CreateContestRequest,
	userId string,
) (*domain.CreateContestResponse, error) {
	if req == nil {
		return nil, errors.New("contest request is required")
	}

	contestList, err := s.contestRepo.CheckContestInTimeWindowExcluding(
		ctx,
		req.StartTime,
		req.EndTime,
		"",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check contest time window: %w", err)
	}
	if len(contestList) > 0 {
		return nil, domain.ErrContestOverlap
	}

	if err := s.checkProblemSetters(ctx, req.ProblemSetters); err != nil {
		return nil, err
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

	if err := s.contestRepo.CreateContest(ctx, contest); err != nil {
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
		Status:         domain.DeriveContestStatus(*contest, time.Now()),
		ProblemSetters: contest.ProblemSetters,
		CreatedBy:      contest.CreatedBy,
		CreatedAt:      contest.CreatedAt,
		UpdatedAt:      contest.UpdatedAt,
	}

	return response, nil
}

func (s *contestService) UpdateContest(
	ctx context.Context,
	req *domain.UpdateContestRequest,
) (*domain.UpdateContestResponse, error) {
	if req == nil {
		return nil, errors.New("contest request is required")
	}

	contest, err := s.getContest(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	// Once the contest has started (now >= StartTime), its time window is
	// locked: changing StartTime or EndTime is rejected. Other fields may
	// still be updated, and re-saving the unchanged times is allowed.
	started := !time.Now().Before(contest.StartTime)
	timesChanged := !req.StartTime.Equal(contest.StartTime) || !req.EndTime.Equal(contest.EndTime)
	if started && timesChanged {
		return nil, domain.ErrContestTimeLocked
	}

	// Pre Update validations for time window and problem setters. The
	// contest being updated is excluded from the overlap check so it
	// does not collide with its own existing window.
	contestList, err := s.contestRepo.CheckContestInTimeWindowExcluding(
		ctx,
		req.StartTime,
		req.EndTime,
		req.Id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check contest time window: %w", err)
	}
	if len(contestList) > 0 {
		return nil, domain.ErrContestOverlap
	}

	if err := s.checkProblemSetters(ctx, req.ProblemSetters); err != nil {
		return nil, err
	}

	contest.Name = req.Name
	contest.Description = req.Description
	contest.StartTime = req.StartTime
	contest.EndTime = req.EndTime
	contest.Duration = req.Duration
	contest.Visible = req.Visible
	contest.IsActive = req.IsActive
	contest.ProblemSetters = req.ProblemSetters

	if err := s.contestRepo.UpdateContest(ctx, contest); err != nil {
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
		Status:         domain.DeriveContestStatus(*contest, time.Now()),
		ProblemSetters: contest.ProblemSetters,
		CreatedBy:      contest.CreatedBy,
		CreatedAt:      contest.CreatedAt,
		UpdatedAt:      contest.UpdatedAt,
	}

	return response, nil
}

func (s *contestService) GetContestDetails(ctx context.Context, id string) (*domain.CreateContestResponse, error) {
	contest, err := s.getContest(ctx, id)
	if err != nil {
		return nil, err
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
		Status:         domain.DeriveContestStatus(*contest, time.Now()),
		ProblemSetters: contest.ProblemSetters,
		CreatedBy:      contest.CreatedBy,
		CreatedAt:      contest.CreatedAt,
		UpdatedAt:      contest.UpdatedAt,
	}

	return response, nil
}

func (s *contestService) DeleteContest(ctx context.Context, req *domain.DeleteContestRequest) error {
	if req == nil {
		return errors.New("delete contest request is required")
	}

	contest, err := s.getContest(ctx, req.Id)
	if err != nil {
		return err
	}

	// A live contest cannot be deleted.
	if domain.DeriveContestStatus(*contest, time.Now()) == domain.ContestStatusLive {
		return domain.ErrContestLive
	}

	if err := s.contestRepo.DeleteContest(ctx, req.Id); err != nil {
		return fmt.Errorf("failed to delete contest: %w", err)
	}
	return nil
}

func (s *contestService) StartContest(ctx context.Context, id string) error {
	contest, err := s.getContest(ctx, id)
	if err != nil {
		return err
	}

	// An ended contest (now > EndTime) cannot be started again.
	if time.Now().After(contest.EndTime) {
		return domain.ErrContestEnded
	}

	if s.contestProblemsRepo == nil {
		return domain.ErrContestProblemsNotConfigured
	}

	problems, err := s.contestProblemsRepo.GetContestProblemsByContestID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get contest problems: %w", err)
	}
	if len(problems) == 0 {
		return domain.ErrContestHasNoProblems
	}

	contest.IsActive = true

	if err := s.contestRepo.UpdateContest(ctx, contest); err != nil {
		return fmt.Errorf("failed to update contest: %w", err)
	}
	return nil
}

// EndContest deactivates a contest without changing EndTime; the
// derived status then reports it as ended.
func (s *contestService) EndContest(ctx context.Context, id string) error {
	contest, err := s.getContest(ctx, id)
	if err != nil {
		return err
	}

	contest.IsActive = false

	if err := s.contestRepo.UpdateContest(ctx, contest); err != nil {
		return fmt.Errorf("failed to update contest: %w", err)
	}
	return nil
}

func toContestResponse(contest domain.Contest, now time.Time) domain.CreateContestResponse {
	return domain.CreateContestResponse{
		Id:             contest.Id,
		Name:           contest.Name,
		Description:    contest.Description,
		StartTime:      contest.StartTime,
		EndTime:        contest.EndTime,
		Duration:       contest.Duration,
		Visible:        contest.Visible,
		IsActive:       contest.IsActive,
		Status:         domain.DeriveContestStatus(contest, now),
		ProblemSetters: contest.ProblemSetters,
		CreatedBy:      contest.CreatedBy,
		CreatedAt:      contest.CreatedAt,
		UpdatedAt:      contest.UpdatedAt,
	}
}

func (s *contestService) GetAllContests(ctx context.Context) ([]domain.CreateContestResponse, error) {
	contests, err := s.contestRepo.GetAllContests(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get contests: %w", err)
	}

	now := time.Now()
	contestResponses := make([]domain.CreateContestResponse, 0, len(contests))
	for _, contest := range contests {
		contestResponses = append(contestResponses, toContestResponse(contest, now))
	}

	return contestResponses, nil
}

func (s *contestService) GetVisibleContests(ctx context.Context) ([]domain.CreateContestResponse, error) {
	contests, err := s.contestRepo.GetVisibleContests(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get contests: %w", err)
	}

	now := time.Now()
	contestResponses := make([]domain.CreateContestResponse, 0, len(contests))
	for _, contest := range contests {
		contestResponses = append(contestResponses, toContestResponse(contest, now))
	}

	return contestResponses, nil
}
