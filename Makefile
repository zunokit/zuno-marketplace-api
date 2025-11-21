.PHONY: help dev migrate test build

# ============================================================
# Configuration
# ============================================================

SHELL := /bin/bash
GOPATH := $(shell go env GOPATH)

# Migrate tool path (Windows/Unix compatible)
ifeq ($(OS),Windows_NT)
	MIGRATE := "$(GOPATH)\bin\migrate.exe"
else
	MIGRATE := $(GOPATH)/bin/migrate
endif

# Database configuration
DB_HOST ?= localhost
DB_PORT ?= 5433
DB_USER ?= postgres
DB_PASSWORD ?= postgres
DB_NAME ?= nft_marketplace
DB_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

.DEFAULT_GOAL := help

# ============================================================
# Help
# ============================================================

help: ## Show help
	@echo ============================================================
	@echo   Zuno NFT Marketplace - Quick Start
	@echo ============================================================
	@echo
	@echo OPTION 1 - Docker Compose - Recommended:
	@echo   make dev           - Start all services
	@echo   make dev-stop      - Stop all services
	@echo   make dev-logs      - View logs
	@echo
	@echo OPTION 2 - Tilt/Kubernetes - Advanced:
	@echo   See TILT.md for instructions
	@echo
	@echo Common Commands:
	@echo   make test          - Run tests
	@echo   make build         - Build services
	@echo   make migrate       - Run migrations
	@echo   make proto         - Generate protobuf
	@echo   make lint          - Run linter
	@echo   make format        - Format code
	@echo
	@echo Tools:
	@echo   make install-tools - Install dev tools
	@echo ============================================================

# ============================================================
# Docker Compose (Primary Development Method)
# ============================================================

dev: ## Start Docker Compose (one command does everything)
	@echo ============================================================
	@echo   Starting Docker Compose environment...
	@echo ============================================================
	@docker compose down -v 2>nul >nul || true
	@echo [1/3] Starting services...
	docker compose up -d
	@echo [2/3] Waiting for PostgreSQL to be ready...
	@timeout /t 8 /nobreak > nul 2>&1
	@echo [3/3] Running database migrations...
	@$(MAKE) migrate
	@echo ============================================================
	@echo   Ready!
	@echo ============================================================
	@echo GraphQL Playground: http://localhost:8081/graphql
	@echo PostgreSQL:         localhost:5433
	@echo RabbitMQ UI:        http://localhost:15672 (guest/guest)
	@echo
	@echo View logs:  make dev-logs
	@echo Stop:       make dev-stop
	@echo ============================================================

dev-stop: ## Stop Docker Compose
	@echo Stopping Docker Compose...
	docker compose down
	@echo Stopped!

dev-logs: ## View Docker Compose logs
	docker compose logs -f

dev-clean: ## Stop and remove all data
	@echo Cleaning up...
	docker compose down -v
	@echo Done!

# ============================================================
# Database Migrations
# ============================================================

migrate: ## Run database migrations
	@echo Running migrations...
	$(MIGRATE) -path db/migrations -database "$(DB_URL)" up
	@echo Migrations complete!

migrate-status: ## Show migration status
	@echo Migration status:
	-@$(MIGRATE) -path db/migrations -database "$(DB_URL)" version
	@echo
	@echo Available migrations:
	@dir /b db\migrations\*.up.sql 2>nul || ls -1 db/migrations/*.up.sql 2>/dev/null

migrate-create: ## Create new migration (make migrate-create NAME=add_feature)
	@echo Creating migration: $(NAME)
	$(MIGRATE) create -ext sql -dir db/migrations -seq $(NAME)

migrate-down: ## Rollback last migration
	@echo Rolling back last migration...
	$(MIGRATE) -path db/migrations -database "$(DB_URL)" down 1
	@echo Rollback complete!

# ============================================================
# Testing
# ============================================================

test: ## Run all tests
	@echo Running tests...
	go test ./...

test-coverage: ## Run tests with coverage report
	@echo Running tests with coverage...
	go test -cover -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo Coverage report: coverage.html

test-verbose: ## Run tests with verbose output
	go test -v ./...

# ============================================================
# Building
# ============================================================

build: ## Build all services
	@echo Building all services...
	@if not exist "build" mkdir build
	cd services/auth-service/cmd && go build -o ../../../build/auth-service.exe .
	cd services/user-service/cmd && go build -o ../../../build/user-service.exe .
	cd services/wallet-service/cmd && go build -o ../../../build/wallet-service.exe .
	cd services/collection-service/cmd && go build -o ../../../build/collection-service.exe .
	cd services/graphql-gateway && go build -o ../build/graphql-gateway.exe .
	@echo Build complete! Binaries in ./build/

clean: ## Clean build artifacts
	@echo Cleaning build artifacts...
	@if exist "build" rmdir /s /q build
	@if not exist "build" mkdir build

# ============================================================
# Code Generation
# ============================================================

proto: ## Generate protobuf code
	@echo Generating protobuf code...
	@mkdir -p shared/proto/pb
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/*.proto
	@if exist proto\*.pb.go move /Y proto\*.pb.go shared\proto\pb\
	@echo Protobuf generation complete!

# ============================================================
# Code Quality
# ============================================================

lint: ## Run linter
	@echo Running linter...
	golangci-lint run ./...

format: ## Format code
	@echo Formatting code...
	gofmt -w -s .
	goimports -w .

# ============================================================
# Tools
# ============================================================

install-tools: ## Install development tools
	@echo Installing development tools...
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo Tools installed!
