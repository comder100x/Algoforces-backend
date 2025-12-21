package domain

import (
	"context"
	"time"
)

// Actual reponse returned to user initially
type User struct {
	Id        string    `json:"id" gorm:"primaryKey;type:uuid"`
	Email     string    `json:"email" gorm:"uniqueIndex;not null"` // Fix json tag
	Password  string    `json:"-" gorm:"not null"`
	Role      string    `json:"role" gorm:"default:user"`
	Username  string    `json:"username" gorm:""`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// SignupRequest is what the frontend sends to register.
type SignupRequest struct {
	Username string `json:"username" binding:"required,min=3"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest is what the frontend sends to log in.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse is what we send back after successful login/signup.
type AuthResponse struct {
	AccessToken string `json:"access_token"`
	User        User   `json:"user"` // Returns user info (without password)
}

type UserProfileResponse struct {
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type UpdateUserProfileRequest struct {
	Username string `json:"username" binding:"required,min=3"`
	Email    string `json:"email" binding:"required,email"`
}

// UserRepository defines how we talk to the Database.
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	UpdateByID(ctx context.Context, id string, user *User) error
}

// UserUseCase defines the business logic (Hashing passwords, Generating tokens).
type UserUseCase interface {
	Signup(ctx context.Context, req *SignupRequest) (*AuthResponse, error)
	Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error)
	GetUserProfile(ctx context.Context, userID string) (*UserProfileResponse, error)
	UpdateUserProfile(ctx context.Context, userID string, req *UpdateUserProfileRequest) (*UserProfileResponse, error)
}
