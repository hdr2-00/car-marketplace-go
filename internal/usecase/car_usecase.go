package usecase

import (
	"car-marketplace/internal/domain"
	"car-marketplace/internal/repository"
	"context"
	"time"
)

type CarUseCase interface {
	Create(ctx context.Context, req *domain.CreateCarListingRequest, sellerID int64) (*domain.CarListing, error)
	GetByID(ctx context.Context, id int64, incrementViews bool) (*domain.CarListing, error)
	Update(ctx context.Context, id int64, req *domain.UpdateCarListingRequest, userID int64) (*domain.CarListing, error)
	Delete(ctx context.Context, id int64, userID int64) error
	List(ctx context.Context, filter *domain.CarListingFilter) ([]domain.CarListing, int64, error)
	MarkAsSold(ctx context.Context, id int64, userID int64) error

	// Reference data
	GetAllMakes(ctx context.Context) ([]domain.CarMake, error)
	GetModelsByMakeID(ctx context.Context, makeID int64) ([]domain.CarModel, error)
	GetAllBodyTypes(ctx context.Context) ([]domain.BodyType, error)
	GetAllFuelTypes(ctx context.Context) ([]domain.FuelType, error)
	GetAllTransmissions(ctx context.Context) ([]domain.TransmissionType, error)
	GetAllColors(ctx context.Context) ([]domain.Color, error)
	GetAllConditions(ctx context.Context) ([]domain.CarCondition, error)
	GetAllFeatures(ctx context.Context) ([]domain.Feature, error)

	// Image operations
	AddImage(ctx context.Context, listingID int64, imageURL string, userID int64) (*domain.CarImage, error)
	DeleteImage(ctx context.Context, imageID int64, userID int64) error
	SetPrimaryImage(ctx context.Context, listingID int64, imageID int64, userID int64) error
}

type carUseCase struct {
	carRepo repository.CarRepository
}

func NewCarUseCase(carRepo repository.CarRepository) CarUseCase {
	return &carUseCase{
		carRepo: carRepo,
	}
}

