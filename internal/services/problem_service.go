package services

import (
	"algoforces/internal/domain"
	"context"
	"errors"

	"github.com/google/uuid"
)

type problemService struct {
	problemRepo domain.ProblemRepository
	userRepo    domain.UserRepository
}

func NewProblemService(problemRepo domain.ProblemRepository, userRepo domain.UserRepository) domain.ProblemUseCase {
	return &problemService{
		problemRepo: problemRepo,
		userRepo:    userRepo,
	}
}

func (s *problemService) CreateProblem(ctx context.Context, req *domain.ProblemCreationRequest, createdBy string) (*domain.ProblemCreationResponse, error) {
	// Verify user exists and has permission
	user, err := s.userRepo.GetByID(ctx, createdBy)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if user.Role != "admin" && user.Role != "problem_setter" {
		return nil, errors.New("user does not have permission to create problems")
	}

	// Set defaults if not provided
	if req.TimeLimitInSeconds == 0 {
		req.TimeLimitInSeconds = 1 // default 1 second
	}
	if req.MemoryLimitInMB == 0 {
		req.MemoryLimitInMB = 256 // default 256 MB
	}

	problem := &domain.Problem{
		UniqueID:           uuid.New().String(),
		Title:              req.Title,
		Statement:          req.Statement,
		Difficulty:         req.Difficulty,
		TimeLimitInSeconds: req.TimeLimitInSeconds,
		MemoryLimitInMB:    req.MemoryLimitInMB,
		CreatedBy:          createdBy,
	}

	err = s.problemRepo.CreateProblem(ctx, problem)
	if err != nil {
		return nil, err
	}

	return &domain.ProblemCreationResponse{
		UniqueID:           problem.UniqueID,
		Title:              problem.Title,
		Statement:          problem.Statement,
		Difficulty:         problem.Difficulty,
		TimeLimitInSeconds: problem.TimeLimitInSeconds,
		MemoryLimitInMB:    problem.MemoryLimitInMB,
		CreatedBy:          problem.CreatedBy,
		CreatedAt:          problem.CreatedAt,
		UpdatedAt:          problem.UpdatedAt,
	}, nil
}

func (s *problemService) CreateProblemsInBulk(ctx context.Context, req *domain.BulkProblemCreationRequest, createdBy string) (*domain.BulkProblemCreationResponse, error) {
	// Verify user exists and has permission
	user, err := s.userRepo.GetByID(ctx, createdBy)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if user.Role != "admin" && user.Role != "problem_setter" {
		return nil, errors.New("user does not have permission to create problems")
	}

	response := &domain.BulkProblemCreationResponse{
		Problems: []domain.ProblemCreationResponse{},
		Errors:   []string{},
	}

	for i, problemReq := range req.Problems {
		// Set defaults if not provided
		if problemReq.TimeLimitInSeconds == 0 {
			problemReq.TimeLimitInSeconds = 1
		}
		if problemReq.MemoryLimitInMB == 0 {
			problemReq.MemoryLimitInMB = 256
		}

		problem := &domain.Problem{
			UniqueID:           uuid.New().String(),
			Title:              problemReq.Title,
			Statement:          problemReq.Statement,
			Difficulty:         problemReq.Difficulty,
			TimeLimitInSeconds: problemReq.TimeLimitInSeconds,
			MemoryLimitInMB:    problemReq.MemoryLimitInMB,
			CreatedBy:          createdBy,
		}

		err = s.problemRepo.CreateProblem(ctx, problem)
		if err != nil {
			response.FailedCount++
			response.Errors = append(response.Errors, "problem '"+problemReq.Title+"' at index "+string(rune(i+'0'))+" failed: "+err.Error())
			continue
		}

		response.SuccessCount++
		response.Problems = append(response.Problems, domain.ProblemCreationResponse{
			UniqueID:           problem.UniqueID,
			Title:              problem.Title,
			Statement:          problem.Statement,
			Difficulty:         problem.Difficulty,
			TimeLimitInSeconds: problem.TimeLimitInSeconds,
			MemoryLimitInMB:    problem.MemoryLimitInMB,
			CreatedBy:          problem.CreatedBy,
			CreatedAt:          problem.CreatedAt,
			UpdatedAt:          problem.UpdatedAt,
		})
	}

	return response, nil
}

