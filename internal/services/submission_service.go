package services

import (
	"algoforces/internal/domain"
	"algoforces/pkg/queue"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type SubmissionService struct {
	submissionRepo domain.SubmissionRepository
	queue          queue.SubmissionQueueInterface
}

func NewSubmissionService(submissionRepo domain.SubmissionRepository, queue queue.SubmissionQueueInterface) domain.SubmissionUseCase {
	return &SubmissionService{
		submissionRepo: submissionRepo,
		queue:          queue,
	}
}

func (s *SubmissionService) CreateNewSubmission(ctx context.Context, req *domain.CreateSubmissionRequest) (*domain.CreateSubmissionResponse, error) {
	//Get all TestCases for the problem
	testCases, err := s.submissionRepo.GetAllTestCasesForProblem(ctx, req.ProblemID)
	if err != nil {
		return nil, err
	}

	submissionID := uuid.New().String()
	timNow := time.Now()
	//Update the DB Status
	submission := &domain.Submission{
		UniqueID:    submissionID,
		UserId:      req.UserID,
		ContestID:   req.ContestID,
		ProblemID:   req.ProblemID,
		Code:        req.Code,
		Language:    req.Language,
		SubmittedAt: timNow,
		Verdict:     string(domain.VerdictPending),
	}
	err = s.submissionRepo.CreateNewSubmission(ctx, submission)
	if err != nil {
		return nil, err
	}

	//Enqueue the submission to redis queue
	payload := queue.SubmissionPayload{
		SubmissionID:      submissionID,
		ProblemID:         req.ProblemID,
		UserID:            req.UserID,
		ContestID:         req.ContestID,
		Code:              req.Code,
		Language:          req.Language,
		TestCases:         testCases,
		TimeLimitInSecond: req.TimeLimitInSecond,
		MemoryLimitInMB:   req.MemoryLimitInMB,
	}

	//Push to Redis Queue
	err = s.queue.EnqueueSubmission(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to enqueue submission: %w", err)
	}

	// 7. Update DB status to "Queued"
	queuedTime := time.Now()
	submission.Verdict = string(domain.VerdictQueued)
	submission.QueuedAt = &queuedTime
	err = s.submissionRepo.UpdateSubmissionStatus(ctx, submissionID, string(domain.VerdictQueued))
	if err != nil {
		return nil, err
	}
	// 8. Return response to user
	return &domain.CreateSubmissionResponse{
		UniqueID:    submissionID,
		UserID:      req.UserID,
		ContestID:   req.ContestID,
		ProblemID:   req.ProblemID,
		Language:    req.Language,
		Verdict:     string(domain.VerdictQueued),
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
