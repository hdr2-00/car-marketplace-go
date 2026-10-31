package handler

import (
	"car-marketplace/internal/domain"
	"car-marketplace/pkg/response"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// errorStatuses maps domain errors to HTTP status codes. Errors not listed
// here are treated as internal errors.
var errorStatuses = []struct {
	err    error
	status int
}{
	// 400
	{domain.ErrInvalidInput, http.StatusBadRequest},
	{domain.ErrBadRequest, http.StatusBadRequest},
	{domain.ErrInvalidEmail, http.StatusBadRequest},
	{domain.ErrWeakPassword, http.StatusBadRequest},
	{domain.ErrInvalidUserType, http.StatusBadRequest},
	{domain.ErrInvalidMake, http.StatusBadRequest},
	{domain.ErrInvalidModel, http.StatusBadRequest},
	{domain.ErrInvalidBodyType, http.StatusBadRequest},
	{domain.ErrInvalidFuelType, http.StatusBadRequest},
	{domain.ErrInvalidTransmission, http.StatusBadRequest},
	{domain.ErrInvalidColor, http.StatusBadRequest},
	{domain.ErrInvalidCondition, http.StatusBadRequest},
	{domain.ErrInvalidFeature, http.StatusBadRequest},
	{domain.ErrInvalidStatus, http.StatusBadRequest},
	{domain.ErrCarListingCannotEdit, http.StatusBadRequest},
	{domain.ErrCannotInquireOwn, http.StatusBadRequest},
	{domain.ErrCannotReviewSelf, http.StatusBadRequest},
	{domain.ErrCannotFavoriteOwn, http.StatusBadRequest},

	// 401
	{domain.ErrInvalidCredentials, http.StatusUnauthorized},
	{domain.ErrUnauthorized, http.StatusUnauthorized},
	{domain.ErrInvalidToken, http.StatusUnauthorized},
	{domain.ErrTokenExpired, http.StatusUnauthorized},
	{domain.ErrInvalidRefreshToken, http.StatusUnauthorized},
	{domain.ErrRevokedToken, http.StatusUnauthorized},

	// 403
	{domain.ErrForbidden, http.StatusForbidden},
	{domain.ErrUserNotActive, http.StatusForbidden},
	{domain.ErrUserNotVerified, http.StatusForbidden},
	{domain.ErrCarListingNotOwner, http.StatusForbidden},
	{domain.ErrInquiryNotOwner, http.StatusForbidden},
	{domain.ErrReviewNotOwner, http.StatusForbidden},

	// 404
	{domain.ErrNotFound, http.StatusNotFound},
	{domain.ErrRecordNotFound, http.StatusNotFound},
	{domain.ErrUserNotFound, http.StatusNotFound},
	{domain.ErrCarListingNotFound, http.StatusNotFound},
	{domain.ErrImageNotFound, http.StatusNotFound},
	{domain.ErrInquiryNotFound, http.StatusNotFound},
	{domain.ErrReviewNotFound, http.StatusNotFound},
	{domain.ErrFavoriteNotFound, http.StatusNotFound},

	// 409
	{domain.ErrConflict, http.StatusConflict},
	{domain.ErrDuplicateEntry, http.StatusConflict},
	{domain.ErrUserAlreadyExists, http.StatusConflict},
	{domain.ErrCarListingAlreadyExists, http.StatusConflict},
	{domain.ErrCarListingAlreadySold, http.StatusConflict},
	{domain.ErrDuplicateVIN, http.StatusConflict},
	{domain.ErrReviewAlreadyExists, http.StatusConflict},
	{domain.ErrFavoriteAlreadyExists, http.StatusConflict},
}

// respondError writes the response for an error returned by a use case.
// Domain errors get their status and message; anything else is logged and
// reported as a generic 500, so internal details never reach the client.
func respondError(c *gin.Context, err error) {
	for _, e := range errorStatuses {
		if errors.Is(err, e.err) {
			response.Error(c, e.status, e.err.Error(), e.err)
			return
		}
	}

	log.Printf("internal error on %s %s: %v", c.Request.Method, c.FullPath(), err)
	response.Error(c, http.StatusInternalServerError, "Internal server error", nil)
}

// bindError responds to a request body or query that failed to bind.
func bindError(c *gin.Context, message string, err error) {
	response.Error(c, http.StatusBadRequest, message, err)
}

// currentUserID returns the authenticated user's ID, or 0 if there is none.
func currentUserID(c *gin.Context) int64 {
	id, _ := c.Get("user_id")
	userID, _ := id.(int64)
	return userID
}

// requireUserID returns the authenticated user's ID, responding 401 if absent.
func requireUserID(c *gin.Context) (int64, bool) {
	userID := currentUserID(c)
	if userID == 0 {
		respondError(c, domain.ErrUnauthorized)
		return 0, false
	}
	return userID, true
}

// pathID parses a positive int64 path parameter, responding 400 if invalid.
func pathID(c *gin.Context, name, label string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || id <= 0 {
		response.Error(c, http.StatusBadRequest, "Invalid "+label, nil)
		return 0, false
	}
	return id, true
}

// totalPages returns the number of pages needed for total items.
func totalPages(total int64, limit int) int {
	if limit <= 0 {
		return 0
	}
	return int((total + int64(limit) - 1) / int64(limit))
}
