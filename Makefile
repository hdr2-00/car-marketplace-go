.PHONY: help run build test clean migrate swagger deps dev docker-build docker-run docker-compose-up docker-compose-down

# Variables
APP_NAME=car-marketplace
MAIN_PATH=./cmd/api
BUILD_DIR=./bin
DB_PATH=./data/marketplace.db
MIGRATIONS_PATH=./migrations

# Default target
help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

deps: ## Install dependencies
	@echo "📦 Installing dependencies..."
	go mod download
	go mod tidy
	go install github.com/swaggo/swag/cmd/swag@latest
	@echo "✅ Dependencies installed"

run: ## Run the application
	@echo "🚀 Starting application..."
	go run $(MAIN_PATH)/main.go

dev: ## Run the application with hot reload (requires air)
	@echo "🔥 Starting application with hot reload..."
	@if ! command -v air > /dev/null; then \
		echo "Installing air..."; \
		go install github.com/air-verse/air@latest; \
	fi
	go dev $(MAIN_PATH)/main.go

build: ## Build the application
	@echo "🔨 Building application..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)/main.go
	@echo "✅ Build complete: $(BUILD_DIR)/$(APP_NAME)"

test: ## Run tests
	@echo "🧪 Running tests..."
	go test -v -cover ./...

test-coverage: ## Run tests with coverage report
	@echo "🧪 Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

migrate: ## Run database migrations
	@echo "🗄️  Running database migrations..."
	@mkdir -p ./data
	sqlite3 $(DB_PATH) < $(MIGRATIONS_PATH)/001_initial_schema.sql
	@echo "✅ Migrations complete"

migrate-with-data: ## Run migrations and load sample data
	@echo "🗄️  Running database migrations with sample data..."
	@mkdir -p ./data
	sqlite3 $(DB_PATH) < $(MIGRATIONS_PATH)/001_initial_schema.sql
	sqlite3 $(DB_PATH) < $(MIGRATIONS_PATH)/002_sample_data.sql
	@echo "✅ Migrations and sample data loaded"

migrate-fresh: ## Drop all tables and re-run migrations
	@echo "🗄️  Fresh migration..."
	@rm -f $(DB_PATH)
	@mkdir -p ./data
	sqlite3 $(DB_PATH) < $(MIGRATIONS_PATH)/001_initial_schema.sql
	@echo "✅ Fresh migrations complete"

migrate-fresh-with-data: ## Fresh migration with sample data
	@echo "🗄️  Fresh migration with sample data..."
	@rm -f $(DB_PATH)
	@mkdir -p ./data
	sqlite3 $(DB_PATH) < $(MIGRATIONS_PATH)/001_initial_schema.sql
	sqlite3 $(DB_PATH) < $(MIGRATIONS_PATH)/002_sample_data.sql
	@echo "✅ Fresh migrations with sample data complete"

swagger: ## Generate Swagger documentation
	@echo "📚 Generating Swagger documentation..."
	swag init -g $(MAIN_PATH)/main.go -o ./docs/swagger
	@echo "✅ Swagger documentation generated"

clean: ## Clean build artifacts
	@echo "🧹 Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "✅ Clean complete"

clean-db: ## Remove database file
	@echo "🗑️  Removing database..."
	@rm -f $(DB_PATH)
	@echo "✅ Database removed"

lint: ## Run linter
	@echo "🔍 Running linter..."
	@if ! command -v golangci-lint > /dev/null; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	golangci-lint run ./...

format: ## Format code
	@echo "💅 Formatting code..."
	go fmt ./...
	@echo "✅ Code formatted"

setup: deps migrate-fresh swagger ## Complete project setup
	@echo "✅ Project setup complete!"
	@echo "Run 'make run' to start the application"

all: clean deps swagger build ## Clean, install deps, generate swagger, and build
	@echo "✅ All tasks complete"