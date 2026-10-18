package repository

import (
	"car-marketplace/internal/domain"
	"context"
)

// InquiryRepository defines the interface for inquiry data operations
type InquiryRepository interface {
	Create(ctx context.Context, inquiry *domain.Inquiry) error
	GetByID(ctx context.Context, id int64) (*domain.Inquiry, error)
	List(ctx context.Context, filter *domain.InquiryFilter) ([]domain.Inquiry, int64, error)
	MarkAsRead(ctx context.Context, id int64) error
	Delete(ctx context.Context, id int64) error
}

// ReviewRepository defines the interface for review data operations
type ReviewRepository interface {
	Create(ctx context.Context, review *domain.Review) error
	GetByID(ctx context.Context, id int64) (*domain.Review, error)
	Update(ctx context.Context, review *domain.Review) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, filter *domain.ReviewFilter) ([]domain.Review, int64, error)
	GetUserStats(ctx context.Context, userID int64) (*domain.ReviewStats, error)
	HasReviewed(ctx context.Context, reviewerID, reviewedUserID int64, carListingID *int64) (bool, error)
}

// FavoriteRepository defines the interface for favorite data operations
type FavoriteRepository interface {
	Create(ctx context.Context, favorite *domain.Favorite) error
	Delete(ctx context.Context, userID, carListingID int64) error
	Exists(ctx context.Context, userID, carListingID int64) (bool, error)
	List(ctx context.Context, filter *domain.FavoriteFilter) ([]domain.Favorite, int64, error)
	GetByID(ctx context.Context, id int64) (*domain.Favorite, error)
}