func (uc *carUseCase) Create(ctx context.Context, req *domain.CreateCarListingRequest, sellerID int64) (*domain.CarListing, error) {
	// Validate reference data exists
	if _, err := uc.carRepo.GetMakeByID(ctx, req.MakeID); err != nil {
		return nil, domain.ErrInvalidMake
	}

	if _, err := uc.carRepo.GetModelByID(ctx, req.ModelID); err != nil {
		return nil, domain.ErrInvalidModel
	}

	if _, err := uc.carRepo.GetBodyTypeByID(ctx, req.BodyTypeID); err != nil {
		return nil, domain.ErrInvalidBodyType
	}

	if _, err := uc.carRepo.GetFuelTypeByID(ctx, req.FuelTypeID); err != nil {
		return nil, domain.ErrInvalidFuelType
	}

	if _, err := uc.carRepo.GetTransmissionByID(ctx, req.TransmissionID); err != nil {
		return nil, domain.ErrInvalidTransmission
	}

	if _, err := uc.carRepo.GetColorByID(ctx, req.ColorID); err != nil {
		return nil, domain.ErrInvalidColor
	}

	if _, err := uc.carRepo.GetConditionByID(ctx, req.ConditionID); err != nil {
		return nil, domain.ErrInvalidCondition
	}

	// Create listing
	now := time.Now()
	listing := &domain.CarListing{
		SellerID:        sellerID,
		MakeID:          req.MakeID,
		ModelID:         req.ModelID,
		Year:            req.Year,
		BodyTypeID:      req.BodyTypeID,
		FuelTypeID:      req.FuelTypeID,
		TransmissionID:  req.TransmissionID,
		ColorID:         req.ColorID,
		ConditionID:     req.ConditionID,
		VIN:             req.VIN,
		Mileage:         req.Mileage,
		EngineSize:      req.EngineSize,
		Horsepower:      req.Horsepower,
		NumberOfDoors:   req.NumberOfDoors,
		SeatingCapacity: req.SeatingCapacity,
		Price:           req.Price,
		Currency:        req.Currency,
		IsNegotiable:    req.IsNegotiable,
		Title:           req.Title,
		Description:     req.Description,
		City:            req.City,
		State:           req.State,
		Country:         req.Country,
		PostalCode:      req.PostalCode,
		Status:          "active",
		IsFeatured:      false,
		ViewsCount:      0,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := uc.carRepo.Create(ctx, listing); err != nil {
		return nil, err
	}

	// Add features
	if len(req.FeatureIDs) > 0 {
		if err := uc.carRepo.AddFeaturesToListing(ctx, listing.ID, req.FeatureIDs); err != nil {
			return nil, err
		}
	}

	// Load full listing with related data
	return uc.carRepo.GetByID(ctx, listing.ID)
}

func (uc *carUseCase) GetByID(ctx context.Context, id int64, incrementViews bool) (*domain.CarListing, error) {
	listing, err := uc.carRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Increment views if requested
	if incrementViews {
		_ = uc.carRepo.IncrementViews(ctx, id)
	}

	return listing, nil
}

func (uc *carUseCase) Update(ctx context.Context, id int64, req *domain.UpdateCarListingRequest, userID int64) (*domain.CarListing, error) {
	// Get existing listing
	listing, err := uc.carRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check ownership
	if !listing.IsOwner(userID) {
		return nil, domain.ErrCarListingNotOwner
	}

	// Check if can edit
	if !listing.CanEdit() {
		return nil, domain.ErrCarListingCannotEdit
	}

	// Update fields
	if req.MakeID != nil {
		if _, err := uc.carRepo.GetMakeByID(ctx, *req.MakeID); err != nil {
			return nil, domain.ErrInvalidMake
		}
		listing.MakeID = *req.MakeID
	}

	if req.ModelID != nil {
		if _, err := uc.carRepo.GetModelByID(ctx, *req.ModelID); err != nil {
			return nil, domain.ErrInvalidModel
		}
		listing.ModelID = *req.ModelID
	}

	if req.Year != nil {
		listing.Year = *req.Year
	}

	if req.BodyTypeID != nil {
		if _, err := uc.carRepo.GetBodyTypeByID(ctx, *req.BodyTypeID); err != nil {
			return nil, domain.ErrInvalidBodyType
		}
		listing.BodyTypeID = *req.BodyTypeID
	}

	if req.FuelTypeID != nil {
		if _, err := uc.carRepo.GetFuelTypeByID(ctx, *req.FuelTypeID); err != nil {
			return nil, domain.ErrInvalidFuelType
		}
		listing.FuelTypeID = *req.FuelTypeID
	}

	if req.TransmissionID != nil {
		if _, err := uc.carRepo.GetTransmissionByID(ctx, *req.TransmissionID); err != nil {
			return nil, domain.ErrInvalidTransmission
		}
		listing.TransmissionID = *req.TransmissionID
	}

	if req.ColorID != nil {
		if _, err := uc.carRepo.GetColorByID(ctx, *req.ColorID); err != nil {
			return nil, domain.ErrInvalidColor
		}
		listing.ColorID = *req.ColorID
	}

	if req.ConditionID != nil {
		if _, err := uc.carRepo.GetConditionByID(ctx, *req.ConditionID); err != nil {
			return nil, domain.ErrInvalidCondition
		}
		listing.ConditionID = *req.ConditionID
	}

	if req.VIN != nil {
		listing.VIN = *req.VIN
	}

	if req.Mileage != nil {
		listing.Mileage = *req.Mileage
	}

	if req.EngineSize != nil {
		listing.EngineSize = *req.EngineSize
	}

	if req.Horsepower != nil {
		listing.Horsepower = *req.Horsepower
	}

	if req.NumberOfDoors != nil {
		listing.NumberOfDoors = *req.NumberOfDoors
	}

	if req.SeatingCapacity != nil {
		listing.SeatingCapacity = *req.SeatingCapacity
	}

	if req.Price != nil {
		listing.Price = *req.Price
	}

	if req.Currency != nil {
		listing.Currency = *req.Currency
	}

	if req.IsNegotiable != nil {
		listing.IsNegotiable = *req.IsNegotiable
	}

	if req.Title != nil {
		listing.Title = *req.Title
	}

	if req.Description != nil {
		listing.Description = *req.Description
	}

	if req.City != nil {
		listing.City = *req.City
	}

	if req.State != nil {
		listing.State = *req.State
	}

	if req.Country != nil {
		listing.Country = *req.Country
	}

	if req.PostalCode != nil {
		listing.PostalCode = *req.PostalCode
	}

	if req.Status != nil {
		listing.Status = *req.Status
	}

	listing.UpdatedAt = time.Now()

	// Update listing
	if err := uc.carRepo.Update(ctx, listing); err != nil {
		return nil, err
	}

	// Update features if provided
	if req.FeatureIDs != nil {
		// Remove old features
		if err := uc.carRepo.RemoveAllFeaturesFromListing(ctx, listing.ID); err != nil {
			return nil, err
		}

		// Add new features
		if len(req.FeatureIDs) > 0 {
			if err := uc.carRepo.AddFeaturesToListing(ctx, listing.ID, req.FeatureIDs); err != nil {
				return nil, err
			}
		}
	}

	// Load full listing with related data
	return uc.carRepo.GetByID(ctx, listing.ID)
}

func (uc *carUseCase) Delete(ctx context.Context, id int64, userID int64) error {
	// Get existing listing
	listing, err := uc.carRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check ownership
	if !listing.IsOwner(userID) {
		return domain.ErrCarListingNotOwner
	}

	return uc.carRepo.Delete(ctx, id)
}

func (uc *carUseCase) List(ctx context.Context, filter *domain.CarListingFilter) ([]domain.CarListing, int64, error) {
	filter.SetDefaults()
	return uc.carRepo.List(ctx, filter)
}

func (uc *carUseCase) MarkAsSold(ctx context.Context, id int64, userID int64) error {
	// Get existing listing
	listing, err := uc.carRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check ownership
	if !listing.IsOwner(userID) {
		return domain.ErrCarListingNotOwner
	}

	// Check if already sold
	if listing.Status == "sold" {
		return domain.ErrCarListingAlreadySold
	}

	// Mark as sold
	listing.MarkAsSold()
	return uc.carRepo.Update(ctx, listing)
}

// Reference data methods

func (uc *carUseCase) GetAllMakes(ctx context.Context) ([]domain.CarMake, error) {
	return uc.carRepo.GetAllMakes(ctx)
}

func (uc *carUseCase) GetModelsByMakeID(ctx context.Context, makeID int64) ([]domain.CarModel, error) {
	return uc.carRepo.GetModelsByMakeID(ctx, makeID)
}

func (uc *carUseCase) GetAllBodyTypes(ctx context.Context) ([]domain.BodyType, error) {
	return uc.carRepo.GetAllBodyTypes(ctx)
}

func (uc *carUseCase) GetAllFuelTypes(ctx context.Context) ([]domain.FuelType, error) {
	return uc.carRepo.GetAllFuelTypes(ctx)
}

func (uc *carUseCase) GetAllTransmissions(ctx context.Context) ([]domain.TransmissionType, error) {
	return uc.carRepo.GetAllTransmissions(ctx)
}

func (uc *carUseCase) GetAllColors(ctx context.Context) ([]domain.Color, error) {
	return uc.carRepo.GetAllColors(ctx)
}

func (uc *carUseCase) GetAllConditions(ctx context.Context) ([]domain.CarCondition, error) {
	return uc.carRepo.GetAllConditions(ctx)
}

func (uc *carUseCase) GetAllFeatures(ctx context.Context) ([]domain.Feature, error) {
	return uc.carRepo.GetAllFeatures(ctx)
}

// Image operations

func (uc *carUseCase) AddImage(ctx context.Context, listingID int64, imageURL string, userID int64) (*domain.CarImage, error) {
	// Get listing to check ownership
	listing, err := uc.carRepo.GetByID(ctx, listingID)
	if err != nil {
		return nil, err
	}

	if !listing.IsOwner(userID) {
		return nil, domain.ErrCarListingNotOwner
	}

	// Get existing images to determine display order
	existingImages, _ := uc.carRepo.GetListingImages(ctx, listingID)

	image := &domain.CarImage{
		CarListingID: listingID,
		ImageURL:     imageURL,
		DisplayOrder: len(existingImages),
		IsPrimary:    len(existingImages) == 0, // First image is primary
		CreatedAt:    time.Now(),
	}

	if err := uc.carRepo.CreateImage(ctx, image); err != nil {
		return nil, err
	}

	return image, nil
}

func (uc *carUseCase) DeleteImage(ctx context.Context, imageID int64, userID int64) error {
	// Get the image to find its listing
	images, err := uc.carRepo.GetListingImages(ctx, 0)
	if err != nil {
		return err
	}

	var image *domain.CarImage
	for _, img := range images {
		if img.ID == imageID {
			image = &img
			break
		}
	}

	if image == nil {
		return domain.ErrNotFound
	}

	// Check listing ownership
	listing, err := uc.carRepo.GetByID(ctx, image.CarListingID)
	if err != nil {
		return err
	}

	if !listing.IsOwner(userID) {
		return domain.ErrCarListingNotOwner
	}

	return uc.carRepo.DeleteImage(ctx, imageID)
}

func (uc *carUseCase) SetPrimaryImage(ctx context.Context, listingID int64, imageID int64, userID int64) error {
	// Check listing ownership
	listing, err := uc.carRepo.GetByID(ctx, listingID)
	if err != nil {
		return err
	}

	if !listing.IsOwner(userID) {
		return domain.ErrCarListingNotOwner
	}

	return uc.carRepo.SetPrimaryImage(ctx, listingID, imageID)
}
