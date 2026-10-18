# Go Online Car Marketplace

A RESTful API for an online car marketplace built with Go, Gin framework, and SQLite. This application provides comprehensive functionality for managing car listings, user authentication, inquiries, reviews, and favorites across three user types: individuals, dealerships, and showrooms.

## Project Structure

```
go-online-marketplace/
├── cmd/
│   └── api/
│       └── main.go                 # Entry point (-migrate-only, -seed flags)
├── internal/
│   ├── app/
│   │   ├── app.go                  # Dependency wiring and background jobs
│   │   └── app_test.go             # HTTP integration tests
│   ├── config/
│   │   └── config.go               # Configuration management
│   ├── domain/
│   │   ├── user.go                 # User entities
│   │   ├── car.go                  # Car entities
│   │   ├── auth.go                 # Auth entities
│   │   ├── inquiry.go              # Inquiry entities
│   │   ├── review.go               # Review entities
│   │   ├── favorite.go             # Favorite entities
│   │   └── errors.go               # Domain errors
│   ├── repository/
│   │   ├── user_repository.go      # User repo interface
│   │   ├── car_repository.go       # Car repo interface
│   │   ├── other_repositories.go   # Other repo interfaces
│   │   └── transactor.go           # Transaction interface
│   ├── usecase/
│   │   ├── auth_usecase.go         # Auth business logic
│   │   ├── car_usecase.go          # Car business logic
│   │   ├── other_usecases.go       # Other business logic
│   │   ├── reference_cache.go      # In-memory reference data
│   │   └── view_counter.go         # Batched listing view counts
│   ├── delivery/
│   │   └── http/
│   │       ├── handler/
│   │       │   ├── auth_handler.go # Auth HTTP handlers
│   │       │   ├── car_handler.go  # Car HTTP handlers
│   │       │   ├── other_handlers.go # Other HTTP handlers
│   │       │   └── errors.go       # Domain error → HTTP status mapping
│   │       ├── middleware/
│   │       │   ├── auth_middleware.go # JWT validation
│   │       │   ├── cors_middleware.go # CORS support
│   │       │   └── rate_limit.go   # Rate limiting
│   │       └── router/
│   │           └── router.go       # Route definitions
│   └── infrastructure/
│       ├── database/
│       │   └── sqlite/
│       │       ├── sqlite.go       # DB connection
│       │       ├── migrate.go      # Embedded migration runner
│       │       ├── tx.go           # Context-scoped transactions
│       │       ├── user_repository.go # User data access
│       │       ├── car_repository.go # Car data access
│       │       └── other_repositories.go # Other data access
│       └── security/
│           ├── jwt.go              # JWT utilities
│           └── password.go         # Password hashing
├── pkg/
│   └── response/
│       └── response.go             # Response helpers
├── migrations/
│   ├── embed.go                    # Embeds the SQL files into the binary
│   ├── 001_initial_schema.sql      # Database schema
│   ├── 002_search_and_integrity.sql # Full-text search, indexes, constraints
│   └── seeds/
│       └── sample_data.sql         # Sample data (loaded with -seed)
├── docs/
│   └── swagger/                    # Swagger documentation
├── .env.example                    # Example environment file
├── Makefile                        # Build commands
├── go.mod                          # Go modules
└── README.md                       # This file
```



## Features

### Authentication & Security
- JWT-based authentication with access and refresh tokens
- HTTP-only cookie support for secure token storage
- Bcrypt password hashing (cost 12)
- Token refresh mechanism (15-minute access tokens, 7-day refresh tokens)
- Multi-device logout support
- Secure password change functionality
- Rate limiting (10 req/sec global, 5 req/sec on auth endpoints)

### User Management
- Three user types: **Individual**, **Dealership**, and **Showroom**
- Flexible user profiles with type-specific schemas
- Business profile support with multi-location addresses
- Email-based authentication and verification

### Car Listing Management
- Complete CRUD operations for car listings
- Owner verification and authorization
- Advanced filtering by:
  - Make and model
  - Year range (min/max)
  - Price range (min/max)
  - Location
  - Car condition (new, used, certified pre-owned)
  - Fuel type, transmission type, body type
