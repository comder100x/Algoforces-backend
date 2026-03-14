package utils_test

import (
	"algoforces/internal/utils"
	"testing"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		expectError bool
	}{
		{
			name:        "Valid password hash",
			password:    "mySecurePassword123!",
			expectError: false,
		},
		{
			name:        "Empty password",
			password:    "",
			expectError: false,
		},
		{
			name:        "Long password (within bcrypt limit)",
			password:    "this_is_a_long_password_within_bcrypt_limit_123456789",
			expectError: false,
		},
		{
			name:        "Password exceeding bcrypt limit (72 bytes)",
			password:    "this_is_a_very_long_password_that_exceeds_bcrypt_72_byte_limit_1234567890abcdefghijklmnop",
			expectError: true,
		},
		{
			name:        "Password with special characters",
			password:    "P@$$w0rd!#%^&*()",
			expectError: false,
		},
		{
			name:        "Password with unicode",
			password:    "пароль密码🔒",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := utils.HashPassword(tt.password)

			if (err != nil) != tt.expectError {
				t.Errorf("HashPassword() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError && hash == "" {
				t.Error("HashPassword() returned empty hash for valid password")
			}

			if hash == tt.password {
				t.Error("HashPassword() hash equals original password")
			}
		})
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "testPassword123!"
	hash, err := utils.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password for test: %v", err)
	}

	tests := []struct {
		name           string
		password       string
		hash           string
		expectMatch    bool
		expectError    bool
	}{
		{
			name:        "Correct password",
			password:    password,
			hash:        hash,
			expectMatch: true,
			expectError: false,
		},
		{
			name:        "Wrong password",
			password:    "wrongPassword123!",
			hash:        hash,
			expectMatch: false,
			expectError: false,
		},
		{
			name:        "Empty password vs non-empty hash",
			password:    "",
			hash:        hash,
			expectMatch: false,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.VerifyPassword(tt.password, tt.hash)

			if result != tt.expectMatch {
				t.Errorf("VerifyPassword() = %v, expect %v", result, tt.expectMatch)
			}
		})
	}
}

func TestHashPasswordDifferentHashes(t *testing.T) {
	password := "samePassword123!"

	hash1, _ := utils.HashPassword(password)
	hash2, _ := utils.HashPassword(password)

	if hash1 == hash2 {
		t.Error("HashPassword() should generate different hashes for same password")
	}

	if !utils.VerifyPassword(password, hash1) || !utils.VerifyPassword(password, hash2) {
		t.Error("VerifyPassword() failed to verify hashes generated from same password")
	}
}

func TestVerifyPasswordWithInvalidHash(t *testing.T) {
	password := "myPassword"
	invalidHash := "notAValidBcryptHash"

	result := utils.VerifyPassword(password, invalidHash)

	if result {
		t.Error("VerifyPassword() should return false for invalid hash")
	}
}

func BenchmarkHashPassword(b *testing.B) {
	password := "benchmarkPassword123!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utils.HashPassword(password)
	}
}

func BenchmarkVerifyPassword(b *testing.B) {
	password := "benchmarkPassword123!"
	hash, _ := utils.HashPassword(password)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utils.VerifyPassword(password, hash)
	}
}
