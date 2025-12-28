package worker

import (
	"algoforces/internal/domain"
	"algoforces/pkg/judge0"
	"algoforces/pkg/queue"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
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

	TotalTestCases := len(payload.TestCases)
	passedTests := 0
	var testResults []string
	var maxTime float64
	var maxMemory int

	for i, testCase := range payload.TestCases {
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

		//Wait for result
		_, err = jw.Judge0Client.WaitForCompletion(submissionResponse.Token, time.Duration(payload.TimeLimitInSecond)*time.Second)
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
		testResults = append(testResults, fmt.Sprintf("Test %d: %s", i+1, verdict))

		if verdict == domain.VerdictAccepted {
			passedTests++
		} else {
			// Stop on first failure and update submission with error verdict
			errorMsg := jw.formatFailedTestCase(testCase, finalResponse, i+1)
			return jw.updateSubmissionError(ctx, payload.SubmissionID, verdict, errorMsg, passedTests, TotalTestCases, maxTime, maxMemory, testResults)
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

// formatFailedTestCase creates a detailed failure message
func (w *JudgeWorker) formatFailedTestCase(testCase domain.TestCase, status *judge0.SubmissionStatus, testNum int) string {
	var details strings.Builder
	details.WriteString(fmt.Sprintf("Test Case #%d Failed\n", testNum))

	if !testCase.IsHidden {
		details.WriteString(fmt.Sprintf("Input: %s\n", testCase.Input))
		details.WriteString(fmt.Sprintf("Expected: %s\n", testCase.ExpectedOutput))
		if status.Stdout != nil {
			details.WriteString(fmt.Sprintf("Got: %s\n", *status.Stdout))
		}
	}

	if status.Stderr != nil && *status.Stderr != "" {
		details.WriteString(fmt.Sprintf("Stderr: %s\n", *status.Stderr))
	}

	if status.CompileOutput != nil && *status.CompileOutput != "" {
		details.WriteString(fmt.Sprintf("Compilation: %s\n", *status.CompileOutput))
	}

	if status.Message != nil && *status.Message != "" {
		details.WriteString(fmt.Sprintf("Message: %s\n", *status.Message))
	}

	return details.String()
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
	verdict domain.VerdictStatus, errorMsg string, passed, total int,
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
		FailedTestCase:    &errorMsg,
		JudgeCompletedAt:  &now,
	})
	if err != nil {
		log.Printf("Failed to update error status: %v", err)
		return fmt.Errorf("failed to update submission result: %w", err)
	}

	log.Printf("Submission %s completed with verdict: %s (%d/%d tests passed) - %s",
		submissionID, verdict, passed, total, errorMsg)

	return nil
}
