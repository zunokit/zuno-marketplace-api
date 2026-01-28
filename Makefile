.PHONY: help dev migrate test build

# ============================================================
# Configuration
# ============================================================

# Load environment variables from .env.development (if exists)
-include .env.development
export

SHELL := /bin/bash
GOPATH := $(shell go env GOPATH)

# Version information for Sentry releases
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || echo "unknown")
LDFLAGS = -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Migrate tool path (cross-platform)
MIGRATE := $(shell which migrate)

# Database configuration (local Docker)
DB_HOST ?= localhost
DB_PORT ?= 5433
DB_USER ?= postgres
DB_PASSWORD ?= postgres
DB_NAME ?= nft_marketplace
DB_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

# DATABASE_URL for serverless (Neon) - loaded from .env.development or environment
DATABASE_URL ?=

.DEFAULT_GOAL := help

# ============================================================
# Help
# ============================================================

help: ## Show help
	@echo ============================================================
	@echo   Zuno NFT Marketplace - Quick Start
	@echo ============================================================
	@echo
	@echo Development:
	@echo   make dev         - Start all services with Air
	@echo   make dev-stop    - Stop Air services
	@echo   make dev-logs    - View combined logs
	@echo   make dev-logs-all - View all logs separately
	@echo
	@echo Common Commands:
	@echo   make test          - Run tests
	@echo   make build         - Build all services
	@echo   make build-version - Show build version info
	@echo   make migrate       - Run migrations (Docker)
	@echo   make migrate-dev - Run migrations (Neon)
	@echo   make proto         - Generate protobuf
	@echo   make lint          - Run linter
	@echo   make format        - Format code
	@echo
	@echo Tools:
	@echo   make install-tools - Install dev tools
	@echo ============================================================

# ============================================================
# Air Hot-Reload (Development Mode)
# ============================================================

dev: ## Start Air development environment (serverless infra)
	@echo ============================================================
	@echo   Starting Air Development Environment...
	@echo ============================================================
	@./scripts/dev.sh all

dev-auth: ## Start auth-service with Air
	@./scripts/dev.sh auth

dev-user: ## Start user-service with Air
	@./scripts/dev.sh user

dev-wallet: ## Start wallet-service with Air
	@./scripts/dev.sh wallet

dev-gateway: ## Start graphql-gateway with Air
	@./scripts/dev.sh gateway

dev-stop: ## Stop Air services
	@./scripts/stop-air.sh

dev-logs: ## View Air logs (combined)
	@tail -f logs/all-services.log 2>/dev/null || echo "No logs found. Start services first."

dev-logs-all: ## View all Air logs separately
	@tail -f logs/*.log 2>/dev/null || echo "No logs found. Start services first."



# ============================================================
# Database Migrations production
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
	@$(FIND_CMD) || echo "No migrations found"

migrate-create: ## Create new migration (make migrate-create NAME=add_feature)
	@echo Creating migration: $(NAME)
	$(MIGRATE) create -ext sql -dir db/migrations -seq $(NAME)

migrate-down: ## Rollback last migration
	@echo Rolling back last migration...
	$(MIGRATE) -path db/migrations -database "$(DB_URL)" down 1
	@echo Rollback complete!

# ============================================================
# Serverless Database Migrations (Neon Development Mode)
# ============================================================

migrate-dev: ## Run migrations on Neon serverless database
	@echo Running migrations on Neon serverless... && \
	[ -n "$(DATABASE_URL)" ] || (echo "Error: DATABASE_URL not set. Please set in .env.development or environment." && exit 1) && \
	$(MIGRATE) -path db/migrations -database "$(DATABASE_URL)" up && \
	echo "Migrations complete!"

migrate-dev-status: ## Show Neon migration status
	@echo Neon migration status: && \
	[ -n "$(DATABASE_URL)" ] || (echo "Error: DATABASE_URL not set." && exit 1) && \
	$(MIGRATE) -path db/migrations -database "$(DATABASE_URL)" version

migrate-dev-down: ## Rollback last Neon migration
	@echo Rolling back last Neon migration... && \
	[ -n "$(DATABASE_URL)" ] || (echo "Error: DATABASE_URL not set." && exit 1) && \
	$(MIGRATE) -path db/migrations -database "$(DATABASE_URL)" down 1 && \
	echo "Rollback complete!"

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

# Detect OS for Windows compatibility
ifeq ($(OS),Windows_NT)
	MKDIR_CMD = if not exist build mkdir build
	RM_CMD = if exist build rmdir /s /q build
	MKDIR_PROTO = if not exist shared\proto\pb mkdir shared\proto\pb
	FIND_CMD = dir /b db\migrations\*.up.sql 2>nul
else
	MKDIR_CMD = mkdir -p build
	RM_CMD = rm -rf build
	MKDIR_PROTO = mkdir -p shared/proto/pb
	FIND_CMD = find db/migrations -name "*.up.sql" 2>/dev/null || ls -1 db/migrations/*.up.sql 2>/dev/null
endif

build-auth: ## Build auth service
	@echo Building auth-service...
	@$(MKDIR_CMD)
	cd services/auth-service/cmd && go build $(LDFLAGS) -o ../../build/auth-service$(if $(filter $(OS),Windows_NT),.exe,) .

build-user: ## Build user service
	@echo Building user-service...
	@$(MKDIR_CMD)
	cd services/user-service/cmd && go build $(LDFLAGS) -o ../../build/user-service$(if $(filter $(OS),Windows_NT),.exe,) .

build-wallet: ## Build wallet service
	@echo Building wallet-service...
	@$(MKDIR_CMD)
	cd services/wallet-service/cmd && go build $(LDFLAGS) -o ../../build/wallet-service$(if $(filter $(OS),Windows_NT),.exe,) .

build-collection: ## Build collection service
	@echo Building collection-service...
	@$(MKDIR_CMD)
	cd services/collection-service/cmd && go build $(LDFLAGS) -o ../../../build/collection-service$(if $(filter $(OS),Windows_NT),.exe,) .

build-gateway: ## Build graphql gateway
	@echo Building graphql-gateway...
	@$(MKDIR_CMD)
	cd services/graphql-gateway/cmd && go build $(LDFLAGS) -o ../../../build/graphql-gateway$(if $(filter $(OS),Windows_NT),.exe,) .

build: build-auth build-user build-wallet build-collection build-gateway ## Build all services
	@echo Build complete! Binaries in ./build/
	@echo Version: $(VERSION) BuildTime: $(BUILD_TIME)

build-version: ## Show build version info
	@echo Version: $(VERSION)
	@echo BuildTime: $(BUILD_TIME)
	@echo LDFLAGS: $(LDFLAGS)

clean: ## Clean build artifacts
	@echo Cleaning build artifacts...
	@$(RM_CMD)
	@$(MKDIR_CMD)

# ============================================================
# Code Generation
# ============================================================

proto: ## Generate protobuf code
	@echo Generating protobuf code...
	@$(MKDIR_PROTO)
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
	go install github.com/air-verse/air@latest
	@echo Tools installed!
