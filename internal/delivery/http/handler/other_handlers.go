package handler

import (
	"car-marketplace/internal/domain"
	"car-marketplace/internal/usecase"
	"car-marketplace/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ================= INQUIRY HANDLER =================

type InquiryHandler struct {
	inquiryUseCase usecase.InquiryUseCase
}

func NewInquiryHandler(inquiryUseCase usecase.InquiryUseCase) *InquiryHandler {
	return &InquiryHandler{
		inquiryUseCase: inquiryUseCase,
	}
}

// Create godoc
// @Summary Create an inquiry
// @Description Send an inquiry about a car listing
// @Tags inquiries
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.CreateInquiryRequest true "Inquiry details"
// @Success 201 {object} response.Response{data=domain.Inquiry}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /inquiries [post]
func (h *InquiryHandler) Create(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req domain.CreateInquiryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	inquiry, err := h.inquiryUseCase.Create(c.Request.Context(), &req, userID.(int64))
	if err != nil {
		switch err {
		case domain.ErrCannotInquireOwn:
			response.Error(c, http.StatusBadRequest, "You cannot inquire about your own listing", err)
		case domain.ErrCarListingNotFound:
			response.Error(c, http.StatusNotFound, "Car listing not found", err)
		default:
			response.Error(c, http.StatusInternalServerError, "Failed to create inquiry", err)
		}
		return
	}

	response.Success(c, http.StatusCreated, "Inquiry sent successfully", inquiry)
}

// List godoc
// @Summary List inquiries
// @Description Get list of inquiries (sent or received)
// @Tags inquiries
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param car_listing_id query int false "Car listing ID"
// @Param is_read query boolean false "Read status"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} response.PaginatedResponse{data=[]domain.Inquiry}
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /inquiries [get]
func (h *InquiryHandler) List(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var filter domain.InquiryFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid query parameters", err)
		return
	}

	filter.SetDefaults()

	inquiries, total, err := h.inquiryUseCase.List(c.Request.Context(), &filter, userID.(int64))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to list inquiries", err)
		return
	}

	pagination := response.PaginationMeta{
		Page:       filter.Page,
		PerPage:    filter.Limit,
		Total:      total,
		TotalPages: int((total + int64(filter.Limit) - 1) / int64(filter.Limit)),
	}

	response.Paginated(c, http.StatusOK, "Inquiries retrieved successfully", inquiries, pagination)
}

// MarkAsRead godoc
// @Summary Mark inquiry as read
// @Description Mark an inquiry as read (seller only)
// @Tags inquiries
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Inquiry ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /inquiries/{id}/read [put]
func (h *InquiryHandler) MarkAsRead(c *gin.Context) {
	userID, _ := c.Get("user_id")

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid inquiry ID", err)
		return
	}

	if err := h.inquiryUseCase.MarkAsRead(c.Request.Context(), id, userID.(int64)); err != nil {
		switch err {
		case domain.ErrInquiryNotFound:
			response.Error(c, http.StatusNotFound, "Inquiry not found", err)
		case domain.ErrInquiryNotOwner:
			response.Error(c, http.StatusForbidden, "Not authorized", err)
		default:
			response.Error(c, http.StatusInternalServerError, "Failed to mark as read", err)
		}
		return
	}

	response.Success(c, http.StatusOK, "Inquiry marked as read", nil)
}

// ================= REVIEW HANDLER =================

type ReviewHandler struct {
	reviewUseCase usecase.ReviewUseCase
}

func NewReviewHandler(reviewUseCase usecase.ReviewUseCase) *ReviewHandler {
	return &ReviewHandler{
		reviewUseCase: reviewUseCase,
	}
}

// Create godoc
// @Summary Create a review
// @Description Create a review for a user
// @Tags reviews
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.CreateReviewRequest true "Review details"
// @Success 201 {object} response.Response{data=domain.Review}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /reviews [post]
func (h *ReviewHandler) Create(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req domain.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	review, err := h.reviewUseCase.Create(c.Request.Context(), &req, userID.(int64))
	if err != nil {
		switch err {
		case domain.ErrCannotReviewSelf:
			response.Error(c, http.StatusBadRequest, "You cannot review yourself", err)
		case domain.ErrReviewAlreadyExists:
			response.Error(c, http.StatusConflict, "You have already reviewed this user", err)
		default:
			response.Error(c, http.StatusInternalServerError, "Failed to create review", err)
		}
		return
	}

	response.Success(c, http.StatusCreated, "Review created successfully", review)
}

// List godoc
// @Summary List reviews
// @Description Get list of reviews with optional filters
// @Tags reviews
// @Accept json
// @Produce json
// @Param reviewed_user_id query int false "Reviewed user ID"
// @Param min_rating query int false "Minimum rating (1-5)"
// @Param max_rating query int false "Maximum rating (1-5)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} response.PaginatedResponse{data=[]domain.Review}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /reviews [get]
func (h *ReviewHandler) List(c *gin.Context) {
	var filter domain.ReviewFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid query parameters", err)
		return
	}

	filter.SetDefaults()

	reviews, total, err := h.reviewUseCase.List(c.Request.Context(), &filter)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to list reviews", err)
		return
	}

	pagination := response.PaginationMeta{
		Page:       filter.Page,
		PerPage:    filter.Limit,
		Total:      total,
		TotalPages: int((total + int64(filter.Limit) - 1) / int64(filter.Limit)),
	}

	response.Paginated(c, http.StatusOK, "Reviews retrieved successfully", reviews, pagination)
}

