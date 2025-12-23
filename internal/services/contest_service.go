package services

import (
	"algoforces/internal/domain"
	"context"
	"errors"

	"github.com/google/uuid"
)

type contestService struct {
	contestRepo domain.ContestRepository
}

func NewContestService(contestRepo domain.ContestRepository) domain.ContestUseCase {
	return &contestService{
		contestRepo: contestRepo,
	}
}

func (s *contestService) CreateContest(ctx context.Context, req *domain.CreateContestRequest, userId string) (*domain.CreateContestResponse, error) {
	contestList, err := s.contestRepo.CheckContestInTimeWindow(ctx, req.StartTime, req.EndTime)
	if err != nil {
		return nil, err
	}

	if len(contestList) > 0 {
		return nil, errors.New("already a contest exists in the given time window")
	}

	contest := &domain.Contest{
		Id:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Duration:    req.Duration,
		Visible:     req.Visible,
		CreatedBy:   userId,
	}

	err = s.contestRepo.CreateContest(ctx, contest)
	if err != nil {
		return nil, err
	}

	response := &domain.CreateContestResponse{
		Id:          contest.Id,
		Name:        contest.Name,
		Description: contest.Description,
		StartTime:   contest.StartTime,
		EndTime:     contest.EndTime,
		Duration:    contest.Duration,
		Visible:     contest.Visible,
		CreatedBy:   contest.CreatedBy,
		CreatedAt:   contest.CreatedAt,
		UpdatedAt:   contest.UpdatedAt,
	}

	return response, nil
}

func (s *contestService) UpdateContest(ctx context.Context, req *domain.UpdateContestRequest) (*domain.UpdateContestResponse, error) {
	contest, err := s.contestRepo.GetByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	contest.Name = req.Name
	contest.Description = req.Description
	contest.StartTime = req.StartTime
	contest.EndTime = req.EndTime
	contest.Duration = req.Duration
	contest.Visible = req.Visible

	err = s.contestRepo.UpdateContest(ctx, contest)
	if err != nil {
		return nil, err
	}

	response := &domain.UpdateContestResponse{
		Id:          contest.Id,
		Name:        contest.Name,
		Description: contest.Description,
		StartTime:   contest.StartTime,
		EndTime:     contest.EndTime,
		Duration:    contest.Duration,
		Visible:     contest.Visible,
		CreatedBy:   contest.CreatedBy,
		CreatedAt:   contest.CreatedAt,
		UpdatedAt:   contest.UpdatedAt,
	}

	return response, nil
}

func (s *contestService) GetContestDetails(ctx context.Context, id string) (*domain.CreateContestResponse, error) {
	contest, err := s.contestRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	response := &domain.CreateContestResponse{
		Id:          contest.Id,
		Name:        contest.Name,
		Description: contest.Description,
		StartTime:   contest.StartTime,
		EndTime:     contest.EndTime,
		Duration:    contest.Duration,
		Visible:     contest.Visible,
		CreatedBy:   contest.CreatedBy,
		CreatedAt:   contest.CreatedAt,
		UpdatedAt:   contest.UpdatedAt,
	}

	return response, nil
}

func (s *contestService) DeleteContest(ctx context.Context, req *domain.DeleteContestRequest) error {
	return s.contestRepo.DeleteContest(ctx, req.Id)
}
