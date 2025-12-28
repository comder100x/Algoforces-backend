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
		utils.SendError(ctx, http.StatusInternalServerError, err, "Invalid Request Body")
		return
	}

	// Call the use case to create a new submission
	createSubmissionResponse, err := h.submissionUseCase.CreateNewSubmission(ctx.Request.Context(), &createSubmissionRequest)
	if err != nil {
		utils.SendError(ctx, http.StatusInternalServerError, err, "Failed to create submission")
		return
	}

	utils.SendSuccess(ctx, http.StatusOK, createSubmissionResponse, "Submission created successfully")
}
