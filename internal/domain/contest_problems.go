package domain

import (
	"context"
	"time"
)

type ContestProblems struct {
	UniqueID      string    `json:"unique_id" gorm:"type:uuid;primaryKey"`
	ContestID     string    `json:"contest_id" gorm:"type:uuid;not null"`
	ProblemID     string    `json:"problem_id" gorm:"type:uuid;not null"`
	OrderPosition int       `json:"order_position" gorm:"not null"`
	MaxPoints     int       `json:"max_points" gorm:"not null"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// Create
type CreateContestProblemRequest struct {
	ContestID     string `json:"contest_id" binding:"required"`
	ProblemID     string `json:"problem_id" binding:"required"`
	OrderPosition int    `json:"order_position" binding:"required"`
	MaxPoints     int    `json:"max_points" binding:"required"`
}

type CreateContestProblemResponse struct {
	UniqueID      string    `json:"unique_id"`
	ContestID     string    `json:"contest_id"`
	ProblemID     string    `json:"problem_id"`
	OrderPosition int       `json:"order_position"`
	MaxPoints     int       `json:"max_points"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Update
type UpdateContestProblemRequest struct {
	UniqueID      string `json:"unique_id" binding:"required"`
	ContestID     string `json:"contest_id"`
	ProblemID     string `json:"problem_id"`
	OrderPosition *int   `json:"order_position"`
	MaxPoints     *int   `json:"max_points"`
}

type UpdateContestProblemResponse struct {
	UniqueID      string    `json:"unique_id"`
	ContestID     string    `json:"contest_id"`
	ProblemID     string    `json:"problem_id"`
	OrderPosition int       `json:"order_position"`
	MaxPoints     int       `json:"max_points"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Get Single
type GetContestProblemRequest struct {
	UniqueID string `json:"unique_id" uri:"id" binding:"required"`
}

type GetContestProblemResponse struct {
	UniqueID      string    `json:"unique_id"`
	ContestID     string    `json:"contest_id"`
	ProblemID     string    `json:"problem_id"`
	OrderPosition int       `json:"order_position"`
	MaxPoints     int       `json:"max_points"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Get All (by Contest)
type GetContestProblemsRequest struct {
	ContestID string `json:"contest_id" uri:"contestId" binding:"required"`
}

type GetContestProblemsResponse struct {
	Problems []ContestProblemDetail `json:"problems"`
}

type ContestProblemDetail struct {
	UniqueID      string    `json:"unique_id"`
	ContestID     string    `json:"contest_id"`
	ProblemID     string    `json:"problem_id"`
	OrderPosition int       `json:"order_position"`
	MaxPoints     int       `json:"max_points"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	// Optional: include problem details
	Problem *Problem `json:"problem,omitempty"`
}

// Bulk Create
type BulkCreateContestProblemRequest struct {
	ContestID       string                   `json:"contest_id" binding:"required"`
	ContestProblems []BulkContestProblemItem `json:"problems" binding:"required,dive"`
}

type BulkContestProblemItem struct {
	ProblemID     string `json:"problem_id" binding:"required"`
	OrderPosition int    `json:"order_position" binding:"required"`
	MaxPoints     int    `json:"max_points" binding:"required"`
}

type BulkCreateContestProblemResponse struct {
	SuccessCount    int                            `json:"success_count"`
	FailedCount     int                            `json:"failed_count"`
	CreatedProblems []CreateContestProblemResponse `json:"created_problems"`
	Errors          []string                       `json:"errors,omitempty"`
}

// Delete
type DeleteContestProblemRequest struct {
	UniqueID string `json:"unique_id" uri:"id" binding:"required"`
}

type DeleteContestProblemResponse struct {
	Message string `json:"message"`
}

// Repository Interface
type ContestProblemsRepository interface {
	CreateContestProblem(ctx context.Context, contestProblem *ContestProblems) error
	GetContestProblemByID(ctx context.Context, uniqueID string) (*ContestProblems, error)
	GetContestProblemsByContestID(ctx context.Context, contestID string) ([]ContestProblems, error)
	UpdateContestProblem(ctx context.Context, contestProblem *ContestProblems) error
	DeleteContestProblem(ctx context.Context, uniqueID string) error
	GetContestProblemsByContestIDWithDetails(ctx context.Context, contestID string) ([]ContestProblemDetail, error)
}

// Use Case Interface
type ContestProblemsUseCase interface {
	CreateContestProblem(ctx context.Context, req *CreateContestProblemRequest) (*CreateContestProblemResponse, error)
	BulkCreateContestProblems(ctx context.Context, req *BulkCreateContestProblemRequest) (*BulkCreateContestProblemResponse, error)
	GetContestProblem(ctx context.Context, req *GetContestProblemRequest) (*GetContestProblemResponse, error)
	GetContestProblems(ctx context.Context, req *GetContestProblemsRequest) (*GetContestProblemsResponse, error)
	UpdateContestProblem(ctx context.Context, req *UpdateContestProblemRequest) (*UpdateContestProblemResponse, error)
	DeleteContestProblem(ctx context.Context, req *DeleteContestProblemRequest) (*DeleteContestProblemResponse, error)
}
