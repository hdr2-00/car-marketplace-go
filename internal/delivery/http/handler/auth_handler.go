package handler

import (
	"car-marketplace/internal/delivery/http/middleware"
	"car-marketplace/internal/domain"
	"car-marketplace/internal/usecase"
	"car-marketplace/pkg/response"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const refreshTokenCookie = "refresh_token"

// AuthHandler handles authentication related HTTP requests
type AuthHandler struct {
	authUseCase     usecase.AuthUseCase
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	secureCookies   bool
}

// NewAuthHandler creates a new auth handler instance. secureCookies sets the
// Secure flag on auth cookies, so browsers only send them over HTTPS.
func NewAuthHandler(authUseCase usecase.AuthUseCase, accessTokenTTL, refreshTokenTTL time.Duration, secureCookies bool) *AuthHandler {
	return &AuthHandler{
		authUseCase:     authUseCase,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
		secureCookies:   secureCookies,
	}
}

// setCookie sets an HTTP-only, SameSite=Lax cookie; a negative maxAge deletes it.
func (h *AuthHandler) setCookie(c *gin.Context, name, value string, maxAge int) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(name, value, maxAge, "/", "", h.secureCookies, true)
}

func (h *AuthHandler) setAuthCookies(c *gin.Context, tokens *domain.TokenPair) {
	h.setCookie(c, middleware.AccessTokenCookie, tokens.AccessToken, int(h.accessTokenTTL.Seconds()))
	h.setCookie(c, refreshTokenCookie, tokens.RefreshToken, int(h.refreshTokenTTL.Seconds()))
}

func (h *AuthHandler) clearAuthCookies(c *gin.Context) {
	h.setCookie(c, middleware.AccessTokenCookie, "", -1)
	h.setCookie(c, refreshTokenCookie, "", -1)
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
		bindError(c, "Invalid request body", err)
		return
	}

	user, err := h.authUseCase.Register(c.Request.Context(), &req)
	if err != nil {
		respondError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "User registered successfully", user)
}

// Login godoc
// @Summary User login
// @Description Authenticate user and set access and refresh tokens as HTTP-only cookies. Protected endpoints accept the access_token cookie or an "Authorization: Bearer" header.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body domain.LoginRequest true "Login credentials"
// @Success 200 {object} response.Response{data=domain.User}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response "Account not active"
// @Failure 500 {object} response.Response
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		bindError(c, "Invalid request body", err)
		return
	}

	tokens, user, err := h.authUseCase.Login(c.Request.Context(), &req)
	if err != nil {
		respondError(c, err)
		return
	}

	h.setAuthCookies(c, tokens)
	response.Success(c, http.StatusOK, "Login successful", user)
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Exchange the refresh token cookie for new access and refresh tokens. Each refresh token can be used once.
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	refreshToken, err := c.Cookie(refreshTokenCookie)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Refresh token not found", nil)
		return
	}

	tokens, err := h.authUseCase.RefreshToken(c.Request.Context(), refreshToken)
	if err != nil {
		respondError(c, err)
		return
	}

	h.setAuthCookies(c, tokens)
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
	refreshToken, err := c.Cookie(refreshTokenCookie)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Refresh token not found", nil)
		return
	}

	if err := h.authUseCase.Logout(c.Request.Context(), refreshToken); err != nil {
		respondError(c, err)
		return
	}

	h.clearAuthCookies(c)
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
	userID, ok := requireUserID(c)
	if !ok {
		return
	}

	if err := h.authUseCase.LogoutAll(c.Request.Context(), userID); err != nil {
		respondError(c, err)
		return
	}

	h.clearAuthCookies(c)
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
	userID, ok := requireUserID(c)
	if !ok {
		return
	}

	var req domain.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindError(c, "Invalid request body", err)
		return
	}

	if err := h.authUseCase.ChangePassword(c.Request.Context(), userID, &req); err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			response.Error(c, http.StatusUnauthorized, "Current password is incorrect", err)
			return
		}
		respondError(c, err)
		return
	}

	h.clearAuthCookies(c)
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
	userID, ok := requireUserID(c)
	if !ok {
		return
	}

	user, err := h.authUseCase.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		respondError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Profile retrieved successfully", user)
}
