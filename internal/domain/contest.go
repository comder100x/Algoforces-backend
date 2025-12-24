package domain

import (
	"context"
	"time"
)

type Contest struct {
	Id             string    `json:"id" gorm:"primaryKey";type:uuid"`
	Name           string    `json:"name" gorm:"not null"`
	Description    string    `json:"description" gorm:""`
	StartTime      time.Time `json:"start_time" gorm:"not null"`
	EndTime        time.Time `json:"end_time" gorm:"not null"`
	Duration       int       `json:"duration" gorm:"not null"` // in minutes
	Visible        bool      `json:"visible" gorm:"default:false"`
	CreatedBy      string    `json:"created_by" gorm:"type:uuid;not null"` //refrences User(Id)
	IsActive       bool      `json:"is_active" gorm:"default:false"`
	ProblemSetters []string  `json:"problem_setters" gorm:"type:uuid[]"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type CreateContestRequest struct {
	Name           string    `json:"name" binding:"required"`
	Description    string    `json:"description"`
	StartTime      time.Time `json:"start_time" binding:"required"`
	EndTime        time.Time `json:"end_time" binding:"required,gtfield=StartTime"`
	Duration       int       `json:"duration" binding:"required,gt=0"` // in minutes
	Visible        bool      `json:"visible"`
	IsActive       bool      `json:"is_active"`
	ProblemSetters []string  `json:"problem_setters"`
}

type CreateContestResponse struct {
	Id             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	Duration       int       `json:"duration"` // in minutes
	Visible        bool      `json:"visible"`
	IsActive       bool      `json:"is_active"`
	ProblemSetters []string  `json:"problem_setters"`
	CreatedBy      string    `json:"created_by"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type UpdateContestRequest struct {
	Id             string    `json:"id" binding:"required,uuid"`
	Name           string    `json:"name" binding:"required"`
	Description    string    `json:"description"`
	StartTime      time.Time `json:"start_time" binding:"required"`
	EndTime        time.Time `json:"end_time" binding:"required,gtfield=StartTime"`
	Duration       int       `json:"duration" binding:"required,gt=0"` // in minutes
	Visible        bool      `json:"visible"`
	ProblemSetters []string  `json:"problem_setters"`
	IsActive       bool      `json:"is_active"`
}

type UpdateContestResponse struct {
	Id             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	Duration       int       `json:"duration"` // in minutes
	Visible        bool      `json:"visible"`
	ProblemSetters []string  `json:"problem_setters"`
	IsActive       bool      `json:"is_active"`
	CreatedBy      string    `json:"created_by"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type DeleteContestRequest struct {
	Id string `json:"id" binding:"required,uuid"`
}

type ContestRepository interface {
	CreateContest(ctx context.Context, contest *Contest) error
	GetByID(ctx context.Context, id string) (*Contest, error)
	UpdateContest(ctx context.Context, contest *Contest) error
	DeleteContest(ctx context.Context, id string) error
	CheckContestInTimeWindow(ctx context.Context, startTime, endTime time.Time) ([]Contest, error)
}

type ContestUseCase interface {
	CreateContest(ctx context.Context, req *CreateContestRequest, userId string) (*CreateContestResponse, error)
	UpdateContest(ctx context.Context, req *UpdateContestRequest) (*UpdateContestResponse, error)
	GetContestDetails(ctx context.Context, id string) (*CreateContestResponse, error)
	DeleteContest(ctx context.Context, req *DeleteContestRequest) error
	StartContest(ctx context.Context, id string) error
}
