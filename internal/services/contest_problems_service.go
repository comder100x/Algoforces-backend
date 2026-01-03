package services

import (
	"algoforces/internal/domain"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type contestProblemsService struct {
	contestProblemsRepo domain.ContestProblemsRepository
	contestRepo         domain.ContestRepository
	problemRepo         domain.ProblemRepository
}

func NewContestProblemsService(
	contestProblemsRepo domain.ContestProblemsRepository,
	contestRepo domain.ContestRepository,
	problemRepo domain.ProblemRepository,
) domain.ContestProblemsUseCase {
	return &contestProblemsService{
		contestProblemsRepo: contestProblemsRepo,
		contestRepo:         contestRepo,
		problemRepo:         problemRepo,
	}
}

func (s *contestProblemsService) CreateContestProblem(ctx context.Context, req *domain.CreateContestProblemRequest) (*domain.CreateContestProblemResponse, error) {
	// Validate contest exists
	_, err := s.contestRepo.GetByID(ctx, req.ContestID)
	if err != nil {
		return nil, errors.New("contest not found")
	}

	// Validate problem exists
	_, err = s.problemRepo.GetProblemByID(ctx, req.ProblemID)
	if err != nil {
		return nil, errors.New("problem not found")
	}

	contestProblem := &domain.ContestProblems{
		UniqueID:      uuid.New().String(),
		ContestID:     req.ContestID,
		ProblemID:     req.ProblemID,
		OrderPosition: req.OrderPosition,
		MaxPoints:     req.MaxPoints,
	}

	err = s.contestProblemsRepo.CreateContestProblem(ctx, contestProblem)
	if err != nil {
		return nil, err
	}

	response := &domain.CreateContestProblemResponse{
		UniqueID:      contestProblem.UniqueID,
		ContestID:     contestProblem.ContestID,
		ProblemID:     contestProblem.ProblemID,
		OrderPosition: contestProblem.OrderPosition,
		MaxPoints:     contestProblem.MaxPoints,
		CreatedAt:     contestProblem.CreatedAt,
		UpdatedAt:     contestProblem.UpdatedAt,
	}

	return response, nil
}

func (s *contestProblemsService) BulkCreateContestProblems(ctx context.Context, req *domain.BulkCreateContestProblemRequest) (*domain.BulkCreateContestProblemResponse, error) {
	// Validate contest exists
	_, err := s.contestRepo.GetByID(ctx, req.ContestID)
	if err != nil {
		return nil, errors.New("contest not found")
	}

	response := &domain.BulkCreateContestProblemResponse{
		CreatedProblems: []domain.CreateContestProblemResponse{},
		Errors:          []string{},
	}

	for i, item := range req.ContestProblems {
		// Validate problem exists
		_, err := s.problemRepo.GetProblemByID(ctx, item.ProblemID)
		if err != nil {
			response.FailedCount++
			response.Errors = append(response.Errors,
				fmt.Sprintf("problem at index %d (problem_id: %s) failed: problem not found", i, item.ProblemID))
			continue
		}

		contestProblem := &domain.ContestProblems{
			UniqueID:      uuid.New().String(),
			ContestID:     req.ContestID,
			ProblemID:     item.ProblemID,
			OrderPosition: item.OrderPosition,
			MaxPoints:     item.MaxPoints,
		}

		err = s.contestProblemsRepo.CreateContestProblem(ctx, contestProblem)
		if err != nil {
			response.FailedCount++
			response.Errors = append(response.Errors,
				fmt.Sprintf("problem at index %d (problem_id: %s) failed: %s", i, item.ProblemID, err.Error()))
			continue
		}

		response.SuccessCount++
		response.CreatedProblems = append(response.CreatedProblems, domain.CreateContestProblemResponse{
			UniqueID:      contestProblem.UniqueID,
			ContestID:     contestProblem.ContestID,
			ProblemID:     contestProblem.ProblemID,
			OrderPosition: contestProblem.OrderPosition,
			MaxPoints:     contestProblem.MaxPoints,
			CreatedAt:     contestProblem.CreatedAt,
			UpdatedAt:     contestProblem.UpdatedAt,
		})
	}

	return response, nil
}

func (s *contestProblemsService) GetContestProblem(ctx context.Context, req *domain.GetContestProblemRequest) (*domain.GetContestProblemResponse, error) {
	contestProblem, err := s.contestProblemsRepo.GetContestProblemByID(ctx, req.UniqueID)
	if err != nil {
		return nil, err
	}

	response := &domain.GetContestProblemResponse{
		UniqueID:      contestProblem.UniqueID,
		ContestID:     contestProblem.ContestID,
		ProblemID:     contestProblem.ProblemID,
		OrderPosition: contestProblem.OrderPosition,
		MaxPoints:     contestProblem.MaxPoints,
		CreatedAt:     contestProblem.CreatedAt,
		UpdatedAt:     contestProblem.UpdatedAt,
	}

	return response, nil
}

func (s *contestProblemsService) GetContestProblems(ctx context.Context, req *domain.GetContestProblemsRequest) (*domain.GetContestProblemsResponse, error) {
	// Validate contest exists
	_, err := s.contestRepo.GetByID(ctx, req.ContestID)
	if err != nil {
		return nil, errors.New("contest not found")
	}

	problems, err := s.contestProblemsRepo.GetContestProblemsByContestIDWithDetails(ctx, req.ContestID)
	if err != nil {
		return nil, err
	}

	response := &domain.GetContestProblemsResponse{
		Problems: problems,
	}

	return response, nil
}

func (s *contestProblemsService) UpdateContestProblem(ctx context.Context, req *domain.UpdateContestProblemRequest) (*domain.UpdateContestProblemResponse, error) {
	contestProblem, err := s.contestProblemsRepo.GetContestProblemByID(ctx, req.UniqueID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.ContestID != "" {
		// Validate new contest exists
		_, err := s.contestRepo.GetByID(ctx, req.ContestID)
		if err != nil {
			return nil, errors.New("contest not found")
		}
		contestProblem.ContestID = req.ContestID
	}

	if req.ProblemID != "" {
		// Validate new problem exists
		_, err := s.problemRepo.GetProblemByID(ctx, req.ProblemID)
		if err != nil {
			return nil, errors.New("problem not found")
		}
		contestProblem.ProblemID = req.ProblemID
	}

	if req.OrderPosition != nil {
		contestProblem.OrderPosition = *req.OrderPosition
	}

	if req.MaxPoints != nil {
		contestProblem.MaxPoints = *req.MaxPoints
	}

	err = s.contestProblemsRepo.UpdateContestProblem(ctx, contestProblem)
	if err != nil {
		return nil, err
	}

	response := &domain.UpdateContestProblemResponse{
		UniqueID:      contestProblem.UniqueID,
		ContestID:     contestProblem.ContestID,
		ProblemID:     contestProblem.ProblemID,
		OrderPosition: contestProblem.OrderPosition,
		MaxPoints:     contestProblem.MaxPoints,
		CreatedAt:     contestProblem.CreatedAt,
		UpdatedAt:     contestProblem.UpdatedAt,
	}

	return response, nil
}

func (s *contestProblemsService) DeleteContestProblem(ctx context.Context, req *domain.DeleteContestProblemRequest) (*domain.DeleteContestProblemResponse, error) {
	// Check if exists
	_, err := s.contestProblemsRepo.GetContestProblemByID(ctx, req.UniqueID)
	if err != nil {
		return nil, errors.New("contest problem not found")
	}

	err = s.contestProblemsRepo.DeleteContestProblem(ctx, req.UniqueID)
	if err != nil {
		return nil, err
	}

	response := &domain.DeleteContestProblemResponse{
		Message: "Contest problem deleted successfully",
	}

	return response, nil
}
