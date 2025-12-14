package security

import (
	"golang.org/x/crypto/bcrypt"
)

// PasswordHasher defines the interface for password hashing operations
type PasswordHasher interface {
	HashPassword(password string) (string, error)
	ComparePassword(hashedPassword, password string) error
}

// bcryptHasher implements PasswordHasher using bcrypt
type bcryptHasher struct {
	cost int
}

// NewBcryptHasher creates a new bcrypt password hasher
func NewBcryptHasher(cost int) PasswordHasher {
	// Use default cost if invalid cost provided
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}

	return &bcryptHasher{
		cost: cost,
	}
}

// HashPassword hashes a plain text password
func (h *bcryptHasher) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// ComparePassword compares a hashed password with a plain text password
func (h *bcryptHasher) ComparePassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
