package repository

import (
	"car-marketplace/internal/domain"
	"context"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// User CRUD operations
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id int64) error

	// User type operations
	GetUserTypeByName(ctx context.Context, name string) (*domain.UserTypeEntity, error)
	GetUserTypeByID(ctx context.Context, id int64) (*domain.UserTypeEntity, error)

	// Profile operations
	CreateUserProfile(ctx context.Context, profile *domain.UserProfile) error
	GetUserProfile(ctx context.Context, userID int64) (*domain.UserProfile, error)
	UpdateUserProfile(ctx context.Context, profile *domain.UserProfile) error

	// Business profile operations
	CreateBusinessProfile(ctx context.Context, profile *domain.BusinessProfile) error
	GetBusinessProfile(ctx context.Context, userID int64) (*domain.BusinessProfile, error)
	UpdateBusinessProfile(ctx context.Context, profile *domain.BusinessProfile) error

	// Refresh token operations
	CreateRefreshToken(ctx context.Context, token *domain.RefreshToken) error
	// ConsumeRefreshToken atomically revokes a valid token and returns it;
	// it fails with domain.ErrInvalidRefreshToken if the token is unknown,
	// expired or already used.
	ConsumeRefreshToken(ctx context.Context, token string) (*domain.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, token string) error
	RevokeAllUserTokens(ctx context.Context, userID int64) error
	DeleteExpiredTokens(ctx context.Context) error
}
