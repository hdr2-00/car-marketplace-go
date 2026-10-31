package main

import (
	_ "car-marketplace/docs/swagger" // Import swagger docs
	"car-marketplace/internal/app"
	"car-marketplace/internal/config"
	"car-marketplace/internal/infrastructure/database/sqlite"
	"car-marketplace/migrations"
	"context"
	"flag"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"
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
	migrateOnly := flag.Bool("migrate-only", false, "apply database migrations and exit")
	seed := flag.Bool("seed", false, "load sample data after migrating")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize database, migrations and HTTP handler
	application, err := app.New(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	defer application.Close()

	log.Println("✅ Database connected and migrated")

	if *seed {
		if err := sqlite.Seed(ctx, application.DB, migrations.Seeds); err != nil {
			log.Fatalf("Failed to load sample data: %v", err)
		}
		log.Println("✅ Sample data loaded")
	}
	if *migrateOnly {
		return
	}

	application.Start()

	// Configure HTTP server
	srv := &http.Server{
		Addr:         cfg.GetServerAddress(),
		Handler:      application.Handler,
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

	// Wait for shutdown signal
	<-ctx.Done()
	stop()

	log.Println("🛑 Shutting down server...")

	// Graceful shutdown with 5 second timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited successfully")
}
