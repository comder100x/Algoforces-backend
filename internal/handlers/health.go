package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetHealth godoc
// @Summary      Health Check
// @Description  Get the health status of the API
// @Tags         Health
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /api/health [get]
func GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":     "active",
		"message":    "Algoforces API is running smoothly",
		"statusCode": http.StatusOK,
	})
}
