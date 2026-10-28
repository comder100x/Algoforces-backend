package handlers

import (
	"algoforces/internal/domain"
	"algoforces/internal/middleware"
	"algoforces/internal/utils"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type ContestHandler struct {
	contestUseCase domain.ContestUseCase
}

func NewContestHandler(contestUseCase domain.ContestUseCase) *ContestHandler {
	return &ContestHandler{
		contestUseCase: contestUseCase,
	}
}

// sendContestError maps contest sentinels to client statuses. Anything
// else is a sanitised 500, logged once at the handler boundary.
func sendContestError(c *gin.Context, err error, fallbackMsg string) {
	switch {
	case errors.Is(err, domain.ErrContestNotFound):
		utils.SendError(c, http.StatusNotFound, err, "Contest not found")
	case errors.Is(err, domain.ErrContestHasNoProblems),
		errors.Is(err, domain.ErrProblemSetterNotFound),
		errors.Is(err, domain.ErrProblemSetterNotAuthorized):
		utils.SendError(c, http.StatusBadRequest, err, "Invalid contest request")
	case errors.Is(err, domain.ErrContestEnded),
		errors.Is(err, domain.ErrContestLive),
		errors.Is(err, domain.ErrContestOverlap),
		errors.Is(err, domain.ErrContestTimeLocked):
		utils.SendError(c, http.StatusConflict, err, "Contest state conflict")
	default:
		log.Error().Err(err).Msg(fallbackMsg)
		utils.SendError(c, http.StatusInternalServerError, errors.New("internal server error"), fallbackMsg)
	}
}

// parseContestID validates the :id path param. It reports 400 itself
// and returns false when the id is missing or not a UUID.
func parseContestID(c *gin.Context) (string, bool) {
	contestId := c.Param("id")
	if contestId == "" {
		utils.SendError(c, http.StatusBadRequest, errors.New("contest ID is required"), "Contest ID is required")
		return "", false
	}
	if _, err := uuid.Parse(contestId); err != nil {
		utils.SendError(c, http.StatusBadRequest, errors.New("contest ID must be a valid UUID"), "Invalid contest ID")
		return "", false
	}
	return contestId, true
}

// CreateContest godoc
//
//	@Summary		Create a new Contest
//	@Description	Create a new contest (admin or problem-setter only)
//	@Tags			Contest
//	@Accept			json
//	@Produce		json
//	@Param			createContestRequest	body	domain.CreateContestRequest	true	"Create Contest Request"
//	@Security		BearerAuth
//	@Success		201	{object}	domain.CreateContestResponse
//	@Failure		400	{object}	utils.ErrorResponse
//	@Failure		401	{object}	utils.ErrorResponse
//	@Failure		409	{object}	utils.ErrorResponse
//	@Failure		500	{object}	utils.ErrorResponse
//	@Router			/api/contest/create [post]
func (h *ContestHandler) CreateContest(c *gin.Context) {
	var createContestRequest domain.CreateContestRequest
	if err := c.ShouldBindJSON(&createContestRequest); err != nil {
		utils.SendError(c, http.StatusBadRequest, err, "Invalid request body")
		return
	}
	userId, err := middleware.GetUserID(c)
	if err != nil {
		utils.SendError(c, http.StatusUnauthorized, err, "Failed to get user ID")
		return
	}

	contestResponse, err := h.contestUseCase.CreateContest(c.Request.Context(), &createContestRequest, userId)
	if err != nil {
		sendContestError(c, err, "Failed to create contest")
		return
	}

	log.Info().
		Str("user_id", userId).
		Str("contest_id", contestResponse.Id).
		Str("name", createContestRequest.Name).
		Msg("Contest created successfully")
	utils.SendSuccess(c, http.StatusCreated, contestResponse, "Contest created successfully")

}

// GetContestDetails godoc
//
//	@Summary		Get Contest Details
//	@Description	Get details of a specific contest by ID
//	@Tags			Contest
//	@Produce		json
//	@Param			id	path	string	true	"Contest ID"
//	@Security		BearerAuth
//	@Success		200	{object}	domain.CreateContestResponse
//	@Failure		400	{object}	utils.ErrorResponse
//	@Failure		404	{object}	utils.ErrorResponse
//	@Failure		500	{object}	utils.ErrorResponse
//	@Router			/api/contest/{id} [get]
func (h *ContestHandler) GetContestDetails(c *gin.Context) {
	contestId, ok := parseContestID(c)
	if !ok {
		return
	}

	contestResponse, err := h.contestUseCase.GetContestDetails(c.Request.Context(), contestId)
	if err != nil {
		sendContestError(c, err, "Failed to get contest")
		return
	}

	utils.SendSuccess(c, http.StatusOK, contestResponse, "Contest retrieved successfully")
}

// UpdateContest godoc
//
//	@Summary		Update a Contest
//	@Description	Update an existing contest (admin or problem-setter only)
//	@Tags			Contest
//	@Accept			json
//	@Produce		json
//	@Param			updateContestRequest	body	domain.UpdateContestRequest	true	"Update Contest Request"
//	@Security		BearerAuth
//	@Success		200	{object}	domain.UpdateContestResponse
//	@Failure		400	{object}	utils.ErrorResponse
//	@Failure		404	{object}	utils.ErrorResponse
//	@Failure		409	{object}	utils.ErrorResponse
//	@Failure		500	{object}	utils.ErrorResponse
//	@Router			/api/contest/update [put]
func (h *ContestHandler) UpdateContest(c *gin.Context) {
	var updateContestRequest domain.UpdateContestRequest
	if err := c.ShouldBindJSON(&updateContestRequest); err != nil {
		utils.SendError(c, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	contestResponse, err := h.contestUseCase.UpdateContest(c.Request.Context(), &updateContestRequest)
	if err != nil {
		sendContestError(c, err, "Failed to update contest")
		return
	}

	utils.SendSuccess(c, http.StatusOK, contestResponse, "Contest updated successfully")
}

// DeleteContest godoc
//
//	@Summary		Delete a Contest
//	@Description	Delete a contest by ID (admin only)
//	@Tags			Contest
//	@Produce		json
//	@Param			id	path	string	true	"Contest ID"
//	@Security		BearerAuth
//	@Success		200	{object}	utils.SuccessResponse
//	@Failure		400	{object}	utils.ErrorResponse
//	@Failure		404	{object}	utils.ErrorResponse
//	@Failure		409	{object}	utils.ErrorResponse
//	@Failure		500	{object}	utils.ErrorResponse
//	@Router			/api/contest/{id} [delete]
func (h *ContestHandler) DeleteContest(c *gin.Context) {
	contestId, ok := parseContestID(c)
	if !ok {
		return
	}

	deleteRequest := &domain.DeleteContestRequest{Id: contestId}
	err := h.contestUseCase.DeleteContest(c.Request.Context(), deleteRequest)
	if err != nil {
		sendContestError(c, err, "Failed to delete contest")
		return
	}

	utils.SendSuccess(c, http.StatusOK, nil, "Contest deleted successfully")
}

// StartContest godoc
//
//	@Summary		Start a Contest
//	@Description	Start a contest (admin or problem-setter only). The contest must not have ended and must have at least one problem.
//	@Tags			Contest
//	@Produce		json
//	@Param			id	path	string	true	"Contest ID"
//	@Security		BearerAuth
//	@Success		200	{object}	utils.SuccessResponse
//	@Failure		400	{object}	utils.ErrorResponse
//	@Failure		401	{object}	utils.ErrorResponse
//	@Failure		404	{object}	utils.ErrorResponse
//	@Failure		409	{object}	utils.ErrorResponse
//	@Failure		500	{object}	utils.ErrorResponse
//	@Router			/api/contest/{id}/start [post]
func (h *ContestHandler) StartContest(c *gin.Context) {
	contestId, ok := parseContestID(c)
	if !ok {
		return
	}
	userId, err := middleware.GetUserID(c)
	if err != nil {
		utils.SendError(c, http.StatusUnauthorized, err, "Failed to get user ID")
		return
	}

	if err := h.contestUseCase.StartContest(c.Request.Context(), contestId); err != nil {
		sendContestError(c, err, "Failed to start contest")
		return
	}

	log.Info().Str("user_id", userId).Str("contest_id", contestId).Msg("Contest started successfully")
	utils.SendSuccess(c, http.StatusOK, nil, "Contest started successfully")
}

// EndContest godoc
//
//	@Summary		End a Contest
//	@Description	End a contest (admin or problem-setter only). Sets the contest inactive; its end time is left unchanged.
//	@Tags			Contest
//	@Produce		json
//	@Param			id	path	string	true	"Contest ID"
//	@Security		BearerAuth
//	@Success		200	{object}	utils.SuccessResponse
//	@Failure		400	{object}	utils.ErrorResponse
//	@Failure		401	{object}	utils.ErrorResponse
//	@Failure		404	{object}	utils.ErrorResponse
//	@Failure		500	{object}	utils.ErrorResponse
//	@Router			/api/contest/{id}/end [post]
func (h *ContestHandler) EndContest(c *gin.Context) {
	contestId, ok := parseContestID(c)
	if !ok {
		return
	}
	userId, err := middleware.GetUserID(c)
	if err != nil {
		utils.SendError(c, http.StatusUnauthorized, err, "Failed to get user ID")
		return
	}

	if err := h.contestUseCase.EndContest(c.Request.Context(), contestId); err != nil {
		sendContestError(c, err, "Failed to end contest")
		return
	}

	log.Info().Str("user_id", userId).Str("contest_id", contestId).Msg("Contest ended successfully")
	utils.SendSuccess(c, http.StatusOK, nil, "Contest ended successfully")
}

// GetAllContests godoc
//
//	@Summary		Get All Contests (Admin)
//	@Description	Get all contests (admin only)
//	@Tags			Contest
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{array}		domain.CreateContestResponse
//	@Failure		500	{object}	utils.ErrorResponse
//	@Router			/api/admin/contests [get]
func (h *ContestHandler) GetAllContests(c *gin.Context) {
	contests, err := h.contestUseCase.GetAllContests(c.Request.Context())
	if err != nil {
		sendContestError(c, err, "Failed to get contests")
		return
	}

	utils.SendSuccess(c, http.StatusOK, contests, "Contests retrieved successfully")
}

// GetVisibleContests godoc
//
//	@Summary		Get Visible Contests
//	@Description	Get all visible contests, ordered by start time (any authenticated user)
//	@Tags			Contest
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{array}		domain.CreateContestResponse
//	@Failure		500	{object}	utils.ErrorResponse
//	@Router			/api/contest/all [get]
func (h *ContestHandler) GetVisibleContests(c *gin.Context) {
	contests, err := h.contestUseCase.GetVisibleContests(c.Request.Context())
	if err != nil {
		sendContestError(c, err, "Failed to get contests")
		return
	}

	utils.SendSuccess(c, http.StatusOK, contests, "Contests retrieved successfully")
}