func (s *problemService) GetProblemByID(ctx context.Context, id string) (*domain.ProblemCreationResponse, error) {
	problem, err := s.problemRepo.GetProblemByID(ctx, id)
	if err != nil {
		return nil, errors.New("problem not found")
	}

	return &domain.ProblemCreationResponse{
		UniqueID:           problem.UniqueID,
		Title:              problem.Title,
		Statement:          problem.Statement,
		Difficulty:         problem.Difficulty,
		TimeLimitInSeconds: problem.TimeLimitInSeconds,
		MemoryLimitInMB:    problem.MemoryLimitInMB,
		CreatedBy:          problem.CreatedBy,
		CreatedAt:          problem.CreatedAt,
		UpdatedAt:          problem.UpdatedAt,
	}, nil
}

func (s *problemService) UpdateProblem(ctx context.Context, req *domain.ProblemUpdateRequest, userID string) (*domain.ProblemUpdateResponse, error) {
	// Verify user exists and has permission
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if user.Role != "admin" && user.Role != "problem_setter" {
		return nil, errors.New("user does not have permission to update problems")
	}

	// Get existing problem
	existingProblem, err := s.problemRepo.GetProblemByID(ctx, req.UniqueID)
	if err != nil {
		return nil, errors.New("problem not found")
	}

	// Check if user is the creator or admin
	if user.Role != "admin" && existingProblem.CreatedBy != userID {
		return nil, errors.New("user can only update their own problems")
	}

	// Set defaults if not provided
	if req.TimeLimitInSeconds == 0 {
		req.TimeLimitInSeconds = existingProblem.TimeLimitInSeconds
	}
	if req.MemoryLimitInMB == 0 {
		req.MemoryLimitInMB = existingProblem.MemoryLimitInMB
	}

	// Update problem
	existingProblem.Title = req.Title
	existingProblem.Statement = req.Statement
	existingProblem.Difficulty = req.Difficulty
	existingProblem.TimeLimitInSeconds = req.TimeLimitInSeconds
	existingProblem.MemoryLimitInMB = req.MemoryLimitInMB

	err = s.problemRepo.UpdateProblem(ctx, existingProblem)
	if err != nil {
		return nil, err
	}

	return &domain.ProblemUpdateResponse{
		UniqueID:           existingProblem.UniqueID,
		Title:              existingProblem.Title,
		Statement:          existingProblem.Statement,
		Difficulty:         existingProblem.Difficulty,
		TimeLimitInSeconds: existingProblem.TimeLimitInSeconds,
		MemoryLimitInMB:    existingProblem.MemoryLimitInMB,
		CreatedBy:          existingProblem.CreatedBy,
		CreatedAt:          existingProblem.CreatedAt,
		UpdatedAt:          existingProblem.UpdatedAt,
	}, nil
}

func (s *problemService) DeleteProblem(ctx context.Context, id string, userID string) error {
	// Verify user exists and has permission
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.New("user not found")
	}

	if user.Role != "admin" && user.Role != "problem_setter" {
		return errors.New("user does not have permission to delete problems")
	}

	// Get existing problem
	existingProblem, err := s.problemRepo.GetProblemByID(ctx, id)
	if err != nil {
		return errors.New("problem not found")
	}

	// Check if user is the creator or admin
	if user.Role != "admin" && existingProblem.CreatedBy != userID {
		return errors.New("user can only delete their own problems")
	}

	return s.problemRepo.DeleteProblem(ctx, id)
}

func (s *problemService) GetAllProblems(ctx context.Context) ([]domain.ProblemCreationResponse, error) {
	problems, err := s.problemRepo.GetAllProblems(ctx)
	if err != nil {
		return nil, err
	}

	var problemResponses []domain.ProblemCreationResponse
	for _, problem := range problems {
		problemResponses = append(problemResponses, domain.ProblemCreationResponse{
			UniqueID:           problem.UniqueID,
			Title:              problem.Title,
			Statement:          problem.Statement,
			Difficulty:         problem.Difficulty,
			TimeLimitInSeconds: problem.TimeLimitInSeconds,
			MemoryLimitInMB:    problem.MemoryLimitInMB,
			CreatedBy:          problem.CreatedBy,
			CreatedAt:          problem.CreatedAt,
			UpdatedAt:          problem.UpdatedAt,
		})
	}

	return problemResponses, nil
}
