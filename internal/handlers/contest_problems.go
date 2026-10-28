package handlers

import (
	"algoforces/internal/domain"
	"algoforces/internal/middleware"
	"algoforces/internal/utils"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type ContestProblemsHandler struct {
	contestProblemsUseCase domain.ContestProblemsUseCase
}

func NewContestProblemsHandler(contestProblemsUseCase domain.ContestProblemsUseCase) *ContestProblemsHandler {
	return &ContestProblemsHandler{
		contestProblemsUseCase: contestProblemsUseCase,
	}
}

// CreateContestProblem godoc
//
//	@Summary		Add a problem to a contest
//	@Description	Add a problem to a contest with order position and max points (admin or problem-setter only)
//	@Tags			Contest Problems
//	@Accept			json
//	@Produce		json
//	@Param			createContestProblemRequest	body	domain.CreateContestProblemRequest	true	"Create Contest Problem Request"
//	@Security		BearerAuth
//	@Success		201	{object}	domain.CreateContestProblemResponse
//	@Failure		400	{object}	utils.ErrorResponse
//	@Failure		500	{object}	utils.ErrorResponse
//	@Router			/api/contest-problem/create [post]
func (h *ContestProblemsHandler) CreateContestProblem(c *gin.Context) {
	var createRequest domain.CreateContestProblemRequest
	if err := c.ShouldBindJSON(&createRequest); err != nil {
		utils.SendError(c, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	response, err := h.contestProblemsUseCase.CreateContestProblem(c.Request.Context(), &createRequest)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err, "Failed to create contest problem")
		return
	}

	utils.SendSuccess(c, http.StatusCreated, response, "Contest problem created successfully")
}

// BulkCreateContestProblems godoc
//
//	@Summary		Add multiple problems to a contest
//	@Description	Add multiple problems to a contest in bulk with order positions and max points (admin or problem-setter only)
//	@Tags			Contest Problems
//	@Accept			json
//	@Produce		json
//	@Param			bulkCreateRequest	body	domain.BulkCreateContestProblemRequest	true	"Bulk Create Contest Problems Request"
//	@Security		BearerAuth
//	@Success		201	{object}	domain.BulkCreateContestProblemResponse
//	@Failure		400	{object}	utils.ErrorResponse
//	@Failure		500	{object}	utils.ErrorResponse
//	@Router			/api/contest-problem/bulk [post]
func (h *ContestProblemsHandler) BulkCreateContestProblems(c *gin.Context) {
	var bulkRequest domain.BulkCreateContestProblemRequest
	if err := c.ShouldBindJSON(&bulkRequest); err != nil {
		utils.SendError(c, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	response, err := h.contestProblemsUseCase.BulkCreateContestProblems(c.Request.Context(), &bulkRequest)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err, "Failed to bulk create contest problems")
		return
	}

	utils.SendSuccess(c, http.StatusCreated, response, "Contest problems created in bulk successfully")
}

// GetContestProblem godoc
//
//	@Summary		Get a contest problem by ID
//	@Description	Get details of a specific contest problem by its unique ID
//	@Tags			Contest Problems
//	@Produce		json
//	@Param			id	path	string	true	"Contest Problem ID"
//	@Security		BearerAuth
//	@Success		200	{object}	domain.GetContestProblemResponse
//	@Failure		400	{object}	utils.ErrorResponse
//	@Failure		404	{object}	utils.ErrorResponse
//	@Router			/api/contest-problem/{id} [get]
func (h *ContestProblemsHandler) GetContestProblem(c *gin.Context) {
	uniqueID := c.Param("id")
	if uniqueID == "" {
		utils.SendError(c, http.StatusBadRequest, nil, "Contest problem ID is required")
		return
	}

	req := &domain.GetContestProblemRequest{UniqueID: uniqueID}
	response, err := h.contestProblemsUseCase.GetContestProblem(c.Request.Context(), req)
	if err != nil {
		utils.SendError(c, http.StatusNotFound, err, "Contest problem not found")
		return
	}

	utils.SendSuccess(c, http.StatusOK, response, "Contest problem retrieved successfully")
}

// GetContestProblems godoc
//
//	@Summary		Get all problems for a contest
//	@Description	Get all problems associated with a specific contest, under the same access gate as GET /api/contest/{id}/problems: admins, problem-setters, the contest creator, and the contest's listed problem-setters may view anytime; registered participants may view only during the live window. Everyone else receives 403. Problems are ordered by order position and include the problem statement; no test-case input/output is returned.
//	@Tags			Contest Problems
//	@Produce		json
//	@Param			contestId	path	string	true	"Contest ID"
//	@Security		BearerAuth
//	@Success		200	{object}	domain.GetContestProblemsResponse
//	@Failure		400	{object}	utils.ErrorResponse
//	@Failure		401	{object}	utils.ErrorResponse
//	@Failure		403	{object}	utils.ErrorResponse
//	@Failure		404	{object}	utils.ErrorResponse
//	@Failure		500	{object}	utils.ErrorResponse
//	@Router			/api/contest-problem/contest/{contestId} [get]
func (h *ContestProblemsHandler) GetContestProblems(c *gin.Context) {
	contestID := c.Param("contestId")
	if contestID == "" {
		utils.SendError(c, http.StatusBadRequest, errors.New("contest ID is required"), "Contest ID is required")
		return
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		utils.SendError(c, http.StatusUnauthorized, err, "User ID not found in token")
		return
	}

	role, err := middleware.GetUserRole(c)
	if err != nil {
		utils.SendError(c, http.StatusForbidden, err, "User role not found")
		return
	}

	// Same gate as the participant endpoint: an unregistered caller must
	// not fetch problems through this legacy route either.
	response, err := h.contestProblemsUseCase.GetContestProblemsForParticipant(
		c.Request.Context(),
		contestID,
		userID,
		role,
	)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrContestProblemsForbidden):
			utils.SendError(c, http.StatusForbidden, err, "You do not have access to this contest's problems")
		case errors.Is(err, domain.ErrContestNotFound):
			utils.SendError(c, http.StatusNotFound, err, "Contest not found")
		default:
			log.Error().
				Err(err).
				Str("contest_id", contestID).
				Str("user_id", userID).
				Msg("Failed to get contest problems")
			utils.SendError(
				c,
				http.StatusInternalServerError,
				errors.New("internal server error"),
				"Failed to get contest problems",
			)
		}
		return
	}

	utils.SendSuccess(c, http.StatusOK, response, "Contest problems retrieved successfully")
}

