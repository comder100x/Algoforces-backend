package judge0

import (
	"algoforces/internal/conf"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var Judg0BaseURL = conf.JUDGE0_URL

// Language IDs for Judge0
const (
	LanguagePython = 71 // Python 3.8.1
	LanguageCPP    = 54 // C++ (GCC 9.2.0)
	LanguageJava   = 62 // Java (OpenJDK 13.0.1)
)

// SubmissionRequest represents a Judge0 submission
type SubmissionRequest struct {
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
}

// SubmissionResponse represents Judge0's response to a submission
type SubmissionResponse struct {
	Token string `json:"token"`
}

// SubmissionStatus represents the status of a Judge0 submission
type SubmissionStatus struct {
	Stdout        *string `json:"stdout"`
	Time          *string `json:"time"`   // execution time in seconds
	Memory        *int    `json:"memory"` // memory in kilobytes
	Stderr        *string `json:"stderr"`
	Token         string  `json:"token"`
	CompileOutput *string `json:"compile_output"`
	Message       *string `json:"message"`
	Status        Status  `json:"status"`
}

// Status represents Judge0 status codes
type Status struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
}

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

// Client is the Judge0 API client
type Judge0Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new Judge0 client
func NewClient(baseURL string) *Judge0Client {
	return &Judge0Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type Submission struct {
	SourceCode     string `json:"source_code"`
	LanguageID     int    `json:"language_id"`
	Stdin          string `json:"stdin"`
	ExpectedOutput string `json:"expected_output"`
}

func (c *Judge0Client) CreateSubmission(req *SubmissionRequest) (*SubmissionResponse, error) {
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

	var submissionResp SubmissionResponse
	if err := json.Unmarshal(body, &submissionResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &submissionResp, nil
}

func (c *Judge0Client) GetSubmissionStatus(token string) (*SubmissionStatus, error) {
	httpReq, err := http.NewRequest("GET", c.baseURL+"/submissions/"+token+"?base64_encoded=false", nil)
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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var statusResp SubmissionStatus
	if err := json.Unmarshal(body, &statusResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &statusResp, nil
}

// GetLanguageID converts language string to Judge0 language ID
func GetLanguageID(language string) (int, error) {
	switch language {
	case "python":
		return LanguagePython, nil
	case "cpp":
		return LanguageCPP, nil
	case "java":
		return LanguageJava, nil
	default:
		return 0, fmt.Errorf("unsupported language: %s", language)
	}
}

// WaitForCompletion polls Judge0 until the submission is complete
func (c *Judge0Client) WaitForCompletion(token string, maxWaitTime time.Duration) (*SubmissionStatus, error) {
	startTime := time.Now()
	pollInterval := 1 * time.Second

	for {
		if time.Since(startTime) > maxWaitTime {
			return nil, fmt.Errorf("timeout waiting for submission completion")
		}

		status, err := c.GetSubmissionStatus(token)
		if err != nil {
			return nil, err
		}

		// Check if processing is complete (not in queue or processing)
		if status.Status.ID != StatusInQueue && status.Status.ID != StatusProcessing {
			return status, nil
		}

		time.Sleep(pollInterval)
	}
}
