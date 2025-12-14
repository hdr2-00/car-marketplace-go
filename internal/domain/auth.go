package domain

import "time"

// RefreshToken represents a refresh token in the system
type RefreshToken struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	IsRevoked bool      `json:"is_revoked"`
	CreatedAt time.Time `json:"created_at"`
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"` // seconds until access token expires
}

// AuthClaims represents the claims stored in JWT
type AuthClaims struct {
	UserID   int64  `json:"user_id"`
	Email    string `json:"email"`
	UserType string `json:"user_type"`
	Role     string `json:"role"`
}

// LoginRequest represents login credentials
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// RegisterRequest represents registration data
//
// Conditional Required Fields:
// - When user_type = "individual": first_name and last_name are REQUIRED
// - When user_type = "dealership" or "showroom": business_name is REQUIRED
// - phone_number is always optional
//
// The API will return a 400 error if required fields for the selected user_type are missing.
type RegisterRequest struct {
	Email    string   `json:"email" binding:"required,email" example:"user@example.com"`
	Password string   `json:"password" binding:"required,min=8" example:"SecurePass123!"`
	UserType UserType `json:"user_type" binding:"required,oneof=individual dealership showroom" example:"individual" enums:"individual,dealership,showroom"`

	// Required when user_type is "individual"
	FirstName string `json:"first_name,omitempty" example:"John"`
	LastName  string `json:"last_name,omitempty" example:"Doe"`

	// Required when user_type is "dealership" or "showroom"
	BusinessName string `json:"business_name,omitempty" example:"Premium Auto Dealership"`

	PhoneNumber string `json:"phone_number,omitempty" binding:"omitempty,e164" example:"+1234567890"`
}

// RefreshTokenRequest represents a request to refresh access token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// IsExpired checks if the refresh token is expired
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

// IsValid checks if the refresh token is valid (not expired and not revoked)
func (rt *RefreshToken) IsValid() bool {
	return !rt.IsExpired() && !rt.IsRevoked
}

// Validate validates the registration request based on user type
func (r *RegisterRequest) Validate() error {
	switch r.UserType {
	case UserTypeIndividual:
		if r.FirstName == "" || r.LastName == "" {
			return ErrInvalidInput
		}
	case UserTypeDealership, UserTypeShowroom:
		if r.BusinessName == "" {
			return ErrInvalidInput
		}
	default:
		return ErrInvalidUserType
	}
	return nil
}
