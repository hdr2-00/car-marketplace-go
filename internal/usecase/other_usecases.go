package usecase

import (
	"car-marketplace/internal/domain"
	"car-marketplace/internal/repository"
	"context"
	"time"
)

// ================= INQUIRY USE CASE =================

type InquiryUseCase interface {
	Create(ctx context.Context, req *domain.CreateInquiryRequest, buyerID int64) (*domain.Inquiry, error)
	GetByID(ctx context.Context, id int64, userID int64) (*domain.Inquiry, error)
	List(ctx context.Context, filter *domain.InquiryFilter, userID int64) ([]domain.Inquiry, int64, error)
	MarkAsRead(ctx context.Context, id int64, userID int64) error
}

type inquiryUseCase struct {
	inquiryRepo repository.InquiryRepository
	carRepo     repository.CarRepository
}

func NewInquiryUseCase(inquiryRepo repository.InquiryRepository, carRepo repository.CarRepository) InquiryUseCase {
	return &inquiryUseCase{
		inquiryRepo: inquiryRepo,
		carRepo:     carRepo,
	}
}

func (uc *inquiryUseCase) Create(ctx context.Context, req *domain.CreateInquiryRequest, buyerID int64) (*domain.Inquiry, error) {
	// Get car listing
	listing, err := uc.carRepo.GetByID(ctx, req.CarListingID)
	if err != nil {
		return nil, err
	}

	// Check if buyer is not the seller
	if listing.SellerID == buyerID {
		return nil, domain.ErrCannotInquireOwn
	}

	inquiry := &domain.Inquiry{
		CarListingID: req.CarListingID,
		BuyerID:      buyerID,
		SellerID:     listing.SellerID,
		Subject:      req.Subject,
		Message:      req.Message,
		IsRead:       false,
		CreatedAt:    time.Now(),
	}

	if err := uc.inquiryRepo.Create(ctx, inquiry); err != nil {
		return nil, err
	}

	return inquiry, nil
}

func (uc *inquiryUseCase) GetByID(ctx context.Context, id int64, userID int64) (*domain.Inquiry, error) {
	inquiry, err := uc.inquiryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if user is buyer or seller
	if inquiry.BuyerID != userID && inquiry.SellerID != userID {
		return nil, domain.ErrInquiryNotOwner
	}

	return inquiry, nil
}

func (uc *inquiryUseCase) List(ctx context.Context, filter *domain.InquiryFilter, userID int64) ([]domain.Inquiry, int64, error) {
	filter.SetDefaults()

	// Ensure user can only see their own inquiries
	// If neither buyer_id nor seller_id is set, filter by user_id
	if filter.BuyerID == nil && filter.SellerID == nil {
		filter.BuyerID = &userID
		// Also include inquiries where user is the seller
		// This requires a more complex query, so we'll handle it differently
	}

	return uc.inquiryRepo.List(ctx, filter)
}

func (uc *inquiryUseCase) MarkAsRead(ctx context.Context, id int64, userID int64) error {
	inquiry, err := uc.inquiryRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Only seller can mark as read
	if inquiry.SellerID != userID {
		return domain.ErrInquiryNotOwner
	}

	return uc.inquiryRepo.MarkAsRead(ctx, id)
}

// ================= REVIEW USE CASE =================

type ReviewUseCase interface {
	Create(ctx context.Context, req *domain.CreateReviewRequest, reviewerID int64) (*domain.Review, error)
	GetByID(ctx context.Context, id int64) (*domain.Review, error)
	Update(ctx context.Context, id int64, req *domain.UpdateReviewRequest, userID int64) (*domain.Review, error)
	Delete(ctx context.Context, id int64, userID int64) error
	List(ctx context.Context, filter *domain.ReviewFilter) ([]domain.Review, int64, error)
	GetUserStats(ctx context.Context, userID int64) (*domain.ReviewStats, error)
}

type reviewUseCase struct {
	reviewRepo repository.ReviewRepository
}

func NewReviewUseCase(reviewRepo repository.ReviewRepository) ReviewUseCase {
	return &reviewUseCase{
		reviewRepo: reviewRepo,
	}
}

