package utils_test

import (
	"algoforces/internal/utils"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateToken(t *testing.T) {
	tests := []struct {
		name      string
		userId    string
		role      string
		email     string
		expectErr bool
	}{
		{
			name:      "Valid token generation",
			userId:    "user123",
			role:      "admin",
			email:     "admin@example.com",
			expectErr: false,
		},
		{
			name:      "Token with empty fields",
			userId:    "",
			role:      "",
			email:     "",
			expectErr: false,
		},
		{
			name:      "Token with special characters",
			userId:    "user@domain.com",
			role:      "moderator/editor",
			email:     "test+tag@example.com",
			expectErr: false,
		},
		{
			name:      "Token with unicode",
			userId:    "用户123",
			role:      "管理员",
			email:     "тест@example.com",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenString, err := utils.GenerateToken(tt.userId, tt.role, tt.email)

			if (err != nil) != tt.expectErr {
				t.Errorf("GenerateToken() error = %v, expectErr %v", err, tt.expectErr)
				return
			}

			if !tt.expectErr {
				if tokenString == "" {
					t.Error("GenerateToken() returned empty token")
					return
				}

				parts := strings.Split(tokenString, ".")
				if len(parts) != 3 {
					t.Errorf("GenerateToken() token has %d parts, expected 3", len(parts))
				}
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	userId := "testUser123"
	role := "admin"
	email := "test@example.com"

	tokenString, err := utils.GenerateToken(userId, role, email)
	if err != nil {
		t.Fatalf("Failed to generate token for test: %v", err)
	}

	tests := []struct {
		name      string
		token     string
		expectErr bool
		validate  func(*utils.JWTToken) bool
	}{
		{
			name:      "Valid token",
			token:     tokenString,
			expectErr: false,
			validate: func(claims *utils.JWTToken) bool {
				return claims.UserID == userId && claims.Role == role && claims.Email == email
			},
		},
		{
			name:      "Invalid token format",
			token:     "invalid.token.here",
			expectErr: true,
		},
		{
			name:      "Empty token",
			token:     "",
			expectErr: true,
		},
		{
			name:      "Malformed token",
			token:     "malformed",
			expectErr: true,
		},
		{
			name:      "Token with tampered signature",
			token:     tokenString + "tampered",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := utils.ValidateToken(tt.token)

			if (err != nil) != tt.expectErr {
				t.Errorf("ValidateToken() error = %v, expectErr %v", err, tt.expectErr)
				return
			}

			if !tt.expectErr && tt.validate != nil {
				if !tt.validate(claims) {
					t.Errorf("ValidateToken() claims validation failed. Got UserID=%s, Role=%s, Email=%s",
						claims.UserID, claims.Role, claims.Email)
				}
			}
		})
	}
}

func TestTokenExpiration(t *testing.T) {
	userId := "testUser"
	role := "user"
	email := "user@example.com"

	tokenString, err := utils.GenerateToken(userId, role, email)
	if err != nil {
		t.Fatalf("Failed to generatetoken: %v", err)
	}

	token, _ := jwt.ParseWithClaims(tokenString, &utils.JWTToken{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("dev-secret-key-change-in-production"), nil
	})

	claims := token.Claims.(*utils.JWTToken)

	now := time.Now()
	expiresAt := claims.ExpiresAt.Time
	duration := expiresAt.Sub(now)

	expectedDuration := 24 * time.Hour
	tolerance := 2 * time.Minute

	if duration > expectedDuration+tolerance || duration < expectedDuration-tolerance {
		t.Errorf("Token expiration = %v, expecting approximately %v", duration, expectedDuration)
	}
}

func TestTokenStructure(t *testing.T) {
	userId := "structUser"
	role := "admin"
	email := "struct@example.com"

	tokenString, _ := utils.GenerateToken(userId, role, email)
	claims, _ := utils.ValidateToken(tokenString)

	t.Run("Check UserID claim", func(t *testing.T) {
		if claims.UserID != userId {
			t.Errorf("UserID = %s, want %s", claims.UserID, userId)
		}
	})

	t.Run("Check Role claim", func(t *testing.T) {
		if claims.Role != role {
			t.Errorf("Role = %s, want %s", claims.Role, role)
		}
	})

	t.Run("Check Email claim", func(t *testing.T) {
		if claims.Email != email {
			t.Errorf("Email = %s, want %s", claims.Email, email)
		}
	})
}

func BenchmarkGenerateToken(b *testing.B) {
	userId := "benchUser"
	role := "admin"
	email := "bench@example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utils.GenerateToken(userId, role, email)
	}
}

func BenchmarkValidateToken(b *testing.B) {
	tokenString, _ := utils.GenerateToken("user", "admin", "test@example.com")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utils.ValidateToken(tokenString)
	}
}
