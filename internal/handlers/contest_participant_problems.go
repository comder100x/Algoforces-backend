package handlers

import (
	"algoforces/internal/domain"
	"algoforces/internal/middleware"
	"algoforces/internal/utils"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetContestProblemsForParticipant godoc
//
//	@Summary		Get contest problems for a participant
//	@Description	Get all problems for a contest as a participant. Admins, problem-setters, the contest creator, and the contest's listed problem-setters may view anytime. Registered participants may view only during the live window (start time to end time, while the contest is active). Everyone else receives 403. Problems are ordered by order position and include the problem statement; no test-case input/output is returned.
//	@Tags			Contest Problems
//	@Produce		json
//	@Param			id	path	string	true	"Contest ID"
//	@Security		BearerAuth
//	@Success		200	{object}	domain.GetContestProblemsResponse
//	@Failure		400	{object}	utils.ErrorResponse
//	@Failure		403	{object}	utils.ErrorResponse
//	@Failure		404	{object}	utils.ErrorResponse
//	@Router			/api/contest/{id}/problems [get]
func (h *ContestProblemsHandler) GetContestProblemsForParticipant(c *gin.Context) {
	contestID := c.Param("id")
	if contestID == "" {
		utils.SendError(c, http.StatusBadRequest, errors.New("contest ID is required"), "Contest ID is required")
		return
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err, "Failed to get user ID")
		return
	}

	role, err := middleware.GetUserRole(c)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err, "Failed to get user role")
		return
	}

	response, err := h.contestProblemsUseCase.GetContestProblemsForParticipant(c.Request.Context(), contestID, userID, role)
	if err != nil {
		if errors.Is(err, domain.ErrContestProblemsForbidden) {
			utils.SendError(c, http.StatusForbidden, err, "You do not have access to this contest's problems")
			return
		}
		utils.SendError(c, http.StatusNotFound, err, "Failed to get contest problems")
		return
	}

	utils.SendSuccess(c, http.StatusOK, response, "Contest problems retrieved successfully")
}
