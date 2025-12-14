package router

import (
	"car-marketplace/internal/delivery/http/handler"
	"car-marketplace/internal/delivery/http/middleware"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"golang.org/x/time/rate"
)

// RouterConfig holds dependencies for router setup
type RouterConfig struct {
	AuthHandler     *handler.AuthHandler
	CarHandler      *handler.CarHandler
	InquiryHandler  *handler.InquiryHandler
	ReviewHandler   *handler.ReviewHandler
	FavoriteHandler *handler.FavoriteHandler
	AuthMiddleware  *middleware.AuthMiddleware
	RateLimiter     *middleware.RateLimiter
}

// SetupRouter configures and returns the Gin router
func SetupRouter(config RouterConfig) *gin.Engine {
	router := gin.New()

	// Global middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Custom CORS configuration
	corsConfig := middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"}, // Your frontend URLs
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * 3600,
	}
	router.Use(middleware.CORSWithConfig(corsConfig))

	// Health check endpoint (no versioning, no rate limit)
	router.GET("/health", healthCheck)
	router.GET("/ping", ping)

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Apply global rate limiting to all API routes
		if config.RateLimiter != nil {
			v1.Use(config.RateLimiter.Limit())
		}

		// Public routes (no authentication required)
		setupPublicRoutes(v1, config)

		// Protected routes (authentication required)
		setupProtectedRoutes(v1, config)
	}

	// 404 handler
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Route not found",
			"error": gin.H{
				"code":    "ROUTE_NOT_FOUND",
				"details": "The requested endpoint does not exist",
			},
		})
	})

	return router
}

// setupPublicRoutes configures public routes (no authentication)
func setupPublicRoutes(rg *gin.RouterGroup, config RouterConfig) {
	auth := rg.Group("/auth")
	{
		// Stricter rate limiting for auth endpoints
		authLimiter := middleware.NewRateLimiter(
			rate.Limit(5), // 5 requests per second
			10,            // burst of 10
			5*time.Minute, // 5 minutes cleanup
		)
		auth.Use(authLimiter.Limit())

		auth.POST("/register", config.AuthHandler.Register)
		auth.POST("/login", config.AuthHandler.Login)
		auth.POST("/refresh", config.AuthHandler.RefreshToken)
		auth.POST("/logout", config.AuthHandler.Logout)
	}

	// Public car listings (optional auth to show if favorited)
	cars := rg.Group("/cars")
	{
		cars.Use(config.AuthMiddleware.OptionalAuth())
		cars.GET("", config.CarHandler.List)
		cars.GET("/:id", config.CarHandler.GetByID)
	}

	// Reference data (public)
	ref := rg.Group("/reference")
	{
		ref.GET("/makes", config.CarHandler.GetMakes)
		ref.GET("/makes/:make_id/models", config.CarHandler.GetModels)
		ref.GET("/data", config.CarHandler.GetReferenceData)
	}

	// Public reviews
	reviews := rg.Group("/reviews")
	{
		reviews.GET("", config.ReviewHandler.List)
		reviews.GET("/users/:user_id/stats", config.ReviewHandler.GetUserStats)
	}
}

// setupProtectedRoutes configures protected routes (authentication required)
func setupProtectedRoutes(rg *gin.RouterGroup, config RouterConfig) {
	// All routes in this group require authentication
	protected := rg.Group("")
	protected.Use(config.AuthMiddleware.Authenticate())

	// Auth-related protected routes
	auth := protected.Group("/auth")
	{
		auth.POST("/logout-all", config.AuthHandler.LogoutAll)
		auth.POST("/change-password", config.AuthHandler.ChangePassword)
		auth.GET("/profile", config.AuthHandler.GetProfile)
	}

	// Car listing management
	cars := protected.Group("/cars")
	{
		cars.POST("", config.CarHandler.Create)
		cars.PUT("/:id", config.CarHandler.Update)
		cars.DELETE("/:id", config.CarHandler.Delete)
		cars.POST("/:id/sold", config.CarHandler.MarkAsSold)
	}

	// Favorites
	favorites := protected.Group("/favorites")
	{
		favorites.GET("", config.FavoriteHandler.List)
		favorites.POST("/:id", config.FavoriteHandler.Add)
		favorites.DELETE("/:id", config.FavoriteHandler.Remove)
	}

	// Inquiries/Messages
	inquiries := protected.Group("/inquiries")
	{
		inquiries.POST("", config.InquiryHandler.Create)
		inquiries.GET("", config.InquiryHandler.List)
		inquiries.PUT("/:id/read", config.InquiryHandler.MarkAsRead)
	}

	// Reviews
	reviews := protected.Group("/reviews")
	{
		reviews.POST("", config.ReviewHandler.Create)
		reviews.PUT("/:id", config.ReviewHandler.Update)
		reviews.DELETE("/:id", config.ReviewHandler.Delete)
	}
}

// healthCheck returns the health status of the API
// @Summary Health check
// @Description Check if the API is running
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"message": "Car Marketplace API is running",
		"version": "1.0.0",
	})
}

// ping returns a simple pong response
// @Summary Ping
// @Description Simple ping endpoint
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router /ping [get]
func ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}
