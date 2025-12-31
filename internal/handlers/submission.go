package handlers

import (
	"algoforces/internal/domain"
	"algoforces/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SubmissionHandler struct {
	submissionUseCase domain.SubmissionUseCase
}

func NewSubmissionHandler(submissionUseCase domain.SubmissionUseCase) *SubmissionHandler {
	return &SubmissionHandler{
		submissionUseCase: submissionUseCase,
	}
}

// CreateSubmission godoc
//
//	@Summary		Create a new submission
//	@Description	Create a new code submission for evaluation
//	@Tags			Submission
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			createSubmissionRequest	body		domain.CreateSubmissionRequest	true	"Create Submission Request"
//	@Success		201							{object}	utils.SuccessResponse{data=domain.CreateSubmissionResponse}
//	@Failure		400							{object}	utils.ErrorResponse
//	@Failure		500							{object}	utils.ErrorResponse
//	@Router			/api/submission/create [post]
func (h *SubmissionHandler) CreateSubmission(ctx *gin.Context) {
	var createSubmissionRequest domain.CreateSubmissionRequest
	if err := ctx.ShouldBindJSON(&createSubmissionRequest); err != nil {
		utils.SendError(ctx, http.StatusBadRequest, err, "Invalid Request Body")
		return
	}

	// Call the use case to create a new submission
	createSubmissionResponse, err := h.submissionUseCase.CreateNewSubmission(ctx.Request.Context(), &createSubmissionRequest)
	if err != nil {
		utils.SendError(ctx, http.StatusInternalServerError, err, "Failed to create submission")
		return
	}

	utils.SendSuccess(ctx, http.StatusCreated, createSubmissionResponse, "Submission created successfully")
}

// GetSubmissionDetails godoc
//
//	@Summary		Get submission details
//	@Description	Get details of a specific submission by its ID
//	@Tags			Submission
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		string	true	"Submission ID"
//	@Success		200	{object}	utils.SuccessResponse{data=object}
//	@Failure		400	{object}	utils.ErrorResponse
//	@Failure		404	{object}	utils.ErrorResponse
//	@Router			/api/submission/{id} [get]
func (h *SubmissionHandler) GetSubmissionDetails(ctx *gin.Context) {
	submissionID := ctx.Param("id")
	if submissionID == "" {
		utils.SendError(ctx, http.StatusBadRequest, nil, "Submission ID is required")
		return
	}

	// Call the use case to get submission details
	submission, err := h.submissionUseCase.GetSubmissionDetails(ctx.Request.Context(), submissionID)
	if err != nil {
		utils.SendError(ctx, http.StatusNotFound, err, "Submission not found")
		return
	}

	utils.SendSuccess(ctx, http.StatusOK, submission, "Submission details retrieved successfully")
}

// UpdateSubmissionStatus godoc
//
//	@Summary		Update submission status
//	@Description	Update the status of a specific submission
//	@Tags			Submission
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"Submission ID"
//	@Param			request	body		map[string]string		true	"Status Update Request"
//	@Success		200		{object}	utils.SuccessResponse
//	@Failure		400		{object}	utils.ErrorResponse
//	@Failure		500		{object}	utils.ErrorResponse
//	@Router			/api/submission/{id}/update [put]
func (h *SubmissionHandler) UpdateSubmissionStatus(ctx *gin.Context) {
	submissionID := ctx.Param("id")
	if submissionID == "" {
		utils.SendError(ctx, http.StatusBadRequest, nil, "Submission ID is required")
		return
	}

	var updateRequest struct {
		Status string `json:"status" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&updateRequest); err != nil {
		utils.SendError(ctx, http.StatusBadRequest, err, "Invalid Request Body")
		return
	}

	// Call the use case to update submission status
	err := h.submissionUseCase.UpdateSubmissionStatus(ctx.Request.Context(), submissionID, updateRequest.Status)
	if err != nil {
		utils.SendError(ctx, http.StatusInternalServerError, err, "Failed to update submission status")
		return
	}

	utils.SendSuccess(ctx, http.StatusOK, nil, "Submission status updated successfully")
}

// JudgeSubmissionCallback godoc
//
//	@Summary		Judge submission callback
//	@Description	Judge a submission callback
//	@Tags			Submission
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			judgeSubmissionCallbackRequest	body		domain.JudgeSubmissionCallbackRequest	true	"Judge Submission Callback Request"
//	@Success		200								{object}	utils.SuccessResponse
//	@Failure		400								{object}	utils.ErrorResponse
//	@Failure		500								{object}	utils.ErrorResponse
//	@Router			/api/submission/callback [put]
func (h *SubmissionHandler) JudgeSubmissionCallback(ctx *gin.Context) {
	var judgeSubmissionCallbackRequest domain.JudgeSubmissionCallbackRequest
	if err := ctx.ShouldBindJSON(&judgeSubmissionCallbackRequest); err != nil {
		utils.SendError(ctx, http.StatusBadRequest, err, "Invalid Request Body")
		return
	}

	// Call the use case to judge submission callback
	err := h.submissionUseCase.JudgeSubmissionCallback(ctx.Request.Context(), &judgeSubmissionCallbackRequest)
	if err != nil {
		utils.SendError(ctx, http.StatusInternalServerError, err, "Failed to judge submission callback")
		return
	}

	utils.SendSuccess(ctx, http.StatusOK, nil, "Submission callback judged successfully")
}
