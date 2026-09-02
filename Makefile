# Load DB settings for migration-* and any recipe that uses DB_DSN (not gated on a repo-root .env).
ifneq (,$(wildcard cmd/server/.env))
    include cmd/server/.env
    export $(shell sed -n 's/^\([A-Za-z_][A-Za-z0-9_]*\)=.*/\1/p' cmd/server/.env)
endif

.PHONY: lint lint-fix mocks tests test-services test-utils test-handlers test-repositories test-coverage test-coverage-html test-race test-verbose test-specific test-specific-verbose test-specific-coverage docker-build security-fs security-image security
DB_DSN ?= postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable$(if $(POSTGRES_SCHEMA),&search_path=$(POSTGRES_SCHEMA))
MIGRATION_DIR ?= file://cmd/migrations/sql
DB_DEV_URL ?= docker://postgres/18/dev

bootstrap: container-up migrate-up up

.PHONY: lint
lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.2 run ./... --config .golangci.yml

.PHONY: lint-fix
lint-fix:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.2 run ./... --config .golangci.yml --fix

mocks:
	mockery --case snake --dir ./repositories --all --output ./mocks/repositories
	mockery --case snake --dir ./adapters --all --output ./mocks/adapters

build:
	@cd cmd/${cmd} && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ${service_name} .

dep:
	@go mod tidy

docker-build:
	DOCKER_BUILDKIT=1 docker build -f Dockerfile -t golang-boilerplate:latest .

# Match .github/workflows/ci.yml security job (requires: brew install trivy / apt install trivy)
# Uses repo-root trivy.yaml (skip-files for local gitignored secrets).
TRIVY_FLAGS = --format table --exit-code 1 --ignore-unfixed --vuln-type os,library --severity CRITICAL,HIGH --config trivy.yaml

security-fs:
	trivy fs . $(TRIVY_FLAGS) --scanners vuln,secret,misconfig

security-image: docker-build
	trivy image golang-boilerplate:latest $(TRIVY_FLAGS)

security: security-fs security-image

container-up:
	set -a && source cmd/server/.env && set +a && docker compose up -d

container-down:
	docker compose down

otel-up:
	docker compose up -d otel-collector jaeger

otel-down:
	docker compose stop otel-collector jaeger

up:
	cd cmd/server && go run main.go

major-version-update:
	go get -u -t ./...

minor-version-update:
	go get -u ./...

.PHONY: migrate-inspect
migrate-inspect:
	atlas schema inspect --url "$(DB_DSN)"

.PHONY: migrate-up
migrate-up:
	atlas migrate apply --dir "$(MIGRATION_DIR)" --url "$(DB_DSN)"

.PHONY: migrate-up-preview
migrate-up-preview:
	atlas migrate apply --dir "$(MIGRATION_DIR)" --url "$(DB_DSN)" --dry-run

.PHONY: migrate-down
migrate-down:
	atlas migrate down --dir "$(MIGRATION_DIR)" --url "$(DB_DSN)" --dev-url "$(DB_DEV_URL)"

.PHONY: migrate-down-preview
migrate-down-preview:
	atlas migrate down --dir "$(MIGRATION_DIR)" --url "$(DB_DSN)" --dev-url "$(DB_DEV_URL)" --dry-run

.PHONY: migrate-create
migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "❌ Missing migration name. Usage: make migrate_create name=<migration_name>"; \
		exit 1; \
	fi
	atlas migrate new $(name)

.PHONY: migrate-reset
migrate-reset:
	atlas schema clean --url "$(DB_DSN)"

.PHONY: migrate-generate
migrate-generate:
	atlas migrate diff --env gorm ${name}

.PHONY: migrate-status
migrate-status:
	atlas migrate status --dir "$(MIGRATION_DIR)" --url "$(DB_DSN)"

.PHONY: migrate-hash
migrate-hash:
	atlas migrate hash

format:
	go fmt ./...

swagger-load:
	swag init \
		-g main.go \
		-d ./cmd/server,./internal/handlers,./internal/middlewares,./internal/services,./internal/repositories,./internal/models,./internal/utils,./internal/config,./internal/constants,./internal/dtos,./internal/logger,./internal/db \
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