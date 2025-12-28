package domain

import (
	"context"
	"time"

	"github.com/lib/pq"
)

// VerdictStatus represents the status of a submission
type VerdictStatus string

const (
	VerdictPending             VerdictStatus = "Pending"
	VerdictQueued              VerdictStatus = "Queued"
	VerdictProcessing          VerdictStatus = "Processing"
	VerdictAccepted            VerdictStatus = "Accepted"
	VerdictWrongAnswer         VerdictStatus = "Wrong Answer"
	VerdictTimeLimitExceeded   VerdictStatus = "Time Limit Exceeded"
	VerdictMemoryLimitExceeded VerdictStatus = "Memory Limit Exceeded"
	VerdictRuntimeError        VerdictStatus = "Runtime Error"
	VerdictCompilationError    VerdictStatus = "Compilation Error"
	VerdictSystemError         VerdictStatus = "System Error"
)

type Submission struct {
	// Problem related stuff
	UniqueID    string     `json:"unique_id" gorm:"primaryKey;type:uuid"`
	UserId      string     `json:"user_id" gorm:"type:uuid;not null"`    // references User(Id)
	ContestID   string     `json:"contest_id" gorm:"type:uuid;not null"` // references Contest(Id)
	ProblemID   string     `json:"problem_id" gorm:"type:uuid;not null"` // references Problem(Id)
	Code        string     `json:"code" gorm:"type:text;not null"`
	Language    string     `json:"language" gorm:"type:varchar(20);not null"`
	SubmittedAt time.Time  `json:"submitted_at"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	QueuedAt    *time.Time `json:"queued_at"`

	//Verdict status update
	Verdict           string         `json:"verdict" gorm:"type:varchar(50);not null;default:'Pending';index"`
	Score             int            `json:"score" gorm:"default:0"`
	TestCasesPassed   int            `json:"test_cases_passed" gorm:"default:0"`
	TotalTestCases    int            `json:"total_test_cases" gorm:"default:0"`
	ExecutionTimeInMS float64        `json:"execution_time"`
	MemoryUsedInKB    float64        `json:"memory_used_in_kb"`
	CompilationError  string         `json:"compilation_error" gorm:"type:text"`
	RuntimeError      string         `json:"runtime_error" gorm:"type:text"`
	TestCaseResults   pq.StringArray `json:"test_case_results" gorm:"type:text[]"` // JSON array of test results
	FailedTestCase    *string        `json:"failed_test_case" gorm:"type:text"`    // First failed test case details
	JudgeCompletedAt  *time.Time     `json:"judge_completed_at"`
}

type CreateSubmissionRequest struct {
	UserID            string `json:"user_id" binding:"required,uuid"`
	ContestID         string `json:"contest_id" binding:"required,uuid"`
	ProblemID         string `json:"problem_id" binding:"required,uuid"`
	Code              string `json:"code" binding:"required"`
	Language          string `json:"language" binding:"required,oneof=python cpp java"`
	TimeLimitInSecond int    `json:"time_limit" binding:"required,gt=0"`
	MemoryLimitInMB   int    `json:"memory_limit" binding:"required,gt=0"`
}

type CreateSubmissionResponse struct {
	UniqueID    string    `json:"unique_id"`
	UserID      string    `json:"user_id"`
	ContestID   string    `json:"contest_id"`
	ProblemID   string    `json:"problem_id"`
	Language    string    `json:"language"`
	Verdict     string    `json:"verdict"`
	SubmittedAt time.Time `json:"submitted_at"`
	Message     string    `json:"message"`
}

type UpdateSubmissionResultRequest struct {
	Verdict           string         `json:"verdict" gorm:"type:varchar(50);not null;default:'Pending';index"`
	Score             int            `json:"score" gorm:"default:0"`
	TestCasesPassed   int            `json:"test_cases_passed" gorm:"default:0"`
	TotalTestCases    int            `json:"total_test_cases" gorm:"default:0"`
	ExecutionTimeInMS float64        `json:"execution_time"`
	MemoryUsedInKB    float64        `json:"memory_used_in_kb"`
	CompilationError  string         `json:"compilation_error" gorm:"type:text"`
	RuntimeError      string         `json:"runtime_error" gorm:"type:text"`
	TestCaseResults   pq.StringArray `json:"test_case_results" gorm:"type:text[]"` // JSON array of test results
	FailedTestCase    *string        `json:"failed_test_case" gorm:"type:text"`    // First failed test case details
	JudgeCompletedAt  *time.Time     `json:"judge_completed_at"`
}
type UpdateSubmissionResultResponse struct {
	UniqueID          string         `json:"unique_id"`
	UserID            string         `json:"user_id"`
	ContestID         string         `json:"contest_id"`
	ProblemID         string         `json:"problem_id"`
	Language          string         `json:"language"`
	Verdict           string         `json:"verdict"`
	SubmittedAt       time.Time      `json:"submitted_at"`
	Message           string         `json:"message"`
	Score             int            `json:"score" gorm:"default:0"`
	TestCasesPassed   int            `json:"test_cases_passed" gorm:"default:0"`
	TotalTestCases    int            `json:"total_test_cases" gorm:"default:0"`
	ExecutionTimeInMS float64        `json:"execution_time"`
	MemoryUsedInKB    float64        `json:"memory_used_in_kb"`
	CompilationError  string         `json:"compilation_error" gorm:"type:text"`
	RuntimeError      string         `json:"runtime_error" gorm:"type:text"`
	TestCaseResults   pq.StringArray `json:"test_case_results" gorm:"type:text[]"` // JSON array of test results
	FailedTestCase    *string        `json:"failed_test_case" gorm:"type:text"`    // First failed test case details
	JudgeCompletedAt  *time.Time     `json:"judge_completed_at"`
}

type SubmissionRepository interface {
	GetAllTestCasesForProblem(ctx context.Context, problemID string) ([]TestCase, error)
	CreateNewSubmission(ctx context.Context, submission *Submission) error
	GetSubmissionDetails(ctx context.Context, uniqueID string) (*Submission, error)
	UpdateSubmissionStatus(ctx context.Context, submissionID string, status string) error
	UpdateSubmissionResult(ctx context.Context, submissionID string, result *Submission) error
}

type SubmissionUseCase interface {
	CreateNewSubmission(ctx context.Context, req *CreateSubmissionRequest) (*CreateSubmissionResponse, error)
	GetSubmissionDetails(ctx context.Context, uniqueID string) (*Submission, error)
	UpdateSubmissionStatus(ctx context.Context, submissionID string, status string) error
	UpdateSubmissionResult(ctx context.Context, submissionID string, req *UpdateSubmissionResultRequest) (*UpdateSubmissionResultResponse, error)
}
