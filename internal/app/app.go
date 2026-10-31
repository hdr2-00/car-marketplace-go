// Package app wires the application's dependencies together. It is shared by
// cmd/api and the HTTP integration tests.
package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"car-marketplace/internal/config"
	"car-marketplace/internal/delivery/http/handler"
	"car-marketplace/internal/delivery/http/middleware"
	"car-marketplace/internal/delivery/http/router"
	"car-marketplace/internal/infrastructure/database/sqlite"
	"car-marketplace/internal/infrastructure/security"
	"car-marketplace/internal/repository"
	"car-marketplace/internal/usecase"
	"car-marketplace/migrations"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

const (
	viewFlushInterval    = 10 * time.Second
	tokenCleanupInterval = time.Hour

	// Stricter limits for the public auth endpoints (login, register, ...)
	authRateLimit = rate.Limit(5)
	authBurst     = 10
)

// App holds the wired application.
type App struct {
	DB      *sql.DB
	Handler http.Handler

	views        *usecase.ViewCounter
	userRepo     repository.UserRepository
	rateLimiters []*middleware.RateLimiter

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// New opens the database, applies pending migrations and builds the HTTP handler.
func New(ctx context.Context, cfg *config.Config) (*App, error) {
	gin.SetMode(cfg.Server.Mode)

	db, err := sqlite.NewDatabase(sqlite.Config{
		DatabasePath:    cfg.Database.Path,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := sqlite.Migrate(ctx, db, migrations.Schema); err != nil {
		db.Close()
		return nil, err
	}

	// Repositories
	tx := sqlite.NewTransactor(db)
	userRepo := sqlite.NewUserRepository(db)
	carRepo := sqlite.NewCarRepository(db)
	inquiryRepo := sqlite.NewInquiryRepository(db)
	reviewRepo := sqlite.NewReviewRepository(db)
	favoriteRepo := sqlite.NewFavoriteRepository(db)

	// Security services
	jwtService := security.NewJWTService(cfg.JWT.SecretKey)
	passwordHasher := security.NewBcryptHasher(10)

	// Use cases
	views := usecase.NewViewCounter(carRepo)
	authUseCase := usecase.NewAuthUseCase(tx, userRepo, jwtService, passwordHasher, cfg.JWT.AccessTokenTTL, cfg.JWT.RefreshTokenTTL)
	carUseCase := usecase.NewCarUseCase(tx, carRepo, views)
	inquiryUseCase := usecase.NewInquiryUseCase(inquiryRepo, carRepo)
	reviewUseCase := usecase.NewReviewUseCase(reviewRepo, userRepo, carRepo)
	favoriteUseCase := usecase.NewFavoriteUseCase(favoriteRepo, carRepo)

	a := &App{DB: db, views: views, userRepo: userRepo}

	var rateLimiter, authRateLimiter *middleware.RateLimiter
	if cfg.RateLimit.Enabled {
		rateLimiter = middleware.NewRateLimiter(rate.Limit(cfg.RateLimit.RequestsPerSec), cfg.RateLimit.Burst, cfg.RateLimit.CleanupInterval)
		authRateLimiter = middleware.NewRateLimiter(authRateLimit, authBurst, cfg.RateLimit.CleanupInterval)
		a.rateLimiters = append(a.rateLimiters, rateLimiter, authRateLimiter)
	}

	a.Handler, err = router.SetupRouter(router.RouterConfig{
		AuthHandler:     handler.NewAuthHandler(authUseCase, cfg.JWT.AccessTokenTTL, cfg.JWT.RefreshTokenTTL, cfg.Server.CookieSecure),
		CarHandler:      handler.NewCarHandler(carUseCase),
		InquiryHandler:  handler.NewInquiryHandler(inquiryUseCase),
		ReviewHandler:   handler.NewReviewHandler(reviewUseCase),
		FavoriteHandler: handler.NewFavoriteHandler(favoriteUseCase),
		AuthMiddleware:  middleware.NewAuthMiddleware(jwtService),
		RateLimiter:     rateLimiter,
		AuthRateLimiter: authRateLimiter,
		AllowedOrigins:  cfg.CORS.AllowedOrigins,
		TrustedProxies:  cfg.Server.TrustedProxies,
		HealthCheck:     db.PingContext,
	})
	if err != nil {
		db.Close()
		return nil, err
	}

	return a, nil
}

// Start launches background jobs: flushing buffered listing views, removing
// dead refresh tokens and forgetting idle rate-limit clients. They stop on Close.
func (a *App) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel

	a.goRun(func() { a.views.Run(ctx, viewFlushInterval) })
	a.goRun(func() { a.cleanupTokens(ctx) })
	for _, rl := range a.rateLimiters {
		a.goRun(func() { rl.Run(ctx) })
	}
}

func (a *App) goRun(fn func()) {
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		fn()
	}()
}

func (a *App) cleanupTokens(ctx context.Context) {
	ticker := time.NewTicker(tokenCleanupInterval)
	defer ticker.Stop()

	for {
		if err := a.userRepo.DeleteExpiredTokens(ctx); err != nil && ctx.Err() == nil {
			log.Printf("failed to delete expired refresh tokens: %v", err)
		}

		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}
	}
}

// Close stops background jobs, writes buffered views and closes the database.
func (a *App) Close() error {
	if a.cancel != nil {
		a.cancel()
		a.wg.Wait() // the view counter flushes once more as it stops
	} else if err := a.views.Flush(context.Background()); err != nil {
		log.Printf("failed to flush listing views: %v", err)
	}

	return a.DB.Close()
}
