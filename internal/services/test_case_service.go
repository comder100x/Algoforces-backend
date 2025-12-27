package services

import (
	"algoforces/internal/domain"
	"context"
	"errors"

	"github.com/google/uuid"
)

type TestCaseService struct {
	testCaseRepo domain.TestCaseRepository
}

func NewTestCaseService(testCaseRepo domain.TestCaseRepository) domain.TestCaseUseCase {
	return &TestCaseService{
		testCaseRepo: testCaseRepo,
	}
}

func (s *TestCaseService) CreateNewTestCase(ctx context.Context, req *domain.CreateTestCaseRequest) (*domain.CreateTestCaseResponse, error) {
	testCase := &domain.TestCase{
		UniqueID:       uuid.New().String(),
		ProblemID:      req.ProblemID,
		Input:          req.Input,
		ExpectedOutput: req.ExpectedOutput,
		IsHidden:       req.IsHidden,
		OrderPosition:  req.OrderPosition,
	}

	err := s.testCaseRepo.CreateTestCase(ctx, testCase)
	if err != nil {
		return nil, err
	}

	return &domain.CreateTestCaseResponse{
		UniqueID:       testCase.UniqueID,
		ProblemID:      testCase.ProblemID,
		Input:          testCase.Input,
		ExpectedOutput: testCase.ExpectedOutput,
		IsHidden:       testCase.IsHidden,
		OrderPosition:  testCase.OrderPosition,
	}, nil
}

func (s *TestCaseService) UpdateSingleTestCase(ctx context.Context, req *domain.UpdateTestCaseRequest) (*domain.UpdateTestCaseResponse, error) {
	testCase, err := s.testCaseRepo.GetTestCaseByUniqueID(ctx, req.UniqueID)

	if err != nil {
		return nil, errors.New("Error  in getting the test Case")
	}

	testCase.Input = req.Input
	testCase.ExpectedOutput = req.ExpectedOutput
	testCase.IsHidden = req.IsHidden
	testCase.OrderPosition = req.OrderPosition

	err = s.testCaseRepo.UpdateTestCase(ctx, testCase)
	if err != nil {
		return nil, err
	}

	return &domain.UpdateTestCaseResponse{
		UniqueID:       testCase.UniqueID,
		ProblemID:      testCase.ProblemID,
		Input:          testCase.Input,
		ExpectedOutput: testCase.ExpectedOutput,
		IsHidden:       testCase.IsHidden,
		OrderPosition:  testCase.OrderPosition,
	}, nil
}

func (s *TestCaseService) DeleteSingleTestCase(ctx context.Context, uniqueId string) error {
	err := s.testCaseRepo.DeleteTestCase(ctx, uniqueId)
	if err != nil {
		return err
	}
	return nil
}

func (s *TestCaseService) GetAllTestCasesForProblem(ctx context.Context, problemId string) ([]*domain.TestCase, error) {
	testCases, err := s.testCaseRepo.GetTestCasesByProblemID(ctx, problemId)
	if err != nil {
		return nil, err
	}
	return testCases, nil
}

func (s *TestCaseService) GetTestCaseDetails(ctx context.Context, uniqueId string) (*domain.TestCase, error) {
	testCase, err := s.testCaseRepo.GetTestCaseByUniqueID(ctx, uniqueId)
	if err != nil {
		return nil, err
	}
	return testCase, nil
}

func (s *TestCaseService) UploadTestCasesInBulk(ctx context.Context, req *domain.BulkTestCaseUploadRequest) (*domain.BulkTestCaseUploadResponse, error) {

	response := &domain.BulkTestCaseUploadResponse{
		CreatedTestCases: []domain.CreateTestCaseResponse{},
	}
	for _, testCase := range req.TestCases {
		testcase := &domain.TestCase{
			UniqueID:       uuid.New().String(),
			ProblemID:      testCase.ProblemID,
			Input:          testCase.Input,
			ExpectedOutput: testCase.ExpectedOutput,
			IsHidden:       testCase.IsHidden,
			OrderPosition:  testCase.OrderPosition,
		}

		err := s.testCaseRepo.CreateTestCase(ctx, testcase)
		if err != nil {
			return nil, err
		}

		response.CreatedTestCases = append(response.CreatedTestCases, domain.CreateTestCaseResponse{
			UniqueID:       testcase.UniqueID,
			ProblemID:      testcase.ProblemID,
			Input:          testcase.Input,
			ExpectedOutput: testcase.ExpectedOutput,
			IsHidden:       testcase.IsHidden,
			OrderPosition:  testcase.OrderPosition,
		})
	}

	return response, nil

}
