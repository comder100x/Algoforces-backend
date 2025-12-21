package utils

import (
	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Error      string `json:"error"`
	Message    string `json:"message,omitempty"`
	StatusCode int    `json:"status_code"`
}

type SuccessResponse struct {
	Data       interface{} `json:"data"`
	Message    string      `json:"message,omitempty"`
	StatusCode int         `json:"status_code"`
}

// SendError sends an error response
func SendError(c *gin.Context, statusCode int, err error, message string) {
	c.JSON(statusCode, ErrorResponse{
		Error:      err.Error(),
		Message:    message,
		StatusCode: statusCode,
	})
}

// SendSuccess sends a success response
func SendSuccess(c *gin.Context, statusCode int, data interface{}, message string) {
	c.JSON(statusCode, SuccessResponse{
		Data:       data,
		Message:    message,
		StatusCode: statusCode,
	})
}
