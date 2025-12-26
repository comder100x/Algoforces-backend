package handlers

import (
	"algoforces/internal/domain"
	"algoforces/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	adminUseCase domain.AdminUseCase
}

func NewAdminHandler(adminUseCase domain.AdminUseCase) *AdminHandler {
	return &AdminHandler{
		adminUseCase: adminUseCase,
	}
}

// AddRole godoc
//
//	@Summary		Add Role to User
//	@Description	Assign a new role to a user
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			addRoleRequest	body	domain.AddRoleRequest	true	"Add Role Request"
//	@Security		BearerAuth
//	@Success		200	{object}	domain.AddRoleResponse
//	@Failure		400	{object}	utils.ErrorResponse
//	@Failure		500	{object}	utils.ErrorResponse
//	@Router			/api/admin/addrole [put]
func (h *AdminHandler) AddRole(c *gin.Context) {
	var addRoleRequest domain.AddRoleRequest
	if err := c.ShouldBindJSON(&addRoleRequest); err != nil {
		utils.SendError(c, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	addRoleResponse, err := h.adminUseCase.AddRole(c.Request.Context(), &addRoleRequest)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err, "Failed to add role")
		return
	}

	utils.SendSuccess(c, http.StatusOK, addRoleResponse, "Role added successfully")
}

// RemoveRole godoc
//
//	@Summary		Remove Role from User
//	@Description	Remove a role from a user
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			removeRoleRequest	body	domain.RemoveRoleRequest	true	"Remove Role Request"
//	@Security		BearerAuth
//	@Success		200	{object}	domain.RemoveRoleResponse
//	@Failure		400	{object}	utils.ErrorResponse
//	@Failure		500	{object}	utils.ErrorResponse
//	@Router			/api/admin/removerole [put]
func (h *AdminHandler) RemoveRole(c *gin.Context) {
	var removeRoleRequest domain.RemoveRoleRequest
	if err := c.ShouldBindJSON(&removeRoleRequest); err != nil {
		utils.SendError(c, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	removeRoleResponse, err := h.adminUseCase.RemoveRole(c.Request.Context(), &removeRoleRequest)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err, "Failed to remove role")
		return
	}

	utils.SendSuccess(c, http.StatusOK, removeRoleResponse, "Role removed successfully")
}

// GetAllUsers godoc
//
//	@Summary		Get All Users
//	@Description	Get list of all users (admin only)
//	@Tags			Admin
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	domain.UserListResponse
//	@Failure		500	{object}	utils.ErrorResponse
//	@Router			/api/admin/users [get]
func (h *AdminHandler) GetAllUsers(c *gin.Context) {
	response, err := h.adminUseCase.GetAllUsers(c.Request.Context())
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err, "Failed to fetch users")
		return
	}

	utils.SendSuccess(c, http.StatusOK, response, "Users fetched successfully")
}

// GetAdmins godoc
//
//	@Summary		Get All Admins
//	@Description	Get list of all admin users (admin only)
//	@Tags			Admin
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	domain.UserListResponse
//	@Failure		500	{object}	utils.ErrorResponse
//	@Router			/api/admin/admins [get]
func (h *AdminHandler) GetAdmins(c *gin.Context) {
	response, err := h.adminUseCase.GetUsersByRole(c.Request.Context(), "admin")
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err, "Failed to fetch admins")
		return
	}

	utils.SendSuccess(c, http.StatusOK, response, "Admins fetched successfully")
}

// GetProblemSetters godoc
//
//	@Summary		Get All Problem Setters
//	@Description	Get list of all problem setter users (admin only)
//	@Tags			Admin
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	domain.UserListResponse
//	@Failure		500	{object}	utils.ErrorResponse
//	@Router			/api/admin/problem-setters [get]
func (h *AdminHandler) GetProblemSetters(c *gin.Context) {
	response, err := h.adminUseCase.GetUsersByRole(c.Request.Context(), "problem_setter")
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err, "Failed to fetch problem setters")
		return
	}

	utils.SendSuccess(c, http.StatusOK, response, "Problem setters fetched successfully")
}
