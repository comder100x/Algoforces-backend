package handlers

import (
	"algoforces/internal/domain"
	"algoforces/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TestCaseHandler struct {
	testCaseUseCase domain.TestCaseUseCase
}

func NewTestCaseHandler(testCaseUseCase domain.TestCaseUseCase) *TestCaseHandler {
	return &TestCaseHandler{
		testCaseUseCase: testCaseUseCase,
	}
}

// CreateTestCase godoc
// @Summary		Create a new test case
// @Description	Creates a new test case for a problem
// @Tags			TestCase
// @Accept			json
// @Produce		json
// @Param			request	body		domain.CreateTestCaseRequest	true	"Test case creation request"
// @Success		200		{object}	utils.SuccessResponse{data=domain.CreateTestCaseResponse}
// @Failure		400		{object}	utils.ErrorResponse
// @Failure		500		{object}	utils.ErrorResponse
// @Security		BearerAuth
// @Router			/api/testcase/create [post]
func (h *TestCaseHandler) CreateTestCase(ctx *gin.Context) {
	var createTestCaseRequest domain.CreateTestCaseRequest
	if err := ctx.ShouldBindJSON(&createTestCaseRequest); err != nil {
		utils.SendError(ctx, http.StatusInternalServerError, err, "Invalid Request Body")
		return
	}

	createTestCaseResponse, err := h.testCaseUseCase.CreateNewTestCase(ctx.Request.Context(), &createTestCaseRequest)
	if err != nil {
		utils.SendError(ctx, http.StatusInternalServerError, err, "Failed to create test case")
		return
	}

	utils.SendSuccess(ctx, http.StatusOK, createTestCaseResponse, "Test case created successfully")
}

// GetAllTestCasesForProblem godoc
// @Summary		Get all test cases for a problem
// @Description	Retrieves all test cases associated with a specific problem
// @Tags			TestCase
// @Accept			json
// @Produce		json
// @Param			problemId	path		string	true	"Problem ID"
// @Success		200			{object}	utils.SuccessResponse{data=[]domain.TestCase}
// @Failure		400			{object}	utils.ErrorResponse
// @Failure		500			{object}	utils.ErrorResponse
// @Security		BearerAuth
// @Router			/api/testcase/problem/{problemId} [get]
func (h *TestCaseHandler) GetAllTestCasesForProblem(ctx *gin.Context) {
	problemID := ctx.Param("problemId")

	if problemID == "" {
		utils.SendError(ctx, http.StatusBadRequest, nil, "Problem ID is required")
		return
	}

	testCases, err := h.testCaseUseCase.GetAllTestCasesForProblem(ctx.Request.Context(), problemID)
	if err != nil {
		utils.SendError(ctx, http.StatusInternalServerError, err, "Failed to fetch test cases")
		return
	}

	utils.SendSuccess(ctx, http.StatusOK, testCases, "Test cases fetched successfully")
}

// GetTestCaseDetails godoc
// @Summary		Get test case details
// @Description	Retrieves details of a specific test case by its unique ID
// @Tags			TestCase
// @Accept			json
// @Produce		json
// @Param			id	path		string	true	"Test Case Unique ID"
// @Success		200	{object}	utils.SuccessResponse{data=domain.TestCase}
// @Failure		400	{object}	utils.ErrorResponse
// @Failure		500	{object}	utils.ErrorResponse
// @Security		BearerAuth
// @Router			/api/testcase/{id} [get]
func (h *TestCaseHandler) GetTestCaseDetails(ctx *gin.Context) {
	uniqueID := ctx.Param("id")

	if uniqueID == "" {
		utils.SendError(ctx, http.StatusBadRequest, nil, "Test Case Unique ID is required")
		return
	}

	testCase, err := h.testCaseUseCase.GetTestCaseDetails(ctx.Request.Context(), uniqueID)
	if err != nil {
		utils.SendError(ctx, http.StatusInternalServerError, err, "Failed to fetch test case details")
		return
	}

	utils.SendSuccess(ctx, http.StatusOK, testCase, "Test case details fetched successfully")
}

