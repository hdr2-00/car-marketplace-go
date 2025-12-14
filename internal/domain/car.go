package domain

import "time"

// CarListing represents a car listing in the system
type CarListing struct {
	ID              int64      `json:"id"`
	SellerID        int64      `json:"seller_id"`
	MakeID          int64      `json:"make_id"`
	ModelID         int64      `json:"model_id"`
	Year            int        `json:"year"`
	BodyTypeID      int64      `json:"body_type_id"`
	FuelTypeID      int64      `json:"fuel_type_id"`
	TransmissionID  int64      `json:"transmission_type_id"`
	ColorID         int64      `json:"color_id"`
	ConditionID     int64      `json:"condition_id"`
	VIN             string     `json:"vin,omitempty"`
	Mileage         int        `json:"mileage"`
	EngineSize      float64    `json:"engine_size"`
	Horsepower      int        `json:"horsepower,omitempty"`
	NumberOfDoors   int        `json:"number_of_doors,omitempty"`
	SeatingCapacity int        `json:"seating_capacity,omitempty"`
	Price           float64    `json:"price"`
	Currency        string     `json:"currency"`
	IsNegotiable    bool       `json:"is_negotiable"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	City            string     `json:"city"`
	State           string     `json:"state"`
	Country         string     `json:"country"`
	PostalCode      string     `json:"postal_code,omitempty"`
	Status          string     `json:"status"`
	IsFeatured      bool       `json:"is_featured"`
	ViewsCount      int        `json:"views_count"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	SoldAt          *time.Time `json:"sold_at,omitempty"`

	// Related entities
	Seller       *User             `json:"seller,omitempty"`
	Make         *CarMake          `json:"make,omitempty"`
	Model        *CarModel         `json:"model,omitempty"`
	BodyType     *BodyType         `json:"body_type,omitempty"`
	FuelType     *FuelType         `json:"fuel_type,omitempty"`
	Transmission *TransmissionType `json:"transmission,omitempty"`
	Color        *Color            `json:"color,omitempty"`
	Condition    *CarCondition     `json:"condition,omitempty"`
	Images       []CarImage        `json:"images,omitempty"`
	Features     []Feature         `json:"features,omitempty"`
}

// CarMake represents a car manufacturer
type CarMake struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	LogoURL         string    `json:"logo_url,omitempty"`
	CountryOfOrigin string    `json:"country_of_origin,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// CarModel represents a car model
type CarModel struct {
	ID        int64     `json:"id"`
	MakeID    int64     `json:"make_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BodyType represents a car body type
type BodyType struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// FuelType represents a fuel type
type FuelType struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// TransmissionType represents a transmission type
type TransmissionType struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// Color represents a color
type Color struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	HexCode   string    `json:"hex_code,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// CarCondition represents a car condition
type CarCondition struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// Feature represents a car feature
type Feature struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Category  string    `json:"category,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// CarImage represents a car image
type CarImage struct {
	ID           int64     `json:"id"`
	CarListingID int64     `json:"car_listing_id"`
	ImageURL     string    `json:"image_url"`
	DisplayOrder int       `json:"display_order"`
	IsPrimary    bool      `json:"is_primary"`
	CreatedAt    time.Time `json:"created_at"`
}

// CreateCarListingRequest represents a request to create a car listing
type CreateCarListingRequest struct {
	MakeID          int64   `json:"make_id" binding:"required"`
	ModelID         int64   `json:"model_id" binding:"required"`
	Year            int     `json:"year" binding:"required,min=1900,max=2100"`
	BodyTypeID      int64   `json:"body_type_id" binding:"required"`
	FuelTypeID      int64   `json:"fuel_type_id" binding:"required"`
	TransmissionID  int64   `json:"transmission_type_id" binding:"required"`
	ColorID         int64   `json:"color_id" binding:"required"`
	ConditionID     int64   `json:"condition_id" binding:"required"`
	VIN             string  `json:"vin,omitempty"`
	Mileage         int     `json:"mileage" binding:"required,min=0"`
	EngineSize      float64 `json:"engine_size" binding:"required,min=0"`
	Horsepower      int     `json:"horsepower,omitempty" binding:"min=0"`
	NumberOfDoors   int     `json:"number_of_doors,omitempty" binding:"min=2,max=6"`
	SeatingCapacity int     `json:"seating_capacity,omitempty" binding:"min=1,max=20"`
	Price           float64 `json:"price" binding:"required,min=0"`
	Currency        string  `json:"currency" binding:"required,len=3"`
	IsNegotiable    bool    `json:"is_negotiable"`
	Title           string  `json:"title" binding:"required,min=10,max=255"`
	Description     string  `json:"description" binding:"required,min=20"`
	City            string  `json:"city" binding:"required"`
	State           string  `json:"state" binding:"required"`
	Country         string  `json:"country" binding:"required"`
	PostalCode      string  `json:"postal_code,omitempty"`
	FeatureIDs      []int64 `json:"feature_ids,omitempty"`
}

