.PHONY: lint mocks tests test-services test-utils test-handlers test-repositories test-coverage test-coverage-html test-race test-verbose test-specific test-specific-verbose test-specific-coverage

bootstrap: container-up migration-up up

lint:
	golangci-lint run

mocks:
	mockery --case snake --dir ./repositories --all --output ./mocks/repositories
	mockery --case snake --dir ./adapters --all --output ./mocks/adapters


build:
	@cd cmd/${cmd} && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ${service_name} .

dep:
	@go mod tidy

container-up:
	set -a && source cmd/server/.env && set +a && docker compose up -d

container-down:
	docker compose down

up:
	cd cmd/server && go run main.go

major-version-update:
	go get -u -t ./...

minor-version-update:
	go get -u ./...

migration-status:
	set -a && source cmd/migrations/.env && set +a && goose --dir ./cmd/migrations/sql status

migration-up:
	set -a && source cmd/migrations/.env && set +a && goose --dir ./cmd/migrations/sql up

migration-create:
	set -a && source cmd/migrations/.env && set +a && goose --dir ./cmd/migrations/sql create ${name} sql

migration-down:
	set -a && source cmd/migrations/.env && set +a && goose --dir ./cmd/migrations/sql down

seed-up:
	set -a && source cmd/migrations/.env && set +a && goose --dir ./cmd/migrations/seed up

seed-down:
	set -a && source cmd/migrations/.env && set +a && goose --dir ./cmd/migrations/seed down

format:
	go fmt ./...

swagger-load:
	swag init \
		-g main.go \
		-d ./cmd/server,./internal/handlers,./internal/middlewares,./internal/services,./internal/repositories,./internal/models,./internal/utils,./internal/config,./internal/constants,./internal/integration,./internal/dtos,./internal/logger,./internal/db,./internal/cache \
		--output ./docs

# Test targets
tests:
	@echo "Running all tests with coverage and race detection..."
	@go test -v -cover -race -timeout 300s -count=1 ./...

test-services:
	@echo "Running service layer tests..."
	@go test -v -cover ./internal/services

test-utils:
	@echo "Running utility tests..."
	@go test -v -cover ./internal/utils

test-handlers:
	@echo "Running handler tests..."
	@go test -v -cover ./internal/handlers

test-repositories:
	@echo "Running repository tests..."
	@go test -v -cover ./internal/repositories

test-coverage:
	@echo "Running tests with coverage report..."
	@go test -cover ./...

test-coverage-html:
	@echo "Generating HTML coverage report..."
	@go test -coverprofile=coverage.out ./... || true
	@if [ -f coverage.out ]; then \
		go tool cover -html=coverage.out -o coverage.html; \
		echo ""; \
		echo "✓ Coverage report generated successfully!"; \
		echo "  File: coverage.html ($$(pwd)/coverage.html)"; \
		echo "  Open it in your browser to view the report."; \
	else \
		echo ""; \
		echo "✗ Error: coverage.out was not generated."; \
		echo "  Some tests may have failed before coverage data could be collected."; \
		exit 1; \
	fi

test-race:
	@echo "Running tests with race detection..."
	@go test -race -timeout 300s ./...

test-verbose:
	@echo "Running tests with verbose output..."
	@go test -v ./...

test-specific:
	@echo "Running specific test: $(TEST)"
	@if [ -z "$(TEST)" ]; then \
		echo "Usage: make test-specific TEST=TestUserService_Create"; \
		exit 1; \
	fi
	@go test -run $(TEST) ./internal/services

test-specific-verbose:
	@echo "Running specific test with verbose output: $(TEST)"
	@if [ -z "$(TEST)" ]; then \
		echo "Usage: make test-specific-verbose TEST=TestUserService_Create"; \
		exit 1; \
	fi
	@go test -v -run $(TEST) ./internal/services

test-specific-coverage:
	@echo "Running specific test with coverage: $(TEST)"
	@if [ -z "$(TEST)" ]; then \
		echo "Usage: make test-specific-coverage TEST=TestUserService_Create"; \
		exit 1; \
	fi
	@go test -cover -run $(TEST) ./internal/services