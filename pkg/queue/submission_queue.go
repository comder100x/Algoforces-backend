package queue

import (
	"algoforces/internal/domain"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"
)

// Task types
const (
	TypeSubmissionJudge = "submission:judge"
)

type SubmissionPayload struct {
	SubmissionID      string            `json:"submission_id"`
	ProblemID         string            `json:"problem_id"`
	UserID            string            `json:"user_id"`
	ContestID         string            `json:"contest_id"`
	Code              string            `json:"code"`
	Language          string            `json:"language"`
	VisibleTestCases  []domain.TestCase `json:"visible_test_cases"`
	HiddenTestCases   []domain.TestCase `json:"hidden_test_cases"`
	TimeLimitInSecond int               `json:"time_limit"`
	MemoryLimitInMB   int               `json:"memory_limit"`
}

// SubmissionQueue manages the Redis queue for submissions
type SubmissionQueue struct {
	client    *asynq.Client
	inspector *asynq.Inspector
}

// NewSubmissionQueue creates a new submission queue client
func NewSubmissionQueue(redisURL string) (*SubmissionQueue, error) {
	redisOpt := asynq.RedisClientOpt{
		Addr: redisURL,
	}

	client := asynq.NewClient(redisOpt)
	inspector := asynq.NewInspector(redisOpt)

	return &SubmissionQueue{
		client:    client,
		inspector: inspector,
	}, nil
}

// EnqueueSubmission adds a new submission to the queue
func (sq *SubmissionQueue) EnqueueSubmission(ctx context.Context, payload SubmissionPayload) error {
	// Serialize payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create a new task
	task := asynq.NewTask(TypeSubmissionJudge, payloadBytes,
		asynq.MaxRetry(3),                  // Retry up to 3 times on failure
		asynq.Timeout(5*time.Minute),       // Task timeout
		asynq.Queue("submission"),          // Queue name
		asynq.Retention(24*time.Hour),      // Keep completed tasks for 24 hours
		asynq.ProcessIn(1*time.Second),     // Process immediately (or delay if needed)
		asynq.TaskID(payload.SubmissionID), // Unique task ID
	)

	// Enqueue the task
	info, err := sq.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	log.Printf("Enqueued submission %s to queue: %s", payload.SubmissionID, info.Queue)
	return nil
}

// GetQueueInfo returns information about the queue
func (sq *SubmissionQueue) GetQueueInfo(queueName string) (*asynq.QueueInfo, error) {
	return sq.inspector.GetQueueInfo(queueName)
}

// ListPendingTasks returns pending tasks in the queue
func (sq *SubmissionQueue) ListPendingTasks(queueName string) ([]*asynq.TaskInfo, error) {
	return sq.inspector.ListPendingTasks(queueName)
}

// Close closes the queue client
func (sq *SubmissionQueue) Close() error {
	return sq.client.Close()
}

// SubmissionQueueInterface defines the interface for submission queue operations
type SubmissionQueueInterface interface {
	EnqueueSubmission(ctx context.Context, payload SubmissionPayload) error
	GetQueueInfo(queueName string) (*asynq.QueueInfo, error)
	Close() error
}