// UpdateCarListingRequest represents a request to update a car listing
type UpdateCarListingRequest struct {
	MakeID          *int64   `json:"make_id,omitempty"`
	ModelID         *int64   `json:"model_id,omitempty"`
	Year            *int     `json:"year,omitempty" binding:"omitempty,min=1900,max=2100"`
	BodyTypeID      *int64   `json:"body_type_id,omitempty"`
	FuelTypeID      *int64   `json:"fuel_type_id,omitempty"`
	TransmissionID  *int64   `json:"transmission_type_id,omitempty"`
	ColorID         *int64   `json:"color_id,omitempty"`
	ConditionID     *int64   `json:"condition_id,omitempty"`
	VIN             *string  `json:"vin,omitempty"`
	Mileage         *int     `json:"mileage,omitempty" binding:"omitempty,min=0"`
	EngineSize      *float64 `json:"engine_size,omitempty" binding:"omitempty,min=0"`
	Horsepower      *int     `json:"horsepower,omitempty" binding:"omitempty,min=0"`
	NumberOfDoors   *int     `json:"number_of_doors,omitempty" binding:"omitempty,min=2,max=6"`
	SeatingCapacity *int     `json:"seating_capacity,omitempty" binding:"omitempty,min=1,max=20"`
	Price           *float64 `json:"price,omitempty" binding:"omitempty,min=0"`
	Currency        *string  `json:"currency,omitempty" binding:"omitempty,len=3"`
	IsNegotiable    *bool    `json:"is_negotiable,omitempty"`
	Title           *string  `json:"title,omitempty" binding:"omitempty,min=10,max=255"`
	Description     *string  `json:"description,omitempty" binding:"omitempty,min=20"`
	City            *string  `json:"city,omitempty"`
	State           *string  `json:"state,omitempty"`
	Country         *string  `json:"country,omitempty"`
	PostalCode      *string  `json:"postal_code,omitempty"`
	Status          *string  `json:"status,omitempty" binding:"omitempty,oneof=active sold pending expired inactive"`
	FeatureIDs      []int64  `json:"feature_ids,omitempty"`
}

// CarListingFilter represents filters for searching car listings
type CarListingFilter struct {
	MakeID         *int64   `form:"make_id"`
	ModelID        *int64   `form:"model_id"`
	MinYear        *int     `form:"min_year"`
	MaxYear        *int     `form:"max_year"`
	BodyTypeID     *int64   `form:"body_type_id"`
	FuelTypeID     *int64   `form:"fuel_type_id"`
	TransmissionID *int64   `form:"transmission_type_id"`
	ColorID        *int64   `form:"color_id"`
	ConditionID    *int64   `form:"condition_id"`
	MinPrice       *float64 `form:"min_price"`
	MaxPrice       *float64 `form:"max_price"`
	MinMileage     *int     `form:"min_mileage"`
	MaxMileage     *int     `form:"max_mileage"`
	City           *string  `form:"city"`
	State          *string  `form:"state"`
	Country        *string  `form:"country"`
	Status         *string  `form:"status"`
	SellerID       *int64   `form:"seller_id"`
	IsFeatured     *bool    `form:"is_featured"`
	Search         *string  `form:"search"`
	SortBy         string   `form:"sort_by" binding:"omitempty,oneof=price_asc price_desc year_asc year_desc mileage_asc mileage_desc created_asc created_desc"`
	Page           int      `form:"page" binding:"omitempty,min=1"`
	Limit          int      `form:"limit" binding:"omitempty,min=1,max=100"`
}

// SetDefaults sets default values for pagination
func (f *CarListingFilter) SetDefaults() {
	if f.Page == 0 {
		f.Page = 1
	}
	if f.Limit == 0 {
		f.Limit = 20
	}
	if f.SortBy == "" {
		f.SortBy = "created_desc"
	}
}

// IsOwner checks if the user is the owner of the listing
func (c *CarListing) IsOwner(userID int64) bool {
	return c.SellerID == userID
}

// CanEdit checks if the listing can be edited
func (c *CarListing) CanEdit() bool {
	return c.Status == "active" || c.Status == "inactive"
}

// MarkAsSold marks the listing as sold
func (c *CarListing) MarkAsSold() {
	c.Status = "sold"
	now := time.Now()
	c.SoldAt = &now
	c.UpdatedAt = now
}
