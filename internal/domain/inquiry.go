package domain

import "time"

// Inquiry represents a message/inquiry about a car listing
type Inquiry struct {
	ID           int64     `json:"id"`
	CarListingID int64     `json:"car_listing_id"`
	BuyerID      int64     `json:"buyer_id"`
	SellerID     int64     `json:"seller_id"`
	Subject      string    `json:"subject,omitempty"`
	Message      string    `json:"message"`
	IsRead       bool      `json:"is_read"`
	CreatedAt    time.Time `json:"created_at"`

	// Related entities
	CarListing *CarListing `json:"car_listing,omitempty"`
	Buyer      *User       `json:"buyer,omitempty"`
	Seller     *User       `json:"seller,omitempty"`
}

// CreateInquiryRequest represents a request to create an inquiry
type CreateInquiryRequest struct {
	CarListingID int64  `json:"car_listing_id" binding:"required"`
	Subject      string `json:"subject,omitempty" binding:"max=255"`
	Message      string `json:"message" binding:"required,min=10,max=1000"`
}

// InquiryFilter represents filters for listing inquiries
type InquiryFilter struct {
	CarListingID *int64 `form:"car_listing_id"`
	BuyerID      *int64 `form:"buyer_id"`
	SellerID     *int64 `form:"seller_id"`
	IsRead       *bool  `form:"is_read"`
	Page         int    `form:"page" binding:"min=1"`
	Limit        int    `form:"limit" binding:"min=1,max=100"`
}

// SetDefaults sets default values for pagination
func (f *InquiryFilter) SetDefaults() {
	if f.Page == 0 {
		f.Page = 1
	}
	if f.Limit == 0 {
		f.Limit = 20
	}
}

// MarkAsRead marks the inquiry as read
func (i *Inquiry) MarkAsRead() {
	i.IsRead = true
}