// UpdateContestProblem godoc
//
//	@Summary		Update a contest problem
//	@Description	Update an existing contest problem (admin or problem-setter only)
//	@Tags			Contest Problems
//	@Accept			json
//	@Produce		json
//	@Param			updateContestProblemRequest	body	domain.UpdateContestProblemRequest	true	"Update Contest Problem Request"
//	@Security		BearerAuth
//	@Success		200	{object}	domain.UpdateContestProblemResponse
//	@Failure		400	{object}	utils.ErrorResponse
//	@Failure		500	{object}	utils.ErrorResponse
//	@Router			/api/contest-problem/update [put]
func (h *ContestProblemsHandler) UpdateContestProblem(c *gin.Context) {
	var updateRequest domain.UpdateContestProblemRequest
	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		utils.SendError(c, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	response, err := h.contestProblemsUseCase.UpdateContestProblem(c.Request.Context(), &updateRequest)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err, "Failed to update contest problem")
		return
	}

	utils.SendSuccess(c, http.StatusOK, response, "Contest problem updated successfully")
}

// DeleteContestProblem godoc
//
//	@Summary		Delete a contest problem
//	@Description	Remove a problem from a contest (admin only)
//	@Tags			Contest Problems
//	@Produce		json
//	@Param			id	path	string	true	"Contest Problem ID"
//	@Security		BearerAuth
//	@Success		200	{object}	domain.DeleteContestProblemResponse
//	@Failure		400	{object}	utils.ErrorResponse
//	@Failure		500	{object}	utils.ErrorResponse
//	@Router			/api/contest-problem/{id} [delete]
func (h *ContestProblemsHandler) DeleteContestProblem(c *gin.Context) {
	uniqueID := c.Param("id")
	if uniqueID == "" {
		utils.SendError(c, http.StatusBadRequest, nil, "Contest problem ID is required")
		return
	}

	req := &domain.DeleteContestProblemRequest{UniqueID: uniqueID}
	response, err := h.contestProblemsUseCase.DeleteContestProblem(c.Request.Context(), req)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err, "Failed to delete contest problem")
		return
	}

	utils.SendSuccess(c, http.StatusOK, response, response.Message)
}
