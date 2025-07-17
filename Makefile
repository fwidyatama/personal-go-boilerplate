# Variables
APP_NAME := go-boilerplate
APP_VERSION := 1.0.0
DOCKER_IMAGE := $(APP_NAME):$(APP_VERSION)
DOCKER_REGISTRY := localhost:5000

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOMOD := $(GOCMD) mod
BINARY_NAME := main
BINARY_UNIX := $(BINARY_NAME)_unix


# DB Variables
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_USER ?= postgres
DB_PASSWORD ?= password
DB_NAME ?= microservice_db
DB_SSLMODE ?= disable

DB_URL = "host=$(DB_HOST) port=$(DB_PORT) user=$(DB_USER) password=$(DB_PASSWORD) dbname=$(DB_NAME) sslmode=$(DB_SSLMODE)"


# Build flags
LDFLAGS := -ldflags "-X main.Version=$(APP_VERSION) -X main.BuildTime=$(shell date -u '+%Y-%m-%d_%H:%M:%S')"

.PHONY: all build clean test coverage deps tidy docker-build docker-compose-up docker-compose-down docker-logs migrate-up migrate-down migrate-status migrate-create run dev install-tools swagger lint fmt vet security bench build-all health load-test help

# Default target
all: clean deps test build

# Build the application
build:
	@echo "Building $(APP_NAME)..."
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) ./cmd
	@echo "Build complete!"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
	@echo "Clean complete!"

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	$(GOMOD) tidy

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE) .
	@echo "Docker image built: $(DOCKER_IMAGE)"

# Start all services with Docker Compose
docker-compose-up:
	@echo "Starting all services..."
	docker compose up --build -d

# Stop all services with Docker Compose
docker-compose-down:
	@echo "Stopping all services..."
	docker compose down

# View logs
docker-logs:
	docker compose logs -f

# Database migrations
migrate-up:
	@echo "Running database migrations..."
	goose -dir migrations postgres $(DB_URL) up

migrate-down:
	@echo "Rolling back database migrations..."
	goose -dir migrations postgres $(DB_URL) down

migrate-status:
	@echo "Checking migration status..."
	goose -dir migrations postgres $(DB_URL) status

migrate-create:
	@echo "Creating new migration file..."
	@read -p "Enter migration name: " name; \
	goose -dir migrations create $$name sql

# Run the application locally
run:
	@echo "Running application..."
	$(GOBUILD) -o $(BINARY_NAME) ./cmd
	./$(BINARY_NAME)

# Run with hot reload (requires air)
dev:
	@echo "Running in development mode..."
	air

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/air-verse/air@latest
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install github.com/swaggo/swag/cmd/swag@latest

# Generate Swagger documentation
swagger:
	@echo "Generating OpenAPI types and server/client code with oapi-codegen..."
	oapi-codegen --config codegen.yaml ./docs/openapi.yaml

# Lint code
lint:
	@echo "Linting code..."
	golangci-lint run

# Format code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

# Vet code
vet:
	@echo "Vetting code..."
	$(GOCMD) vet ./...

# Security scan
security:
	@echo "Running security scan..."
	gosec ./...

# Performance benchmark
bench:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

# Build for multiple platforms
build-all: clean
	@echo "Building for multiple platforms..."
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_UNIX)_linux_amd64 ./cmd
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_UNIX)_windows_amd64.exe ./cmd
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_UNIX)_darwin_amd64 ./cmd

# Health check
health:
	@echo "Checking application health..."
	curl -f http://localhost:8080/api/v1/health || echo "Application is not healthy"

# Load test (requires hey)
load-test:
	@echo "Running load test..."
	hey -n 1000 -c 10 http://localhost:8080/api/v1/health

# Show help
help:
	@echo "Available commands:"
	@echo "  build          - Build the application"
	@echo "  clean          - Clean build artifacts"
	@echo "  test           - Run tests"
	@echo "  coverage       - Run tests with coverage"
	@echo "  deps           - Download dependencies"
	@echo "  tidy           - Tidy dependencies"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-compose-up   - Start all services"
	@echo "  docker-compose-down - Stop all services"
	@echo "  docker-logs    - View logs"
	@echo "  migrate-up     - Run database migrations"
	@echo "  migrate-down   - Rollback migrations"
	@echo "  migrate-status - Check migration status"
	@echo "  migrate-create - Create new migration"
	@echo "  run            - Run application locally"
	@echo "  dev            - Run with hot reload"
	@echo "  install-tools  - Install development tools"
	@echo "  swagger        - Generate Swagger docs"
	@echo "  lint           - Lint code"
	@echo "  fmt            - Format code"
	@echo "  vet            - Vet code"
	@echo "  security       - Security scan"
	@echo "  bench          - Run benchmarks"
	@echo "  build-all      - Build for multiple platforms"
	@echo "  health         - Health check"
	@echo "  load-test      - Load test"
	@echo "  help           - Show this help" 

