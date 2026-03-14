package middleware_test

import (
	"algoforces/internal/middleware"
	"algoforces/internal/utils"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAuthMiddlewareValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	token, err := utils.GenerateToken("user123", "admin", "admin@example.com")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	tests := []struct {
		name      string
		authHeader string
		expectAuth bool
		expectStatus int
	}{
		{
			name:      "Valid Bearer token",
			authHeader: "Bearer " + token,
			expectAuth: true,
			expectStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(middleware.AuthMiddleware())
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req, _ := http.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", tt.authHeader)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectStatus {
				t.Errorf("Expected status %d, got %d", tt.expectStatus, w.Code)
			}
		})
	}
}

func TestAuthMiddlewareInvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name      string
		authHeader string
		expectStatus int
	}{
		{
			name:      "No authorization header",
			authHeader: "",
			expectStatus: http.StatusUnauthorized,
		},
		{
			name:      "Invalid token",
			authHeader: "Bearer invalidtoken",
			expectStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(middleware.AuthMiddleware())
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req, _ := http.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectStatus {
				t.Errorf("Expected status %d, got %d", tt.expectStatus, w.Code)
			}
		})
	}
}

func TestGetUserID(t *testing.T) {
	tests := []struct {
		name    string
		userId  string
		hasError bool
	}{
		{
			name:    "Get existing user ID",
			userId:  "user123",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())

			if tt.userId != "" {
				c.Set("user_id", tt.userId)
			}

			id, err := middleware.GetUserID(c)

			if (err != nil) != tt.hasError {
				t.Errorf("Expected error=%v, got error=%v", tt.hasError, err != nil)
			}

			if !tt.hasError && id != tt.userId {
				t.Errorf("Expected user ID %s, got %s", tt.userId, id)
			}
		})
	}
}

func BenchmarkAuthMiddleware(b *testing.B) {
	gin.SetMode(gin.TestMode)

	token, _ := utils.GenerateToken("user", "admin", "test@example.com")
	router := gin.New()
	router.Use(middleware.AuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
	}
}
