package handler

import (
	"car-marketplace/internal/domain"
	"car-marketplace/internal/usecase"
	"car-marketplace/pkg/response"
	"net/http"

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
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /inquiries [post]
func (h *InquiryHandler) Create(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}

	var req domain.CreateInquiryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindError(c, "Invalid request body", err)
		return
	}

	inquiry, err := h.inquiryUseCase.Create(c.Request.Context(), &req, userID)
	if err != nil {
		respondError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "Inquiry sent successfully", inquiry)
}

// List godoc
// @Summary List inquiries
// @Description List inquiries the authenticated user sent or received. buyer_id / seller_id may only be the user's own ID.
// @Tags inquiries
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param car_listing_id query int false "Car listing ID"
// @Param buyer_id query int false "Only inquiries sent by this user (must be yourself)"
// @Param seller_id query int false "Only inquiries received by this user (must be yourself)"
// @Param is_read query boolean false "Read status"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} response.PaginatedResponse{data=[]domain.Inquiry}
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /inquiries [get]
func (h *InquiryHandler) List(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}

	var filter domain.InquiryFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		bindError(c, "Invalid query parameters", err)
		return
	}

	inquiries, total, err := h.inquiryUseCase.List(c.Request.Context(), &filter, userID)
	if err != nil {
		respondError(c, err)
		return
	}

	response.Paginated(c, http.StatusOK, "Inquiries retrieved successfully", inquiries, response.PaginationMeta{
		Page:       filter.Page,
		PerPage:    filter.Limit,
		Total:      total,
		TotalPages: totalPages(total, filter.Limit),
	})
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
	userID, ok := requireUserID(c)
	if !ok {
		return
	}

	id, ok := pathID(c, "id", "inquiry ID")
	if !ok {
		return
	}

	if err := h.inquiryUseCase.MarkAsRead(c.Request.Context(), id, userID); err != nil {
		respondError(c, err)
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
// @Description Create a review for a user, optionally tied to one of their listings
// @Tags reviews
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.CreateReviewRequest true "Review details"
// @Success 201 {object} response.Response{data=domain.Review}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response "Reviewed user or listing not found"
// @Failure 409 {object} response.Response "Already reviewed"
// @Failure 500 {object} response.Response
// @Router /reviews [post]
func (h *ReviewHandler) Create(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}

	var req domain.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindError(c, "Invalid request body", err)
		return
	}

	review, err := h.reviewUseCase.Create(c.Request.Context(), &req, userID)
	if err != nil {
		respondError(c, err)
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
		bindError(c, "Invalid query parameters", err)
		return
	}

	reviews, total, err := h.reviewUseCase.List(c.Request.Context(), &filter)
	if err != nil {
		respondError(c, err)
		return
	}

	response.Paginated(c, http.StatusOK, "Reviews retrieved successfully", reviews, response.PaginationMeta{
		Page:       filter.Page,
		PerPage:    filter.Limit,
		Total:      total,
		TotalPages: totalPages(total, filter.Limit),
	})
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
	userID, ok := requireUserID(c)
	if !ok {
		return
	}

	id, ok := pathID(c, "id", "review ID")
	if !ok {
		return
	}

	var req domain.UpdateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindError(c, "Invalid request body", err)
		return
	}

	review, err := h.reviewUseCase.Update(c.Request.Context(), id, &req, userID)
	if err != nil {
		respondError(c, err)
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
	userID, ok := requireUserID(c)
	if !ok {
		return
	}

	id, ok := pathID(c, "id", "review ID")
	if !ok {
		return
	}

	if err := h.reviewUseCase.Delete(c.Request.Context(), id, userID); err != nil {
		respondError(c, err)
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
	userID, ok := pathID(c, "user_id", "user ID")
	if !ok {
		return
	}

	stats, err := h.reviewUseCase.GetUserStats(c.Request.Context(), userID)
	if err != nil {
		respondError(c, err)
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
// @Failure 404 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /favorites/{id} [post]
func (h *FavoriteHandler) Add(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}

	carListingID, ok := pathID(c, "id", "car listing ID")
	if !ok {
		return
	}

	favorite, err := h.favoriteUseCase.Add(c.Request.Context(), carListingID, userID)
	if err != nil {
		respondError(c, err)
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
	userID, ok := requireUserID(c)
	if !ok {
		return
	}

	carListingID, ok := pathID(c, "id", "car listing ID")
	if !ok {
		return
	}

	if err := h.favoriteUseCase.Remove(c.Request.Context(), carListingID, userID); err != nil {
		respondError(c, err)
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
	userID, ok := requireUserID(c)
	if !ok {
		return
	}

	var filter domain.FavoriteFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		bindError(c, "Invalid query parameters", err)
		return
	}
	filter.UserID = userID

	favorites, total, err := h.favoriteUseCase.List(c.Request.Context(), &filter)
	if err != nil {
		respondError(c, err)
		return
	}

	response.Paginated(c, http.StatusOK, "Favorites retrieved successfully", favorites, response.PaginationMeta{
		Page:       filter.Page,
		PerPage:    filter.Limit,
		Total:      total,
		TotalPages: totalPages(total, filter.Limit),
	})
}
