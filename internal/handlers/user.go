package handlers

import (
	"algoforces/internal/domain"
	"algoforces/internal/middleware"
	"algoforces/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userUseCase domain.UserUseCase
}

func NewUserHandler(userUseCase domain.UserUseCase) *UserHandler {
	return &UserHandler{
		userUseCase: userUseCase,
	}
}

func (h *UserHandler) GetUserProfile(c *gin.Context) {
	// Extract user ID from middleware context
	userID, err := middleware.GetUserID(c)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err, "Failed to get user ID")
		return
	}

	profileResponse, err := h.userUseCase.GetUserProfile(c.Request.Context(), userID)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err, "Failed to get user profile")
		return
	}

	utils.SendSuccess(c, http.StatusOK, profileResponse, "User profile fetched successfully")
}

func (h *UserHandler) UpdateUserProfile(c *gin.Context) {

	userID, err := middleware.GetUserID(c)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err, "Failed to get user ID")
		return
	}

	var updateUserProfileRequest domain.UpdateUserProfileRequest
	if err := c.ShouldBindJSON(&updateUserProfileRequest); err != nil {
		utils.SendError(c, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	updateProfileResponse, err := h.userUseCase.UpdateUserProfile(c.Request.Context(), userID, &updateUserProfileRequest)

	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err, "Failed to update user profile")
		return
	}

	utils.SendSuccess(c, http.StatusOK, updateProfileResponse, "User profile updated successfully")

}
