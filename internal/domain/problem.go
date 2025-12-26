package domain

import (
	"context"
	"time"
)

type Problem struct {
	UniqueID           string    `json:"unique_id" gorm:"primaryKey;type:uuid"`
	Title              string    `json:"title" gorm:"not null"`
	Statement          string    `json:"statement" gorm:"type:text;not null"`
	Difficulty         string    `json:"difficulty" gorm:"not null"`
	TimeLimitInSeconds int       `json:"time_limit_in_seconds" gorm:"not null"`
	MemoryLimitInMB    int       `json:"memory_limit_in_mb" gorm:"not null"`
	CreatedBy          string    `json:"created_by" gorm:"type:uuid;not null"` // references User(Id)
	CreatedAt          time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type ProblemCreationRequest struct {
	Title              string `json:"title" binding:"required"`
	Statement          string `json:"statement" binding:"required"`
	Difficulty         string `json:"difficulty" binding:"required,oneof=easy medium hard"`
	TimeLimitInSeconds int    `json:"time_limit_in_seconds,omitempty" binding:"omitempty,gt=0"` // default: 1 second
	MemoryLimitInMB    int    `json:"memory_limit_in_mb,omitempty" binding:"omitempty,gt=0"`    // default: 256 MB
}

type ProblemCreationResponse struct {
	UniqueID           string    `json:"unique_id"`
	Title              string    `json:"title"`
	Statement          string    `json:"statement"`
	Difficulty         string    `json:"difficulty"`
	TimeLimitInSeconds int       `json:"time_limit_in_seconds"`
	MemoryLimitInMB    int       `json:"memory_limit_in_mb"`
	CreatedBy          string    `json:"created_by"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type ProblemUpdateRequest struct {
	UniqueID           string `json:"unique_id" binding:"required,uuid"`
	Title              string `json:"title" binding:"required"`
	Statement          string `json:"statement" binding:"required"`
	Difficulty         string `json:"difficulty" binding:"required,oneof=easy medium hard"`
	TimeLimitInSeconds int    `json:"time_limit_in_seconds,omitempty" binding:"omitempty,gt=0"` // default: 1 second
	MemoryLimitInMB    int    `json:"memory_limit_in_mb,omitempty" binding:"omitempty,gt=0"`    // default: 256 MB
}

type ProblemUpdateResponse struct {
	UniqueID           string    `json:"unique_id"`
	Title              string    `json:"title"`
	Statement          string    `json:"statement"`
	Difficulty         string    `json:"difficulty"`
	TimeLimitInSeconds int       `json:"time_limit_in_seconds"`
	MemoryLimitInMB    int       `json:"memory_limit_in_mb"`
	CreatedBy          string    `json:"created_by"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type BulkProblemCreationRequest struct {
	Problems []ProblemCreationRequest `json:"problems" binding:"required,min=1,dive"`
}

type BulkProblemCreationResponse struct {
	SuccessCount int                       `json:"success_count"`
	FailedCount  int                       `json:"failed_count"`
	Problems     []ProblemCreationResponse `json:"problems"`
	Errors       []string                  `json:"errors,omitempty"`
}

type ProblemRepository interface {
	CreateProblem(ctx context.Context, problem *Problem) error
	GetProblemByID(ctx context.Context, id string) (*Problem, error)
	UpdateProblem(ctx context.Context, problem *Problem) error
	DeleteProblem(ctx context.Context, id string) error
	GetAllProblems(ctx context.Context) ([]Problem, error)
}

type ProblemUseCase interface {
	CreateProblem(ctx context.Context, req *ProblemCreationRequest, createdBy string) (*ProblemCreationResponse, error)
	CreateProblemsInBulk(ctx context.Context, req *BulkProblemCreationRequest, createdBy string) (*BulkProblemCreationResponse, error)
	GetProblemByID(ctx context.Context, id string) (*ProblemCreationResponse, error)
	UpdateProblem(ctx context.Context, req *ProblemUpdateRequest, userID string) (*ProblemUpdateResponse, error)
	DeleteProblem(ctx context.Context, id string, userID string) error
	GetAllProblems(ctx context.Context) ([]ProblemCreationResponse, error)
}
