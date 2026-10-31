package middleware

import (
	"car-marketplace/internal/domain"
	"car-marketplace/internal/infrastructure/security"
	"car-marketplace/pkg/response"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AccessTokenCookie is the cookie that carries the access token for browser clients.
const AccessTokenCookie = "access_token"

// AuthMiddleware creates a middleware for JWT authentication
type AuthMiddleware struct {
	jwtService security.JWTService
}

// NewAuthMiddleware creates a new auth middleware instance
func NewAuthMiddleware(jwtService security.JWTService) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
	}
}

// errMalformedAuthHeader means an Authorization header was sent but isn't "Bearer <token>".
var errMalformedAuthHeader = errors.New("invalid authorization header format")

// accessToken extracts the token from the Authorization header, falling back
// to the access_token cookie set at login. It returns "" if neither is present.
func accessToken(c *gin.Context) (string, error) {
	if header := c.GetHeader("Authorization"); header != "" {
		scheme, token, ok := strings.Cut(header, " ")
		if !ok || !strings.EqualFold(scheme, "Bearer") || token == "" {
			return "", errMalformedAuthHeader
		}
		return token, nil
	}

	if token, err := c.Cookie(AccessTokenCookie); err == nil {
		return token, nil
	}

	return "", nil
}

func setClaims(c *gin.Context, claims *domain.AuthClaims) {
	c.Set("user_id", claims.UserID)
	c.Set("email", claims.Email)
	c.Set("user_type", claims.UserType)
	c.Set("role", claims.Role)
}

// Authenticate requires a valid access token (Bearer header or cookie).
func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := accessToken(c)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "Invalid authorization header format", domain.ErrUnauthorized)
			c.Abort()
			return
		}
		if tokenString == "" {
			response.Error(c, http.StatusUnauthorized, "Authentication required", domain.ErrUnauthorized)
			c.Abort()
			return
		}

		claims, err := m.jwtService.ValidateAccessToken(tokenString)
		if err != nil {
			if errors.Is(err, domain.ErrTokenExpired) {
				response.Error(c, http.StatusUnauthorized, "Token has expired", err)
			} else {
				response.Error(c, http.StatusUnauthorized, "Invalid token", domain.ErrInvalidToken)
			}
			c.Abort()
			return
		}

		setClaims(c, claims)
		c.Next()
	}
}

// OptionalAuth sets the user in the context if a valid token is present, but
// never rejects the request.
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if tokenString, err := accessToken(c); err == nil && tokenString != "" {
			if claims, err := m.jwtService.ValidateAccessToken(tokenString); err == nil {
				setClaims(c, claims)
			}
		}

		c.Next()
	}
}
