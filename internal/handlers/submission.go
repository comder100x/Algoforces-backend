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
