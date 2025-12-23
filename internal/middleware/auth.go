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

		var token string
		// Support both "Bearer <token>" and just "<token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			token = parts[1]
		} else if len(parts) == 1 {
			// Token provided without Bearer prefix
			token = parts[0]
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid authorization format",
				"message": "Authorization header must be a token or in format: Bearer <token>",
			})
			c.Abort()
			return
		}

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

func RoleMiddleware(allowedRole ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, err := GetUserRole(c)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "User role not found",
				"message": "Access denied",
			})
			c.Abort()
			return
		}
		// Check if user has ANY of the allowed roles
		for _, role := range allowedRole {
			if userRole == role {
				c.Next() // Role matched, allow access
				return
			}
		}

		// No matching role found
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Insufficient permissions",
			"message": "You do not have access to this resource",
		})
		c.Abort()
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
func GetUserRole(c *gin.Context) (string, error) {
	userRole, exists := c.Get("user_role")
	if !exists {
		return "", errors.New("user role not found in context")
	}
	return userRole.(string), nil
}
