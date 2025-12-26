package handlers

import (
	"algoforces/internal/domain"
	"algoforces/internal/middleware"
	"algoforces/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ContestHandler struct {
	contestUseCase domain.ContestUseCase
}

func NewContestHandler(contestUseCase domain.ContestUseCase) *ContestHandler {
	return &ContestHandler{
		contestUseCase: contestUseCase,
	}
}

// CreateContest godoc
// @Summary      Create a new Contest
// @Description  Create a new contest (admin or problem-setter only)
// @Tags         Contest
// @Accept       json
// @Produce      json
// @Param        createContestRequest  body      domain.CreateContestRequest  true  "Create Contest Request"
// @Security     BearerAuth
// @Success      201  {object}  domain.CreateContestResponse
// @Failure      400  {object}  utils.ErrorResponse
// @Failure      500  {object}  utils.ErrorResponse
// @Router       /api/contest/create [post]
func (h *ContestHandler) CreateContest(c *gin.Context) {
	var createContestRequest domain.CreateContestRequest
	if err := c.ShouldBindJSON(&createContestRequest); err != nil {
		utils.SendError(c, http.StatusBadRequest, err, "Invalid request body")
		return
	}
	userId, err := middleware.GetUserID(c)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err, "Failed to get user ID")
		return
	}

	contestResponse, err := h.contestUseCase.CreateContest(c.Request.Context(), &createContestRequest, userId)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err, "Failed to create contest")
		return
	}

	utils.SendSuccess(c, http.StatusCreated, contestResponse, "Contest created successfully")

}

// GetContestDetails godoc
// @Summary      Get Contest Details
// @Description  Get details of a specific contest by ID
// @Tags         Contest
// @Produce      json
// @Param        id   path      string  true  "Contest ID"
// @Security     BearerAuth
// @Success      200  {object}  domain.CreateContestResponse
// @Failure      400  {object}  utils.ErrorResponse
// @Failure      404  {object}  utils.ErrorResponse
// @Router       /api/contest/{id} [get]
func (h *ContestHandler) GetContestDetails(c *gin.Context) {
	contestId := c.Param("id")
	if contestId == "" {
		utils.SendError(c, http.StatusBadRequest, nil, "Contest ID is required")
		return
	}

	contestResponse, err := h.contestUseCase.GetContestDetails(c.Request.Context(), contestId)
	if err != nil {
		utils.SendError(c, http.StatusNotFound, err, "Contest not found")
		return
	}

	utils.SendSuccess(c, http.StatusOK, contestResponse, "Contest retrieved successfully")
}

// UpdateContest godoc
// @Summary      Update a Contest
// @Description  Update an existing contest (admin or problem-setter only)
// @Tags         Contest
// @Accept       json
// @Produce      json
// @Param        updateContestRequest  body      domain.UpdateContestRequest  true  "Update Contest Request"
// @Security     BearerAuth
// @Success      200  {object}  domain.UpdateContestResponse
// @Failure      400  {object}  utils.ErrorResponse
// @Failure      500  {object}  utils.ErrorResponse
// @Router       /api/contest/update [put]
func (h *ContestHandler) UpdateContest(c *gin.Context) {
	var updateContestRequest domain.UpdateContestRequest
	if err := c.ShouldBindJSON(&updateContestRequest); err != nil {
		utils.SendError(c, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	contestResponse, err := h.contestUseCase.UpdateContest(c.Request.Context(), &updateContestRequest)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err, "Failed to update contest")
		return
	}

	utils.SendSuccess(c, http.StatusOK, contestResponse, "Contest updated successfully")
}

// DeleteContest godoc
// @Summary      Delete a Contest
// @Description  Delete a contest by ID (admin only)
// @Tags         Contest
// @Produce      json
// @Param        id   path      string  true  "Contest ID"
// @Security     BearerAuth
// @Success      200  {object}  utils.SuccessResponse
// @Failure      400  {object}  utils.ErrorResponse
// @Failure      500  {object}  utils.ErrorResponse
// @Router       /api/contest/{id} [delete]
func (h *ContestHandler) DeleteContest(c *gin.Context) {
	contestId := c.Param("id")
	if contestId == "" {
		utils.SendError(c, http.StatusBadRequest, nil, "Contest ID is required")
		return
	}

	deleteRequest := &domain.DeleteContestRequest{Id: contestId}
	err := h.contestUseCase.DeleteContest(c.Request.Context(), deleteRequest)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err, "Failed to delete contest")
		return
	}

	utils.SendSuccess(c, http.StatusOK, nil, "Contest deleted successfully")
}

// GetAllContests godoc
// @Summary      Get All Contests (Admin)
// @Description  Get all contests (admin only)
// @Tags         Contest
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}  domain.CreateContestResponse
// @Failure      500  {object}  utils.ErrorResponse
// @Router       /api/admin/contests [get]
func (h *ContestHandler) GetAllContests(c *gin.Context) {
	contests, err := h.contestUseCase.GetAllContests(c.Request.Context())
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err, "Failed to get contests")
		return
	}

	utils.SendSuccess(c, http.StatusOK, contests, "Contests retrieved successfully")
}