- Multiple sorting options (price, year, mileage, created date)
- Image upload and management (multiple images per listing)
- Feature assignment (safety, comfort, entertainment, technology)
- Status tracking (active, sold, pending, expired)
- View counter for listing popularity

### Marketplace Features
- **Inquiries/Messaging**: Direct communication between buyers and sellers
- **Reviews & Ratings**: 5-star rating system with comments
- **User Reputation**: Aggregate rating statistics
- **Favorites/Wishlist**: Save listings for later viewing
- **Reference Data API**: Access to makes, models, colors, features, etc.

### API Features
- RESTful API design
- Pagination on all list endpoints (default 20 items per page)
- CORS support for cross-origin requests
- Swagger/OpenAPI documentation
- Standardized JSON response format
- Comprehensive error handling with domain-specific errors
- Bearer token authentication


## Technology Stack

- **Go**: 1.24.1
- **Web Framework**: Gin (v1.10.0)
- **Database**: SQLite (v1.14.24)
- **Authentication**: JWT (golang-jwt/jwt v5.2.1)
- **Password Hashing**: Bcrypt (golang.org/x/crypto)
- **Configuration**: Viper (v1.19.0)
- **Validation**: Go Playground Validator (v10.22.1)
- **Documentation**: Swag (v1.16.4)
- **Testing**: Testify (v1.10.0)

## Prerequisites

- Go 1.24.1 or higher
- Make (optional, for Makefile commands)
- Git

## Installation

1. **Clone the repository**:
```bash
git clone https://github.com/yourusername/go-online-marketplace.git
cd go-online-marketplace
```

2. **Install dependencies**:
```bash
make deps
# or
go mod download
```

3. **Set up the database and generate Swagger docs**:
```bash
make setup
# This runs: deps, migrate-fresh, and swagger
```

The server applies pending migrations automatically on startup, so the
`sqlite3` CLI is not required. To load the sample data, run
`make migrate-with-data` (or `go run ./cmd/api -migrate-only -seed`).

## Configuration

### Environment Variables

The application uses environment variables for configuration. Copy the example file and customize it:

```bash
cp .env.example .env
```

Then edit `.env` to set your configuration:

```env
# Server Configuration
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
GIN_MODE=debug  # debug, release, test
SERVER_READ_TIMEOUT=10s
SERVER_WRITE_TIMEOUT=10s
SERVER_IDLE_TIMEOUT=60s
COOKIE_SECURE=false        # Secure auth cookies; defaults to true in release mode
TRUSTED_PROXIES=           # Proxies allowed to set X-Forwarded-For (empty = none)

# CORS Configuration
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173

# Database Configuration
DATABASE_PATH=./data/marketplace.db
DATABASE_MAX_OPEN_CONNS=25
DATABASE_MAX_IDLE_CONNS=5
DATABASE_MAX_LIFETIME=5m

# JWT Configuration
# IMPORTANT: Generate a strong random key for production!
# You can use: openssl rand -base64 64
JWT_SECRET_KEY=your-super-secret-key-change-this-in-production
JWT_ACCESS_TOKEN_TTL=15m
JWT_REFRESH_TOKEN_TTL=168h  # 7 days

# Rate Limiting Configuration
RATE_LIMIT_ENABLED=true
RATE_LIMIT_RPS=10.0
RATE_LIMIT_BURST=20
RATE_LIMIT_CLEANUP=5m
```

**CRITICAL for Production**:
- Generate a strong `JWT_SECRET_KEY` (at least 32 characters): `openssl rand -base64 64`.
  The server refuses to start in release mode with the built-in placeholder.
- Set `GIN_MODE=release` for production
- Behind a reverse proxy, set `TRUSTED_PROXIES` so rate limiting sees real client IPs
- Use appropriate database path and backup strategy
- Adjust rate limiting based on your needs

## Running the Application

### Using Make

```bash
# Run the application
make run

# Run with hot reload (requires air)
make dev

```

### Using Go directly

