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

type contestProblemsService struct {
	contestProblemsRepo domain.ContestProblemsRepository
	contestRepo         domain.ContestRepository
	problemRepo         domain.ProblemRepository
	contestRegisterRepo domain.ContestRegisterRepository
}

func NewContestProblemsService(
	contestProblemsRepo domain.ContestProblemsRepository,
	contestRepo domain.ContestRepository,
	problemRepo domain.ProblemRepository,
	contestRegisterRepo domain.ContestRegisterRepository,
) domain.ContestProblemsUseCase {
	return &contestProblemsService{
		contestProblemsRepo: contestProblemsRepo,
		contestRepo:         contestRepo,
		problemRepo:         problemRepo,
		contestRegisterRepo: contestRegisterRepo,
	}
}

// contestLookupErr maps a missing contest row to the not-found sentinel;
// any other lookup failure keeps its cause for the handler to log.
func contestLookupErr(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.ErrContestNotFound
	}
	return fmt.Errorf("get contest by id: %w", err)
}

// problemLookupErr maps a missing problem row to a not-found error;
// any other lookup failure keeps its cause.
func problemLookupErr(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("problem not found")
	}
	return fmt.Errorf("get problem by id: %w", err)
}

func (s *contestProblemsService) CreateContestProblem(
	ctx context.Context,
	req *domain.CreateContestProblemRequest,
) (*domain.CreateContestProblemResponse, error) {
	if req == nil {
		return nil, errors.New("create contest problem request is required")
	}

	// Validate contest exists
	_, err := s.contestRepo.GetByID(ctx, req.ContestID)
	if err != nil {
		return nil, contestLookupErr(err)
	}

	// Validate problem exists
	_, err = s.problemRepo.GetProblemByID(ctx, req.ProblemID)
	if err != nil {
		return nil, problemLookupErr(err)
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
		return nil, fmt.Errorf("create contest problem: %w", err)
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

func (s *contestProblemsService) BulkCreateContestProblems(
	ctx context.Context,
	req *domain.BulkCreateContestProblemRequest,
) (*domain.BulkCreateContestProblemResponse, error) {
	if req == nil {
		return nil, errors.New("bulk create contest problems request is required")
	}

	// Validate contest exists
	_, err := s.contestRepo.GetByID(ctx, req.ContestID)
	if err != nil {
		return nil, contestLookupErr(err)
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
			response.Errors = append(
				response.Errors,
				fmt.Sprintf(
					"problem at index %d (problem_id: %s) failed: %s",
					i,
					item.ProblemID,
					problemLookupErr(err),
				),
			)
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
			response.Errors = append(
				response.Errors,
				fmt.Sprintf(
					"problem at index %d (problem_id: %s) failed: %s",
					i,
					item.ProblemID,
					fmt.Errorf("create contest problem: %w", err),
				),
			)
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

func (s *contestProblemsService) GetContestProblem(
	ctx context.Context,
	req *domain.GetContestProblemRequest,
) (*domain.GetContestProblemResponse, error) {
	if req == nil {
		return nil, errors.New("get contest problem request is required")
	}

	contestProblem, err := s.contestProblemsRepo.GetContestProblemByID(ctx, req.UniqueID)
	if err != nil {
		return nil, fmt.Errorf("get contest problem by id: %w", err)
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

func (s *contestProblemsService) GetContestProblems(
	ctx context.Context,
	req *domain.GetContestProblemsRequest,
) (*domain.GetContestProblemsResponse, error) {
	if req == nil {
		return nil, errors.New("get contest problems request is required")
	}

	// Validate contest exists
	_, err := s.contestRepo.GetByID(ctx, req.ContestID)
	if err != nil {
		return nil, contestLookupErr(err)
	}

	problems, err := s.contestProblemsRepo.GetContestProblemsByContestIDWithDetails(ctx, req.ContestID)
	if err != nil {
		return nil, fmt.Errorf("get contest problems by contest id: %w", err)
	}
	if problems == nil {
		problems = []domain.ContestProblemDetail{}
	}

	response := &domain.GetContestProblemsResponse{
		Problems: problems,
	}

	return response, nil
}

func (s *contestProblemsService) GetContestProblemsForParticipant(
	ctx context.Context,
	contestID, userID, role string,
) (*domain.GetContestProblemsResponse, error) {
	// Validate contest exists.
	contest, err := s.contestRepo.GetByID(ctx, contestID)
	if err != nil {
		return nil, contestLookupErr(err)
	}

	// A missing row is not an error: it means the caller is unregistered.
	registration, err := s.contestRegisterRepo.GetRegistrationByUserAndContest(ctx, userID, contestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			registration = nil
		} else {
			return nil, fmt.Errorf("get registration by user and contest: %w", err)
		}
	}

	// Single access decision point; the live window is evaluated inside the
	// helper directly from the Contest fields (StartTime, EndTime, IsActive).
	if !CanViewContestProblemsForParticipant(
		contest,
		userID,
		role,
		registration,
		time.Now(),
	) {
		return nil, domain.ErrContestProblemsForbidden
	}

	// No test-case fields are carried, so no test-case data can leak here.
	problems, err := s.contestProblemsRepo.GetContestProblemsByContestIDWithDetails(ctx, contestID)
	if err != nil {
		return nil, fmt.Errorf("get contest problems by contest id: %w", err)
	}
	if problems == nil {
		problems = []domain.ContestProblemDetail{}
	}

	return &domain.GetContestProblemsResponse{
		Problems: problems,
	}, nil
}

func (s *contestProblemsService) UpdateContestProblem(
	ctx context.Context,
	req *domain.UpdateContestProblemRequest,
) (*domain.UpdateContestProblemResponse, error) {
	if req == nil {
		return nil, errors.New("update contest problem request is required")
	}

	contestProblem, err := s.contestProblemsRepo.GetContestProblemByID(ctx, req.UniqueID)
	if err != nil {
		return nil, fmt.Errorf("get contest problem by id: %w", err)
	}

	// Update fields if provided
	if req.ContestID != "" {
		// Validate new contest exists
		_, err := s.contestRepo.GetByID(ctx, req.ContestID)
		if err != nil {
			return nil, contestLookupErr(err)
		}
		contestProblem.ContestID = req.ContestID
	}

	if req.ProblemID != "" {
		// Validate new problem exists
		_, err := s.problemRepo.GetProblemByID(ctx, req.ProblemID)
		if err != nil {
			return nil, problemLookupErr(err)
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
		return nil, fmt.Errorf("update contest problem: %w", err)
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

func (s *contestProblemsService) DeleteContestProblem(
	ctx context.Context,
	req *domain.DeleteContestProblemRequest,
) (*domain.DeleteContestProblemResponse, error) {
	if req == nil {
		return nil, errors.New("delete contest problem request is required")
	}

	// Check if exists
	_, err := s.contestProblemsRepo.GetContestProblemByID(ctx, req.UniqueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("contest problem not found")
		}
		return nil, fmt.Errorf("get contest problem by id: %w", err)
	}

	err = s.contestProblemsRepo.DeleteContestProblem(ctx, req.UniqueID)
	if err != nil {
		return nil, fmt.Errorf("delete contest problem: %w", err)
	}

	response := &domain.DeleteContestProblemResponse{
		Message: "Contest problem deleted successfully",
	}

	return response, nil
}
