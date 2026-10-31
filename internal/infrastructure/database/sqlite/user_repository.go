package sqlite

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"car-marketplace/internal/domain"
	"car-marketplace/internal/repository"
)

// userRepository implements repository.UserRepository for SQLite
type userRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new SQLite user repository
func NewUserRepository(db *sql.DB) repository.UserRepository {
	return &userRepository{db: db}
}

// Create inserts a new user into the database
func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (email, password_hash, user_type_id, is_verified, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	result, err := conn(ctx, r.db).ExecContext(ctx, query,
		user.Email,
		user.PasswordHash,
		user.UserTypeID,
		user.IsVerified,
		user.IsActive,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		if isUniqueViolation(err, "users.email") {
			return domain.ErrUserAlreadyExists
		}
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	user.ID = id
	return nil
}

// GetByID retrieves a user by their ID
func (r *userRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	query := `
		SELECT u.id, u.email, u.password_hash, u.user_type_id, u.is_verified, u.is_active,
		       u.created_at, u.updated_at,
		       ut.id, ut.name, COALESCE(ut.description, ''), ut.created_at, ut.updated_at
		FROM users u
		LEFT JOIN user_types ut ON u.user_type_id = ut.id
		WHERE u.id = ?
	`

	user := &domain.User{
		UserType: &domain.UserTypeEntity{},
	}

	err := conn(ctx, r.db).QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.UserTypeID,
		&user.IsVerified,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.UserType.ID,
		&user.UserType.Name,
		&user.UserType.Description,
		&user.UserType.CreatedAt,
		&user.UserType.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

// GetByEmail retrieves a user by their email address
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT u.id, u.email, u.password_hash, u.user_type_id, u.is_verified, u.is_active,
		       u.created_at, u.updated_at,
		       ut.id, ut.name, COALESCE(ut.description, ''), ut.created_at, ut.updated_at
		FROM users u
		LEFT JOIN user_types ut ON u.user_type_id = ut.id
		WHERE u.email = ?
	`

	user := &domain.User{
		UserType: &domain.UserTypeEntity{},
	}

	err := conn(ctx, r.db).QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.UserTypeID,
		&user.IsVerified,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.UserType.ID,
		&user.UserType.Name,
		&user.UserType.Description,
		&user.UserType.CreatedAt,
		&user.UserType.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

// Update updates a user's information
func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET email = ?, password_hash = ?, user_type_id = ?, is_verified = ?,
		    is_active = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := conn(ctx, r.db).ExecContext(ctx, query,
		user.Email,
		user.PasswordHash,
		user.UserTypeID,
		user.IsVerified,
		user.IsActive,
		user.UpdatedAt,
		user.ID,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

// Delete deletes a user from the database
func (r *userRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM users WHERE id = ?`

	result, err := conn(ctx, r.db).ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

// GetUserTypeByName retrieves a user type by name
func (r *userRepository) GetUserTypeByName(ctx context.Context, name string) (*domain.UserTypeEntity, error) {
	query := `SELECT id, name, COALESCE(description, ''), created_at, updated_at FROM user_types WHERE name = ?`

	userType := &domain.UserTypeEntity{}
	err := conn(ctx, r.db).QueryRowContext(ctx, query, name).Scan(
		&userType.ID,
		&userType.Name,
		&userType.Description,
		&userType.CreatedAt,
		&userType.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrInvalidUserType
		}
		return nil, err
	}

	return userType, nil
}

// GetUserTypeByID retrieves a user type by ID
func (r *userRepository) GetUserTypeByID(ctx context.Context, id int64) (*domain.UserTypeEntity, error) {
	query := `SELECT id, name, COALESCE(description, ''), created_at, updated_at FROM user_types WHERE id = ?`

	userType := &domain.UserTypeEntity{}
	err := conn(ctx, r.db).QueryRowContext(ctx, query, id).Scan(
		&userType.ID,
		&userType.Name,
		&userType.Description,
		&userType.CreatedAt,
		&userType.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrInvalidUserType
		}
		return nil, err
	}

	return userType, nil
}