```bash
go run cmd/api/main.go
```

### Hot Reloading with Air (Development)

This project includes Air for hot-reloading during development. Air automatically rebuilds and restarts your application when code changes are detected.

#### Install Air

```bash
# Install Air globally
go install github.com/air-verse/air@latest
```

#### Run with Air

```bash
# Start the application with hot reload
air

# Or use the make command
make dev
```

Air will:
- Watch for changes in `.go` files
- Automatically rebuild the application
- Restart the server with the new changes
- Display build errors in `build-errors.log`

Configuration is in [.air.toml](.air.toml). The compiled binary runs from `./tmp/main.exe`.

The server will start on `http://localhost:8080` (or the port specified in your configuration).

## API Documentation

### Swagger UI

Once the application is running, access the interactive API documentation at:

```
http://localhost:8080/swagger/index.html
```

### Quick Start Guide

1. **Register a new user**:
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "user_type": "individual",
    "first_name": "Jane",
    "last_name": "Doe"
  }'
```

2. **Login** (tokens are set as HTTP-only cookies; `-c` saves them):
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -c cookies.txt \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

3. **Create a car listing** (requires authentication; send the cookies with `-b`,
or an `Authorization: Bearer <access_token>` header):
```bash
curl -X POST http://localhost:8080/api/v1/cars \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -d '{
    "make_id": 1,
    "model_id": 1,
    "year": 2023,
    "body_type_id": 1,
    "fuel_type_id": 1,
    "transmission_type_id": 1,
    "color_id": 1,
    "condition_id": 2,
    "mileage": 15000,
    "engine_size": 2.0,
    "price": 25000,
    "currency": "USD",
    "title": "2023 Toyota Camry SE",
    "description": "Well-maintained car with full service history.",
    "city": "New York",
    "state": "NY",
    "country": "USA"
  }'