// Update godoc
// @Summary Update a review
// @Description Update an existing review (owner only)
// @Tags reviews
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Review ID"
// @Param request body domain.UpdateReviewRequest true "Update data"
// @Success 200 {object} response.Response{data=domain.Review}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /reviews/{id} [put]
func (h *ReviewHandler) Update(c *gin.Context) {
	userID, _ := c.Get("user_id")

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid review ID", err)
		return
	}

	var req domain.UpdateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	review, err := h.reviewUseCase.Update(c.Request.Context(), id, &req, userID.(int64))
	if err != nil {
		switch err {
		case domain.ErrReviewNotFound:
			response.Error(c, http.StatusNotFound, "Review not found", err)
		case domain.ErrReviewNotOwner:
			response.Error(c, http.StatusForbidden, "You are not the owner of this review", err)
		default:
			response.Error(c, http.StatusInternalServerError, "Failed to update review", err)
		}
		return
	}

	response.Success(c, http.StatusOK, "Review updated successfully", review)
}

// Delete godoc
// @Summary Delete a review
// @Description Delete a review (owner only)
// @Tags reviews
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Review ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /reviews/{id} [delete]
func (h *ReviewHandler) Delete(c *gin.Context) {
	userID, _ := c.Get("user_id")

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid review ID", err)
		return
	}

	if err := h.reviewUseCase.Delete(c.Request.Context(), id, userID.(int64)); err != nil {
		switch err {
		case domain.ErrReviewNotFound:
			response.Error(c, http.StatusNotFound, "Review not found", err)
		case domain.ErrReviewNotOwner:
			response.Error(c, http.StatusForbidden, "You are not the owner of this review", err)
		default:
			response.Error(c, http.StatusInternalServerError, "Failed to delete review", err)
		}
		return
	}

	response.Success(c, http.StatusOK, "Review deleted successfully", nil)
}

// GetUserStats godoc
// @Summary Get user review stats
// @Description Get review statistics for a user
// @Tags reviews
// @Accept json
// @Produce json
// @Param user_id path int true "User ID"
// @Success 200 {object} response.Response{data=domain.ReviewStats}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /reviews/users/{user_id}/stats [get]
func (h *ReviewHandler) GetUserStats(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	stats, err := h.reviewUseCase.GetUserStats(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get user stats", err)
		return
	}

	response.Success(c, http.StatusOK, "User stats retrieved successfully", stats)
}

// ================= FAVORITE HANDLER =================

type FavoriteHandler struct {
	favoriteUseCase usecase.FavoriteUseCase
}

func NewFavoriteHandler(favoriteUseCase usecase.FavoriteUseCase) *FavoriteHandler {
	return &FavoriteHandler{
		favoriteUseCase: favoriteUseCase,
	}
}

// Add godoc
// @Summary Add to favorites
// @Description Add a car listing to favorites
// @Tags favorites
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Car Listing ID"
// @Success 201 {object} response.Response{data=domain.Favorite}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /favorites/{id} [post]
func (h *FavoriteHandler) Add(c *gin.Context) {
	userID, _ := c.Get("user_id")

	carListingID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid car listing ID", err)
		return
	}

	favorite, err := h.favoriteUseCase.Add(c.Request.Context(), carListingID, userID.(int64))
	if err != nil {
		switch err {
		case domain.ErrCannotFavoriteOwn:
			response.Error(c, http.StatusBadRequest, "You cannot favorite your own listing", err)
		case domain.ErrFavoriteAlreadyExists:
			response.Error(c, http.StatusConflict, "Already in favorites", err)
		case domain.ErrCarListingNotFound:
			response.Error(c, http.StatusNotFound, "Car listing not found", err)
		default:
			response.Error(c, http.StatusInternalServerError, "Failed to add to favorites", err)
		}
		return
	}

	response.Success(c, http.StatusCreated, "Added to favorites successfully", favorite)
}

// Remove godoc
// @Summary Remove from favorites
// @Description Remove a car listing from favorites
// @Tags favorites
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Car Listing ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /favorites/{id} [delete]
func (h *FavoriteHandler) Remove(c *gin.Context) {
	userID, _ := c.Get("user_id")

	carListingID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid car listing ID", err)
		return
	}

	if err := h.favoriteUseCase.Remove(c.Request.Context(), carListingID, userID.(int64)); err != nil {
		switch err {
		case domain.ErrFavoriteNotFound:
			response.Error(c, http.StatusNotFound, "Favorite not found", err)
		default:
			response.Error(c, http.StatusInternalServerError, "Failed to remove from favorites", err)
		}
		return
	}

	response.Success(c, http.StatusOK, "Removed from favorites successfully", nil)
}

// List godoc
// @Summary List favorites
// @Description Get list of user's favorite car listings
// @Tags favorites
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} response.PaginatedResponse{data=[]domain.Favorite}
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /favorites [get]
func (h *FavoriteHandler) List(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var filter domain.FavoriteFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid query parameters", err)
		return
	}

	filter.UserID = userID.(int64)
	filter.SetDefaults()

	favorites, total, err := h.favoriteUseCase.List(c.Request.Context(), &filter)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to list favorites", err)
		return
	}

	pagination := response.PaginationMeta{
		Page:       filter.Page,
		PerPage:    filter.Limit,
		Total:      total,
		TotalPages: int((total + int64(filter.Limit) - 1) / int64(filter.Limit)),
	}

	response.Paginated(c, http.StatusOK, "Favorites retrieved successfully", favorites, pagination)
}
