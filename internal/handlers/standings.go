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

type StandingsHandler struct {
	standingsUseCase domain.StandingsUseCase
}

func NewStandingsHandler(standingsUseCase domain.StandingsUseCase) *StandingsHandler {
	return &StandingsHandler{
		standingsUseCase: standingsUseCase,
	}
}

// GetStandings godoc
//
//	@Summary		Get contest standings
//	@Description	Get the leaderboard for a contest. Before the contest starts, only admins and problem setters may view it; once live or ended, any authenticated user may.
//	@Tags			Standings
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		string	true	"Contest ID"
//	@Success		200	{object}	utils.SuccessResponse{data=domain.StandingsResponse}
//	@Failure		400	{object}	utils.ErrorResponse
//	@Failure		401	{object}	utils.ErrorResponse
//	@Failure		403	{object}	utils.ErrorResponse
//	@Failure		404	{object}	utils.ErrorResponse
//	@Failure		500	{object}	utils.ErrorResponse
//	@Router			/api/contest/{id}/standings [get]
func (h *StandingsHandler) GetStandings(ctx *gin.Context) {
	if h == nil || h.standingsUseCase == nil {
		utils.SendError(ctx, http.StatusInternalServerError, errors.New("internal server error"), "Failed to get contest standings")
		return
	}

	contestID := ctx.Param("id")
	if contestID == "" {
		utils.SendError(ctx, http.StatusBadRequest, errors.New("contest ID is required"), "Contest ID is required")
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

	standings, err := h.standingsUseCase.GetStandings(ctx.Request.Context(), contestID, userID, userRole)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrContestNotFound):
			utils.SendError(ctx, http.StatusNotFound, err, "Contest not found")
		case errors.Is(err, domain.ErrStandingsForbidden):
			utils.SendError(ctx, http.StatusForbidden, err, "Standings are not available before the contest starts")
		default:
			log.Error().Err(err).Str("contest_id", contestID).Str("user_id", userID).Msg("Failed to get contest standings")
			utils.SendError(ctx, http.StatusInternalServerError, errors.New("internal server error"), "Failed to get contest standings")
		}
		return
	}

	utils.SendSuccess(ctx, http.StatusOK, standings, "Standings retrieved successfully")
}
