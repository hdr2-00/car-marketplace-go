package handler

import (
	"car-marketplace/internal/domain"
	"car-marketplace/internal/usecase"
	"car-marketplace/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CarHandler struct {
	carUseCase usecase.CarUseCase
}

func NewCarHandler(carUseCase usecase.CarUseCase) *CarHandler {
	return &CarHandler{
		carUseCase: carUseCase,
	}
}

// Create godoc
// @Summary Create a new car listing
// @Description Create a new car listing (authentication required)
// @Tags cars
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.CreateCarListingRequest true "Car listing details"
// @Success 201 {object} response.Response{data=domain.CarListing}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 409 {object} response.Response "VIN already listed"
// @Failure 500 {object} response.Response
// @Router /cars [post]
func (h *CarHandler) Create(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}

	var req domain.CreateCarListingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindError(c, "Invalid request body", err)
		return
	}

	listing, err := h.carUseCase.Create(c.Request.Context(), &req, userID)
	if err != nil {
		respondError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "Car listing created successfully", listing)
}

// GetByID godoc
// @Summary Get car listing by ID
// @Description Get detailed information about a car listing, including its make, model, images and features. Views by the owner are not counted.
// @Tags cars
// @Accept json
// @Produce json
// @Param id path int true "Car Listing ID"
// @Success 200 {object} response.Response{data=domain.CarListing}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /cars/{id} [get]
func (h *CarHandler) GetByID(c *gin.Context) {
	id, ok := pathID(c, "id", "car listing ID")
	if !ok {
		return
	}

	listing, err := h.carUseCase.GetByID(c.Request.Context(), id, currentUserID(c))
	if err != nil {
		respondError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Car listing retrieved successfully", listing)
}

// List godoc
// @Summary List car listings
// @Description Get a paginated list of car listings with optional filters
// @Tags cars
// @Accept json
// @Produce json
// @Param make_id query int false "Make ID"
// @Param model_id query int false "Model ID"
// @Param min_year query int false "Minimum year"
// @Param max_year query int false "Maximum year"
// @Param min_price query number false "Minimum price"
// @Param max_price query number false "Maximum price"
// @Param city query string false "City (case-insensitive)"
// @Param state query string false "State (case-insensitive)"
// @Param country query string false "Country (case-insensitive)"
// @Param search query string false "Full-text search in title and description (matches word prefixes)"
// @Param sort_by query string false "Sort by (price_asc, price_desc, year_asc, year_desc, mileage_asc, mileage_desc, created_asc, created_desc)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} response.PaginatedResponse{data=[]domain.CarListing}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /cars [get]
func (h *CarHandler) List(c *gin.Context) {
	var filter domain.CarListingFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		bindError(c, "Invalid query parameters", err)
		return
	}

	listings, total, err := h.carUseCase.List(c.Request.Context(), &filter)
	if err != nil {
		respondError(c, err)
		return
	}

	response.Paginated(c, http.StatusOK, "Car listings retrieved successfully", listings, response.PaginationMeta{
		Page:       filter.Page,
		PerPage:    filter.Limit,
		Total:      total,
		TotalPages: totalPages(total, filter.Limit),
	})
}

// Update godoc
// @Summary Update car listing
// @Description Update an existing car listing (owner only). Status may only be set to active or inactive; use the sold endpoint to mark a car as sold.
// @Tags cars
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Car Listing ID"
// @Param request body domain.UpdateCarListingRequest true "Update data"
// @Success 200 {object} response.Response{data=domain.CarListing}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /cars/{id} [put]
func (h *CarHandler) Update(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}

	id, ok := pathID(c, "id", "car listing ID")
	if !ok {
		return
	}

	var req domain.UpdateCarListingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindError(c, "Invalid request body", err)
		return
	}

	listing, err := h.carUseCase.Update(c.Request.Context(), id, &req, userID)
	if err != nil {
		respondError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Car listing updated successfully", listing)
}

// Delete godoc
// @Summary Delete car listing
// @Description Delete a car listing (owner only)
// @Tags cars
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Car Listing ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /cars/{id} [delete]
func (h *CarHandler) Delete(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}

	id, ok := pathID(c, "id", "car listing ID")
	if !ok {
		return
	}

	if err := h.carUseCase.Delete(c.Request.Context(), id, userID); err != nil {
		respondError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Car listing deleted successfully", nil)
}

