package usecase

import (
	"car-marketplace/internal/domain"
	"car-marketplace/internal/repository"
	"context"
	"time"
)

// referenceCacheTTL bounds how stale cached reference data can get.
const referenceCacheTTL = 10 * time.Minute

type CarUseCase interface {
	Create(ctx context.Context, req *domain.CreateCarListingRequest, sellerID int64) (*domain.CarListing, error)
	// GetByID returns a listing with its related data and records a view,
	// unless viewerID is the listing's owner (viewerID 0 means anonymous).
	GetByID(ctx context.Context, id int64, viewerID int64) (*domain.CarListing, error)
	Update(ctx context.Context, id int64, req *domain.UpdateCarListingRequest, userID int64) (*domain.CarListing, error)
	Delete(ctx context.Context, id int64, userID int64) error
	List(ctx context.Context, filter *domain.CarListingFilter) ([]domain.CarListing, int64, error)
	MarkAsSold(ctx context.Context, id int64, userID int64) error

	// Reference data
	GetAllMakes(ctx context.Context) ([]domain.CarMake, error)
	GetModelsByMakeID(ctx context.Context, makeID int64) ([]domain.CarModel, error)
	GetReferenceData(ctx context.Context) (*domain.ReferenceData, error)

	// Image operations
	AddImage(ctx context.Context, listingID int64, imageURL string, userID int64) (*domain.CarImage, error)
	DeleteImage(ctx context.Context, listingID int64, imageID int64, userID int64) error
	SetPrimaryImage(ctx context.Context, listingID int64, imageID int64, userID int64) error
}

type carUseCase struct {
	tx      repository.Transactor
	carRepo repository.CarRepository
	refs    *referenceCache
	views   *ViewCounter
}

func NewCarUseCase(tx repository.Transactor, carRepo repository.CarRepository, views *ViewCounter) CarUseCase {
	return &carUseCase{
		tx:      tx,
		carRepo: carRepo,
		refs:    newReferenceCache(carRepo, referenceCacheTTL),
		views:   views,
	}
}

