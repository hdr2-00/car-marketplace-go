package domain

import "time"

// Review represents a user review
type Review struct {
	ID                 int64     `json:"id"`
	ReviewerID         int64     `json:"reviewer_id"`
	ReviewedUserID     int64     `json:"reviewed_user_id"`
	CarListingID       *int64    `json:"car_listing_id,omitempty"`
	Rating             int       `json:"rating"`
	Title              string    `json:"title,omitempty"`
	Comment            string    `json:"comment,omitempty"`
	IsVerifiedPurchase bool      `json:"is_verified_purchase"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`

	// Related entities
	Reviewer     *User       `json:"reviewer,omitempty"`
	ReviewedUser *User       `json:"reviewed_user,omitempty"`
	CarListing   *CarListing `json:"car_listing,omitempty"`
}

// CreateReviewRequest represents a request to create a review
type CreateReviewRequest struct {
	ReviewedUserID int64  `json:"reviewed_user_id" binding:"required"`
	CarListingID   *int64 `json:"car_listing_id,omitempty"`
	Rating         int    `json:"rating" binding:"required,min=1,max=5"`
	Title          string `json:"title,omitempty" binding:"max=255"`
	Comment        string `json:"comment,omitempty" binding:"max=1000"`
}

// UpdateReviewRequest represents a request to update a review
type UpdateReviewRequest struct {
	Rating  *int    `json:"rating,omitempty" binding:"omitempty,min=1,max=5"`
	Title   *string `json:"title,omitempty" binding:"omitempty,max=255"`
	Comment *string `json:"comment,omitempty" binding:"omitempty,max=1000"`
}

// ReviewFilter represents filters for listing reviews
type ReviewFilter struct {
	ReviewerID     *int64 `form:"reviewer_id"`
	ReviewedUserID *int64 `form:"reviewed_user_id"`
	CarListingID   *int64 `form:"car_listing_id"`
	MinRating      *int   `form:"min_rating" binding:"omitempty,min=1,max=5"`
	MaxRating      *int   `form:"max_rating" binding:"omitempty,min=1,max=5"`
	Page           int    `form:"page" binding:"min=1"`
	Limit          int    `form:"limit" binding:"min=1,max=100"`
}

// SetDefaults sets default values for pagination
func (f *ReviewFilter) SetDefaults() {
	if f.Page == 0 {
		f.Page = 1
	}
	if f.Limit == 0 {
		f.Limit = 20
	}
}

// IsOwner checks if the user is the owner of the review
func (r *Review) IsOwner(userID int64) bool {
	return r.ReviewerID == userID
}

// ReviewStats represents review statistics for a user
type ReviewStats struct {
	TotalReviews  int         `json:"total_reviews"`
	AverageRating float64     `json:"average_rating"`
	RatingCounts  map[int]int `json:"rating_counts"`
}
