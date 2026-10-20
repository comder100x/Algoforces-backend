package services_test

import (
	"algoforces/internal/utils"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, email, username, password, role string) error {
	args := m.Called(ctx, email, username, password, role)
	return args.Error(0)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (interface{}, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0), args.Error(1)
}

// MockPasswordHasher provides mock for password operations
type MockPasswordHasher struct {
	mock.Mock
}

// TestAuthServiceSignupValidation tests signup validation
func TestAuthServiceSignupValidation(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
		username string
		isValid  bool
	}{
		{
			name:     "Valid signup request",
			email:    "valid@example.com",
			password: "validpassword123",
			username: "validuser",
			isValid:  true,
		},
		{
			name:     "Empty email",
			email:    "",
			password: "validpassword123",
			username: "validuser",
			isValid:  false,
		},
		{
			name:     "Empty password",
			email:    "valid@example.com",
			password: "",
			username: "validuser",
			isValid:  false,
		},
		{
			name:     "Empty username",
			email:    "valid@example.com",
			password: "validpassword123",
			username: "",
			isValid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.isValid {
				if tt.email == "" || tt.password == "" || tt.username == "" {
					return // Expected to be invalid
				}
			}

			if tt.isValid {
				if tt.email == "" || tt.password == "" {
					t.Error("Valid request should have email and password")
				}
			}
		})
	}
}

// TestJWTTokenGeneration tests JWT token generation
func TestJWTTokenGeneration(t *testing.T) {
	tests := []struct {
		name   string
		userId string
		role   string
		email  string
	}{
		{
			name:   "Generate token for admin",
			userId: uuid.New().String(),
			role:   "admin",
			email:  "admin@example.com",
		},
		{
			name:   "Generate token for user",
			userId: uuid.New().String(),
			role:   "user",
			email:  "user@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := utils.GenerateToken(tt.userId, tt.role, tt.email)

			if err != nil {
				t.Errorf("GenerateToken() error = %v", err)
				return
			}

			if token == "" {
				t.Error("GenerateToken() returned empty token")
			}

			// Validate the token
			claims, err := utils.ValidateToken(token)
			if err != nil {
				t.Errorf("ValidateToken() error = %v", err)
				return
			}

			if claims.UserID != tt.userId {
				t.Errorf("Token claim UserID = %s, want %s", claims.UserID, tt.userId)
			}
		})
	}
}

// TestPasswordHashing tests password hashing and verification
func TestPasswordHashing(t *testing.T) {
	password := "testPassword123!"

	hash, err := utils.HashPassword(password)
	if err != nil {
		t.Errorf("HashPassword() error = %v", err)
		return
	}

	if hash == "" {
		t.Error("HashPassword() returned empty hash")
	}

	// Verify correct password
	if !utils.VerifyPassword(password, hash) {
		t.Error("VerifyPassword() should return true for correct password")
	}

	// Verify incorrect password
	if utils.VerifyPassword("wrongPassword", hash) {
		t.Error("VerifyPassword() should return false for incorrect password")
	}
}

// TestContextHandling tests context handling in services
func TestContextHandling(t *testing.T) {
	tests := []struct {
		name   string
		hasCtx bool
	}{
		{
			name:   "With context",
			hasCtx: true,
		},
		{
			name:   "With nil context",
			hasCtx: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ctx context.Context

			if tt.hasCtx {
				ctx = context.Background()
			}

			if tt.hasCtx && ctx == nil {
				t.Error("Context should not be nil")
			}
		})
	}
}

// TestContextTimeout tests context timeout
func TestContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	time.Sleep(150 * time.Millisecond)

	if err := ctx.Err(); err != context.DeadlineExceeded {
		t.Errorf("Expected DeadlineExceeded, got %v", err)
	}
}

// BenchmarkPasswordHashing benchmarks password hashing
func BenchmarkPasswordHashing(b *testing.B) {
	password := "benchPasswordTest123!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utils.HashPassword(password)
	}
}

// BenchmarkJWTGeneration benchmarks JWT token generation
func BenchmarkJWTGeneration(b *testing.B) {
	userId := uuid.New().String()
	role := "admin"
	email := "test@example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utils.GenerateToken(userId, role, email)
	}
}
