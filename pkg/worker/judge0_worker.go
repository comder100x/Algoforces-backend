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
	"os"

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
	// Parse task payload
	var payload queue.SubmissionPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}

	// Update the status to processing
	err := jw.submissionRepo.UpdateSubmissionStatus(ctx, payload.SubmissionID, string(domain.VerdictProcessing))
	if err != nil {
		return err
	}

	// Get language ID
	languageID, err := judge0.GetLanguageID(payload.Language)
	if err != nil {
		return errors.New("failed to get language ID")
	}

	totalTestCases := len(payload.VisibleTestCases) + len(payload.HiddenTestCases)
	log.Printf("Processing submission %s: Language ID: %d, Visible: %d, Hidden: %d",
		payload.SubmissionID, languageID, len(payload.VisibleTestCases), len(payload.HiddenTestCases))

	// Update total test cases count in submission
	err = jw.submissionRepo.UpdateSubmissionResult(ctx, payload.SubmissionID, &domain.Submission{
		TotalTestCases: totalTestCases,
	})
	if err != nil {
		return fmt.Errorf("failed to update total test cases: %w", err)
	}

	callbackURL := fmt.Sprintf("%s/api/submission/callback", os.Getenv("APP_URL"))

	// Submit visible test cases to Judge0
	for i, testCase := range payload.VisibleTestCases {
		log.Printf("Submitting visible test case %d/%d for submission %s", i+1, totalTestCases, payload.SubmissionID)

		submissionReqData := &judge0.SubmissionRequest{
			SourceCode:     payload.Code,
			LanguageID:     languageID,
			Stdin:          testCase.Input,
			CPUTimeLimit:   float64(payload.TimeLimitInSecond),
			MemoryLimit:    payload.MemoryLimitInMB * 1024, // Convert MB to KB
			ExpectedOutput: testCase.ExpectedOutput,
			CallbackURL:    callbackURL,
		}

		// Submit to Judge0
		submissionResponse, err := jw.Judge0Client.CreateSubmission(submissionReqData)
		if err != nil {
			return fmt.Errorf("failed to submit visible test case %d: %w", i+1, err)
		}

		// Save token mapping for callback handling
		err = jw.submissionRepo.UpdateSubmissionTokenMaps(ctx, payload.SubmissionID, submissionResponse.Token, testCase.UniqueID)
		if err != nil {
			return fmt.Errorf("failed to save token mapping for visible test case %d: %w", i+1, err)
		}
	}

	// Submit hidden test cases to Judge0
	for i, testCase := range payload.HiddenTestCases {
		testNum := len(payload.VisibleTestCases) + i + 1
		log.Printf("Submitting hidden test case %d/%d for submission %s", testNum, totalTestCases, payload.SubmissionID)

		submissionReqData := &judge0.SubmissionRequest{
			SourceCode:     payload.Code,
			LanguageID:     languageID,
			Stdin:          testCase.Input,
			CPUTimeLimit:   float64(payload.TimeLimitInSecond),
			MemoryLimit:    payload.MemoryLimitInMB * 1024, // Convert MB to KB
			ExpectedOutput: testCase.ExpectedOutput,
			CallbackURL:    callbackURL,
		}

		// Submit to Judge0
		submissionResponse, err := jw.Judge0Client.CreateSubmission(submissionReqData)
		if err != nil {
			return fmt.Errorf("failed to submit hidden test case %d: %w", testNum, err)
		}

		// Save token mapping for callback handling
		err = jw.submissionRepo.UpdateSubmissionTokenMaps(ctx, payload.SubmissionID, submissionResponse.Token, testCase.UniqueID)
		if err != nil {
			return fmt.Errorf("failed to save token mapping for hidden test case %d: %w", testNum, err)
		}
	}

	log.Printf("Successfully submitted all %d test cases for submission %s", totalTestCases, payload.SubmissionID)
	return nil
}
