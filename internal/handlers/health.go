package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":     "active",
		"message":    "Algoforces API is running smoothly",
		"statusCode": http.StatusOK,
	})
}
