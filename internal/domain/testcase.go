package domain

import "context"

type TestCase struct {
	UniqueID       string `json:"unique_id" gorm:"primaryKey;type:uuid"`
	ProblemID      string `json:"problem_id" gorm:"type:uuid;not null"` // references Problem(UniqueID)
	Input          string `json:"input" gorm:"type:text;not null"`
	ExpectedOutput string `json:"output" gorm:"type:text;not null"`
	IsHidden       bool   `json:"is_hidden" gorm:"not null;default:true"`
	OrderPosition  int    `json:"order_position" gorm:"not null"`
}

type CreateTestCaseRequest struct {
	ProblemID      string `json:"problem_id" binding:"required,uuid"`
	Input          string `json:"input" binding:"required"`
	ExpectedOutput string `json:"output" binding:"required"`
	IsHidden       bool   `json:"is_hidden"`
	OrderPosition  int    `json:"order_position" binding:"required,min=1"`
}

type CreateTestCaseResponse struct {
	UniqueID       string `json:"unique_id"`
	ProblemID      string `json:"problem_id"`
	Input          string `json:"input"`
	ExpectedOutput string `json:"output"`
	IsHidden       bool   `json:"is_hidden"`
	OrderPosition  int    `json:"order_position"`
}

type UpdateTestCaseRequest struct {
	UniqueID       string `json:"unique_id" binding:"required,uuid"`
	ProblemID      string `json:"problem_id" binding:"required,uuid"`
	Input          string `json:"input" binding:"required"`
	ExpectedOutput string `json:"output" binding:"required"`
	IsHidden       bool   `json:"is_hidden"`
	OrderPosition  int    `json:"order_position" binding:"required,min=1"`
}

type UpdateTestCaseResponse struct {
	UniqueID       string `json:"unique_id"`
	ProblemID      string `json:"problem_id"`
	Input          string `json:"input"`
	ExpectedOutput string `json:"output"`
	IsHidden       bool   `json:"is_hidden"`
	OrderPosition  int    `json:"order_position"`
}

type BulkTestCaseUploadRequest struct {
	TestCases []CreateTestCaseRequest `json:"test_cases" binding:"required,dive"`
}

type BulkTestCaseUploadResponse struct {
	CreatedTestCases []CreateTestCaseResponse `json:"created_test_cases"`
}

type TestCaseRepository interface {
	CreateTestCase(ctx context.Context, testCase *TestCase) error
	UpdateTestCase(ctx context.Context, testCase *TestCase) error
	DeleteTestCase(ctx context.Context, uniqueID string) error
	GetTestCasesByProblemID(ctx context.Context, problemID string) ([]*TestCase, error)
	GetTestCaseByUniqueID(ctx context.Context, uniqueID string) (*TestCase, error)
}

type TestCaseUseCase interface {
	CreateNewTestCase(ctx context.Context, req *CreateTestCaseRequest) (*CreateTestCaseResponse, error)
	GetAllTestCasesForProblem(ctx context.Context, problemID string) ([]*TestCase, error)
	GetTestCaseDetails(ctx context.Context, uniqueID string) (*TestCase, error)
	UpdateSingleTestCase(ctx context.Context, req *UpdateTestCaseRequest) (*UpdateTestCaseResponse, error)
	DeleteSingleTestCase(ctx context.Context, uniqueID string) error
	UploadTestCasesInBulk(ctx context.Context, req *BulkTestCaseUploadRequest) (*BulkTestCaseUploadResponse, error)
}
