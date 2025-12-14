package handler

import (
	"car-marketplace/internal/domain"
	"car-marketplace/internal/usecase"
	"car-marketplace/pkg/response"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication related HTTP requests
type AuthHandler struct {
	authUseCase     usecase.AuthUseCase
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

// NewAuthHandler creates a new auth handler instance
func NewAuthHandler(authUseCase usecase.AuthUseCase, accessTokenTTL, refreshTokenTTL time.Duration) *AuthHandler {
	return &AuthHandler{
		authUseCase:     authUseCase,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user account (individual, dealership, or showroom). Note: first_name and last_name are required when user_type is "individual". business_name is required when user_type is "dealership" or "showroom".
// @Tags auth
// @Accept json
// @Produce json
// @Param request body domain.RegisterRequest true "Registration details"
// @Success 201 {object} response.Response{data=domain.User}
// @Failure 400 {object} response.Response "Invalid request or missing required fields for selected user_type"
// @Failure 409 {object} response.Response "User already exists"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req domain.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	user, err := h.authUseCase.Register(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case domain.ErrUserAlreadyExists:
			response.Error(c, http.StatusConflict, "User already exists", err)
		case domain.ErrInvalidInput, domain.ErrInvalidUserType:
			response.Error(c, http.StatusBadRequest, err.Error(), err)
		default:
			response.Error(c, http.StatusInternalServerError, "Failed to register user", err)
		}
		return
	}

	response.Success(c, http.StatusCreated, "User registered successfully", user)
}

// Login godoc
// @Summary User login
// @Description Authenticate user and set access and refresh tokens as HTTP-only cookies
// @Tags auth
// @Accept json
// @Produce json
// @Param request body domain.LoginRequest true "Login credentials"
// @Success 200 {object} response.Response{data=domain.User}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	tokens, user, err := h.authUseCase.Login(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case domain.ErrInvalidCredentials:
			response.Error(c, http.StatusUnauthorized, "Invalid email or password", err)
		case domain.ErrUserNotActive:
			response.Error(c, http.StatusForbidden, "User account is not active", err)
		default:
			response.Error(c, http.StatusInternalServerError, "Failed to login", err)
		}
		return
	}

	// Set access token as HTTP-only cookie
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		"access_token",                  // name
		tokens.AccessToken,              // value
		int(h.accessTokenTTL.Seconds()), // maxAge in seconds from config
		"/",                             // path
		"",                              // domain
		false,                           // secure (set to true in production with HTTPS)
		true,                            // httpOnly
	)

	// Set refresh token as HTTP-only cookie
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		"refresh_token",                  // name
		tokens.RefreshToken,              // value
		int(h.refreshTokenTTL.Seconds()), // maxAge in seconds from config
		"/",                              // path
		"",                               // domain
		false,                            // secure (set to true in production with HTTPS)
		true,                             // httpOnly
	)

	response.Success(c, http.StatusOK, "Login successful", user)
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Generate new access token using refresh token from cookie
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// Get refresh token from cookie
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Refresh token not found", err)
		return
	}

	tokens, err := h.authUseCase.RefreshToken(c.Request.Context(), refreshToken)
	if err != nil {
		switch err {
		case domain.ErrInvalidRefreshToken, domain.ErrTokenExpired:
			response.Error(c, http.StatusUnauthorized, "Invalid or expired refresh token", err)
		case domain.ErrUserNotActive:
			response.Error(c, http.StatusForbidden, "User account is not active", err)
		default:
			response.Error(c, http.StatusInternalServerError, "Failed to refresh token", err)
		}
		return
	}

	// Set new access token as HTTP-only cookie
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		"access_token",                  // name
		tokens.AccessToken,              // value
		int(h.accessTokenTTL.Seconds()), // maxAge in seconds from config
		"/",                             // path
		"",                              // domain
		false,                           // secure (set to true in production with HTTPS)
		true,                            // httpOnly
	)

	// Set new refresh token as HTTP-only cookie
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		"refresh_token",                  // name
		tokens.RefreshToken,              // value
		int(h.refreshTokenTTL.Seconds()), // maxAge in seconds from config
		"/",                              // path
		"",                               // domain
		false,                            // secure (set to true in production with HTTPS)
		true,                             // httpOnly
	)

	response.Success(c, http.StatusOK, "Token refreshed successfully", nil)
}

// Logout godoc
// @Summary User logout
// @Description Revoke refresh token and clear cookies
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// Get refresh token from cookie
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Refresh token not found", err)
		return
	}

	if err := h.authUseCase.Logout(c.Request.Context(), refreshToken); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to logout", err)
		return
	}

	// Clear cookies
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		"access_token",
		"",
		-1, // negative maxAge deletes the cookie
		"/",
		"",
		false,
		true,
	)

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		"refresh_token",
		"",
		-1, // negative maxAge deletes the cookie
		"/",
		"",
		false,
		true,
	)

	response.Success(c, http.StatusOK, "Logout successful", nil)
}

// LogoutAll godoc
// @Summary Logout from all devices
// @Description Revoke all refresh tokens for the authenticated user and clear cookies
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /auth/logout-all [post]
func (h *AuthHandler) LogoutAll(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	if err := h.authUseCase.LogoutAll(c.Request.Context(), userID.(int64)); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to logout from all devices", err)
		return
	}

	// Clear cookies
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		"access_token",
		"",
		-1, // negative maxAge deletes the cookie
		"/",
		"",
		false,
		true,
	)

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		"refresh_token",
		"",
		-1, // negative maxAge deletes the cookie
		"/",
		"",
		false,
		true,
	)

	response.Success(c, http.StatusOK, "Logged out from all devices successfully", nil)
}

// ChangePassword godoc
// @Summary Change user password
// @Description Change password for the authenticated user
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.ChangePasswordRequest true "Password change details"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var req domain.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.authUseCase.ChangePassword(c.Request.Context(), userID.(int64), &req); err != nil {
		switch err {
		case domain.ErrInvalidCredentials:
			response.Error(c, http.StatusUnauthorized, "Current password is incorrect", err)
		case domain.ErrWeakPassword:
			response.Error(c, http.StatusBadRequest, "Password does not meet requirements", err)
		default:
			response.Error(c, http.StatusInternalServerError, "Failed to change password", err)
		}
		return
	}

	response.Success(c, http.StatusOK, "Password changed successfully. Please login again.", nil)
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get authenticated user's profile information
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=domain.User}
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	user, err := h.authUseCase.GetUserByID(c.Request.Context(), userID.(int64))
	if err != nil {
		switch err {
		case domain.ErrUserNotFound:
			response.Error(c, http.StatusNotFound, "User not found", err)
		default:
			response.Error(c, http.StatusInternalServerError, "Failed to get user profile", err)
		}
		return
	}

	response.Success(c, http.StatusOK, "Profile retrieved successfully", user)
}
