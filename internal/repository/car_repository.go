package repository

import (
	"car-marketplace/internal/domain"
	"context"
)

// CarRepository defines the interface for car listing data operations
type CarRepository interface {
	// Car listing CRUD
	Create(ctx context.Context, listing *domain.CarListing) error
	GetByID(ctx context.Context, id int64) (*domain.CarListing, error)
	Update(ctx context.Context, listing *domain.CarListing) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, filter *domain.CarListingFilter) ([]domain.CarListing, int64, error)
	IncrementViews(ctx context.Context, id int64) error

	// Car make operations
	GetMakeByID(ctx context.Context, id int64) (*domain.CarMake, error)
	GetAllMakes(ctx context.Context) ([]domain.CarMake, error)

	// Car model operations
	GetModelByID(ctx context.Context, id int64) (*domain.CarModel, error)
	GetModelsByMakeID(ctx context.Context, makeID int64) ([]domain.CarModel, error)

	// Reference data operations
	GetBodyTypeByID(ctx context.Context, id int64) (*domain.BodyType, error)
	GetAllBodyTypes(ctx context.Context) ([]domain.BodyType, error)

	GetFuelTypeByID(ctx context.Context, id int64) (*domain.FuelType, error)
	GetAllFuelTypes(ctx context.Context) ([]domain.FuelType, error)

	GetTransmissionByID(ctx context.Context, id int64) (*domain.TransmissionType, error)
	GetAllTransmissions(ctx context.Context) ([]domain.TransmissionType, error)

	GetColorByID(ctx context.Context, id int64) (*domain.Color, error)
	GetAllColors(ctx context.Context) ([]domain.Color, error)

	GetConditionByID(ctx context.Context, id int64) (*domain.CarCondition, error)
	GetAllConditions(ctx context.Context) ([]domain.CarCondition, error)

	// Feature operations
	GetFeatureByID(ctx context.Context, id int64) (*domain.Feature, error)
	GetAllFeatures(ctx context.Context) ([]domain.Feature, error)
	GetFeaturesByCategory(ctx context.Context, category string) ([]domain.Feature, error)
	AddFeaturesToListing(ctx context.Context, listingID int64, featureIDs []int64) error
	RemoveAllFeaturesFromListing(ctx context.Context, listingID int64) error
	GetListingFeatures(ctx context.Context, listingID int64) ([]domain.Feature, error)

	// Image operations
	CreateImage(ctx context.Context, image *domain.CarImage) error
	GetListingImages(ctx context.Context, listingID int64) ([]domain.CarImage, error)
	DeleteImage(ctx context.Context, id int64) error
	SetPrimaryImage(ctx context.Context, listingID int64, imageID int64) error
}
