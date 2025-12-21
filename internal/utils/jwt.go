package utils

import (
	"errors"
	"time"
	"algoforces/internal/conf"
	"github.com/golang-jwt/jwt/v5"
)


var jwtSecret = conf.JWT_SECRET


type JWTToken struct {
	UserID string `json:"user_id"`
	Role string `json:"role"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// GenerateToken generates a new JWT token
func GenerateToken(userId string, role string, email string) (string, error) {
	jwtPayload := JWTToken{
		UserID: userId,
		Role: role,
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtPayload)
	signedToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", errors.New("failed to generate token")
	}
	return signedToken, nil
}

func ValidateToken(tokenString string) (*JWTToken, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTToken{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return nil, errors.New("failed to parse token")
	}
	return token.Claims.(*JWTToken), nil
}