// MarkAsSold godoc
// @Summary Mark car listing as sold
// @Description Mark a car listing as sold (owner only)
// @Tags cars
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Car Listing ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 409 {object} response.Response "Already sold"
// @Failure 500 {object} response.Response
// @Router /cars/{id}/sold [post]
func (h *CarHandler) MarkAsSold(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}

	id, ok := pathID(c, "id", "car listing ID")
	if !ok {
		return
	}

	if err := h.carUseCase.MarkAsSold(c.Request.Context(), id, userID); err != nil {
		respondError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Car listing marked as sold", nil)
}

// AddImage godoc
// @Summary Add an image to a listing
// @Description Append an image URL to a listing (owner only). The first image becomes the primary image.
// @Tags cars
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Car Listing ID"
// @Param request body domain.AddImageRequest true "Image"
// @Success 201 {object} response.Response{data=domain.CarImage}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /cars/{id}/images [post]
func (h *CarHandler) AddImage(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}

	id, ok := pathID(c, "id", "car listing ID")
	if !ok {
		return
	}

	var req domain.AddImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindError(c, "Invalid request body", err)
		return
	}

	image, err := h.carUseCase.AddImage(c.Request.Context(), id, req.ImageURL, userID)
	if err != nil {
		respondError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "Image added successfully", image)
}

// DeleteImage godoc
// @Summary Delete a listing image
// @Description Delete an image from a listing (owner only). If it was the primary image, the next image becomes primary.
// @Tags cars
// @Produce json
// @Security BearerAuth
// @Param id path int true "Car Listing ID"
// @Param image_id path int true "Image ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /cars/{id}/images/{image_id} [delete]
func (h *CarHandler) DeleteImage(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}

	id, ok := pathID(c, "id", "car listing ID")
	if !ok {
		return
	}
	imageID, ok := pathID(c, "image_id", "image ID")
	if !ok {
		return
	}

	if err := h.carUseCase.DeleteImage(c.Request.Context(), id, imageID, userID); err != nil {
		respondError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Image deleted successfully", nil)
}

// SetPrimaryImage godoc
// @Summary Set a listing's primary image
// @Description Make an image the listing's primary image (owner only)
// @Tags cars
// @Produce json
// @Security BearerAuth
// @Param id path int true "Car Listing ID"
// @Param image_id path int true "Image ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /cars/{id}/images/{image_id}/primary [put]
func (h *CarHandler) SetPrimaryImage(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}

	id, ok := pathID(c, "id", "car listing ID")
	if !ok {
		return
	}
	imageID, ok := pathID(c, "image_id", "image ID")
	if !ok {
		return
	}

	if err := h.carUseCase.SetPrimaryImage(c.Request.Context(), id, imageID, userID); err != nil {
		respondError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Primary image updated", nil)
}

// GetMakes godoc
// @Summary Get all car makes
// @Description Get list of all car makes
// @Tags reference-data
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]domain.CarMake}
// @Failure 500 {object} response.Response
// @Router /reference/makes [get]
func (h *CarHandler) GetMakes(c *gin.Context) {
	makes, err := h.carUseCase.GetAllMakes(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Makes retrieved successfully", makes)
}

// GetModels godoc
// @Summary Get models by make
// @Description Get all models for a specific make
// @Tags reference-data
// @Accept json
// @Produce json
// @Param make_id path int true "Make ID"
// @Success 200 {object} response.Response{data=[]domain.CarModel}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /reference/makes/{make_id}/models [get]
func (h *CarHandler) GetModels(c *gin.Context) {
	makeID, ok := pathID(c, "make_id", "make ID")
	if !ok {
		return
	}

	models, err := h.carUseCase.GetModelsByMakeID(c.Request.Context(), makeID)
	if err != nil {
		respondError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Models retrieved successfully", models)
}

// GetReferenceData godoc
// @Summary Get all reference data
// @Description Get all reference data (body types, fuel types, etc.)
// @Tags reference-data
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=domain.ReferenceData}
// @Failure 500 {object} response.Response
// @Router /reference/data [get]
func (h *CarHandler) GetReferenceData(c *gin.Context) {
	data, err := h.carUseCase.GetReferenceData(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Reference data retrieved successfully", data)
}
