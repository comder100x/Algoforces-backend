package utils_test

import (
	"algoforces/internal/utils"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSendError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		err        error
		message    string
		validate   func(*utils.ErrorResponse) bool
	}{
		{
			name:       "Bad request error",
			statusCode: http.StatusBadRequest,
			err:        &errTest{msg: "invalid input"},
			message:    "Validation failed",
			validate: func(resp *utils.ErrorResponse) bool {
				return resp.StatusCode == http.StatusBadRequest &&
					resp.Error == "invalid input" &&
					resp.Message == "Validation failed"
			},
		},
		{
			name:       "Unauthorized error",
			statusCode: http.StatusUnauthorized,
			err:        &errTest{msg: "invalid credentials"},
			message:    "Authentication required",
			validate: func(resp *utils.ErrorResponse) bool {
				return resp.StatusCode == http.StatusUnauthorized &&
					resp.Error == "invalid credentials"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			utils.SendError(c, tt.statusCode, tt.err, tt.message)

			if w.Code != tt.statusCode {
				t.Errorf("Status code = %d, want %d", w.Code, tt.statusCode)
			}

			var response utils.ErrorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if !tt.validate(&response) {
				t.Errorf("Response validation failed. Got: %+v", response)
			}
		})
	}
}

func TestSendSuccess(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		data       interface{}
		message    string
		validate   func(*utils.SuccessResponse) bool
	}{
		{
			name:       "Success with data",
			statusCode: http.StatusOK,
			data:       map[string]interface{}{"id": 1, "name": "test"},
			message:    "Request successful",
			validate: func(resp *utils.SuccessResponse) bool {
				return resp.StatusCode == http.StatusOK &&
					resp.Message == "Request successful" &&
					resp.Data != nil
			},
		},
		{
			name:       "Success with nil data",
			statusCode: http.StatusNoContent,
			data:       nil,
			message:    "Operation completed",
			validate: func(resp *utils.SuccessResponse) bool {
				return resp.StatusCode == http.StatusNoContent
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			utils.SendSuccess(c, tt.statusCode, tt.data, tt.message)

			if w.Code != tt.statusCode {
				t.Errorf("Status code = %d, want %d", w.Code, tt.statusCode)
			}

			if w.Body.Len() > 0 {
				var response utils.SuccessResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if !tt.validate(&response) {
					t.Errorf("Response validation failed. Got: %+v", response)
				}
			} else if tt.statusCode != http.StatusNoContent {
				t.Error("Expected response body but got empty")
			}
		})
	}
}

type errTest struct {
	msg string
}

func (e *errTest) Error() string {
	return e.msg
}

func BenchmarkSendError(b *testing.B) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	err := &errTest{msg: "benchmark error"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utils.SendError(c, http.StatusBadRequest, err, "Test message")
	}
}

func BenchmarkSendSuccess(b *testing.B) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	data := map[string]interface{}{"id": 1, "name": "test"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utils.SendSuccess(c, http.StatusOK, data, "Test message")
	}
}
