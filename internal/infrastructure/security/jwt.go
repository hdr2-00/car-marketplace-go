package security

import (
	"car-marketplace/internal/domain"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTService defines the interface for JWT operations
type JWTService interface {
	GenerateAccessToken(claims *domain.AuthClaims, ttl time.Duration) (string, error)
	GenerateRefreshToken() (string, error)
	ValidateAccessToken(tokenString string) (*domain.AuthClaims, error)
}

// issuer is set on every access token and required when validating.
const issuer = "car-marketplace-api"

// jwtService implements JWTService interface
type jwtService struct {
	secretKey []byte
}

// JWTClaims represents the JWT claims structure
type JWTClaims struct {
	UserID   int64  `json:"user_id"`
	Email    string `json:"email"`
	UserType string `json:"user_type"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// NewJWTService creates a new JWT service instance
func NewJWTService(secretKey string) JWTService {
	return &jwtService{
		secretKey: []byte(secretKey),
	}
}

// GenerateAccessToken creates a new JWT access token
func (s *jwtService) GenerateAccessToken(authClaims *domain.AuthClaims, ttl time.Duration) (string, error) {
	now := time.Now()
	expiresAt := now.Add(ttl)

	claims := JWTClaims{
		UserID:   authClaims.UserID,
		Email:    authClaims.Email,
		UserType: authClaims.UserType,
		Role:     authClaims.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    issuer,
			Subject:   authClaims.Email,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GenerateRefreshToken creates a new random refresh token
func (s *jwtService) GenerateRefreshToken() (string, error) {
	// Generate 32 random bytes
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	// Encode to base64
	return base64.URLEncoding.EncodeToString(b), nil
}

// ValidateAccessToken validates and parses a JWT access token
func (s *jwtService) ValidateAccessToken(tokenString string) (*domain.AuthClaims, error) {
	claims := &JWTClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return s.secretKey, nil
	},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(issuer),
		jwt.WithExpirationRequired(),
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, domain.ErrTokenExpired
		}
		return nil, domain.ErrInvalidToken
	}

	return &domain.AuthClaims{
		UserID:   claims.UserID,
		Email:    claims.Email,
		UserType: claims.UserType,
		Role:     claims.Role,
	}, nil
}
