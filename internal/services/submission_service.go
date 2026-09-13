package services

import (
	"algoforces/internal/conf"
	"algoforces/internal/domain"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
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

func submissionScore(passed, total, maxPoints int) int {
	if total > 0 && passed == total {
		return maxPoints
	}
	return 0
}

func (s *SubmissionService) CreateNewSubmission(ctx context.Context, req *domain.CreateSubmissionRequest) (*domain.CreateSubmissionResponse, error) {
	// Get all testCases for the problem
	testCases, err := s.submissionRepo.GetAllTestCasesForProblem(ctx, req.ProblemID)
	if err != nil {
		return nil, err
	}

	maxPoints, err := s.submissionRepo.GetContestProblemMaxPoints(ctx, req.ContestID, req.ProblemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get contest problem max points: %w", err)
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
		MaxPoints:      maxPoints,
		SubmittedAt:    timNow,
		Verdict:        string(domain.VerdictPending),
	}
	err = s.submissionRepo.CreateNewSubmission(ctx, submission)
	if err != nil {
		return nil, err
	}

	callbackBase := strings.TrimRight(conf.APP_URL, "/")
	if callbackBase == "" {
		callbackBase = "http://localhost:8080"
	}
	callbackURL := callbackBase + "/api/submission/callback"
	//Get the Language ID
	languageID, err := GetLanguageID(req.Language)
	if err != nil {
		return nil, err
	}

	// Build batch submission request for all test cases
	batchReq := &domain.Judge0BatchSubmissionRequest{
		Submissions: make([]domain.Judge0SubmissionRequest, 0, len(testCases)),
	}

	for _, testCase := range testCases {
		batchReq.Submissions = append(batchReq.Submissions, domain.Judge0SubmissionRequest{
			SourceCode:     req.Code,
			LanguageID:     languageID,
			Stdin:          testCase.Input,
			CPUTimeLimit:   float64(req.TimeLimitInSecond),
			MemoryLimit:    req.MemoryLimitInMB * 1024,
			ExpectedOutput: testCase.ExpectedOutput,
			CallbackURL:    callbackURL,
		})
	}

	// Make batch submission to Judge0
	batchResp, err := s.CreateJudge0BatchSubmission(batchReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create batch submission: %w", err)
	}

	// Create token mappings for each test case
	for i, testCase := range testCases {
		tokenMappingId := uuid.New().String()
		tokenMapping := &domain.SubmissionTestCaseMapping{
			UniqueID:           tokenMappingId,
			Token:              batchResp[i].Token,
			SubmissionID:       submissionID,
			TestCaseID:         testCase.UniqueID,
			TestOrderPosition:  testCase.OrderPosition,
			TestCaseInput:      testCase.Input,
			TestExpectedOutput: testCase.ExpectedOutput,
			IsHidden:           testCase.IsHidden,
			Status:             string(domain.VerdictProcessing),
		}
		err = s.submissionRepo.CreateTokenMapping(ctx, tokenMapping)
		if err != nil {
			return nil, fmt.Errorf("failed to create token mapping: %w", err)
		}
	}

	// Return response to user
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
func (s *SubmissionService) GetSubmissionDetails(ctx context.Context, uniqueID string, includeHidden bool) (*domain.SubmissionDetailsResponse, error) {
	submission, err := s.submissionRepo.GetSubmissionDetails(ctx, uniqueID)
	if err != nil {
		return nil, err
	}

	results := s.parseSubmissionResults(submission.TestCaseResults, includeHidden)
	var failedTestCase *domain.SubmissionTestCaseResult
	if submission.FailedTestCase != nil && *submission.FailedTestCase != "" {
		parsedFailed, err := s.parseStoredTestResult(*submission.FailedTestCase)
		if err == nil && (includeHidden || !parsedFailed.IsHidden) {
			if !includeHidden {
				parsedFailed = s.sanitizeTestResult(parsedFailed)
			}
			failedTestCase = &parsedFailed
		}
	}

	tokenList := submission.TokenList
	if !includeHidden {
		tokenList = nil
	}

	return &domain.SubmissionDetailsResponse{
		UniqueID:          submission.UniqueID,
		UserID:            submission.UserId,
		ContestID:         submission.ContestID,
		ProblemID:         submission.ProblemID,
		Code:              submission.Code,
		Language:          submission.Language,
		SubmittedAt:       submission.SubmittedAt,
		CreatedAt:         submission.CreatedAt,
		UpdatedAt:         submission.UpdatedAt,
		QueuedAt:          submission.QueuedAt,
		TokenList:         tokenList,
		Verdict:           submission.Verdict,
		Score:             submission.Score,
		TestCasesPassed:   submission.TestCasesPassed,
		TotalTestCases:    submission.TotalTestCases,
		ExecutionTimeInMS: submission.ExecutionTimeInMS,
		MemoryUsedInKB:    submission.MemoryUsedInKB,
		CompilationError:  submission.CompilationError,
		RuntimeError:      submission.RuntimeError,
		TestCaseResults:   results,
		FailedTestCase:    failedTestCase,
		JudgeCompletedAt:  submission.JudgeCompletedAt,
		MaxPoints:         submission.MaxPoints,
	}, nil
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
		TokenList:         req.TokenList,
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
		TokenList:         result.TokenList,
	}, nil
}

// JudgeSubmissionCallback is the callback function for the submission
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
	// Parse time string from Judge0 (e.g., "0.002")
	timeInSeconds, _ := strconv.ParseFloat(req.Time, 64)
	timeInMS := timeInSeconds * 1000.0
	memoryInKB := req.Memory
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
	testResult := s.formatTestResult(testMapping, req, testNum, testMapping.IsHidden, verdict)
	testResultJSON, _ := json.Marshal(testResult)

	if s.hasToken(submission.TokenList, testMapping.Token) {
		return s.submissionRepo.UpdateMappingStatus(ctx, testMapping.Token, string(verdict))
	}

	submission.TestCaseResults = append(submission.TestCaseResults, string(testResultJSON))

	// Append the token to the list before any verdict check
	submission.TokenList = append(submission.TokenList, testMapping.Token)
	if err := s.submissionRepo.UpdateMappingStatus(ctx, testMapping.Token, string(verdict)); err != nil {
		return err
	}

	results := s.parseSubmissionResults(submission.TestCaseResults, true)
	submission.TestCasesPassed = s.countAcceptedResults(results)
	firstFailed := s.firstFailedResult(results)
	if firstFailed != nil {
		failedTestCaseJSON, _ := json.Marshal(firstFailed)
		failedTestCase := string(failedTestCaseJSON)
		submission.FailedTestCase = &failedTestCase
	} else {
		submission.FailedTestCase = nil
	}

	if len(submission.TokenList) == submission.TotalTestCases {
		finalVerdict := domain.VerdictAccepted
		if firstFailed != nil {
			finalVerdict = domain.VerdictStatus(firstFailed.Verdict)
		}
		return s.updateSubmissionFinal(ctx, submission.UniqueID, finalVerdict, submission.TestCasesPassed, submission.TotalTestCases, submission.MaxPoints, submission.ExecutionTimeInMS, submission.MemoryUsedInKB, submission.TestCaseResults, submission.TokenList, submission.FailedTestCase)
	}
	_, err = s.UpdateSubmissionResult(ctx, submission.UniqueID, &domain.UpdateSubmissionResultRequest{
		TokenList:         submission.TokenList,
		Verdict:           string(domain.VerdictProcessing),
		TestCaseResults:   submission.TestCaseResults,
		Score:             0,
		TestCasesPassed:   submission.TestCasesPassed,
		TotalTestCases:    submission.TotalTestCases,
		ExecutionTimeInMS: submission.ExecutionTimeInMS,
		MemoryUsedInKB:    submission.MemoryUsedInKB,
		FailedTestCase:    submission.FailedTestCase,
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

// decodeBase64 safely decodes a base64 string, returning the original if decoding fails
func (s *SubmissionService) decodeBase64(encoded string) string {
	if encoded == "" {
		return ""
	}
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return encoded // fallback to original if not valid base64
	}
	return strings.TrimSpace(string(decoded))
}

func (s *SubmissionService) hasToken(tokens []string, token string) bool {
	for _, existingToken := range tokens {
		if existingToken == token {
			return true
		}
	}
	return false
}

func (s *SubmissionService) parseStoredTestResult(raw string) (domain.SubmissionTestCaseResult, error) {
	var result domain.SubmissionTestCaseResult
	err := json.Unmarshal([]byte(raw), &result)
	return result, err
}

func (s *SubmissionService) parseSubmissionResults(rawResults []string, includeHidden bool) []domain.SubmissionTestCaseResult {
	results := make([]domain.SubmissionTestCaseResult, 0, len(rawResults))
	for _, rawResult := range rawResults {
		result, err := s.parseStoredTestResult(rawResult)
		if err != nil {
			log.Warn().Err(err).Str("test_result", rawResult).Msg("failed to parse submission test result")
			continue
		}
		if !includeHidden && result.IsHidden {
			continue
		}
		if !includeHidden {
			result = s.sanitizeTestResult(result)
		}
		results = append(results, result)
	}

	sort.SliceStable(results, func(i, j int) bool {
		return results[i].TestNumber < results[j].TestNumber
	})

	return results
}

func (s *SubmissionService) sanitizeTestResult(result domain.SubmissionTestCaseResult) domain.SubmissionTestCaseResult {
	result.TestCaseID = ""
	result.TestInput = ""
	return result
}

func (s *SubmissionService) countAcceptedResults(results []domain.SubmissionTestCaseResult) int {
	passed := 0
	for _, result := range results {
		if result.Verdict == string(domain.VerdictAccepted) {
			passed++
		}
	}
	return passed
}

func (s *SubmissionService) firstFailedResult(results []domain.SubmissionTestCaseResult) *domain.SubmissionTestCaseResult {
	for _, result := range results {
		if result.Verdict != string(domain.VerdictAccepted) {
			failed := result
			return &failed
		}
	}
	return nil
}

// formatTestResult creates a comprehensive test result for callback requests
func (s *SubmissionService) formatTestResult(testMapping *domain.SubmissionTestCaseMapping, status *domain.JudgeSubmissionCallbackRequest, testNum int, isHidden bool, verdict domain.VerdictStatus) *domain.Judge0FormattedResult {
	// Parse time string from Judge0 (e.g., "0.002")
	timeInSeconds, _ := strconv.ParseFloat(status.Time, 64)
	// Convert to milliseconds
	timeInMS := timeInSeconds * 1000

	return &domain.Judge0FormattedResult{
		TestCaseID:         testMapping.TestCaseID,
		TestNumber:         testNum,
		Verdict:            string(verdict),
		StatusID:           status.Status.ID,
		StatusDescription:  status.Status.Description,
		TestInput:          strings.TrimSpace(testMapping.TestCaseInput),
		TestExpectedOutput: strings.TrimSpace(testMapping.TestExpectedOutput),
		IsHidden:           isHidden,
		ExecutionTimeMS:    timeInMS,
		MemoryUsedKB:       status.Memory,
		Stdout:             s.decodeBase64(status.Stdout),
		Stderr:             s.decodeBase64(status.Stderr),
		CompileOutput:      s.decodeBase64(status.CompileOutput),
		Message:            status.Message,
	}
}

// updateSubmissionFinal updates the submission after all testcase callbacks arrive.
func (s *SubmissionService) updateSubmissionFinal(ctx context.Context, submissionID string,
	verdict domain.VerdictStatus, passed, total, maxPoints int,
	maxTime float64, maxMemory float64, results []string, tokenList []string, failedTestCase *string) error {

	now := time.Now()
	score := submissionScore(passed, total, maxPoints)

	// Update submission result
	err := s.submissionRepo.UpdateSubmissionResult(ctx, submissionID, &domain.Submission{
		Verdict:           string(verdict),
		Score:             score,
		TestCasesPassed:   passed,
		TotalTestCases:    total,
		ExecutionTimeInMS: maxTime,
		MemoryUsedInKB:    maxMemory,
		TestCaseResults:   results,
		TokenList:         tokenList,
		FailedTestCase:    failedTestCase,
		JudgeCompletedAt:  &now,
	})
	if err != nil {
		return fmt.Errorf("failed to update submission result: %w", err)
	}

	log.Info().
		Str("submission_id", submissionID).
		Str("verdict", string(verdict)).
		Int("passed", passed).
		Int("total", total).
		Int("score", score).
		Msg("submission completed")

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

func (c *SubmissionService) CreateJudge0BatchSubmission(req *domain.Judge0BatchSubmissionRequest) (domain.Judge0BatchSubmissionResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal batch request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/submissions/batch?base64_encoded=false", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create batch request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("X-Auth-Token", c.apiKey)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send batch request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read batch response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var batchResp domain.Judge0BatchSubmissionResponse
	if err := json.Unmarshal(body, &batchResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal batch response: %w", err)
	}

	return batchResp, nil
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
