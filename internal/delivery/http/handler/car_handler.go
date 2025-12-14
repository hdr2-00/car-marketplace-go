package handler

import (
	"car-marketplace/internal/domain"
	"car-marketplace/internal/usecase"
	"car-marketplace/pkg/response"
	"net/http"
	"strconv"

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
// @Failure 500 {object} response.Response
// @Router /cars [post]
func (h *CarHandler) Create(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req domain.CreateCarListingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	listing, err := h.carUseCase.Create(c.Request.Context(), &req, userID.(int64))
	if err != nil {
		switch err {
		case domain.ErrInvalidMake, domain.ErrInvalidModel, domain.ErrInvalidBodyType,
			domain.ErrInvalidFuelType, domain.ErrInvalidTransmission, domain.ErrInvalidColor,
			domain.ErrInvalidCondition:
			response.Error(c, http.StatusBadRequest, err.Error(), err)
		default:
			response.Error(c, http.StatusInternalServerError, "Failed to create listing", err)
		}
		return
	}

	response.Success(c, http.StatusCreated, "Car listing created successfully", listing)
}

// GetByID godoc
// @Summary Get car listing by ID
// @Description Get detailed information about a car listing
// @Tags cars
// @Accept json
// @Produce json
// @Param id path int true "Car Listing ID"
// @Success 200 {object} response.Response{data=domain.CarListing}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /cars/{id} [get]
func (h *CarHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid car listing ID", err)
		return
	}

	listing, err := h.carUseCase.GetByID(c.Request.Context(), id, true)
	if err != nil {
		if err == domain.ErrCarListingNotFound {
			response.Error(c, http.StatusNotFound, "Car listing not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to get car listing", err)
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
// @Param city query string false "City"
// @Param state query string false "State"
// @Param country query string false "Country"
// @Param search query string false "Search in title and description"
// @Param sort_by query string false "Sort by (price_asc, price_desc, year_asc, year_desc, created_asc, created_desc)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} response.PaginatedResponse{data=[]domain.CarListing}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /cars [get]
func (h *CarHandler) List(c *gin.Context) {
	var filter domain.CarListingFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid query parameters", err)
		return
	}

	filter.SetDefaults()

	listings, total, err := h.carUseCase.List(c.Request.Context(), &filter)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to list car listings", err)
		return
	}

	pagination := response.PaginationMeta{
		Page:       filter.Page,
		PerPage:    filter.Limit,
		Total:      total,
		TotalPages: int((total + int64(filter.Limit) - 1) / int64(filter.Limit)),
	}

	response.Paginated(c, http.StatusOK, "Car listings retrieved successfully", listings, pagination)
}

// Update godoc
// @Summary Update car listing
// @Description Update an existing car listing (owner only)
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
// @Failure 500 {object} response.Response
// @Router /cars/{id} [put]
func (h *CarHandler) Update(c *gin.Context) {
	userID, _ := c.Get("user_id")

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid car listing ID", err)
		return
	}

	var req domain.UpdateCarListingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	listing, err := h.carUseCase.Update(c.Request.Context(), id, &req, userID.(int64))
	if err != nil {
		switch err {
		case domain.ErrCarListingNotFound:
			response.Error(c, http.StatusNotFound, "Car listing not found", err)
		case domain.ErrCarListingNotOwner:
			response.Error(c, http.StatusForbidden, "You are not the owner of this listing", err)
		case domain.ErrCarListingCannotEdit:
			response.Error(c, http.StatusBadRequest, "This listing cannot be edited", err)
		default:
			response.Error(c, http.StatusInternalServerError, "Failed to update listing", err)
		}
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
	userID, _ := c.Get("user_id")

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid car listing ID", err)
		return
	}

	if err := h.carUseCase.Delete(c.Request.Context(), id, userID.(int64)); err != nil {
		switch err {
		case domain.ErrCarListingNotFound:
			response.Error(c, http.StatusNotFound, "Car listing not found", err)
		case domain.ErrCarListingNotOwner:
			response.Error(c, http.StatusForbidden, "You are not the owner of this listing", err)
		default:
			response.Error(c, http.StatusInternalServerError, "Failed to delete listing", err)
		}
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
// @Failure 500 {object} response.Response
// @Router /cars/{id}/sold [post]
func (h *CarHandler) MarkAsSold(c *gin.Context) {
	userID, _ := c.Get("user_id")

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid car listing ID", err)
		return
	}

	if err := h.carUseCase.MarkAsSold(c.Request.Context(), id, userID.(int64)); err != nil {
		switch err {
		case domain.ErrCarListingNotFound:
			response.Error(c, http.StatusNotFound, "Car listing not found", err)
		case domain.ErrCarListingNotOwner:
			response.Error(c, http.StatusForbidden, "You are not the owner of this listing", err)
		case domain.ErrCarListingAlreadySold:
			response.Error(c, http.StatusBadRequest, "This car has already been sold", err)
		default:
			response.Error(c, http.StatusInternalServerError, "Failed to mark as sold", err)
		}
		return
	}

	response.Success(c, http.StatusOK, "Car listing marked as sold", nil)
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
		response.Error(c, http.StatusInternalServerError, "Failed to get makes", err)
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
	makeID, err := strconv.ParseInt(c.Param("make_id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid make ID", err)
		return
	}

	models, err := h.carUseCase.GetModelsByMakeID(c.Request.Context(), makeID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get models", err)
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
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /reference/data [get]
func (h *CarHandler) GetReferenceData(c *gin.Context) {
	bodyTypes, _ := h.carUseCase.GetAllBodyTypes(c.Request.Context())
	fuelTypes, _ := h.carUseCase.GetAllFuelTypes(c.Request.Context())
	transmissions, _ := h.carUseCase.GetAllTransmissions(c.Request.Context())
	colors, _ := h.carUseCase.GetAllColors(c.Request.Context())
	conditions, _ := h.carUseCase.GetAllConditions(c.Request.Context())
	features, _ := h.carUseCase.GetAllFeatures(c.Request.Context())

	data := gin.H{
		"body_types":    bodyTypes,
		"fuel_types":    fuelTypes,
		"transmissions": transmissions,
		"colors":        colors,
		"conditions":    conditions,
		"features":      features,
	}

	response.Success(c, http.StatusOK, "Reference data retrieved successfully", data)
}
