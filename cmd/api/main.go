package main

import (
	_ "car-marketplace/docs/swagger" // Import swagger docs
	"car-marketplace/internal/config"
	"car-marketplace/internal/delivery/http/handler"
	"car-marketplace/internal/delivery/http/middleware"
	"car-marketplace/internal/delivery/http/router"
	"car-marketplace/internal/infrastructure/database/sqlite"
	"car-marketplace/internal/infrastructure/security"
	"car-marketplace/internal/usecase"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// @title Car Marketplace API
// @version 1.0
// @description REST API for an online car marketplace where users can buy and sell cars
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@carmarketplace.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	// Initialize database
	dbConfig := sqlite.Config{
		DatabasePath:    cfg.Database.Path,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	}

	db, err := sqlite.NewDatabase(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("✅ Database connected successfully")

	// Initialize repositories
	userRepo := sqlite.NewUserRepository(db)
	carRepo := sqlite.NewCarRepository(db)
	inquiryRepo := sqlite.NewInquiryRepository(db)
	reviewRepo := sqlite.NewReviewRepository(db)
	favoriteRepo := sqlite.NewFavoriteRepository(db)

	// Initialize security services
	jwtService := security.NewJWTService(cfg.JWT.SecretKey)
	passwordHasher := security.NewBcryptHasher(10)

	// Initialize use cases
	authUseCase := usecase.NewAuthUseCase(
		userRepo,
		jwtService,
		passwordHasher,
		cfg.JWT.AccessTokenTTL,
		cfg.JWT.RefreshTokenTTL,
	)

	carUseCase := usecase.NewCarUseCase(carRepo)
	inquiryUseCase := usecase.NewInquiryUseCase(inquiryRepo, carRepo)
	reviewUseCase := usecase.NewReviewUseCase(reviewRepo)
	favoriteUseCase := usecase.NewFavoriteUseCase(favoriteRepo, carRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authUseCase, cfg.JWT.AccessTokenTTL, cfg.JWT.RefreshTokenTTL)
	carHandler := handler.NewCarHandler(carUseCase)
	inquiryHandler := handler.NewInquiryHandler(inquiryUseCase)
	reviewHandler := handler.NewReviewHandler(reviewUseCase)
	favoriteHandler := handler.NewFavoriteHandler(favoriteUseCase)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtService)

	var rateLimiter *middleware.RateLimiter
	if cfg.RateLimit.Enabled {
		rateLimiter = middleware.NewRateLimiter(
			rate.Limit(cfg.RateLimit.RequestsPerSec),
			cfg.RateLimit.Burst,
			cfg.RateLimit.CleanupInterval,
		)
	}

	// Setup router
	routerConfig := router.RouterConfig{
		AuthHandler:     authHandler,
		CarHandler:      carHandler,
		InquiryHandler:  inquiryHandler,
		ReviewHandler:   reviewHandler,
		FavoriteHandler: favoriteHandler,
		AuthMiddleware:  authMiddleware,
		RateLimiter:     rateLimiter,
	}

	r := router.SetupRouter(routerConfig)

	// Configure HTTP server
	srv := &http.Server{
		Addr:         cfg.GetServerAddress(),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("🚀 Server starting on %s", cfg.GetServerAddress())
		log.Printf("📚 Swagger documentation: http://%s/swagger/index.html", cfg.GetServerAddress())
		log.Printf("🏥 Health check: http://%s/health", cfg.GetServerAddress())

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	// Graceful shutdown with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited successfully")
}
