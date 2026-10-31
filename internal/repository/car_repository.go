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
	// IncrementViews adds the given view counts, keyed by listing ID.
	IncrementViews(ctx context.Context, counts map[int64]int) error

	// Reference data
	GetAllMakes(ctx context.Context) ([]domain.CarMake, error)
	GetAllModels(ctx context.Context) ([]domain.CarModel, error)
	GetAllBodyTypes(ctx context.Context) ([]domain.BodyType, error)
	GetAllFuelTypes(ctx context.Context) ([]domain.FuelType, error)
	GetAllTransmissions(ctx context.Context) ([]domain.TransmissionType, error)
	GetAllColors(ctx context.Context) ([]domain.Color, error)
	GetAllConditions(ctx context.Context) ([]domain.CarCondition, error)
	GetAllFeatures(ctx context.Context) ([]domain.Feature, error)

	// Feature operations
	ReplaceListingFeatures(ctx context.Context, listingID int64, featureIDs []int64) error
	GetListingFeatures(ctx context.Context, listingID int64) ([]domain.Feature, error)

	// Image operations
	CreateImage(ctx context.Context, image *domain.CarImage) error
	GetImageByID(ctx context.Context, id int64) (*domain.CarImage, error)
	GetListingImages(ctx context.Context, listingID int64) ([]domain.CarImage, error)
	DeleteImage(ctx context.Context, id int64) error
	SetPrimaryImage(ctx context.Context, listingID int64, imageID int64) error
}