// UpdateTestCase godoc
// @Summary		Update a test case
// @Description	Updates an existing test case
// @Tags			TestCase
// @Accept			json
// @Produce		json
// @Param			request	body		domain.UpdateTestCaseRequest	true	"Test case update request"
// @Success		200		{object}	utils.SuccessResponse{data=domain.UpdateTestCaseResponse}
// @Failure		400		{object}	utils.ErrorResponse
// @Failure		500		{object}	utils.ErrorResponse
// @Security		BearerAuth
// @Router			/api/testcase/update [put]
func (h *TestCaseHandler) UpdateTestCase(ctx *gin.Context) {
	var updateTestCaseRequest domain.UpdateTestCaseRequest
	if err := ctx.ShouldBindJSON(&updateTestCaseRequest); err != nil {
		utils.SendError(ctx, http.StatusInternalServerError, err, "Invalid Request Body")
		return
	}

	updateTestCaseResponse, err := h.testCaseUseCase.UpdateSingleTestCase(ctx.Request.Context(), &updateTestCaseRequest)
	if err != nil {
		utils.SendError(ctx, http.StatusInternalServerError, err, "Failed to update test case")
		return
	}

	utils.SendSuccess(ctx, http.StatusOK, updateTestCaseResponse, "Test case updated successfully")
}

// DeleteTestCase godoc
// @Summary		Delete a test case
// @Description	Deletes a test case by its unique ID
// @Tags			TestCase
// @Accept			json
// @Produce		json
// @Param			id	path		string	true	"Test Case Unique ID"
// @Success		200	{object}	utils.SuccessResponse
// @Failure		400	{object}	utils.ErrorResponse
// @Failure		500	{object}	utils.ErrorResponse
// @Security		BearerAuth
// @Router			/api/testcase/{id} [delete]
func (h *TestCaseHandler) DeleteTestCase(ctx *gin.Context) {
	uniqueID := ctx.Param("id")

	if uniqueID == "" {
		utils.SendError(ctx, http.StatusBadRequest, nil, "Test Case Unique ID is required")
		return
	}

	err := h.testCaseUseCase.DeleteSingleTestCase(ctx.Request.Context(), uniqueID)
	if err != nil {
		utils.SendError(ctx, http.StatusInternalServerError, err, "Failed to delete test case")
		return
	}

	utils.SendSuccess(ctx, http.StatusOK, nil, "Test case deleted successfully")
}

// UploadTestCasesInBulk godoc
// @Summary		Upload test cases in bulk
// @Description	Creates multiple test cases at once for a problem
// @Tags			TestCase
// @Accept			json
// @Produce		json
// @Param			request	body		domain.BulkTestCaseUploadRequest	true	"Bulk test case upload request"
// @Success		200		{object}	utils.SuccessResponse{data=domain.BulkTestCaseUploadResponse}
// @Failure		400		{object}	utils.ErrorResponse
// @Failure		500		{object}	utils.ErrorResponse
// @Security		BearerAuth
// @Router			/api/testcase/bulk [post]
func (h *TestCaseHandler) UploadTestCasesInBulk(ctx *gin.Context) {
	var BulkTestCaseUploadRequest domain.BulkTestCaseUploadRequest
	if err := ctx.ShouldBindJSON(&BulkTestCaseUploadRequest); err != nil {
		utils.SendError(ctx, http.StatusInternalServerError, err, "Invalid Request Body")
		return
	}

	BulkTestCaseUploadResponse, err := h.testCaseUseCase.UploadTestCasesInBulk(ctx.Request.Context(), &BulkTestCaseUploadRequest)
	if err != nil {
		utils.SendError(ctx, http.StatusInternalServerError, err, "Failed to upload test cases in bulk")
		return
	}

	utils.SendSuccess(ctx, http.StatusOK, BulkTestCaseUploadResponse, "Test cases uploaded successfully")
}
