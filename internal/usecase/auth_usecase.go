package usecase

import (
	"car-marketplace/internal/domain"
	"car-marketplace/internal/infrastructure/security"
	"car-marketplace/internal/repository"
	"context"
	"time"
)

// AuthUseCase defines the interface for authentication business logic
type AuthUseCase interface {
	Register(ctx context.Context, req *domain.RegisterRequest) (*domain.User, error)
	Login(ctx context.Context, req *domain.LoginRequest) (*domain.TokenPair, *domain.User, error)
	RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error)
	Logout(ctx context.Context, refreshToken string) error
	LogoutAll(ctx context.Context, userID int64) error
	ChangePassword(ctx context.Context, userID int64, req *domain.ChangePasswordRequest) error
	GetUserByID(ctx context.Context, userID int64) (*domain.User, error)
}

// authUseCase implements AuthUseCase interface
type authUseCase struct {
	tx              repository.Transactor
	userRepo        repository.UserRepository
	jwtService      security.JWTService
	passwordHasher  security.PasswordHasher
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

// NewAuthUseCase creates a new auth use case instance
func NewAuthUseCase(
	tx repository.Transactor,
	userRepo repository.UserRepository,
	jwtService security.JWTService,
	passwordHasher security.PasswordHasher,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
) AuthUseCase {
	return &authUseCase{
		tx:              tx,
		userRepo:        userRepo,
		jwtService:      jwtService,
		passwordHasher:  passwordHasher,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

// Register creates a new user account
func (uc *authUseCase) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.User, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Check if user already exists
	existingUser, err := uc.userRepo.GetByEmail(ctx, req.Email)
	if err != nil && err != domain.ErrUserNotFound {
		return nil, err
	}
	if existingUser != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := uc.passwordHasher.HashPassword(req.Password)
	if err != nil {
		return nil, domain.ErrInternal
	}

	// Get user type ID
	userType, err := uc.userRepo.GetUserTypeByName(ctx, string(req.UserType))
	if err != nil {
		return nil, domain.ErrInvalidUserType
	}

	// Create user and profile together, so a failed profile insert
	// doesn't leave an account without a profile.
	user := &domain.User{
		Email:        req.Email,
		PasswordHash: hashedPassword,
		UserTypeID:   userType.ID,
		IsVerified:   false, // Email verification required
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := uc.userRepo.Create(ctx, user); err != nil {
			return err
		}

		switch req.UserType {
		case domain.UserTypeIndividual:
			return uc.userRepo.CreateUserProfile(ctx, &domain.UserProfile{
				UserID:      user.ID,
				FirstName:   req.FirstName,
				LastName:    req.LastName,
				PhoneNumber: req.PhoneNumber,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			})
		case domain.UserTypeDealership, domain.UserTypeShowroom:
			return uc.userRepo.CreateBusinessProfile(ctx, &domain.BusinessProfile{
				UserID:       user.ID,
				BusinessName: req.BusinessName,
				PhoneNumber:  req.PhoneNumber,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Load user type for response
	user.UserType = userType

	return user, nil
}

// Login authenticates a user and returns tokens
func (uc *authUseCase) Login(ctx context.Context, req *domain.LoginRequest) (*domain.TokenPair, *domain.User, error) {
	// Get user by email
	user, err := uc.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if err == domain.ErrUserNotFound {
			return nil, nil, domain.ErrInvalidCredentials
		}
		return nil, nil, err
	}

	// Verify password
	if err := uc.passwordHasher.ComparePassword(user.PasswordHash, req.Password); err != nil {
		return nil, nil, domain.ErrInvalidCredentials
	}

	// Check if user is active
	if !user.IsActive {
		return nil, nil, domain.ErrUserNotActive
	}

	// Generate tokens
	tokenPair, err := uc.generateTokenPair(ctx, user)
	if err != nil {
		return nil, nil, err
	}

	return tokenPair, user, nil
}

// RefreshToken exchanges a refresh token for a new token pair. The old token
// is consumed atomically, so a token can be used at most once.
func (uc *authUseCase) RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	var tokenPair *domain.TokenPair

	err := uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		token, err := uc.userRepo.ConsumeRefreshToken(ctx, refreshToken)
		if err != nil {
			return err
		}

		user, err := uc.userRepo.GetByID(ctx, token.UserID)
		if err != nil {
			return err
		}

		if !user.IsActive {
			return domain.ErrUserNotActive
		}

		tokenPair, err = uc.generateTokenPair(ctx, user)
		return err
	})
	if err != nil {
		return nil, err
	}

	return tokenPair, nil
}

// Logout revokes a specific refresh token
func (uc *authUseCase) Logout(ctx context.Context, refreshToken string) error {
	return uc.userRepo.RevokeRefreshToken(ctx, refreshToken)
}

// LogoutAll revokes all refresh tokens for a user
func (uc *authUseCase) LogoutAll(ctx context.Context, userID int64) error {
	return uc.userRepo.RevokeAllUserTokens(ctx, userID)
}

// ChangePassword changes user's password
func (uc *authUseCase) ChangePassword(ctx context.Context, userID int64, req *domain.ChangePasswordRequest) error {
	// Get user
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// Verify current password
	if err := uc.passwordHasher.ComparePassword(user.PasswordHash, req.CurrentPassword); err != nil {
		return domain.ErrInvalidCredentials
	}

	// Hash new password
	newHash, err := uc.passwordHasher.HashPassword(req.NewPassword)
	if err != nil {
		return domain.ErrInternal
	}

	// Update password
	user.PasswordHash = newHash
	user.UpdatedAt = time.Now()

	// Update the password and revoke all existing tokens to force re-login
	return uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := uc.userRepo.Update(ctx, user); err != nil {
			return err
		}
		return uc.userRepo.RevokeAllUserTokens(ctx, userID)
	})
}

// GetUserByID retrieves a user by ID
func (uc *authUseCase) GetUserByID(ctx context.Context, userID int64) (*domain.User, error) {
	return uc.userRepo.GetByID(ctx, userID)
}

// generateTokenPair generates access and refresh tokens for a user
func (uc *authUseCase) generateTokenPair(ctx context.Context, user *domain.User) (*domain.TokenPair, error) {
	// Create claims
	claims := &domain.AuthClaims{
		UserID:   user.ID,
		Email:    user.Email,
		UserType: user.UserType.Name,
		Role:     user.UserType.Name, // Using user type as role for now
	}

	// Generate access token
	accessToken, err := uc.jwtService.GenerateAccessToken(claims, uc.accessTokenTTL)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken, err := uc.jwtService.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	// Store refresh token in database
	refreshTokenEntity := &domain.RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(uc.refreshTokenTTL),
		IsRevoked: false,
		CreatedAt: time.Now(),
	}

	if err := uc.userRepo.CreateRefreshToken(ctx, refreshTokenEntity); err != nil {
		return nil, err
	}

	return &domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(uc.accessTokenTTL.Seconds()),
	}, nil
}
