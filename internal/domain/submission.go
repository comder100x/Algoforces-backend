package domain

import (
	"context"
	"time"

	"github.com/lib/pq"
)

const (
	LanguagePython = 71 // Python 3.8.1
	LanguageCPP    = 54 // C++ (GCC 9.2.0)
	LanguageJava   = 62 // Java (OpenJDK 13.0.1)
)

// Judge0 Status IDs
const (
	StatusInQueue             = 1
	StatusProcessing          = 2
	StatusAccepted            = 3
	StatusWrongAnswer         = 4
	StatusTimeLimitExceeded   = 5
	StatusCompilationError    = 6
	StatusRuntimeError        = 7  // SIGSEGV
	StatusRuntimeErrorOther   = 8  // SIGXFSZ
	StatusRuntimeErrorSIGFPE  = 9  // SIGFPE
	StatusRuntimeErrorSIGABRT = 10 // SIGABRT
	StatusRuntimeErrorNZEC    = 11 // Non-zero exit code
	StatusRuntimeErrorOther2  = 12 // Other
	StatusInternalError       = 13
	StatusExecFormatError     = 14
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

type SubmissionTestCaseMapping struct {
	UniqueID           string    `json:"id" gorm:"primaryKey;type:uuid"`
	Token              string    `json:"token" gorm:"uniqueIndex;not null"`
	SubmissionID       string    `json:"submission_id" gorm:"not null"`
	TestCaseID         string    `json:"testcase_id" gorm:"not null"`
	Status             string    `json:"status" gorm:"default:'pending'"`
	TestOrderPosition  int       `json:"test_order_position"`
	TestCaseInput      string    `json:"test_case_input"`
	TestExpectedOutput string    `json:"test_expected_output"`
	IsHidden           bool      `json:"is_hidden" gorm:"not null"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

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
	TokenList         pq.StringArray `json:"token_list" gorm:"type:text[]"`
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
	TokenList         pq.StringArray `json:"token_list" gorm:"type:text[]"`
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
	TokenList         pq.StringArray `json:"token_list" gorm:"type:text[]"`
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

// -----------------------------------Related to JUDGE0---------------------------------------------------------//
type Judge0SubmissionRequest struct {
	SourceCode                           string  `json:"source_code"`
	LanguageID                           int     `json:"language_id"`
	Stdin                                string  `json:"stdin,omitempty"`
	ExpectedOutput                       string  `json:"expected_output,omitempty"`
	CPUTimeLimit                         float64 `json:"cpu_time_limit,omitempty"`  // seconds
	CPUExtraTime                         float64 `json:"cpu_extra_time,omitempty"`  // seconds
	WallTimeLimit                        float64 `json:"wall_time_limit,omitempty"` // seconds
	MemoryLimit                          int     `json:"memory_limit,omitempty"`    // kilobytes
	StackLimit                           int     `json:"stack_limit,omitempty"`     // kilobytes
	MaxProcessesAndFiles                 int     `json:"max_processes_and_or_files,omitempty"`
	EnablePerProcessAndThreadTimeLimit   bool    `json:"enable_per_process_and_thread_time_limit,omitempty"`
	EnablePerProcessAndThreadMemoryLimit bool    `json:"enable_per_process_and_thread_memory_limit,omitempty"`
	CallbackURL                          string  `json:"callback_url,omitempty"`
}

type Judge0SubmissionResponse struct {
	Token string `json:"token"`
}

// Batch submission types for Judge0
type Judge0BatchSubmissionRequest struct {
	Submissions []Judge0SubmissionRequest `json:"submissions"`
}

type Judge0BatchSubmissionResponse []Judge0SubmissionResponse

// SubmissionStatus represents the status of a Judge0 submission

type Judge0Status struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
}
type Judge0SubmissionStatus struct {
	Stdout        *string      `json:"stdout"`
	Time          *string      `json:"time"`   // execution time in seconds
	Memory        *int         `json:"memory"` // memory in kilobytes
	Stderr        *string      `json:"stderr"`
	Token         string       `json:"token"`
	CompileOutput *string      `json:"compile_output"`
	Message       *string      `json:"message"`
	Status        Judge0Status `json:"status"`
}

// Status represents Judge0 status codes
type JudgeSubmissionCallbackRequest struct {
	Token         string       `json:"token"`
	Status        Judge0Status `json:"status"`
	Time          string       `json:"time"`   // Judge0 sends time as string (e.g., "0.002")
	Memory        int          `json:"memory"` // Memory in KB
	Stderr        string       `json:"stderr"`
	Stdout        string       `json:"stdout"`
	CompileOutput string       `json:"compile_output"`
	Message       string       `json:"message"`
	FinishedAt    *time.Time   `json:"finished_at,omitempty"`
}

//-----------------------------------Related to JUDGE0---------------------------------------------------------//

type SubmissionRepository interface {
	GetAllTestCasesForProblem(ctx context.Context, problemID string) ([]TestCase, error)
	GetTestCaseByID(ctx context.Context, testCaseID string) (*TestCase, error)
	CreateNewSubmission(ctx context.Context, submission *Submission) error
	GetSubmissionDetails(ctx context.Context, uniqueID string) (*Submission, error)
	UpdateSubmissionStatus(ctx context.Context, submissionID string, status string) error
	UpdateSubmissionResult(ctx context.Context, submissionID string, result *Submission) error
	CreateTokenMapping(ctx context.Context, mapping *SubmissionTestCaseMapping) error
	GetMappingByToken(ctx context.Context, token string) (*SubmissionTestCaseMapping, error)
	UpdateMappingStatus(ctx context.Context, token string, status string) error
}

type SubmissionUseCase interface {
	CreateNewSubmission(ctx context.Context, req *CreateSubmissionRequest) (*CreateSubmissionResponse, error)
	GetSubmissionDetails(ctx context.Context, uniqueID string) (*Submission, error)
	UpdateSubmissionStatus(ctx context.Context, submissionID string, status string) error
	UpdateSubmissionResult(ctx context.Context, submissionID string, req *UpdateSubmissionResultRequest) (*UpdateSubmissionResultResponse, error)
	JudgeSubmissionCallback(ctx context.Context, req *JudgeSubmissionCallbackRequest) error
}
