package services

import (
	"algoforces/internal/domain"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

type SubmissionService struct {
	submissionRepo domain.SubmissionRepository
	apiKey         string
	baseURL        string
	httpClient     *http.Client
}

func NewSubmissionService(submissionRepo domain.SubmissionRepository, apiKey string, baseURL string) domain.SubmissionUseCase {
	return &SubmissionService{
		submissionRepo: submissionRepo,
		apiKey:         apiKey,
		baseURL:        baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func GetLanguageID(language string) (int, error) {
	switch language {
	case "python":
		return domain.LanguagePython, nil
	case "cpp":
		return domain.LanguageCPP, nil
	case "java":
		return domain.LanguageJava, nil
	default:
		return 0, fmt.Errorf("unsupported language: %s", language)
	}
}

func (s *SubmissionService) CreateNewSubmission(ctx context.Context, req *domain.CreateSubmissionRequest) (*domain.CreateSubmissionResponse, error) {
	// Get all testCases for the problem
	testCases, err := s.submissionRepo.GetAllTestCasesForProblem(ctx, req.ProblemID)
	if err != nil {
		return nil, err
	}

	submissionID := uuid.New().String()
	timNow := time.Now()
	//Update the DB Status
	submission := &domain.Submission{
		UniqueID:       submissionID,
		UserId:         req.UserID,
		ContestID:      req.ContestID,
		ProblemID:      req.ProblemID,
		Code:           req.Code,
		Language:       req.Language,
		TotalTestCases: len(testCases),
		SubmittedAt:    timNow,
		Verdict:        string(domain.VerdictPending),
	}
	err = s.submissionRepo.CreateNewSubmission(ctx, submission)
	if err != nil {
		return nil, err
	}

	// Get the testCases
	var hiddenTestCases []domain.TestCase
	var visibleTestCases []domain.TestCase
	for _, testCase := range testCases {
		if testCase.IsHidden {
			hiddenTestCases = append(hiddenTestCases, testCase)
		} else {
			visibleTestCases = append(visibleTestCases, testCase)
		}
	}
	callbackURL := os.Getenv("APP_URL") + "/api/submission/callback"
	//Get the Language ID
	languageID, err := GetLanguageID(req.Language)

	if err != nil {
		return nil, err
	}

	for _, testCase := range visibleTestCases {
		//Make a submission to api

		submissionReqData := &domain.Judge0SubmissionRequest{
			SourceCode:     req.Code,
			LanguageID:     languageID,
			Stdin:          testCase.Input,
			CPUTimeLimit:   float64(req.TimeLimitInSecond),
			MemoryLimit:    req.MemoryLimitInMB * 1024,
			ExpectedOutput: testCase.ExpectedOutput,
			CallbackURL:    callbackURL,
		}

		// Make a submission
		var submissionResponseData *domain.Judge0SubmissionResponse
		submissionResponseData, err := s.CreateJudge0Submission(submissionReqData)
		if err != nil {
			return nil, err
		}

		// Make Token ---> [SubmissionID, testCaseID] mapping
		tokenMappingId := uuid.New().String()
		tokenMapping := &domain.SubmissionTestCaseMapping{
			UniqueID:           tokenMappingId,
			Token:              submissionResponseData.Token,
			SubmissionID:       submissionID,
			TestCaseID:         testCase.UniqueID,
			TestOrderPosition:  testCase.OrderPosition,
			TestCaseInput:      testCase.Input,
			TestExpectedOutput: testCase.ExpectedOutput,
			IsHidden:           false,
			Status:             string(domain.VerdictProcessing),
		}
		err = s.submissionRepo.CreateTokenMapping(ctx, tokenMapping)
		if err != nil {
			return nil, fmt.Errorf("failed to create token mapping: %w", err)
		}

	}

	for _, testCase := range hiddenTestCases {
		//Make a submission to api

		submissionReqData := &domain.Judge0SubmissionRequest{
			SourceCode:     req.Code,
			LanguageID:     languageID,
			Stdin:          testCase.Input,
			CPUTimeLimit:   float64(req.TimeLimitInSecond),
			MemoryLimit:    req.MemoryLimitInMB * 1024,
			ExpectedOutput: testCase.ExpectedOutput,
			CallbackURL:    callbackURL,
		}

		// Make a submission
		submissionResponseData, err := s.CreateJudge0Submission(submissionReqData)
		if err != nil {
			return nil, fmt.Errorf("error while creating the submission: %w", err)
		}

		// Make Token ---> [SubmissionID, testCaseID] mapping
		tokenMappingId := uuid.New().String()
		tokenMapping := &domain.SubmissionTestCaseMapping{
			UniqueID:           tokenMappingId,
			Token:              submissionResponseData.Token,
			SubmissionID:       submissionID,
			TestCaseID:         testCase.UniqueID,
			TestOrderPosition:  testCase.OrderPosition,
			TestCaseInput:      testCase.Input,
			TestExpectedOutput: testCase.ExpectedOutput,
			IsHidden:           true,
			Status:             string(domain.VerdictProcessing),
		}
		err = s.submissionRepo.CreateTokenMapping(ctx, tokenMapping)
		if err != nil {
			return nil, fmt.Errorf("failed to create token mapping: %w", err)
		}

	}

	// 8. Return response to user
	return &domain.CreateSubmissionResponse{
		UniqueID:    submissionID,
		UserID:      req.UserID,
		ContestID:   req.ContestID,
		ProblemID:   req.ProblemID,
		Language:    req.Language,
		Verdict:     string(domain.VerdictProcessing),
		SubmittedAt: timNow,
		Message:     "Submission queued successfully for judging",
	}, nil
}

func (s *SubmissionService) UpdateSubmissionStatus(ctx context.Context, submissionID, status string) error {
	return s.submissionRepo.UpdateSubmissionStatus(ctx, submissionID, status)
}
func (s *SubmissionService) GetSubmissionDetails(ctx context.Context, uniqueID string) (*domain.Submission, error) {
	return s.submissionRepo.GetSubmissionDetails(ctx, uniqueID)
}

func (s *SubmissionService) UpdateSubmissionResult(ctx context.Context, submissionID string, req *domain.UpdateSubmissionResultRequest) (*domain.UpdateSubmissionResultResponse, error) {
	result := &domain.Submission{
		Verdict:           req.Verdict,
		Score:             req.Score,
		TestCasesPassed:   req.TestCasesPassed,
		TotalTestCases:    req.TotalTestCases,
		ExecutionTimeInMS: req.ExecutionTimeInMS,
		MemoryUsedInKB:    req.MemoryUsedInKB,
		CompilationError:  req.CompilationError,
		RuntimeError:      req.RuntimeError,
		TestCaseResults:   req.TestCaseResults,
		FailedTestCase:    req.FailedTestCase,
		JudgeCompletedAt:  req.JudgeCompletedAt,
	}
	err := s.submissionRepo.UpdateSubmissionResult(ctx, submissionID, result)
	if err != nil {
		return nil, err
	}
	return &domain.UpdateSubmissionResultResponse{
		UniqueID:          submissionID,
		UserID:            result.UserId,
		ContestID:         result.ContestID,
		ProblemID:         result.ProblemID,
		Language:          result.Language,
		Verdict:           result.Verdict,
		SubmittedAt:       result.SubmittedAt,
		Message:           "Submission result updated successfully",
		Score:             result.Score,
		TestCasesPassed:   result.TestCasesPassed,
		TotalTestCases:    result.TotalTestCases,
		ExecutionTimeInMS: result.ExecutionTimeInMS,
		MemoryUsedInKB:    result.MemoryUsedInKB,
		CompilationError:  result.CompilationError,
		RuntimeError:      result.RuntimeError,
		TestCaseResults:   result.TestCaseResults,
		FailedTestCase:    result.FailedTestCase,
		JudgeCompletedAt:  result.JudgeCompletedAt,
	}, nil
}

func (s *SubmissionService) JudgeSubmissionCallback(ctx context.Context, req *domain.JudgeSubmissionCallbackRequest) error {

	// Get the test case by token
	var testMapping *domain.SubmissionTestCaseMapping
	testMapping, err := s.submissionRepo.GetMappingByToken(ctx, req.Token)
	if err != nil {
		return err
	}
	if testMapping == nil {
		return errors.New("test case not found")
	}
	submission, err := s.submissionRepo.GetSubmissionDetails(ctx, testMapping.SubmissionID)
	if err != nil {
		return err
	}
	// Update the submission results
	timeInMS := req.TimeInSeconds * 1000.0
	memoryInKB := req.MemoryInKB
	if submission.ExecutionTimeInMS < timeInMS {
		submission.ExecutionTimeInMS = timeInMS
	}
	if submission.MemoryUsedInKB < float64(memoryInKB) {
		submission.MemoryUsedInKB = float64(memoryInKB)
	}
	// Map Judge0 status to our verdict
	verdict := s.mapJudge0Status(req.Status.ID)
	// Create comprehensive test result using shared function
	testNum := testMapping.TestOrderPosition
	testResult := s.formatTestResult(testMapping, req, testNum, testMapping.IsHidden)
	submission.TestCaseResults = append(submission.TestCaseResults, testResult)

	if verdict == domain.VerdictAccepted {
		submission.TestCasesPassed++
	} else {
		submission.FailedTestCase = &testMapping.TestCaseID
		return s.updateSubmissionError(ctx, submission.UniqueID, verdict, submission.TestCasesPassed, submission.TotalTestCases, submission.ExecutionTimeInMS, submission.MemoryUsedInKB, submission.TestCaseResults)
	}
	if submission.TestCasesPassed == submission.TotalTestCases {
		return s.updateSubmissionSuccess(ctx, submission.UniqueID, verdict, submission.TestCasesPassed, submission.TotalTestCases, submission.ExecutionTimeInMS, submission.MemoryUsedInKB, submission.TestCaseResults)
	}
	_, err = s.UpdateSubmissionResult(ctx, submission.UniqueID, &domain.UpdateSubmissionResultRequest{
		Verdict:           string(verdict),
		Score:             submission.TestCasesPassed,
		TestCasesPassed:   submission.TestCasesPassed,
		TotalTestCases:    submission.TotalTestCases,
		ExecutionTimeInMS: submission.ExecutionTimeInMS,
		MemoryUsedInKB:    submission.MemoryUsedInKB,
	})
	if err != nil {
		return err
	}
	return nil
}

// mapJudge0Status maps Judge0 status to our verdict
func (s *SubmissionService) mapJudge0Status(statusID int) domain.VerdictStatus {
	switch statusID {
	case domain.StatusAccepted:
		return domain.VerdictAccepted
	case domain.StatusWrongAnswer:
		return domain.VerdictWrongAnswer
	case domain.StatusTimeLimitExceeded:
		return domain.VerdictTimeLimitExceeded
	case domain.StatusCompilationError:
		return domain.VerdictCompilationError
	case domain.StatusRuntimeError,
		domain.StatusRuntimeErrorOther,
		domain.StatusRuntimeErrorSIGFPE,
		domain.StatusRuntimeErrorSIGABRT,
		domain.StatusRuntimeErrorNZEC,
		domain.StatusRuntimeErrorOther2:
		return domain.VerdictRuntimeError
	default:
		return domain.VerdictSystemError
	}
}

// formatTestResult creates a comprehensive single-line test result for callback requests
func (s *SubmissionService) formatTestResult(testMapping *domain.SubmissionTestCaseMapping, status *domain.JudgeSubmissionCallbackRequest, testNum int, isHidden bool) string {
	// Get the verdict for this test
	verdict := s.mapJudge0Status(status.Status.ID)

	// Create comprehensive test result
	testResult := fmt.Sprintf("Test %d (%s): %s", testNum, testMapping.TestCaseID, verdict)

	// Add execution details for visible tests
	if !isHidden {
		testResult += fmt.Sprintf(" | Time: %.2fms", status.TimeInSeconds*1000)
		testResult += fmt.Sprintf(" | Memory: %dKB", status.MemoryInKB)

		// Add input/output details (only for visible tests)
		if !testMapping.IsHidden {
			testResult += fmt.Sprintf(" | Input: %s | Expected: %s", testMapping.TestCaseInput, testMapping.TestExpectedOutput)
			if status.Stdout != "" {
				testResult += fmt.Sprintf(" | Got: %s", status.Stdout)
			}
		}

		// Add error details if any
		if status.Stderr != "" {
			testResult += fmt.Sprintf(" | Stderr: %s", status.Stderr)
		}
		if status.CompileOutput != "" {
			testResult += fmt.Sprintf(" | Compile: %s", status.CompileOutput)
		}
		if status.Message != "" {
			testResult += fmt.Sprintf(" | Message: %s", status.Message)
		}
	}

	return testResult
}

// updateSubmissionSuccess updates the submission with success result
func (s *SubmissionService) updateSubmissionSuccess(ctx context.Context, submissionID string,
	verdict domain.VerdictStatus, passed, total int,
	maxTime float64, maxMemory float64, results []string) error {

	now := time.Now()

	// Update submission result
	err := s.submissionRepo.UpdateSubmissionResult(ctx, submissionID, &domain.Submission{
		Verdict:           string(verdict),
		Score:             passed,
		TestCasesPassed:   passed,
		TotalTestCases:    total,
		ExecutionTimeInMS: maxTime,
		MemoryUsedInKB:    maxMemory,
		TestCaseResults:   results,
		FailedTestCase:    nil, // No failed test case for success
		JudgeCompletedAt:  &now,
	})
	if err != nil {
		return fmt.Errorf("failed to update submission result: %w", err)
	}

	log.Printf("Submission %s completed with verdict: %s (%d/%d tests passed)",
		submissionID, verdict, passed, total)

	return nil
}

// updateSubmissionError updates submission with an error status
func (s *SubmissionService) updateSubmissionError(ctx context.Context, submissionID string,
	verdict domain.VerdictStatus, passed, total int,
	maxTime float64, maxMemory float64, testResults []string) error {

	now := time.Now()

	// Update submission result with error details
	err := s.submissionRepo.UpdateSubmissionResult(ctx, submissionID, &domain.Submission{
		Verdict:           string(verdict),
		Score:             0, // No score for failed submissions
		TestCasesPassed:   passed,
		TotalTestCases:    total,
		ExecutionTimeInMS: maxTime,
		MemoryUsedInKB:    maxMemory,
		TestCaseResults:   testResults,
		FailedTestCase:    nil,
		JudgeCompletedAt:  &now,
	})
	if err != nil {
		log.Printf("Failed to update error status: %v", err)
		return fmt.Errorf("failed to update submission result: %w", err)
	}

	log.Printf("Submission %s completed with verdict: %s (%d/%d tests passed)",
		submissionID, verdict, passed, total)

	return nil
}

func (c *SubmissionService) CreateJudge0Submission(req *domain.Judge0SubmissionRequest) (*domain.Judge0SubmissionResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	httpReq, err := http.NewRequest("POST", c.baseURL+"/submissions?base64_encoded=false&wait=false", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("X-Auth-Token", c.apiKey)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var submissionResp domain.Judge0SubmissionResponse
	if err := json.Unmarshal(body, &submissionResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &submissionResp, nil
}

func (s *SubmissionService) GetJudge0SubmissionStatus(token string) (*domain.Judge0SubmissionStatus, error) {
	httpReq, err := http.NewRequest("GET", s.baseURL+"/submissions/"+token+"?base64_encoded=false", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if s.apiKey != "" {
		httpReq.Header.Set("X-Auth-Token", s.apiKey)
	}

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var statusResp domain.Judge0SubmissionStatus
	if err := json.Unmarshal(body, &statusResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &statusResp, nil
}
