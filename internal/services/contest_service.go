package services

import (
	"algoforces/internal/domain"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type contestService struct {
	contestRepo         domain.ContestRepository
	userRepo            domain.UserRepository
	contestProblemsRepo domain.ContestProblemsRepository
}

func NewContestService(contestRepo domain.ContestRepository, userRepo domain.UserRepository, contestProblemsRepo domain.ContestProblemsRepository) domain.ContestUseCase {
	return &contestService{
		contestRepo:         contestRepo,
		userRepo:            userRepo,
		contestProblemsRepo: contestProblemsRepo,
	}
}

func (s *contestService) CreateContest(ctx context.Context, req *domain.CreateContestRequest, userId string) (*domain.CreateContestResponse, error) {
	contestList, err := s.contestRepo.CheckContestInTimeWindowExcluding(ctx, req.StartTime, req.EndTime, "")
	if err != nil {
		log.Error().Err(err).Str("user_id", userId).Str("contest_name", req.Name).Msg("Failed to check contest time window")
		return nil, err
	}

	if len(contestList) > 0 {
		return nil, errors.New("already a contest exists in the given time window")
	}

	// Check if correct problem setters are provided
	for _, psUserId := range req.ProblemSetters {
		user, err := s.userRepo.GetByID(ctx, psUserId)
		if err != nil {
			log.Error().Err(err).Str("user_id", userId).Str("problem-setter_id", psUserId).Msg("Failed to get problem setter user")
			return nil, err
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
		Status:         domain.ContestStatus(*contest, time.Now()),
		ProblemSetters: contest.ProblemSetters,
		CreatedBy:      contest.CreatedBy,
		CreatedAt:      contest.CreatedAt,
		UpdatedAt:      contest.UpdatedAt,
	}

	return response, nil
}

func (s *contestService) UpdateContest(ctx context.Context, req *domain.UpdateContestRequest) (*domain.UpdateContestResponse, error) {
	contest, err := s.contestRepo.GetByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	// Once the contest has started (now >= StartTime), its time window is
	// locked: changing StartTime or EndTime is rejected. Other fields may
	// still be updated, and re-saving the unchanged times is allowed.
	if !time.Now().Before(contest.StartTime) && (!req.StartTime.Equal(contest.StartTime) || !req.EndTime.Equal(contest.EndTime)) {
		return nil, errors.New("cannot change start time or end time after the contest has started")
	}

	// Pre Update validations for time window and problem setters. The
	// contest being updated is excluded from the overlap check so it
	// does not collide with its own existing window.
	contestList, err := s.contestRepo.CheckContestInTimeWindowExcluding(ctx, req.StartTime, req.EndTime, req.Id)
	if err != nil {
		return nil, err
	}

	if len(contestList) > 0 {
		return nil, errors.New("already a contest exists in the given time window")
	}

	// Check if correct problem setters are provided
	for _, psUserId := range req.ProblemSetters {
		user, err := s.userRepo.GetByID(ctx, psUserId)
		if err != nil {
			return nil, err
		}
		if user.Role != "problem-setter" && user.Role != "admin" {
			return nil, errors.New("user " + psUserId + " is not authorized as problem setter")
		}
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
		return nil, err
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
		Status:         domain.ContestStatus(*contest, time.Now()),
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
		Status:         domain.ContestStatus(*contest, time.Now()),
		ProblemSetters: contest.ProblemSetters,
		CreatedBy:      contest.CreatedBy,
		CreatedAt:      contest.CreatedAt,
		UpdatedAt:      contest.UpdatedAt,
	}

	return response, nil
}

func (s *contestService) DeleteContest(ctx context.Context, req *domain.DeleteContestRequest) error {
	contest, err := s.contestRepo.GetByID(ctx, req.Id)
	if err != nil {
		return err
	}

	// A live contest cannot be deleted.
	if domain.ContestStatus(*contest, time.Now()) == "live" {
		return errors.New("cannot delete a live contest")
	}

	return s.contestRepo.DeleteContest(ctx, req.Id)
}

func (s *contestService) StartContest(ctx context.Context, id string) error {
	contest, err := s.contestRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// An ended contest (now > EndTime) cannot be started again.
	if time.Now().After(contest.EndTime) {
		return errors.New("cannot start a contest that has ended")
	}

	if s.contestProblemsRepo == nil {
		return errors.New("contest problems repository is not configured")
	}

	problems, err := s.contestProblemsRepo.GetContestProblemsByContestID(ctx, id)
	if err != nil {
		return err
	}
	if len(problems) == 0 {
		return errors.New("contest must have at least one problem before it can be started")
	}

	contest.IsActive = true

	return s.contestRepo.UpdateContest(ctx, contest)
}

// EndContest deactivates a contest. EndTime is left unchanged; the derived
// status helper treats a started-but-inactive contest as "ended".
func (s *contestService) EndContest(ctx context.Context, id string) error {
	contest, err := s.contestRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	contest.IsActive = false

	return s.contestRepo.UpdateContest(ctx, contest)
}

func (s *contestService) GetAllContests(ctx context.Context) ([]domain.CreateContestResponse, error) {
	contests, err := s.contestRepo.GetAllContests(ctx)
	if err != nil {
		return nil, err
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
			Status:         domain.ContestStatus(contest, time.Now()),
			ProblemSetters: contest.ProblemSetters,
			CreatedBy:      contest.CreatedBy,
			CreatedAt:      contest.CreatedAt,
			UpdatedAt:      contest.UpdatedAt,
		})
	}

	return contestResponses, nil
}

func (s *contestService) GetVisibleContests(ctx context.Context) ([]domain.CreateContestResponse, error) {
	contests, err := s.contestRepo.GetVisibleContests(ctx)
	if err != nil {
		return nil, err
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
			Status:         domain.ContestStatus(contest, time.Now()),
			ProblemSetters: contest.ProblemSetters,
			CreatedBy:      contest.CreatedBy,
			CreatedAt:      contest.CreatedAt,
			UpdatedAt:      contest.UpdatedAt,
		})
	}

	return contestResponses, nil
}
