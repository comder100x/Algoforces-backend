package handlers

import (
	"algoforces/internal/domain"
	"algoforces/internal/middleware"
	"algoforces/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ContestRegisterHandler struct {
	contestRegisterUseCase domain.ContestRegisterUseCase
}

func NewContestRegisterHandler(contestRegisterUseCase domain.ContestRegisterUseCase) *ContestRegisterHandler {
	return &ContestRegisterHandler{
		contestRegisterUseCase: contestRegisterUseCase,
	}
}

// RegisterContest godoc
// @Summary      Register for a contest
// @Description  Register the current user for a contest
// @Tags         contest-registration
// @Accept       json
// @Produce      json
// @Param        request body domain.ContestRegisterRequest true "Contest registration request"
// @Success      201 {object} domain.ContestRegisterResponse
// @Failure      400 {object} utils.ErrorResponse
// @Failure      500 {object} utils.ErrorResponse
// @Security     BearerAuth
// @Router       /api/contest/register [post]
func (h *ContestRegisterHandler) RegisterContest(c *gin.Context) {
	var contestRegisterRequest domain.ContestRegisterRequest
	if err := c.ShouldBindJSON(&contestRegisterRequest); err != nil {
		utils.SendError(c, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	// get user id from middleware
	userID, err := middleware.GetUserID(c)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err, "Failed to get user ID")
		return
	}

	contestRegisterResponse, err := h.contestRegisterUseCase.RegisterContest(c.Request.Context(), userID, &contestRegisterRequest)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err, "Failed to register for contest")
		return
	}

	utils.SendSuccess(c, http.StatusCreated, contestRegisterResponse, "Registered for contest successfully")
}

// UnregisterContest godoc
// @Summary      Unregister from a contest
// @Description  Unregister the current user from a contest
// @Tags         contest-registration
// @Accept       json
// @Produce      json
// @Param        request body domain.ContestUnregisterRequest true "Contest unregister request"
// @Success      200 {object} utils.SuccessResponse
// @Failure      400 {object} utils.ErrorResponse
// @Failure      500 {object} utils.ErrorResponse
// @Security     BearerAuth
// @Router       /api/contest/unregister [post]
func (h *ContestRegisterHandler) UnregisterContest(c *gin.Context) {
	var contestUnregisterRequest domain.ContestUnregisterRequest
	if err := c.ShouldBindJSON(&contestUnregisterRequest); err != nil {
		utils.SendError(c, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	// get user id from middleware
	userID, err := middleware.GetUserID(c)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err, "Failed to get user ID")
		return
	}

	err = h.contestRegisterUseCase.UnregisterContest(c.Request.Context(), userID, &contestUnregisterRequest)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err, "Failed to unregister from contest")
		return
	}

	utils.SendSuccess(c, http.StatusOK, nil, "Unregistered from contest successfully")
}

// GetAllRegistrations godoc
// @Summary      Get all contest registrations
// @Description  Get all contests the current user is registered for
// @Tags         contest-registration
// @Accept       json
// @Produce      json
// @Success      200 {object} domain.AllRegisteredContestForUserResponse
// @Failure      500 {object} utils.ErrorResponse
// @Security     BearerAuth
// @Router       /api/contest/registrations [get]
func (h *ContestRegisterHandler) GetAllRegistrations(c *gin.Context) {
	// get user id from middleware
	userID, err := middleware.GetUserID(c)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err, "Failed to get user ID")
		return
	}

	allRegistrationsResponse, err := h.contestRegisterUseCase.GetAllRegistrations(c.Request.Context(), userID)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err, "Failed to get registrations")
		return
	}

	utils.SendSuccess(c, http.StatusOK, allRegistrationsResponse, "Retrieved all registrations successfully")
}

// GetAllRegistrationsForAdmin godoc
// @Summary      Get all contest registrations (Admin)
// @Description  Get all contest registrations for all users (admin only)
// @Tags         contest-registration
// @Accept       json
// @Produce      json
// @Success      200 {object} domain.AllRegisteredContestForUserResponse
// @Failure      500 {object} utils.ErrorResponse
// @Security     BearerAuth
// @Router       /api/admin/registrations [get]
func (h *ContestRegisterHandler) GetAllRegistrationsForAdmin(c *gin.Context) {
	allRegistrationsResponse, err := h.contestRegisterUseCase.GetAllRegistrationsForAdmin(c.Request.Context())
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err, "Failed to get all registrations")
		return
	}

	utils.SendSuccess(c, http.StatusOK, allRegistrationsResponse, "Retrieved all registrations successfully")
}