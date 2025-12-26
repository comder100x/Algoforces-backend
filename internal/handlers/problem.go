package handlers

import (
	"algoforces/internal/domain"
	"algoforces/internal/middleware"
	"algoforces/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProblemHandler struct {
	problemUseCase domain.ProblemUseCase
}

func NewProblemHandler(problemUseCase domain.ProblemUseCase) *ProblemHandler {
	return &ProblemHandler{
		problemUseCase: problemUseCase,
	}
}

// CreateProblem godoc
//
//	@Summary		Create a new Problem
//	@Description	Create a new problem (admin or problem-setter only)
//	@Tags			Problem
//	@Accept			json
//	@Produce		json
//	@Param			problemRequest	body	domain.ProblemCreationRequest	true	"Problem Creation Request"
//	@Security		BearerAuth
//	@Success		201	{object}	domain.ProblemCreationResponse
//	@Failure		400	{object}	utils.ErrorResponse
//	@Failure		403	{object}	utils.ErrorResponse
//	@Failure		500	{object}	utils.ErrorResponse
//	@Router			/api/problem/create [post]
func (h *ProblemHandler) CreateProblem(c *gin.Context) {
	var problemRequest domain.ProblemCreationRequest
	if err := c.ShouldBindJSON(&problemRequest); err != nil {
		utils.SendError(c, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err, "Failed to get user ID")
		return
	}

	problemResponse, err := h.problemUseCase.CreateProblem(c.Request.Context(), &problemRequest, userID)
	if err != nil {
		if err.Error() == "user does not have permission to create problems" {
			utils.SendError(c, http.StatusForbidden, err, err.Error())
			return
		}
		utils.SendError(c, http.StatusInternalServerError, err, "Failed to create problem")
		return
	}

	utils.SendSuccess(c, http.StatusCreated, problemResponse, "Problem created successfully")
}

// CreateProblemsInBulk godoc
//
//	@Summary		Create Multiple Problems in Bulk
//	@Description	Create multiple problems at once (admin or problem-setter only)
//	@Tags			Problem
//	@Accept			json
//	@Produce		json
//	@Param			bulkRequest	body	domain.BulkProblemCreationRequest	true	"Bulk Problem Creation Request"
//	@Security		BearerAuth
//	@Success		201	{object}	domain.BulkProblemCreationResponse
//	@Failure		400	{object}	utils.ErrorResponse
//	@Failure		403	{object}	utils.ErrorResponse
//	@Failure		500	{object}	utils.ErrorResponse
//	@Router			/api/problem/bulk [post]
func (h *ProblemHandler) CreateProblemsInBulk(c *gin.Context) {
	var bulkRequest domain.BulkProblemCreationRequest
	if err := c.ShouldBindJSON(&bulkRequest); err != nil {
		utils.SendError(c, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err, "Failed to get user ID")
		return
	}

	bulkResponse, err := h.problemUseCase.CreateProblemsInBulk(c.Request.Context(), &bulkRequest, userID)
	if err != nil {
		if err.Error() == "user does not have permission to create problems" {
			utils.SendError(c, http.StatusForbidden, err, err.Error())
			return
		}
		utils.SendError(c, http.StatusInternalServerError, err, "Failed to create problems in bulk")
		return
	}

	utils.SendSuccess(c, http.StatusCreated, bulkResponse, "Bulk problem creation completed")
}

// GetProblemByID godoc
//
//	@Summary		Get Problem by ID
//	@Description	Get a specific problem by its unique ID
//	@Tags			Problem
//	@Produce		json
//	@Param			id	path	string	true	"Problem ID"
//	@Security		BearerAuth
//	@Success		200	{object}	domain.ProblemCreationResponse
//	@Failure		404	{object}	utils.ErrorResponse
//	@Failure		500	{object}	utils.ErrorResponse
//	@Router			/api/problem/{id} [get]
func (h *ProblemHandler) GetProblemByID(c *gin.Context) {
	problemID := c.Param("id")
	if problemID == "" {
		utils.SendError(c, http.StatusBadRequest, nil, "Problem ID is required")
		return
	}

	problemResponse, err := h.problemUseCase.GetProblemByID(c.Request.Context(), problemID)
	if err != nil {
		utils.SendError(c, http.StatusNotFound, err, "Problem not found")
		return
	}

	utils.SendSuccess(c, http.StatusOK, problemResponse, "Problem retrieved successfully")
}

