package handlers_test

import (
	"algoforces/internal/handlers"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		expectStatus   int
		validateStatus bool
	}{
		{
			name:           "Health check returns OK",
			expectStatus:   http.StatusOK,
			validateStatus: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			handlers.GetHealth(c)

			if w.Code != tt.expectStatus {
				t.Errorf("Status code = %d, want %d", w.Code, tt.expectStatus)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if tt.validateStatus {
				if status, ok := response["status"].(string); !ok || status != "active" {
					t.Errorf("Expected status 'active', got %v", response["status"])
				}
				if code, ok := response["statusCode"].(float64); !ok || int(code) != http.StatusOK {
					t.Errorf("Expected statusCode 200, got %v", response["statusCode"])
				}
			}
		})
	}
}

func TestGetHealthResponseStructure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handlers.GetHealth(c)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	requiredFields := []string{"status", "message", "statusCode"}
	for _, field := range requiredFields {
		if _, ok := response[field]; !ok {
			t.Errorf("Response missing required field: %s", field)
		}
	}
}

func BenchmarkGetHealth(b *testing.B) {
	gin.SetMode(gin.TestMode)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		handlers.GetHealth(c)
	}
}

func TestHandlerContentType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handlers.GetHealth(c)

	contentType := w.Header().Get("Content-Type")
	expectedContentType := "application/json; charset=utf-8"

	if contentType != expectedContentType {
		t.Errorf("Content-Type = %s, want %s", contentType, expectedContentType)
	}
}
