package middleware

import (
	"car-marketplace/internal/domain"
	"car-marketplace/internal/infrastructure/security"
	"car-marketplace/pkg/response"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

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
func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, http.StatusUnauthorized, "Authorization header required", domain.ErrUnauthorized)
			c.Abort()
			return
		}

		// Check if it's a Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(c, http.StatusUnauthorized, "Invalid authorization header format", domain.ErrUnauthorized)
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := m.jwtService.ValidateAccessToken(tokenString)
		if err != nil {
			switch err {
			case domain.ErrTokenExpired:
				response.Error(c, http.StatusUnauthorized, "Token has expired", err)
			case domain.ErrInvalidToken:
				response.Error(c, http.StatusUnauthorized, "Invalid token", err)
			default:
				response.Error(c, http.StatusUnauthorized, "Authentication failed", err)
			}
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("user_type", claims.UserType)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// RequireRole creates a middleware that checks if user has specific role
func (m *AuthMiddleware) RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			response.Error(c, http.StatusUnauthorized, "Unauthorized", domain.ErrUnauthorized)
			c.Abort()
			return
		}

		userRole := role.(string)

		// Check if user's role is in allowed roles
		isAllowed := false
		for _, allowedRole := range allowedRoles {
			if userRole == allowedRole {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			response.Error(c, http.StatusForbidden, "Insufficient permissions", domain.ErrForbidden)
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireUserType creates a middleware that checks if user has specific user type
func (m *AuthMiddleware) RequireUserType(allowedTypes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userType, exists := c.Get("user_type")
		if !exists {
			response.Error(c, http.StatusUnauthorized, "Unauthorized", domain.ErrUnauthorized)
			c.Abort()
			return
		}

		currentUserType := userType.(string)

		// Check if user's type is in allowed types
		isAllowed := false
		for _, allowedType := range allowedTypes {
			if currentUserType == allowedType {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			response.Error(c, http.StatusForbidden, "This action is not allowed for your user type", domain.ErrForbidden)
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAuth validates JWT if present but doesn't require it
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No token provided, continue without authentication
			c.Next()
			return
		}

		// Token provided, try to validate it
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			// Invalid format but don't abort, just continue
			c.Next()
			return
		}

		tokenString := parts[1]
		claims, err := m.jwtService.ValidateAccessToken(tokenString)
		if err == nil {
			// Valid token, set context
			c.Set("user_id", claims.UserID)
			c.Set("email", claims.Email)
			c.Set("user_type", claims.UserType)
			c.Set("role", claims.Role)
		}

		// Continue regardless of validation result
		c.Next()
	}
}