// UpdateProblem godoc
//
//	@Summary		Update a Problem
//	@Description	Update an existing problem (admin or creator only)
//	@Tags			Problem
//	@Accept			json
//	@Produce		json
//	@Param			problemRequest	body	domain.ProblemUpdateRequest	true	"Problem Update Request"
//	@Security		BearerAuth
//	@Success		200	{object}	domain.ProblemUpdateResponse
//	@Failure		400	{object}	utils.ErrorResponse
//	@Failure		403	{object}	utils.ErrorResponse
//	@Failure		404	{object}	utils.ErrorResponse
//	@Failure		500	{object}	utils.ErrorResponse
//	@Router			/api/problem/update [put]
func (h *ProblemHandler) UpdateProblem(c *gin.Context) {
	var problemRequest domain.ProblemUpdateRequest
	if err := c.ShouldBindJSON(&problemRequest); err != nil {
		utils.SendError(c, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err, "Failed to get user ID")
		return
	}

	problemResponse, err := h.problemUseCase.UpdateProblem(c.Request.Context(), &problemRequest, userID)
	if err != nil {
		if err.Error() == "user can only update their own problems" || err.Error() == "user does not have permission to update problems" {
			utils.SendError(c, http.StatusForbidden, err, err.Error())
			return
		}
		if err.Error() == "problem not found" {
			utils.SendError(c, http.StatusNotFound, err, err.Error())
			return
		}
		utils.SendError(c, http.StatusInternalServerError, err, "Failed to update problem")
		return
	}

	utils.SendSuccess(c, http.StatusOK, problemResponse, "Problem updated successfully")
}

// DeleteProblem godoc
//
//	@Summary		Delete a Problem
//	@Description	Delete a problem by ID (admin or creator only)
//	@Tags			Problem
//	@Produce		json
//	@Param			id	path	string	true	"Problem ID"
//	@Security		BearerAuth
//	@Success		200	{object}	utils.SuccessResponse
//	@Failure		400	{object}	utils.ErrorResponse
//	@Failure		403	{object}	utils.ErrorResponse
//	@Failure		404	{object}	utils.ErrorResponse
//	@Failure		500	{object}	utils.ErrorResponse
//	@Router			/api/problem/{id} [delete]
func (h *ProblemHandler) DeleteProblem(c *gin.Context) {
	problemID := c.Param("id")
	if problemID == "" {
		utils.SendError(c, http.StatusBadRequest, nil, "Problem ID is required")
		return
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err, "Failed to get user ID")
		return
	}

	err = h.problemUseCase.DeleteProblem(c.Request.Context(), problemID, userID)
	if err != nil {
		if err.Error() == "user can only delete their own problems" || err.Error() == "user does not have permission to delete problems" {
			utils.SendError(c, http.StatusForbidden, err, err.Error())
			return
		}
		if err.Error() == "problem not found" {
			utils.SendError(c, http.StatusNotFound, err, err.Error())
			return
		}
		utils.SendError(c, http.StatusInternalServerError, err, "Failed to delete problem")
		return
	}

	utils.SendSuccess(c, http.StatusOK, nil, "Problem deleted successfully")
}

// GetAllProblems godoc
//
//	@Summary		Get All Problems
//	@Description	Get all problems (authenticated users only)
//	@Tags			Problem
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{array}		domain.ProblemCreationResponse
//	@Failure		500	{object}	utils.ErrorResponse
//	@Router			/api/problem/all [get]
func (h *ProblemHandler) GetAllProblems(c *gin.Context) {
	problems, err := h.problemUseCase.GetAllProblems(c.Request.Context())
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err, "Failed to get problems")
		return
	}

	utils.SendSuccess(c, http.StatusOK, problems, "Problems retrieved successfully")
}
