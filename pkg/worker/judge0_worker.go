package worker

import (
	"algoforces/internal/domain"
	"algoforces/pkg/judge0"
	"algoforces/pkg/queue"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/hibiken/asynq"
)

type JudgeWorker struct {
	submissionRepo domain.SubmissionRepository
	Judge0Client   *judge0.Judge0Client
}

func NewJudgeWorker(submissionRepo domain.SubmissionRepository, judge0URL string) *JudgeWorker {
	return &JudgeWorker{
		submissionRepo: submissionRepo,
		Judge0Client:   judge0.NewClient(judge0URL),
	}
}

func (jw *JudgeWorker) JudgeSubmission(ctx context.Context, task *asynq.Task) error {
	//Parse task payload
	var payload queue.SubmissionPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}

	// update the status to processing
	err := jw.submissionRepo.UpdateSubmissionStatus(ctx, payload.SubmissionID, string(domain.VerdictProcessing))

	if err != nil {
		return err
	}

	// Get language ID
	languageID, err := judge0.GetLanguageID(payload.Language)
	if err != nil {
		return errors.New("failed to get language ID")
	}
	log.Printf("Language ID: %d", languageID)
	log.Printf("Visible Test Cases: %d", len(payload.VisibleTestCases))
	log.Printf("Hidden Test Cases: %d", len(payload.HiddenTestCases))

	TotalTestCases := len(payload.VisibleTestCases) + len(payload.HiddenTestCases)
	passedTests := 0
	var testResults []string
	var maxTime float64
	var maxMemory int

	isVisibleTestCaseFailed := false
	for i, testCase := range payload.VisibleTestCases {
		log.Printf("Running test case %d/%d for submission %s", i+1, TotalTestCases, payload.SubmissionID)

		// Create submission request
		submissionReqData := &judge0.SubmissionRequest{
			SourceCode:     payload.Code,
			LanguageID:     languageID,
			Stdin:          testCase.Input,
			CPUTimeLimit:   float64(payload.TimeLimitInSecond),
			MemoryLimit:    payload.MemoryLimitInMB * 1024, // Convert MB to KB
			ExpectedOutput: testCase.ExpectedOutput,
		}

		// SUBMIT TO JUDGE0
		submissionResponse, err := jw.Judge0Client.CreateSubmission(submissionReqData)
		if err != nil {
			return err
		}

		// Wait for result - add buffer time for Judge0 queue latency (30 seconds)
		maxWaitTime := time.Duration(payload.TimeLimitInSecond+30) * time.Second
		_, err = jw.Judge0Client.WaitForCompletion(submissionResponse.Token, maxWaitTime)
		if err != nil {
			return err
		}

		// Get the final submission status with details
		finalResponse, err := jw.Judge0Client.GetSubmissionStatus(submissionResponse.Token)
		if err != nil {
			return err
		}

		// Update max time and memory
		if finalResponse.Time != nil {
			if timeVal, err := strconv.ParseFloat(*finalResponse.Time, 64); err == nil {
				timeInMS := timeVal * 1000.0 // Convert seconds to milliseconds
				if timeInMS > maxTime {
					maxTime = timeInMS
				}
			}
		}

		if finalResponse.Memory != nil && *finalResponse.Memory > maxMemory {
			maxMemory = *finalResponse.Memory
		}

		// Map Judge0 status to our verdict
		verdict := jw.mapJudge0Status(finalResponse.Status.ID)

		// Create comprehensive test result using shared function
		testResult := jw.formatTestResult(testCase, finalResponse, i+1, false)
		testResults = append(testResults, testResult)

		if verdict == domain.VerdictAccepted {
			passedTests++
		} else {
			isVisibleTestCaseFailed = true
		}
	}
	if isVisibleTestCaseFailed {
		return jw.updateSubmissionError(ctx, payload.SubmissionID, domain.VerdictWrongAnswer, passedTests, TotalTestCases, maxTime, maxMemory, testResults)
	}
	for i, testCase := range payload.HiddenTestCases {
		log.Printf("Running test case %d/%d for submission %s", i+1, TotalTestCases, payload.SubmissionID)

		// Create submission request
		submissionReqData := &judge0.SubmissionRequest{
			SourceCode:     payload.Code,
			LanguageID:     languageID,
			Stdin:          testCase.Input,
			CPUTimeLimit:   float64(payload.TimeLimitInSecond),
			MemoryLimit:    payload.MemoryLimitInMB * 1024, // Convert MB to KB
			ExpectedOutput: testCase.ExpectedOutput,
		}

		// SUBMIT TO JUDGE0
		submissionResponse, err := jw.Judge0Client.CreateSubmission(submissionReqData)
		if err != nil {
			return err
		}

		// Wait for result - add buffer time for Judge0 queue latency (30 seconds)
		maxWaitTime := time.Duration(payload.TimeLimitInSecond+30) * time.Second
		_, err = jw.Judge0Client.WaitForCompletion(submissionResponse.Token, maxWaitTime)
		if err != nil {
			return err
		}

		// Get the final submission status with details
		finalResponse, err := jw.Judge0Client.GetSubmissionStatus(submissionResponse.Token)
		if err != nil {
			return err
		}

		// Update max time and memory
		if finalResponse.Time != nil {
			if timeVal, err := strconv.ParseFloat(*finalResponse.Time, 64); err == nil {
				timeInMS := timeVal * 1000.0 // Convert seconds to milliseconds
				if timeInMS > maxTime {
					maxTime = timeInMS
				}
			}
		}

		if finalResponse.Memory != nil && *finalResponse.Memory > maxMemory {
			maxMemory = *finalResponse.Memory
		}

		// Map Judge0 status to our verdict
		verdict := jw.mapJudge0Status(finalResponse.Status.ID)

		// Create comprehensive test result using shared function
		testNum := len(payload.VisibleTestCases) + i + 1
		testResult := jw.formatTestResult(testCase, finalResponse, testNum, true)
		testResults = append(testResults, testResult)

		if verdict == domain.VerdictAccepted {
			passedTests++
		} else {
			return jw.updateSubmissionError(ctx, payload.SubmissionID, verdict, passedTests, TotalTestCases, maxTime, maxMemory, testResults)
		}
	}

	// All tests passed - update submission with success
	finalVerdict := domain.VerdictAccepted
	return jw.updateSubmissionSuccess(ctx, payload.SubmissionID, finalVerdict, passedTests, TotalTestCases, testResults, maxTime, maxMemory)
}

