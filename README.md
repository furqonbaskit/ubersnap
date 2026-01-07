# Ubersnap - Coupon Management API

A Go REST API built with Gin and PostgreSQL for managing and claiming coupons with race condition handling.

## Prerequisites

- Docker Desktop (includes Docker and Docker Compose)
- Git
- Go 1.21+ (for local development)

## Project Structure

```
ubersnap/
├── cmd/
│   └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go        # Configuration management
│   ├── database/
│   │   └── postgre.go       # Database initialization
│   ├── handlers/
│   │   └── coupon_handler.go # HTTP request handlers
│   ├── models/
│   │   ├── coupon.go        # Coupon model
│   │   └── claim.go         # Claim model
│   ├── repositories/
│   │   └── coupon_repository.go # Data access layer
│   ├── routes/
│   │   └── routes.go        # Route definitions
│   └── services/
│       └── coupon_service.go # Business logic
├── docker-compose.yml
├── Dockerfile
├── go.mod
├── go.sum
└── README.md
```

## How to Run

### Using Docker Compose (Recommended)

```bash
# Clone the repository
git clone <repository-url>
cd ubersnap

# Set up environment variables
cp .env.example .env

# Start the application and PostgreSQL
docker-compose up --build

# The API will be available at http://localhost:8085
```

### Local Development

```bash
# Install dependencies
go mod download

# Set up environment variables
cp .env.example .env

# Run the application
go run ./cmd/main.go
```

## How to Test

```bash
go test ./...
```

### Run Specific Test Suites

```bash
# Test handlers (integration tests)
go test -v ./internal/handlers/...

#Run Flash Sale Attack test
go test -v -run TestFlashSaleAttackIntegration ./internal/handlers/...

# Run Double Dip Attack test
go test -v -run TestDoubleDipAttackIntegration ./internal/handlers/...

# Run Create Coupon test
go test -v -run TestCreateCouponIntegration ./internal/handlers/...
```

### Create a Coupon

```bash
curl -X POST http://localhost:8085/api/coupons \
  -H "Content-Type: application/json" \
  -d '{
    "name": "SUMMER2024",
    "amount": 100
  }'
```

### Get Coupon by Name

```bash
curl http://localhost:8085/api/coupons/SUMMER2024
```

### Claim a Coupon

```bash
curl -X POST http://localhost:8085/api/coupons/claim \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user_12345",
    "coupon_name": "SUMMER2024"
  }'
```

### Response Example

```json
{
  "name": "SUMMER2024",
  "amount": 100,
  "remaining_amount": 99,
  "claimed_by": ["user_12345"]
}
```

## Architecture Notes

### Database Design

**Coupons Table**
- `id` (UUID): Primary key
- `name` (String): Unique coupon identifier
- `amount` (Int8): Total coupon quantity
- `remaining_amount` (Int8): Available coupons left
- `created_at`, `updated_at`: Timestamps
- `created_by`, `updated_by`: User tracking

**Claims Table**
- `id` (UUID): Primary key
- `user_id` (String): User identifier
- `coupon_id` (UUID): Foreign key to coupons
- `created_at`, `updated_at`: Timestamps
- Unique constraint: `(user_id, coupon_id)` - prevents duplicate claims

### Race Condition Handling

The `ClaimCoupon` operation uses database-level locking to prevent race conditions:

1. **Row Locking**: Uses `FOR UPDATE` clause to lock the coupon row
2. **Transactions**: All operations (check, create claim, update amount) happen atomically
3. **Validation**: Checks remaining amount and existing claims within the transaction
4. **Atomic Update**: Decrements `remaining_amount` only if conditions are met

This ensures:
- No overselling of coupons
- No duplicate claims by the same user
- Data consistency under concurrent load

### Technology Stack

- **Framework**: Gin Web Framework
- **Database**: PostgreSQL with GORM ORM
- **Containerization**: Docker & Docker Compose
- **Language**: Go 1.21

## Stopping the Application

```bash
# Stop Docker Compose services
docker-compose down

# Remove volumes (optional)
docker-compose down -v
```
