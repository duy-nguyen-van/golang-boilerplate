# Golang Boilerplate (Echo + FX)

A production-ready Go web application built on Echo, featuring clean architecture, dependency injection (Uber FX), Keycloak auth integration, Redis caching, PostgreSQL, structured logging, and observability (New Relic + Sentry).

## Table of Contents

- [Features](#features)
- [Project Structure](#project-structure)
- [Quick Start](#quick-start)
  - [Prerequisites](#prerequisites)
  - [Developer tools](#developer-tools)
  - [Setup](#setup)
- [API Endpoints](#api-endpoints)
  - [Public Endpoints](#public-endpoints)
  - [Health Check Endpoints](#health-check-endpoints)
  - [Protected Endpoints](#protected-endpoints-require-jwt)
- [Example API Usage](#example-api-usage)
- [Development](#development)
  - [Available Make Commands](#available-make-commands)
  - [Testing](#testing)
    - [Test Organization](#test-organization)
    - [Running Tests](#running-tests)
    - [Test Structure](#test-structure)
    - [Test Package Naming](#test-package-naming)
    - [Testing Different Layers](#testing-different-layers)
    - [Example Test Files](#example-test-files)
    - [Test Dependencies](#test-dependencies)
    - [Best Practices](#best-practices)
    - [Test Coverage](#test-coverage)
    - [Continuous Integration](#continuous-integration)
- [Architecture](#architecture)
  - [Dependency Injection with Uber FX](#dependency-injection-with-uber-fx)
  - [DTO & Model Layers](#dto--model-layers)
- [Error Handling System](#error-handling-system)
  - [Error Types](#error-types)
  - [Error Structure](#error-structure)
  - [Usage Examples](#usage-examples)
  - [Error Response Format](#error-response-format)
  - [Middleware Integration](#middleware-integration)
  - [Monitoring and Logging](#monitoring-and-logging)
- [Database Connection Management](#database-connection-management)
  - [Database Features](#database-features)
  - [Database Architecture](#database-architecture)
  - [Connection Pooling](#connection-pooling)
  - [Database Health Monitoring](#database-health-monitoring)
  - [Connection Metrics](#connection-metrics)
- [Configuration](#configuration)
  - [Database Configuration Parameters](#database-configuration-parameters)
  - [Rate Limiting](#rate-limiting)
- [Docker](#docker)
  - [Build and Run](#build-and-run)
  - [Services Included](#services-included)
- [Database Monitoring and Troubleshooting](#database-monitoring-and-troubleshooting)
  - [Metrics to Monitor](#metrics-to-monitor)
  - [Recommended Alerts](#recommended-alerts)
  - [Common Issues and Solutions](#common-issues-and-solutions)
  - [Debugging](#debugging)
  - [Performance Considerations](#performance-considerations)
- [Database Migrations](#database-migrations)
- [Production Deployment](#production-deployment)
- [Contributing](#contributing)
- [License](#license)

## Features

- **Clean Architecture**: Layered modules and clear separation of concerns
- **Dependency Injection**: Uber FX for modular dependency management
- **DTO & Model Layers**: Separation between API DTOs and domain models
- **Comprehensive Error Handling**: Structured error system with context, logging, and monitoring
- **Authentication**: JWT-based authentication with Keycloak integration
- **Caching**: Redis cache provider
- **Database**: PostgreSQL with migrations ([Atlas](https://atlasgo.io/))
- **Email**: AWS SES integration
- **Logging**: Structured logging with Zap
- **Observability**: New Relic APM + Sentry error tracking
- **Docker**: Dockerfile and Compose services for Postgres/Redis
- **Middleware**: Auth, CORS, logging, rate limiting, error handling
- **Health Checks**: Built-in health endpoint

## Project Structure

```text
golang-boilerplate/
├─ cmd/
│  ├─ migrations/
│  │  └─ sql/                  # Atlas migration files + atlas.sum
│  │     └─ 20260328081444_init_tables.sql
│  └─ server/
│     ├─ main.go                 # Application entrypoint + FX wiring
│     └─ routes/
│        └─ router.go            # Echo routes and middleware
│
├─ docs/                         # Project documentation (markdown)
│
├─ internal/
│  ├─ cache/                     # Cache abstraction + Redis
│  │  ├─ cache.go
│  │  └─ redis.go
│  ├─ config/                    # Config loader and env bindings
│  │  └─ config.go
│  ├─ constants/                 # Error codes, pagination, providers
│  │  ├─ error_codes.go
│  │  ├─ pagination.go
│  │  └─ third_party_provider.go
│  ├─ db/                        # Database connection management
│  │  ├─ manager.go              # Database manager with connection pooling
│  │  └─ postgres.go             # Postgres connection wrapper
│  ├─ dtos/                      # API DTOs
│  │  ├─ common.go
│  │  ├─ company.go
│  │  ├─ email.go
│  │  ├─ health.go
│  │  └─ user.go
│  ├─ errors/                    # Comprehensive error handling system
│  │  ├─ app_error.go            # Custom error types and structures
│  │  ├─ handler.go              # Error handler utilities
│  │  └─ middleware.go           # Error middleware for panic recovery
│  ├─ handlers/                  # Echo handlers
│  │  ├─ base.go                 # Base handler with error handling
│  │  ├─ company.go              # Company management endpoints
│  │  ├─ health.go               # Health check endpoints
│  │  └─ user.go                 # User management endpoints
│  ├─ httpclient/                # Outbound HTTP client (Resty)
│  │  └─ resty.go
│  ├─ integration/               # External integrations
│  │  ├─ auth/
│  │  │  ├─ auth.go
│  │  │  └─ keycloak.go
│  │  └─ email/
│  │     ├─ email.go
│  │     └─ ses.go
│  ├─ logger/
│  │  └─ logger.go
│  ├─ middlewares/
│  │  ├─ auth.go
│  │  ├─ basic_auth.go
│  │  ├─ cors.go
│  │  ├─ logging.go
│  │  └─ rate_limiter.go
│  ├─ models/
│  │  ├─ auth.go
│  │  ├─ base.go
│  │  ├─ company.go
│  │  ├─ email.go
│  │  └─ user.go
│  ├─ monitoring/
│  │  ├─ newrelic_zap.go
│  │  ├─ new_relic.go
│  │  └─ sentry.go
│  ├─ repositories/
│  │  ├─ abstract.go
│  │  ├─ company.go
│  │  └─ user.go
│  ├─ services/
│  │  ├─ auth.go
│  │  ├─ company.go
│  │  ├─ email.go
│  │  └─ user.go
│  └─ utils/
│     ├─ accent.go
│     ├─ date.go
│     └─ i18n/
│        └─ translator.go
│
├─ atlas.hcl                   # Atlas env (GORM schema → migrate diff)
├─ Dockerfile
├─ docker-compose.yml
├─ go.mod
├─ go.sum
├─ Makefile
└─ README.md
```

## Quick Start

### Prerequisites

- Go 1.25+
- Docker and Docker Compose
- Make (optional)

#### Developer tools

```bash
# Database migrations — install the Atlas CLI (see https://atlasgo.io/getting-started#installation)
# macOS (Homebrew):
brew install ariga/tap/atlas

# Or use the install script from the Atlas docs for Linux/other platforms.
```

Docker is required for some Atlas commands (for example `migrate-down` and `migrate-generate`), which use a temporary dev database (`DB_DEV_URL` defaults to `docker://postgres/18/dev` in the `Makefile`).

### Setup

#### 1. Clone and setup

```bash
git clone <repository-url>
cd golang-boilerplate
go mod tidy
```

#### 2. Configure environment

##### 2.1. Configure server environment (app + migrations)

```bash
cp examples/env/server.env.example cmd/server/.env
```

Migration Make targets read PostgreSQL settings from `cmd/server/.env` (see `Makefile`: `POSTGRES_*` are composed into `DB_DSN` for Atlas).

#### 3. Start dependencies with Docker

```bash
make container-up
```

#### 4. Run the application locally

```bash
# Run DB migrations (uses DB settings from cmd/server/.env)
make migrate-up

# Start the server
make up
```

### API Endpoints

#### Public Endpoints

- `GET /api/v1/` - Health check

#### Health Check Endpoints

- `GET /api/v1/health/database` - Database health status with connection metrics
- `GET /api/v1/health/metrics` - Comprehensive database metrics and configuration

#### Protected Endpoints (require JWT)

**User Management:**

- `POST /api/v1/users` - Create user
- `GET /api/v1/users/{id}` - Get user by ID
- `PUT /api/v1/users/{id}` - Update user
- `DELETE /api/v1/users/{id}` - Delete user
- `GET /api/v1/users` - Get users list
- `GET /api/v1/users/test-rest-client` - Demo endpoint to test outbound REST client

**Company Management:**

- `POST /api/v1/companies` - Create new company
- `GET /api/v1/companies/{id}` - Get company by ID
- `PUT /api/v1/companies/{id}` - Update company
- `DELETE /api/v1/companies/{id}` - Delete company
- `GET /api/v1/companies` - Get companies list

### Example API Usage

#### Create user (with JWT token)

```bash
curl -X POST http://localhost:3000/api/v1/users \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "John Doe", "email": "john@example.com"}'
```

#### Get user by ID (with JWT token)

```bash
curl -X GET http://localhost:3000/api/v1/users/123 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Create company (with JWT token)

```bash
curl -X POST http://localhost:3000/api/v1/companies \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "Acme Corp", "description": "A great company"}'
```

#### Database health check

```bash
curl -X GET http://localhost:3000/api/v1/health/database
```

#### Database metrics

```bash
curl -X GET http://localhost:3000/api/v1/health/metrics
```

## Development

### Available Make Commands

```bash
# Development
make container-up         # Start Docker services (postgres, redis)
make container-down       # Stop Docker services
make up                   # Run the server (cmd/server)
make build cmd=server service_name=main   # Build linux binary for server
make dep                  # go mod tidy
make lint                 # Run golangci-lint
make format               # Format code

# Testing (see Testing section for details)
make tests                # Run all tests with coverage and race detection
make test-services        # Run service layer tests
make test-utils           # Run utility tests
make test-handlers        # Run handler tests
make test-repositories    # Run repository tests
make test-coverage        # Run tests with coverage
make test-coverage-html   # Generate HTML coverage report
make test-race            # Run tests with race detection
make test-verbose         # Run tests with verbose output
make test-specific TEST=TestName  # Run a specific test

# Migrations (Atlas; DB URL from cmd/server/.env via Makefile DB_DSN)
make migrate-status
make migrate-up
make migrate-up-preview      # dry-run apply
make migrate-down          # requires Docker dev URL by default
make migrate-down-preview
make migrate-create name=add_table
make migrate-hash          # refresh atlas.sum after editing migrations
make migrate-inspect       # inspect live schema
make migrate-generate name=my_change   # GORM diff → new migration (atlas.hcl env "gorm")

```

### Testing

This project follows Go testing conventions and best practices for organizing and writing unit tests.

#### Test Organization

Test files are placed in the **same directory** as the source code they test, with the `_test.go` suffix:

```text
internal/
├── services/
│   ├── user.go           # Source code
│   └── user_test.go      # Unit tests for user service
├── handlers/
│   ├── user.go
│   └── user_test.go      # Unit tests for user handler
├── repositories/
│   ├── user.go
│   └── user_test.go      # Unit tests for user repository
└── utils/
    ├── date.go
    └── date_test.go      # Unit tests for date utilities
```

#### Running Tests

The project provides several Makefile targets for running tests:

**Main Test Commands:**

```bash
# Run all tests with coverage and race detection (recommended)
make tests

# Run all tests with verbose output
make test-verbose

# Run all tests with coverage report
make test-coverage

# Generate HTML coverage report
make test-coverage-html

# Run tests with race detection
make test-race
```

**Layer-Specific Test Commands:**

```bash
# Run service layer tests
make test-services

# Run utility tests
make test-utils

# Run handler tests
make test-handlers

# Run repository tests
make test-repositories
```

**Specific Test Commands:**

```bash
# Run a specific test function
make test-specific TEST=TestUserService_Create

# Run a specific test with verbose output
make test-specific-verbose TEST=TestUserService_Create

# Run a specific test with coverage
make test-specific-coverage TEST=TestUserService_Create
```

**Direct Go Commands (alternative to Makefile):**

```bash
# Run tests for a specific package
go test ./internal/services

# Run tests for a specific package with coverage
go test -cover ./internal/services

# Run a specific test function
go test -run TestUserService_Create ./internal/services

# Generate HTML coverage report manually
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

#### Test Structure

Tests follow a table-driven approach for multiple scenarios:

```go
func TestUserService_Create(t *testing.T) {
    tests := []struct {
        name          string
        req           *dtos.CreateUserRequest
        expectedError bool
    }{
        {
            name:          "success - valid request",
            req:           &dtos.CreateUserRequest{...},
            expectedError: false,
        },
        {
            name:          "error - invalid email",
            req:           &dtos.CreateUserRequest{...},
            expectedError: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

#### Test Package Naming

**Same Package (White-box Testing):**

- Uses the same package name as source code (e.g., `package services`)
- Can test unexported functions and internal implementation
- Best for unit tests that need access to internals

**Test Package (Black-box Testing):**

- Uses `_test` suffix (e.g., `package services_test`)
- Tests only the public API
- More resilient to internal refactoring

#### Testing Different Layers

**Repository Tests** (`internal/repositories/*_test.go`):

- Test data access logic and database queries
- Mock or use in-memory databases
- Test CRUD operations, filters, sorting, pagination

**Service Tests** (`internal/services/*_test.go`):

- Test business logic and validation rules
- Mock repository interfaces
- Test error handling and orchestration

**Handler Tests** (`internal/handlers/*_test.go`):

- Test HTTP request/response handling
- Use Echo test utilities
- Mock service interfaces
- Test authentication, validation, status codes

**Utility Tests** (`internal/utils/*_test.go`):

- Test pure functions and helper utilities
- Usually no mocking needed
- Focus on edge cases and correctness

#### Example Test Files

The project includes comprehensive test files demonstrating best practices:

**Service Layer Tests:**

- `internal/services/user_test.go` - User service with mocked repositories
- `internal/services/company_test.go` - Company service tests
- `internal/services/email_test.go` - Email service with mocked email sender
- `internal/services/auth_test.go` - Auth service with mocked auth provider

**Utility Tests:**

- `internal/utils/date_test.go` - Date parsing and validation tests
- `internal/utils/sort_test.go` - Sort validation with table-driven tests

**HTTP Client Tests:**

- `internal/httpclient/resty_test.go` - REST client integration tests

#### Test Dependencies

The project uses [testify](https://github.com/stretchr/testify) for assertions and mocking:

- `assert` - Better assertions with helpful error messages
- `require` - Same as assert but stops test execution on failure
- `mock` - Mock objects for dependencies

#### Best Practices

1. **Keep tests isolated** - Each test should be independent
2. **Test edge cases** - Don't just test the happy path
3. **Use meaningful names** - Test names should describe what is being tested
4. **Mock external dependencies** - Use interfaces to mock databases, APIs, etc.
5. **Keep tests fast** - Unit tests should complete quickly
6. **Test error cases** - Test both success and failure scenarios
7. **Use table-driven tests** - For multiple test cases with similar structure
8. **Avoid testing third-party code** - Focus on your own code

#### Test Coverage

Aim for good test coverage, especially for critical business logic:

```bash
# Check coverage for all packages
go test -cover ./...

# Generate detailed coverage report
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

# View coverage in browser
go tool cover -html=coverage.out
```

#### Continuous Integration

Tests are automatically run in CI/CD pipelines. Ensure all tests pass before committing:

```bash
# Run full test suite before committing (recommended)
make tests

# Or run tests with race detection separately
make test-race

# Check test coverage
make test-coverage-html
```

**Pre-commit Checklist:**

- ✅ All tests pass: `make tests`
- ✅ Code is formatted: `make format`
- ✅ Linter passes: `make lint`
- ✅ Coverage meets requirements: `make test-coverage`

## Architecture

### Dependency Injection with Uber FX

Uber FX wires the application graph (config, logger, monitoring, db, cache, repositories, services, handlers, HTTP server). See `cmd/server/main.go` for providers and lifecycle hooks.

### DTO & Model Layers

The application separates domain models and API DTOs:

- **Models** (`internal/models/`): Domain entities
- **DTOs** (`internal/dtos/`): Request/response structs

Benefits:

- **Type Safety** and **stability** across API boundaries
- **Security**: sensitive fields are not exposed via DTOs
- **Consistency**: standardized response envelope

## Error Handling System

The application features a comprehensive error handling system that provides consistent error responses, structured logging, and monitoring integration.

### Error Types

The system defines several error types for different scenarios:

- `ErrorTypeValidation` - Input validation errors
- `ErrorTypeNotFound` - Resource not found errors
- `ErrorTypeUnauthorized` - Authentication errors
- `ErrorTypeForbidden` - Authorization errors
- `ErrorTypeConflict` - Resource conflict errors
- `ErrorTypeInternal` - Internal server errors
- `ErrorTypeExternal` - External service errors
- `ErrorTypeDatabase` - Database errors
- `ErrorTypeCache` - Cache errors
- `ErrorTypeTimeout` - Timeout errors

### Error Structure

All errors follow a consistent structure with rich context:

```go
type AppError struct {
    Code       string                 // Error code from constants
    Message    string                 // Human-readable error message
    Type       ErrorType              // Error category
    HTTPStatus int                    // HTTP status code
    Cause      error                  // Underlying error
    Context    map[string]interface{} // Additional context
    Timestamp  time.Time              // When error occurred
    StackTrace string                 // Stack trace for debugging
    Operation  string                 // Operation being performed
    Resource   string                 // Resource being accessed
}
```

### Usage Examples

#### Creating Errors

```go
// Simple error creation
err := errors.ValidationError("Invalid email format", nil)

// Error with context
err := errors.DatabaseError("Failed to create user", dbErr).
    WithOperation("create_user").
    WithResource("user").
    WithContext("user_id", userID).
    WithContext("email", email)
```

#### Handler Usage

```go
func (h *UserHandler) CreateUser(c echo.Context) error {
    // Validation errors
    if err := h.validator.Struct(requestDto); err != nil {
        return h.HandleError(c, errors.ValidationError("Validation failed", err))
    }

    // Service errors
    user, err := h.userService.Create(c.Request().Context(), &requestDto)
    if err != nil {
        return h.HandleError(c, err) // Error is already wrapped in service
    }

    return h.SuccessResponse(c, "User created successfully", user, nil)
}
```

### Error Response Format

All errors are returned in a consistent format:

```json
{
  "meta": {
    "error_code": "VALIDATION_ERROR",
    "message": "Invalid email format",
    "code": 400
  },
  "data": null
}
```

### Middleware Integration

The error handling system includes middleware for:

1. **Panic Recovery** - Catches panics and converts them to structured errors
2. **Centralized Error Handling** - Processes all errors and returns consistent responses

### Monitoring and Logging

- **Structured Logging**: All errors are logged with context fields
- **Sentry Integration**: Errors are automatically reported to Sentry with context
- **Stack Traces**: Internal errors include stack traces for debugging

## Database Connection Management

The application features a comprehensive database connection management system that provides enterprise-grade reliability, monitoring, and performance.

### Database Features

- **Advanced Connection Pooling** with configurable parameters
- **Health Monitoring** with automatic checks and retry logic
- **Graceful Shutdown** handling
- **Connection Metrics** and monitoring
- **Error Handling** with structured error reporting
- **Automatic Reconnection** on failures

### Database Architecture

The system consists of:

1. **DatabaseManager** (`internal/db/manager.go`) - Core connection management, health monitoring, metrics collection, and retry logic
2. **PostgresDB** (`internal/db/postgres.go`) - Wrapper around GORM with integration to DatabaseManager
3. **Configuration** (`internal/config/config.go`) - Database connection parameters, pool settings, and timeout configurations

### Connection Pooling

The system implements advanced connection pooling with:

- **Configurable Pool Size**: Set maximum open and idle connections
- **Connection Lifecycle Management**: Automatic cleanup of old connections
- **Idle Connection Management**: Efficient handling of unused connections

```go
// Example usage
db := &db.PostgresDB{}
err := db.NewPostgresDB(cfg)
if err != nil {
    log.Fatal("Failed to connect to database:", err)
}

// Get connection metrics
metrics := db.GetMetrics()
log.Printf("Open connections: %d", metrics.OpenConnections)
```

### Database Health Monitoring

Automatic health checks every 30 seconds:

- **Connection Validation**: Ping database to verify connectivity
- **Response Time Tracking**: Monitor query response times
- **Error Tracking**: Log and report connection issues
- **Retry Logic**: Automatic reconnection on failures

```go
// Manual health check
healthStatus := db.HealthCheck()
if !healthStatus.IsHealthy {
    log.Errorf("Database unhealthy: %s", healthStatus.LastError)
}
```

### Connection Metrics

Real-time monitoring of connection statistics:

- **Pool Statistics**: Open, idle, and in-use connections
- **Wait Metrics**: Connection wait times and counts
- **Configuration**: Current pool settings

## Configuration

Set via `.env` (loaded by viper and godotenv):

- **Server**: `APP_ENV`, `APP_NAME`, `APP_VERSION`, `TIMEZONE`, `APP_HTTP_SERVER` (e.g. `:3000`)
- **Database**: `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`, `DATABASE_DEBUG`
- **Database Connection Pool**: `DATABASE_MAX_OPEN_CONNS` (default: 25), `DATABASE_MAX_IDLE_CONNS` (default: 5), `DATABASE_CONN_MAX_LIFETIME` (default: 5m), `DATABASE_CONN_MAX_IDLE_TIME` (default: 1m)
- **Database Timeouts**: `DATABASE_CONNECT_TIMEOUT` (default: 30s), `DATABASE_QUERY_TIMEOUT` (default: 30s)
- **Database Retry**: `DATABASE_RETRY_ATTEMPTS` (default: 3), `DATABASE_RETRY_DELAY` (default: 1s)
- **Database Health**: `DATABASE_HEALTH_TIMEOUT` (default: 5s)
- **Database SSL**: `DATABASE_SSL_MODE` (default: disable), `DATABASE_TIMEZONE` (default: UTC)
- **Cache**: `CACHE_PROVIDER` (default: redis), `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`, `REDIS_DB`, `REDIS_POOL_SIZE`, `REDIS_DIAL_TIMEOUT`, `REDIS_READ_TIMEOUT`, `REDIS_WRITE_TIMEOUT`, `REDIS_POOL_TIMEOUT`, `REDIS_MAX_RETRIES`, `REDIS_MIN_RETRY_BACKOFF`, `REDIS_MAX_RETRY_BACKOFF`
- **Authentication**: `AUTH_PROVIDER`, `KEYCLOAK_URL`, `KEYCLOAK_REALM`, `KEYCLOAK_CLIENT_ID`, `KEYCLOAK_CLIENT_SECRET`, `KEY_CLAIMS`, `KEYCLOAK_REDIRECT_URI`
- **Email**: `EMAIL_PROVIDER` (ses), `AWS_SES_REGION`, `AWS_SES_ACCESS_KEY`, `AWS_SES_SECRET_KEY`
- **Rate Limiting**: `DEFAULT_RATE_LIMIT`, `AUTH_RATE_LIMIT`, `PUBLIC_RATE_LIMIT`, `RATE_LIMIT`, `RATE_LIMIT_DURATION`
- **Observability**: `NEWRELIC_APP_NAME`, `NEWRELIC_LICENSE`, `SENTRY_DSN`
### Database Configuration Parameters

| Parameter                     | Default | Description                        |
| ----------------------------- | ------- | ---------------------------------- |
| `DATABASE_MAX_OPEN_CONNS`     | 25      | Maximum number of open connections |
| `DATABASE_MAX_IDLE_CONNS`     | 5       | Maximum number of idle connections |
| `DATABASE_CONN_MAX_LIFETIME`  | 5m      | Maximum lifetime of a connection   |
| `DATABASE_CONN_MAX_IDLE_TIME` | 1m      | Maximum idle time of a connection  |
| `DATABASE_CONNECT_TIMEOUT`    | 30s     | Connection timeout                 |
| `DATABASE_QUERY_TIMEOUT`      | 30s     | Query timeout                      |
| `DATABASE_HEALTH_TIMEOUT`     | 5s      | Health check timeout               |
| `DATABASE_RETRY_ATTEMPTS`     | 3       | Number of retry attempts           |
| `DATABASE_RETRY_DELAY`        | 1s      | Delay between retries              |

### Rate Limiting

Default limits are configurable via env. Middleware is applied globally in `router.go`.

## Docker

### Build and Run

```bash
# Build image
docker build -t golang-boilerplate .

# Run container (ensure APP_HTTP_SERVER is set to :3000 in container env)
docker run -p 3000:3000 --env-file .env golang-boilerplate
```

### Services Included

- **PostgreSQL**: Database (5432)
- **Redis**: Cache (6379)
  - App service is commented out in `docker-compose.yml`. Run the app locally with `make up` or create your own app service.

## Database Monitoring and Troubleshooting

### Metrics to Monitor

1. **Connection Pool Utilization**
   - Open connections vs. max connections
   - Idle connections
   - Wait times

2. **Health Status**
   - Connection health
   - Response times
   - Error rates

3. **Performance Metrics**
   - Query response times
   - Connection establishment time
   - Retry attempts

### Recommended Alerts

- Connection pool utilization > 80%
- Health check failures
- Response times > 1 second
- Retry attempts > 2

### Common Issues and Solutions

1. **Connection Pool Exhaustion**
   - Increase `DATABASE_MAX_OPEN_CONNS`
   - Check for connection leaks
   - Monitor connection usage patterns

2. **Slow Queries**
   - Check `DATABASE_QUERY_TIMEOUT`
   - Monitor query performance
   - Optimize database queries

3. **Connection Failures**
   - Check network connectivity
   - Verify database credentials
   - Monitor database server health

4. **High Response Times**
   - Check connection pool settings
   - Monitor database server performance
   - Optimize query performance

### Debugging

1. **Enable Debug Logging**

   ```bash
   DATABASE_DEBUG=true
   LOG_LEVEL=debug
   ```

2. **Check Health Endpoints**

   ```bash
   curl http://localhost:3000/api/v1/health/database
   curl http://localhost:3000/api/v1/health/metrics
   ```

3. **Monitor Logs**
   - Check application logs for database errors
   - Monitor Sentry for error reports
   - Use database monitoring tools

### Performance Considerations

- **Small Applications**: 5-10 connections
- **Medium Applications**: 10-25 connections
- **Large Applications**: 25-50 connections
- **High Traffic**: 50+ connections

## Database Migrations

Migrations are managed with [Atlas](https://atlasgo.io/). SQL files live under `cmd/migrations/sql/`; the checksum file `cmd/migrations/sql/atlas.sum` must stay in sync—run `make migrate-hash` after you add or edit migration files.

Configuration:

- **`atlas.hcl`** — defines the `gorm` env: loads schema from `internal/models` via [atlas-provider-gorm](https://github.com/ariga/atlas-provider-gorm), writes diffs into `cmd/migrations/sql`, and uses a dev database for planning.
- **`Makefile`** — sets `MIGRATION_DIR` (`file://cmd/migrations/sql`), builds `DB_DSN` from `cmd/server/.env`, and sets `DB_DEV_URL` (default `docker://postgres/18/dev`) for commands that need a dev instance.

Common commands:

```bash
# Status and apply
make migrate-status
make migrate-up
make migrate-up-preview

# Roll back (needs a working Docker setup for the default dev URL)
make migrate-down
make migrate-down-preview

# New empty migration file
make migrate-create name=add_users

# Generate a migration from GORM models (run from repo root; set name=...)
make migrate-generate name=describe_your_change

# After hand-editing SQL migrations
make migrate-hash
```

For full Atlas CLI options, see the [Atlas documentation](https://atlasgo.io/docs).

## Production Deployment

1. Build and push Docker image or deploy the binary built from `cmd/server`.
2. Set `APP_ENV=production` and all required env vars.
3. Apply database migrations before or during rollout (for example `atlas migrate apply --dir "file://cmd/migrations/sql" --url "$DATABASE_URL"`, or your orchestrator’s equivalent).
4. Expose the port configured by `APP_HTTP_SERVER` (e.g. `:3000`).

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Run tests and linting
6. Submit a pull request

## License

This project is licensed under the MIT License.