// mapJudge0Status maps Judge0 status to our verdict
func (w *JudgeWorker) mapJudge0Status(statusID int) domain.VerdictStatus {
	switch statusID {
	case judge0.StatusAccepted:
		return domain.VerdictAccepted
	case judge0.StatusWrongAnswer:
		return domain.VerdictWrongAnswer
	case judge0.StatusTimeLimitExceeded:
		return domain.VerdictTimeLimitExceeded
	case judge0.StatusCompilationError:
		return domain.VerdictCompilationError
	case judge0.StatusRuntimeError,
		judge0.StatusRuntimeErrorOther,
		judge0.StatusRuntimeErrorSIGFPE,
		judge0.StatusRuntimeErrorSIGABRT,
		judge0.StatusRuntimeErrorNZEC,
		judge0.StatusRuntimeErrorOther2:
		return domain.VerdictRuntimeError
	default:
		return domain.VerdictSystemError
	}
}

// formatTestResult creates a comprehensive single-line test result
func (w *JudgeWorker) formatTestResult(testCase domain.TestCase, status *judge0.SubmissionStatus, testNum int, isHidden bool) string {
	// Get the verdict for this test
	verdict := w.mapJudge0Status(status.Status.ID)

	// Create comprehensive test result
	testResult := fmt.Sprintf("Test %d (%s): %s", testNum, testCase.UniqueID, verdict)

	// Add execution details for visible tests
	if !isHidden {

		if status.Time != nil {
			if timeVal, err := strconv.ParseFloat(*status.Time, 64); err == nil {
				testResult += fmt.Sprintf(" | Time: %.2fms", timeVal*1000)
			}
		}
		if status.Memory != nil {
			testResult += fmt.Sprintf(" | Memory: %dKB", *status.Memory)
		}

		// Add input/output details (only for visible tests)
		if !testCase.IsHidden {
			testResult += fmt.Sprintf(" | Input: %s | Expected: %s", testCase.Input, testCase.ExpectedOutput)
			if status.Stdout != nil && *status.Stdout != "" {
				testResult += fmt.Sprintf(" | Got: %s", *status.Stdout)
			}
		}

		// Add error details if any
		if status.Stderr != nil && *status.Stderr != "" {
			testResult += fmt.Sprintf(" | Stderr: %s", *status.Stderr)
		}
		if status.CompileOutput != nil && *status.CompileOutput != "" {
			testResult += fmt.Sprintf(" | Compile: %s", *status.CompileOutput)
		}
		if status.Message != nil && *status.Message != "" {
			testResult += fmt.Sprintf(" | Message: %s", *status.Message)
		}
	}

	return testResult
}

// updateSubmissionSuccess updates the submission with success result
func (w *JudgeWorker) updateSubmissionSuccess(ctx context.Context, submissionID string,
	verdict domain.VerdictStatus, passed, total int, results []string,
	maxTime float64, maxMemory int) error {

	now := time.Now()

	// Update submission result
	err := w.submissionRepo.UpdateSubmissionResult(ctx, submissionID, &domain.Submission{
		Verdict:           string(verdict),
		Score:             passed,
		TestCasesPassed:   passed,
		TotalTestCases:    total,
		ExecutionTimeInMS: float64(maxTime),
		MemoryUsedInKB:    float64(maxMemory),
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
func (w *JudgeWorker) updateSubmissionError(ctx context.Context, submissionID string,
	verdict domain.VerdictStatus, passed, total int,
	maxTime float64, maxMemory int, testResults []string) error {

	now := time.Now()

	// Update submission result with error details
	err := w.submissionRepo.UpdateSubmissionResult(ctx, submissionID, &domain.Submission{
		Verdict:           string(verdict),
		Score:             0, // No score for failed submissions
		TestCasesPassed:   passed,
		TotalTestCases:    total,
		ExecutionTimeInMS: float64(maxTime),
		MemoryUsedInKB:    float64(maxMemory),
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
