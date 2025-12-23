package domain

import "context"

type AddRoleRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required,oneof=admin user problem_setter"`
}

type RemoveRoleRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required,oneof=admin user problem_setter"`
}

type AddRoleResponse struct {
	Email       string `json:"email"`
	Role        string `json:"role"`
	CurrentRole string `json:"current_role"`
}

type RemoveRoleResponse struct {
	Email       string `json:"email"`
	Role        string `json:"role"`
	CurrentRole string `json:"current_role"`
}

// UserListItem represents a user in the list response
type UserListItem struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
}

// UserListResponse is the response for listing users
type UserListResponse struct {
	Users []UserListItem `json:"users"`
	Total int            `json:"total"`
}

// AdminUseCase defines the business logic for admin operations.
type AdminUseCase interface {
	AddRole(ctx context.Context, req *AddRoleRequest) (*AddRoleResponse, error)
	RemoveRole(ctx context.Context, req *RemoveRoleRequest) (*RemoveRoleResponse, error)
	GetAllUsers(ctx context.Context) (*UserListResponse, error)
	GetUsersByRole(ctx context.Context, role string) (*UserListResponse, error)
}

// AdminRepository defines how we talk to the Database for admin operations.
type AdminRepository interface {
	UserRepository
	UpdateByEmail(ctx context.Context, email string, user *User) error
	GetAllUsers(ctx context.Context) ([]User, error)
	GetUsersByRole(ctx context.Context, role string) ([]User, error)
}
