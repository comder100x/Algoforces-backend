package middleware

import (
	"algoforces/internal/utils"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT token and sets user context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Authorization header required",
				"message": "Please provide a valid Bearer token",
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid authorization format",
				"message": "Authorization header must be in format: Bearer <token>",
			})
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := utils.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid or expired token",
				"message": "Please login again",
			})
			c.Abort()
			return
		}

		// Set user info in context for handlers to access
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)

		c.Next()
	}
}

// GetUserID extracts user ID from context
func GetUserID(c *gin.Context) (string, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", errors.New("user ID not found in context")
	}
	return userID.(string), nil
}

// GetUserEmail extracts user email from context
func GetUserEmail(c *gin.Context) string {
	userEmail, _ := c.Get("user_email")
	return userEmail.(string)
}

// GetUserRole extracts user role from context
func GetUserRole(c *gin.Context) string {
	userRole, _ := c.Get("user_role")
	return userRole.(string)
}
