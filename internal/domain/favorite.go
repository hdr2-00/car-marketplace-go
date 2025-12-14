package domain

import "time"

// Favorite represents a user's favorite car listing
type Favorite struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	CarListingID int64     `json:"car_listing_id"`
	CreatedAt    time.Time `json:"created_at"`

	// Related entities
	CarListing *CarListing `json:"car_listing,omitempty"`
}

// FavoriteFilter represents filters for listing favorites
type FavoriteFilter struct {
	UserID int64 `form:"user_id"`
	Page   int   `form:"page" binding:"min=1"`
	Limit  int   `form:"limit" binding:"min=1,max=100"`
}

// SetDefaults sets default values for pagination
func (f *FavoriteFilter) SetDefaults() {
	if f.Page == 0 {
		f.Page = 1
	}
	if f.Limit == 0 {
		f.Limit = 20
	}
}
