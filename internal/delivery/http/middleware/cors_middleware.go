package middleware

import (
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int // seconds
}

// CORSWithConfig creates a middleware for Cross-Origin Resource Sharing.
// Only origins in AllowOrigins receive CORS headers; the request's origin is
// echoed back (never "*"), which is what browsers require with credentials.
func CORSWithConfig(config CORSConfig) gin.HandlerFunc {
	methods := strings.Join(config.AllowMethods, ", ")
	headers := strings.Join(config.AllowHeaders, ", ")
	expose := strings.Join(config.ExposeHeaders, ", ")
	maxAge := strconv.Itoa(config.MaxAge)

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		h := c.Writer.Header()
		h.Add("Vary", "Origin")

		if origin != "" && slices.Contains(config.AllowOrigins, origin) {
			h.Set("Access-Control-Allow-Origin", origin)
			if config.AllowCredentials {
				h.Set("Access-Control-Allow-Credentials", "true")
			}
			if methods != "" {
				h.Set("Access-Control-Allow-Methods", methods)
			}
			if headers != "" {
				h.Set("Access-Control-Allow-Headers", headers)
			}
			if expose != "" {
				h.Set("Access-Control-Expose-Headers", expose)
			}
			if config.MaxAge > 0 {
				h.Set("Access-Control-Max-Age", maxAge)
			}
		}

		// Handle preflight requests
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
