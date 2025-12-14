package domain

import "errors"

// Authentication & Authorization Errors
var (
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrUserNotFound        = errors.New("user not found")
	ErrUserAlreadyExists   = errors.New("user with this email already exists")
	ErrInvalidToken        = errors.New("invalid or expired token")
	ErrTokenExpired        = errors.New("token has expired")
	ErrUnauthorized        = errors.New("unauthorized access")
	ErrForbidden           = errors.New("forbidden: insufficient permissions")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrRevokedToken        = errors.New("token has been revoked")
)

// Validation Errors
var (
	ErrInvalidInput    = errors.New("invalid input data")
	ErrInvalidEmail    = errors.New("invalid email format")
	ErrWeakPassword    = errors.New("password does not meet requirements")
	ErrInvalidUserType = errors.New("invalid user type")
)

// User Status Errors
var (
	ErrUserNotActive   = errors.New("user account is not active")
	ErrUserNotVerified = errors.New("user account is not verified")
)

// General Errors
var (
	ErrInternal   = errors.New("internal server error")
	ErrNotFound   = errors.New("resource not found")
	ErrConflict   = errors.New("resource conflict")
	ErrBadRequest = errors.New("bad request")
)

// Database Errors
var (
	ErrDatabaseConnection = errors.New("database connection error")
	ErrDatabaseQuery      = errors.New("database query error")
	ErrRecordNotFound     = errors.New("record not found")
	ErrDuplicateEntry     = errors.New("duplicate entry")
)

// Car Listing Errors
var (
	ErrCarListingNotFound      = errors.New("car listing not found")
	ErrCarListingAlreadyExists = errors.New("car listing already exists")
	ErrCarListingNotOwner      = errors.New("you are not the owner of this listing")
	ErrCarListingCannotEdit    = errors.New("this listing cannot be edited")
	ErrCarListingAlreadySold   = errors.New("this car has already been sold")
)

// Reference Data Errors
var (
	ErrInvalidMake         = errors.New("invalid car make")
	ErrInvalidModel        = errors.New("invalid car model")
	ErrInvalidBodyType     = errors.New("invalid body type")
	ErrInvalidFuelType     = errors.New("invalid fuel type")
	ErrInvalidTransmission = errors.New("invalid transmission type")
	ErrInvalidColor        = errors.New("invalid color")
	ErrInvalidCondition    = errors.New("invalid condition")
	ErrInvalidFeature      = errors.New("invalid feature")
)

// Inquiry Errors
var (
	ErrInquiryNotFound  = errors.New("inquiry not found")
	ErrInquiryNotOwner  = errors.New("you are not authorized to access this inquiry")
	ErrCannotInquireOwn = errors.New("you cannot inquire about your own listing")
)

// Review Errors
var (
	ErrReviewNotFound      = errors.New("review not found")
	ErrReviewNotOwner      = errors.New("you are not the owner of this review")
	ErrReviewAlreadyExists = errors.New("you have already reviewed this user")
	ErrCannotReviewSelf    = errors.New("you cannot review yourself")
)

// Favorite Errors
var (
	ErrFavoriteNotFound      = errors.New("favorite not found")
	ErrFavoriteAlreadyExists = errors.New("listing already in favorites")
	ErrCannotFavoriteOwn     = errors.New("you cannot favorite your own listing")
)
