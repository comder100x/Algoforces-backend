package domain

import (
	"context"
	"time"
)

// ContestRegistration is the database model
type ContestRegistration struct {
	ID           string    `json:"id" gorm:"primaryKey;type:uuid"`
	UserID       string    `json:"user_id" gorm:"type:uuid;not null"`
	ContestID    string    `json:"contest_id" gorm:"type:uuid;not null"`
	RegisteredAt time.Time `json:"registered_at" gorm:"autoCreateTime"`
	Status       string    `json:"status" gorm:"default:registered"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type ContestRegisterRequest struct {
	ContestID string `json:"contest_id" binding:"required,uuid4"`
}

type ContestRegisterResponse struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	ContestID    string    `json:"contest_id"`
	RegisteredAt time.Time `json:"registered_at"`
	Status       string    `json:"status"`
}

type ContestUnregisterRequest struct {
	ContestID string `json:"contest_id" binding:"required,uuid4"`
}

type ContestUnregisterResponse struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	ContestID    string    `json:"contest_id"`
	RegisteredAt time.Time `json:"registered_at"`
	Status       string    `json:"status"`
}

type AllRegisteredContestForUserResponse struct {
	Registrations []ContestRegisterResponse `json:"registrations"`
}

type ContestRegisterRepository interface {
	CreateContestRegistration(ctx context.Context, userID string, contestID string) (*ContestRegistration, error)
	GetRegistrationByUserAndContest(ctx context.Context, userID string, contestID string) (*ContestRegistration, error)
	UpdateRegistrationStatus(ctx context.Context, userID string, contestID string, status string) error
	GetAllRegistrationsByUserID(ctx context.Context, userID string) ([]ContestRegistration, error)
	GetAllRegistrationsForAdmin(ctx context.Context) ([]ContestRegistration, error)
}

type ContestRegisterUseCase interface {
	RegisterContest(ctx context.Context, userID string, req *ContestRegisterRequest) (*ContestRegisterResponse, error)
	UnregisterContest(ctx context.Context, userID string, req *ContestUnregisterRequest) error
	GetAllRegistrations(ctx context.Context, userID string) (*AllRegisteredContestForUserResponse, error)
	GetAllRegistrationsForAdmin(ctx context.Context) (*AllRegisteredContestForUserResponse, error)
}
