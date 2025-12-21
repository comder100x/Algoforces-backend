package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes the password using bcrypt
func HashPassword(password string) (string, error) {
	HashPassword, err :=bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(HashPassword), nil
}

// VerifyPassword checks if the password is correct
func VerifyPassword(password, hashedPassword string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}