```


### API Endpoints Overview

All endpoints except `/health`, `/ping` and `/swagger` are under `/api/v1`.

| Category | Method | Endpoint | Description | Auth Required |
|----------|--------|----------|-------------|---------------|
| **System** | GET | `/health` | Health check | No |
| **System** | GET | `/ping` | Ping endpoint | No |
| **Auth** | POST | `/auth/register` | Register new user | No |
| **Auth** | POST | `/auth/login` | User login | No |
| **Auth** | POST | `/auth/refresh` | Refresh access token | No |
| **Auth** | POST | `/auth/logout` | Logout (current device) | No (refresh cookie) |
| **Auth** | POST | `/auth/logout-all` | Logout (all devices) | Yes |
| **Auth** | POST | `/auth/change-password` | Change password | Yes |
| **Auth** | GET | `/auth/profile` | Get user profile | Yes |
| **Cars** | GET | `/cars` | List car listings | No |
| **Cars** | GET | `/cars/:id` | Get listing details | No |
| **Cars** | POST | `/cars` | Create listing | Yes |
| **Cars** | PUT | `/cars/:id` | Update listing | Yes |
| **Cars** | DELETE | `/cars/:id` | Delete listing | Yes |
| **Cars** | POST | `/cars/:id/sold` | Mark as sold | Yes |
| **Cars** | POST | `/cars/:id/images` | Add image | Yes |
| **Cars** | PUT | `/cars/:id/images/:image_id/primary` | Set primary image | Yes |
| **Cars** | DELETE | `/cars/:id/images/:image_id` | Delete image | Yes |
| **Reference** | GET | `/reference/makes` | Get car makes | No |
| **Reference** | GET | `/reference/makes/:make_id/models` | Get models | No |
| **Reference** | GET | `/reference/data` | Get all reference data | No |
| **Inquiries** | POST | `/inquiries` | Send inquiry | Yes |
| **Inquiries** | GET | `/inquiries` | List inquiries | Yes |
| **Inquiries** | PUT | `/inquiries/:id/read` | Mark as read | Yes |
| **Reviews** | GET | `/reviews` | List reviews | No |
| **Reviews** | GET | `/reviews/users/:user_id/stats` | Get user stats | No |
| **Reviews** | POST | `/reviews` | Create review | Yes |
| **Reviews** | PUT | `/reviews/:id` | Update review | Yes |
| **Reviews** | DELETE | `/reviews/:id` | Delete review | Yes |
| **Favorites** | GET | `/favorites` | List favorites | Yes |
| **Favorites** | POST | `/favorites/:id` | Add to favorites | Yes |
| **Favorites** | DELETE | `/favorites/:id` | Remove from favorites | Yes |

## Database Schema

The application uses SQLite with the following main tables:

### User Management
- `users` - User accounts
- `user_types` - Individual, dealership, showroom
- `user_profiles` - Individual user profiles
- `business_profiles` - Business information
- `business_addresses` - Business locations
- `refresh_tokens` - JWT refresh token storage

### Car Listings
- `car_listings` - Main listing table
- `car_images` - Multiple images per listing
- `car_listing_features` - Feature associations

### Reference Data
- `car_makes` - Manufacturers (Toyota, Honda, etc.)
- `car_models` - Models by make
- `body_types` - Sedan, SUV, Truck, etc.
- `fuel_types` - Petrol, Diesel, Electric, etc.
- `transmission_types` - Manual, Automatic, CVT, etc.
- `colors` - Available colors with hex codes
- `car_conditions` - New, Used, Certified Pre-Owned
- `features` - Safety, comfort, technology features

### Transactions
- `inquiries` - Buyer-seller messages
- `reviews` - User ratings and reviews
- `favorites` - User wishlists

Database migrations are located in the `migrations/` directory.

## Testing

### Run all tests
```bash
make test
```

### Run tests with coverage
```bash
make test-coverage
```

### Run tests with race detection
```bash
make test-race
```

### Run specific test package
```bash
go test ./internal/usecase/...
```

For detailed test documentation, see the test files in each package.



## Development

### Available Make Commands

```bash
make help              # Show all available commands
make deps              # Install dependencies
make run               # Run the application
make dev               # Run with hot reload
make build             # Build executable
make test              # Run tests
make test-coverage     # Run tests with coverage
make test-race         # Run tests with race detection
make lint              # Run linter
make fmt               # Format code
make migrate-fresh     # Fresh database with sample data
make swagger           # Generate Swagger docs
make clean             # Clean build artifacts
make setup             # Complete project setup

```

### Code Style

This project follows standard Go conventions:
- Use `gofmt` for formatting
- Use `golint` or `golangci-lint` for linting
- Follow effective Go practices

### Adding New Features

1. Define domain entities in `internal/domain/`
2. Create repository interface in `internal/repository/`
3. Implement business logic in `internal/usecase/`
4. Create HTTP handlers in `internal/delivery/http/handler/`
5. Add routes in `internal/delivery/http/router/router.go`
6. Implement data access in `internal/infrastructure/database/sqlite/`
7. Write tests for all layers
8. Update Swagger documentation

## Deployment

### Building for Production

```bash
# Build optimized binary
make build

# The binary will be in ./bin/api
./bin/api

```

### Production Deployment Checklist

Before deploying to production:

#### Required Security Updates
- [ ] **Generate a strong `JWT_SECRET_KEY`** (min 32 chars): `openssl rand -base64 64`
- [ ] Set `GIN_MODE=release` for production
- [ ] Remove or secure the `.env` file (never commit it!)
- [ ] Review all environment variables for sensitive data

#### Infrastructure Setup
- [ ] Configure appropriate database path and backup strategy
- [ ] Set up proper logging and log rotation
- [ ] Set `CORS_ALLOWED_ORIGINS` to your frontend origins
- [ ] Set `TRUSTED_PROXIES` to your reverse proxy address
- [ ] Set up SSL/TLS certificates (use reverse proxy like nginx)
- [ ] Configure rate limiting based on expected traffic
- [ ] Set up monitoring and alerting
- [ ] Configure firewall rules

#### Testing & Validation
- [ ] Run all tests: `make test`
- [ ] Test authentication flows
- [ ] Verify API endpoints work correctly
- [ ] Test rate limiting behavior
- [ ] Load test the application