func (uc *carUseCase) Create(ctx context.Context, req *domain.CreateCarListingRequest, sellerID int64) (*domain.CarListing, error) {
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
		Status:          domain.ListingStatusActive,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	refs, err := uc.refs.get(ctx)
	if err != nil {
		return nil, err
	}
	if err := refs.validate(listing, req.FeatureIDs); err != nil {
		return nil, err
	}

	err = uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := uc.carRepo.Create(ctx, listing); err != nil {
			return err
		}
		if len(req.FeatureIDs) > 0 {
			return uc.carRepo.ReplaceListingFeatures(ctx, listing.ID, req.FeatureIDs)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return uc.load(ctx, listing.ID)
}

func (uc *carUseCase) GetByID(ctx context.Context, id int64, viewerID int64) (*domain.CarListing, error) {
	listing, err := uc.load(ctx, id)
	if err != nil {
		return nil, err
	}

	if !listing.IsOwner(viewerID) {
		uc.views.Record(id)
		listing.ViewsCount++
	}

	return listing, nil
}

// load returns a listing with its reference entities, images and features.
// Its view count includes views not yet flushed to the database.
func (uc *carUseCase) load(ctx context.Context, id int64) (*domain.CarListing, error) {
	listing, err := uc.carRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	refs, err := uc.refs.get(ctx)
	if err != nil {
		return nil, err
	}
	refs.attach(listing)

	if listing.Images, err = uc.carRepo.GetListingImages(ctx, id); err != nil {
		return nil, err
	}
	if listing.Features, err = uc.carRepo.GetListingFeatures(ctx, id); err != nil {
		return nil, err
	}
	listing.ViewsCount += uc.views.Pending(id)

	return listing, nil
}

// owned returns a listing if userID owns it.
func (uc *carUseCase) owned(ctx context.Context, id int64, userID int64) (*domain.CarListing, error) {
	listing, err := uc.carRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !listing.IsOwner(userID) {
		return nil, domain.ErrCarListingNotOwner
	}

	return listing, nil
}

func (uc *carUseCase) Update(ctx context.Context, id int64, req *domain.UpdateCarListingRequest, userID int64) (*domain.CarListing, error) {
	listing, err := uc.owned(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	if !listing.CanEdit() {
		return nil, domain.ErrCarListingCannotEdit
	}

	setIfPresent(&listing.MakeID, req.MakeID)
	setIfPresent(&listing.ModelID, req.ModelID)
	setIfPresent(&listing.Year, req.Year)
	setIfPresent(&listing.BodyTypeID, req.BodyTypeID)
	setIfPresent(&listing.FuelTypeID, req.FuelTypeID)
	setIfPresent(&listing.TransmissionID, req.TransmissionID)
	setIfPresent(&listing.ColorID, req.ColorID)
	setIfPresent(&listing.ConditionID, req.ConditionID)
	setIfPresent(&listing.VIN, req.VIN)
	setIfPresent(&listing.Mileage, req.Mileage)
	setIfPresent(&listing.EngineSize, req.EngineSize)
	setIfPresent(&listing.Horsepower, req.Horsepower)
	setIfPresent(&listing.NumberOfDoors, req.NumberOfDoors)
	setIfPresent(&listing.SeatingCapacity, req.SeatingCapacity)
	setIfPresent(&listing.Price, req.Price)
	setIfPresent(&listing.Currency, req.Currency)
	setIfPresent(&listing.IsNegotiable, req.IsNegotiable)
	setIfPresent(&listing.Title, req.Title)
	setIfPresent(&listing.Description, req.Description)
	setIfPresent(&listing.City, req.City)
	setIfPresent(&listing.State, req.State)
	setIfPresent(&listing.Country, req.Country)
	setIfPresent(&listing.PostalCode, req.PostalCode)

	if req.Status != nil {
		// Selling goes through MarkAsSold so sold_at is recorded.
		if *req.Status != domain.ListingStatusActive && *req.Status != domain.ListingStatusInactive {
			return nil, domain.ErrInvalidStatus
		}
		listing.Status = *req.Status
	}

	// Validate the resulting listing, so e.g. a new model must belong to
	// the (possibly unchanged) make.
	refs, err := uc.refs.get(ctx)
	if err != nil {
		return nil, err
	}
	if err := refs.validate(listing, req.FeatureIDs); err != nil {
		return nil, err
	}

	listing.UpdatedAt = time.Now()

	err = uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := uc.carRepo.Update(ctx, listing); err != nil {
			return err
		}
		if req.FeatureIDs != nil {
			return uc.carRepo.ReplaceListingFeatures(ctx, listing.ID, req.FeatureIDs)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return uc.load(ctx, listing.ID)
}

func setIfPresent[T any](dst *T, src *T) {
	if src != nil {
		*dst = *src
	}
}

func (uc *carUseCase) Delete(ctx context.Context, id int64, userID int64) error {
	if _, err := uc.owned(ctx, id, userID); err != nil {
		return err
	}

	return uc.carRepo.Delete(ctx, id)
}

func (uc *carUseCase) List(ctx context.Context, filter *domain.CarListingFilter) ([]domain.CarListing, int64, error) {
	filter.SetDefaults()

	listings, total, err := uc.carRepo.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	refs, err := uc.refs.get(ctx)
	if err != nil {
		return nil, 0, err
	}
	for i := range listings {
		refs.attach(&listings[i])
		listings[i].ViewsCount += uc.views.Pending(listings[i].ID)
	}

	return listings, total, nil
}

func (uc *carUseCase) MarkAsSold(ctx context.Context, id int64, userID int64) error {
	listing, err := uc.owned(ctx, id, userID)
	if err != nil {
		return err
	}

	if listing.Status == domain.ListingStatusSold {
		return domain.ErrCarListingAlreadySold
	}

	listing.MarkAsSold()
	return uc.carRepo.Update(ctx, listing)
}

// Reference data methods

func (uc *carUseCase) GetAllMakes(ctx context.Context) ([]domain.CarMake, error) {
	refs, err := uc.refs.get(ctx)
	if err != nil {
		return nil, err
	}
	return refs.makes, nil
}

func (uc *carUseCase) GetModelsByMakeID(ctx context.Context, makeID int64) ([]domain.CarModel, error) {
	refs, err := uc.refs.get(ctx)
	if err != nil {
		return nil, err
	}

	models := []domain.CarModel{}
	for _, m := range refs.models {
		if m.MakeID == makeID {
			models = append(models, m)
		}
	}
	return models, nil
}

func (uc *carUseCase) GetReferenceData(ctx context.Context) (*domain.ReferenceData, error) {
	refs, err := uc.refs.get(ctx)
	if err != nil {
		return nil, err
	}
	return &refs.data, nil
}

// Image operations

func (uc *carUseCase) AddImage(ctx context.Context, listingID int64, imageURL string, userID int64) (*domain.CarImage, error) {
	if _, err := uc.owned(ctx, listingID, userID); err != nil {
		return nil, err
	}

	image := &domain.CarImage{
		CarListingID: listingID,
		ImageURL:     imageURL,
		CreatedAt:    time.Now(),
	}

	if err := uc.carRepo.CreateImage(ctx, image); err != nil {
		return nil, err
	}

	return image, nil
}

func (uc *carUseCase) DeleteImage(ctx context.Context, listingID int64, imageID int64, userID int64) error {
	if _, err := uc.owned(ctx, listingID, userID); err != nil {
		return err
	}

	image, err := uc.carRepo.GetImageByID(ctx, imageID)
	if err != nil {
		return err
	}
	if image.CarListingID != listingID {
		return domain.ErrImageNotFound
	}

	return uc.carRepo.DeleteImage(ctx, imageID)
}

func (uc *carUseCase) SetPrimaryImage(ctx context.Context, listingID int64, imageID int64, userID int64) error {
	if _, err := uc.owned(ctx, listingID, userID); err != nil {
		return err
	}

	return uc.carRepo.SetPrimaryImage(ctx, listingID, imageID)
}
