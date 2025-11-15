.PHONY: help dev test build clean proto docker-up docker-down tilt-up tilt-down migrate lint format check

# Default target
.DEFAULT_GOAL := help

# Colors for output (Windows compatible)
BLUE :=
GREEN :=
YELLOW :=
RED :=
NC :=

##@ General

help: ## Display this help message
	@echo "=============================================================="
	@echo "  Zuno NFT Marketplace API - Development Commands"
	@echo "=============================================================="
	@echo.
	@echo Usage: make [target]
	@echo.
	@echo Targets:
	@echo   dev              Start development environment with docker-compose
	@echo   dev-stop         Stop development environment
	@echo   test             Run all tests
	@echo   build            Build all services
	@echo   proto            Generate protobuf code
	@echo   lint             Run linter
	@echo   format           Format code
	@echo   docker-up        Start Docker services
	@echo   tilt-up          Start Tilt (hot reload)
	@echo   ci               Run CI pipeline

##@ Development

dev: ## Start development environment
	@echo Starting development environment...
	docker compose up -d
	@echo Services started!
	@echo PostgreSQL: localhost:5432
	@echo Redis: localhost:6379
	@echo RabbitMQ: localhost:5672 (UI: http://localhost:15672)

dev-stop: ## Stop development environment
	@echo Stopping development environment...
	docker compose down

dev-logs: ## Follow logs
	docker compose logs -f

dev-clean: ## Clean environment (remove volumes)
	@echo Cleaning development environment...
	docker compose down -v

##@ Tilt (Kubernetes)

tilt-up: ## Start Tilt
	@echo Starting Tilt...
	@echo Tilt UI: http://localhost:10350
	tilt up

tilt-down: ## Stop Tilt
	tilt down

##@ Testing

test: ## Run all tests
	@echo Running tests...
	go test ./...

test-coverage: ## Run tests with coverage
	@echo Running tests with coverage...
	go test -cover -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo Coverage report: coverage.html

test-verbose: ## Run tests verbose
	go test -v ./...

test-service: ## Run tests for specific service (make test-service SERVICE=auth)
	@echo Running tests for $(SERVICE)-service...
	cd services/$(SERVICE)-service && go test -v ./...

##@ Building

build: ## Build all services
	@echo Building all services...
	@if not exist "build" mkdir build
	cd services/auth-service/cmd && go build -o ../../../build/auth-service.exe .
	cd services/user-service/cmd && go build -o ../../../build/user-service.exe .
	cd services/wallet-service/cmd && go build -o ../../../build/wallet-service.exe .
	cd services/graphql-gateway && go build -o ../build/graphql-gateway.exe .
	@echo Build complete! Binaries in ./build/

build-service: ## Build specific service (make build-service SERVICE=auth)
	@echo Building $(SERVICE)-service...
	cd services/$(SERVICE)-service/cmd && go build -o ../../../build/$(SERVICE)-service.exe .

clean: ## Clean build artifacts
	@echo Cleaning build artifacts...
	@if exist "build" rmdir /s /q build
	@if not exist "build" mkdir build

##@ Code Generation

proto: generate-proto ## Generate protobuf code

generate-proto: ## Generate protobuf code
	@echo Generating protobuf code...
	@if not exist "shared\proto" mkdir shared\proto
	protoc --go_out=shared/proto --go_opt=paths=source_relative --go-grpc_out=shared/proto --go-grpc_opt=paths=source_relative proto/*.proto
	@echo Protobuf generation complete!

##@ Code Quality

lint: ## Run linter
	@echo Running linter...
	golangci-lint run ./...

format: ## Format code
	@echo Formatting code...
	gofmt -w -s .
	goimports -w .

check: lint test ## Run linter and tests

vet: ## Run go vet
	@echo Running go vet...
	go vet ./...

##@ Database

db-reset: ## Reset database
	@echo Resetting database...
	docker compose down postgres
	docker volume rm zuno-marketplace-api_postgres_data
	docker compose up -d postgres

##@ Docker

docker-build: ## Build Docker images
	@echo Building Docker images...
	docker build -f infra/development/docker/auth-service.Dockerfile -t nft-auth-service .
	docker build -f infra/development/docker/user-service.Dockerfile -t nft-user-service .
	docker build -f infra/development/docker/wallet-service.Dockerfile -t nft-wallet-service .
	docker build -f infra/development/docker/graph-gateway.Dockerfile -t nft-graphql-gateway .

docker-up: ## Start Docker services
	docker compose up -d

docker-down: ## Stop Docker services
	docker compose down

docker-logs: ## Follow Docker logs
	docker compose logs -f

docker-ps: ## Show containers
	docker compose ps

##@ Dependencies

deps: ## Download dependencies
	@echo Downloading dependencies...
	go mod download

deps-tidy: ## Tidy dependencies
	@echo Tidying dependencies...
	go mod tidy

deps-update: ## Update dependencies
	@echo Updating dependencies...
	go get -u ./...
	go mod tidy

##@ Tools

install-tools: ## Install dev tools
	@echo Installing development tools...
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest

##@ CI/CD

ci: lint test build ## Run CI pipeline
	@echo CI pipeline complete!

##@ Information

info: ## Show project info
	@echo ============================================================
	@echo   Zuno NFT Marketplace API - Project Info
	@echo ============================================================
	@echo.
	@echo Version: 0.1.0
	@go version
	@echo.
	@echo Services:
	@echo   - auth-service (gRPC: 50051)
	@echo   - user-service (gRPC: 50052)
	@echo   - wallet-service (gRPC: 50053)
	@echo   - graphql-gateway (HTTP: 8081)
	@echo.
	@echo Infrastructure:
	@echo   - PostgreSQL: 5432
	@echo   - Redis: 6379
	@echo   - RabbitMQ: 5672, 15672
