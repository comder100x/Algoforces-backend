package handlers

import (
	"algoforces/internal/domain"
	"algoforces/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type AuthHandler struct {
	authUseCase domain.UserUseCase
}

func NewAuthHandler(authUseCase domain.UserUseCase) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
	}
}

// Signup godoc
//
//	@Summary		User Signup
//	@Description	Register a new user
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			signupRequest	body		domain.SignupRequest	true	"Signup Request"
//	@Success		201				{object}	domain.AuthResponse
//	@Failure		400				{object}	utils.ErrorResponse
//	@Failure		500				{object}	utils.ErrorResponse
//	@Router			/api/auth/signup [post]
func (h *AuthHandler) Signup(c *gin.Context) {
	var signupRequest domain.SignupRequest
	if err := c.ShouldBindJSON(&signupRequest); err != nil {
		log.Error().Err(err).Msg("Invalid signup request body")
		utils.SendError(c, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	authResponse, err := h.authUseCase.Signup(c.Request.Context(), &signupRequest)
	if err != nil {
		log.Error().Err(err).Str("email", signupRequest.Email).Msg("Failed to signup user")
		utils.SendError(c, http.StatusInternalServerError, err, "Failed to signup")
		return
	}

	log.Info().Str("email", signupRequest.Email).Str("username", signupRequest.Username).Msg("User signed up successfully")
	utils.SendSuccess(c, http.StatusCreated, authResponse, "User signed up successfully")
}

// Login godoc
//
//	@Summary		User Login
//	@Description	Authenticate a user and return a token
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			loginRequest	body		domain.LoginRequest	true	"Login Request"
//	@Success		200				{object}	domain.AuthResponse
//	@Failure		400				{object}	utils.ErrorResponse
//	@Failure		500				{object}	utils.ErrorResponse
//	@Router			/api/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var loginRequest domain.LoginRequest
	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		log.Error().Err(err).Msg("Invalid login request body")
		utils.SendError(c, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	authResponse, err := h.authUseCase.Login(c.Request.Context(), &loginRequest)
	if err != nil {
		log.Error().Err(err).Str("email", loginRequest.Email).Msg("Failed to login user")
		utils.SendError(c, http.StatusInternalServerError, err, "Failed to login")
		return
	}

	log.Info().Str("email", loginRequest.Email).Msg("User logged in successfully")
	utils.SendSuccess(c, http.StatusOK, authResponse, "User logged in successfully")
}