func (uc *reviewUseCase) Create(ctx context.Context, req *domain.CreateReviewRequest, reviewerID int64) (*domain.Review, error) {
	// Check if reviewing self
	if req.ReviewedUserID == reviewerID {
		return nil, domain.ErrCannotReviewSelf
	}

	// Check if already reviewed
	hasReviewed, err := uc.reviewRepo.HasReviewed(ctx, reviewerID, req.ReviewedUserID, req.CarListingID)
	if err != nil {
		return nil, err
	}

	if hasReviewed {
		return nil, domain.ErrReviewAlreadyExists
	}

	now := time.Now()
	review := &domain.Review{
		ReviewerID:         reviewerID,
		ReviewedUserID:     req.ReviewedUserID,
		CarListingID:       req.CarListingID,
		Rating:             req.Rating,
		Title:              req.Title,
		Comment:            req.Comment,
		IsVerifiedPurchase: false, // Can be updated later
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if err := uc.reviewRepo.Create(ctx, review); err != nil {
		return nil, err
	}

	return review, nil
}

func (uc *reviewUseCase) GetByID(ctx context.Context, id int64) (*domain.Review, error) {
	return uc.reviewRepo.GetByID(ctx, id)
}

func (uc *reviewUseCase) Update(ctx context.Context, id int64, req *domain.UpdateReviewRequest, userID int64) (*domain.Review, error) {
	review, err := uc.reviewRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check ownership
	if !review.IsOwner(userID) {
		return nil, domain.ErrReviewNotOwner
	}

	// Update fields
	if req.Rating != nil {
		review.Rating = *req.Rating
	}

	if req.Title != nil {
		review.Title = *req.Title
	}

	if req.Comment != nil {
		review.Comment = *req.Comment
	}

	review.UpdatedAt = time.Now()

	if err := uc.reviewRepo.Update(ctx, review); err != nil {
		return nil, err
	}

	return review, nil
}

func (uc *reviewUseCase) Delete(ctx context.Context, id int64, userID int64) error {
	review, err := uc.reviewRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check ownership
	if !review.IsOwner(userID) {
		return domain.ErrReviewNotOwner
	}

	return uc.reviewRepo.Delete(ctx, id)
}

func (uc *reviewUseCase) List(ctx context.Context, filter *domain.ReviewFilter) ([]domain.Review, int64, error) {
	filter.SetDefaults()
	return uc.reviewRepo.List(ctx, filter)
}

func (uc *reviewUseCase) GetUserStats(ctx context.Context, userID int64) (*domain.ReviewStats, error) {
	return uc.reviewRepo.GetUserStats(ctx, userID)
}

// ================= FAVORITE USE CASE =================

type FavoriteUseCase interface {
	Add(ctx context.Context, carListingID int64, userID int64) (*domain.Favorite, error)
	Remove(ctx context.Context, carListingID int64, userID int64) error
	List(ctx context.Context, filter *domain.FavoriteFilter) ([]domain.Favorite, int64, error)
	Exists(ctx context.Context, carListingID int64, userID int64) (bool, error)
}

type favoriteUseCase struct {
	favoriteRepo repository.FavoriteRepository
	carRepo      repository.CarRepository
}

func NewFavoriteUseCase(favoriteRepo repository.FavoriteRepository, carRepo repository.CarRepository) FavoriteUseCase {
	return &favoriteUseCase{
		favoriteRepo: favoriteRepo,
		carRepo:      carRepo,
	}
}

func (uc *favoriteUseCase) Add(ctx context.Context, carListingID int64, userID int64) (*domain.Favorite, error) {
	// Get car listing
	listing, err := uc.carRepo.GetByID(ctx, carListingID)
	if err != nil {
		return nil, err
	}

	// Check if user is not the owner
	if listing.SellerID == userID {
		return nil, domain.ErrCannotFavoriteOwn
	}

	// Check if already favorited
	exists, err := uc.favoriteRepo.Exists(ctx, userID, carListingID)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, domain.ErrFavoriteAlreadyExists
	}

	favorite := &domain.Favorite{
		UserID:       userID,
		CarListingID: carListingID,
		CreatedAt:    time.Now(),
	}

	if err := uc.favoriteRepo.Create(ctx, favorite); err != nil {
		return nil, err
	}

	return favorite, nil
}

func (uc *favoriteUseCase) Remove(ctx context.Context, carListingID int64, userID int64) error {
	return uc.favoriteRepo.Delete(ctx, userID, carListingID)
}

func (uc *favoriteUseCase) List(ctx context.Context, filter *domain.FavoriteFilter) ([]domain.Favorite, int64, error) {
	filter.SetDefaults()
	return uc.favoriteRepo.List(ctx, filter)
}

func (uc *favoriteUseCase) Exists(ctx context.Context, carListingID int64, userID int64) (bool, error) {
	return uc.favoriteRepo.Exists(ctx, userID, carListingID)
}
