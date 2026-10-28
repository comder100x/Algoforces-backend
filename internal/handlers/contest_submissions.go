package handlers

import (
	"algoforces/internal/middleware"
	"algoforces/internal/utils"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// ListContestSubmissions godoc
//
//	@Summary		List contest submissions
//	@Description	List submissions for a contest, newest first. Regular users see only their own submissions and never see hidden test case data; admin and problem-setter callers see all submissions for the contest, including hidden test case data.
//	@Tags			Submission
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		string	true	"Contest ID"
//	@Success		200	{object}	utils.SuccessResponse{data=object}
//	@Failure		400	{object}	utils.ErrorResponse
//	@Failure		401	{object}	utils.ErrorResponse
//	@Failure		403	{object}	utils.ErrorResponse
//	@Failure		500	{object}	utils.ErrorResponse
//	@Router			/api/contest/{id}/submissions [get]
func (h *SubmissionHandler) ListContestSubmissions(ctx *gin.Context) {
	contestID := ctx.Param("id")
	if contestID == "" {
		utils.SendError(ctx, http.StatusBadRequest, errors.New("contest id is required"), "Contest ID is required")
		return
	}

	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		utils.SendError(ctx, http.StatusUnauthorized, err, "User ID not found in token")
		return
	}
	userRole, err := middleware.GetUserRole(ctx)
	if err != nil {
		utils.SendError(ctx, http.StatusForbidden, err, "User role not found")
		return
	}

	submissions, err := h.submissionUseCase.ListContestSubmissions(ctx.Request.Context(), contestID, userID, userRole)
	if err != nil {
		log.Error().Err(err).Str("contest_id", contestID).Str("user_id", userID).Msg("Failed to list contest submissions")
		utils.SendError(ctx, http.StatusInternalServerError, err, "Failed to list contest submissions")
		return
	}

	utils.SendSuccess(ctx, http.StatusOK, submissions, "Contest submissions retrieved successfully")
}
