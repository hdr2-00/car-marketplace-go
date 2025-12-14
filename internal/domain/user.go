package domain

import "time"

// UserType represents the type of user account
type UserType string

const (
	UserTypeIndividual UserType = "individual"
	UserTypeDealership UserType = "dealership"
	UserTypeShowroom   UserType = "showroom"
)

// User represents a user in the system
type User struct {
	ID           int64           `json:"id"`
	Email        string          `json:"email"`
	PasswordHash string          `json:"-"` // Never expose password hash in JSON
	UserTypeID   int64           `json:"user_type_id"`
	UserType     *UserTypeEntity `json:"user_type,omitempty"`
	IsVerified   bool            `json:"is_verified"`
	IsActive     bool            `json:"is_active"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// UserTypeEntity represents the user type reference data
type UserTypeEntity struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UserProfile represents an individual user's profile
type UserProfile struct {
	ID              int64     `json:"id"`
	UserID          int64     `json:"user_id"`
	FirstName       string    `json:"first_name"`
	LastName        string    `json:"last_name"`
	PhoneNumber     string    `json:"phone_number"`
	ProfileImageURL string    `json:"profile_image_url,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// BusinessProfile represents a dealership or showroom profile
type BusinessProfile struct {
	ID              int64     `json:"id"`
	UserID          int64     `json:"user_id"`
	BusinessName    string    `json:"business_name"`
	BusinessLicense string    `json:"business_license,omitempty"`
	TaxID           string    `json:"tax_id,omitempty"`
	PhoneNumber     string    `json:"phone_number"`
	WebsiteURL      string    `json:"website_url,omitempty"`
	LogoURL         string    `json:"logo_url,omitempty"`
	Description     string    `json:"description,omitempty"`
	EstablishedYear int       `json:"established_year,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// IsIndividual checks if user is an individual
func (u *User) IsIndividual() bool {
	return u.UserType != nil && u.UserType.Name == string(UserTypeIndividual)
}

// IsBusiness checks if user is a business (dealership or showroom)
func (u *User) IsBusiness() bool {
	if u.UserType == nil {
		return false
	}
	return u.UserType.Name == string(UserTypeDealership) ||
		u.UserType.Name == string(UserTypeShowroom)
}

// FullName returns the full name for individual users
func (up *UserProfile) FullName() string {
	if up.FirstName == "" && up.LastName == "" {
		return ""
	}
	return up.FirstName + " " + up.LastName
}