// CreateUserProfile creates a profile for an individual user
func (r *userRepository) CreateUserProfile(ctx context.Context, profile *domain.UserProfile) error {
	query := `
		INSERT INTO user_profiles (user_id, first_name, last_name, phone_number, profile_image_url, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	result, err := conn(ctx, r.db).ExecContext(ctx, query,
		profile.UserID,
		profile.FirstName,
		profile.LastName,
		profile.PhoneNumber,
		profile.ProfileImageURL,
		profile.CreatedAt,
		profile.UpdatedAt,
	)

	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	profile.ID = id
	return nil
}

// GetUserProfile retrieves a user's profile
func (r *userRepository) GetUserProfile(ctx context.Context, userID int64) (*domain.UserProfile, error) {
	query := `
		SELECT id, user_id, COALESCE(first_name, ''), COALESCE(last_name, ''),
		       COALESCE(phone_number, ''), COALESCE(profile_image_url, ''), created_at, updated_at
		FROM user_profiles
		WHERE user_id = ?
	`

	profile := &domain.UserProfile{}
	err := conn(ctx, r.db).QueryRowContext(ctx, query, userID).Scan(
		&profile.ID,
		&profile.UserID,
		&profile.FirstName,
		&profile.LastName,
		&profile.PhoneNumber,
		&profile.ProfileImageURL,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return profile, nil
}

// UpdateUserProfile updates a user's profile
func (r *userRepository) UpdateUserProfile(ctx context.Context, profile *domain.UserProfile) error {
	query := `
		UPDATE user_profiles
		SET first_name = ?, last_name = ?, phone_number = ?, profile_image_url = ?, updated_at = ?
		WHERE user_id = ?
	`

	result, err := conn(ctx, r.db).ExecContext(ctx, query,
		profile.FirstName,
		profile.LastName,
		profile.PhoneNumber,
		profile.ProfileImageURL,
		profile.UpdatedAt,
		profile.UserID,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// CreateBusinessProfile creates a profile for a business user
func (r *userRepository) CreateBusinessProfile(ctx context.Context, profile *domain.BusinessProfile) error {
	query := `
		INSERT INTO business_profiles
		(user_id, business_name, business_license, tax_id, phone_number, website_url,
		 logo_url, description, established_year, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := conn(ctx, r.db).ExecContext(ctx, query,
		profile.UserID,
		profile.BusinessName,
		profile.BusinessLicense,
		profile.TaxID,
		profile.PhoneNumber,
		profile.WebsiteURL,
		profile.LogoURL,
		profile.Description,
		profile.EstablishedYear,
		profile.CreatedAt,
		profile.UpdatedAt,
	)

	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	profile.ID = id
	return nil
}

// GetBusinessProfile retrieves a business profile
func (r *userRepository) GetBusinessProfile(ctx context.Context, userID int64) (*domain.BusinessProfile, error) {
	query := `
		SELECT id, user_id, business_name, COALESCE(business_license, ''), COALESCE(tax_id, ''),
		       COALESCE(phone_number, ''), COALESCE(website_url, ''), COALESCE(logo_url, ''),
		       COALESCE(description, ''), COALESCE(established_year, 0), created_at, updated_at
		FROM business_profiles
		WHERE user_id = ?
	`

	profile := &domain.BusinessProfile{}
	err := conn(ctx, r.db).QueryRowContext(ctx, query, userID).Scan(
		&profile.ID,
		&profile.UserID,
		&profile.BusinessName,
		&profile.BusinessLicense,
		&profile.TaxID,
		&profile.PhoneNumber,
		&profile.WebsiteURL,
		&profile.LogoURL,
		&profile.Description,
		&profile.EstablishedYear,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return profile, nil
}

// UpdateBusinessProfile updates a business profile
func (r *userRepository) UpdateBusinessProfile(ctx context.Context, profile *domain.BusinessProfile) error {
	query := `
		UPDATE business_profiles
		SET business_name = ?, business_license = ?, tax_id = ?, phone_number = ?,
		    website_url = ?, logo_url = ?, description = ?, established_year = ?, updated_at = ?
		WHERE user_id = ?
	`

	result, err := conn(ctx, r.db).ExecContext(ctx, query,
		profile.BusinessName,
		profile.BusinessLicense,
		profile.TaxID,
		profile.PhoneNumber,
		profile.WebsiteURL,
		profile.LogoURL,
		profile.Description,
		profile.EstablishedYear,
		profile.UpdatedAt,
		profile.UserID,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// hashToken returns the SHA-256 digest stored in place of a raw refresh
// token, so a database leak does not expose usable tokens.
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// CreateRefreshToken stores the hash of a refresh token
func (r *userRepository) CreateRefreshToken(ctx context.Context, token *domain.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token, expires_at, is_revoked, created_at)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := conn(ctx, r.db).ExecContext(ctx, query,
		token.UserID,
		hashToken(token.Token),
		token.ExpiresAt.UTC(),
		token.IsRevoked,
		token.CreatedAt.UTC(),
	)

	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	token.ID = id
	return nil
}

// ConsumeRefreshToken atomically revokes a valid (unrevoked, unexpired)
// refresh token and returns it. Concurrent calls with the same token cannot
// both succeed, so a token can be exchanged at most once.
func (r *userRepository) ConsumeRefreshToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	query := `
		UPDATE refresh_tokens SET is_revoked = 1
		WHERE token = ? AND is_revoked = 0 AND expires_at > ?
		RETURNING id, user_id, expires_at, created_at
	`

	refreshToken := &domain.RefreshToken{Token: token, IsRevoked: true}
	err := conn(ctx, r.db).QueryRowContext(ctx, query, hashToken(token), time.Now().UTC()).Scan(
		&refreshToken.ID,
		&refreshToken.UserID,
		&refreshToken.ExpiresAt,
		&refreshToken.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrInvalidRefreshToken
		}
		return nil, err
	}

	return refreshToken, nil
}

// RevokeRefreshToken marks a refresh token as revoked
func (r *userRepository) RevokeRefreshToken(ctx context.Context, token string) error {
	query := `UPDATE refresh_tokens SET is_revoked = 1 WHERE token = ?`

	_, err := conn(ctx, r.db).ExecContext(ctx, query, hashToken(token))
	return err
}

// RevokeAllUserTokens revokes all refresh tokens for a user
func (r *userRepository) RevokeAllUserTokens(ctx context.Context, userID int64) error {
	query := `UPDATE refresh_tokens SET is_revoked = 1 WHERE user_id = ?`

	_, err := conn(ctx, r.db).ExecContext(ctx, query, userID)
	return err
}

// DeleteExpiredTokens removes expired and revoked tokens, which can never be used again
func (r *userRepository) DeleteExpiredTokens(ctx context.Context) error {
	query := `DELETE FROM refresh_tokens WHERE expires_at <= ? OR is_revoked = 1`

	_, err := conn(ctx, r.db).ExecContext(ctx, query, time.Now().UTC())
	return err
}
