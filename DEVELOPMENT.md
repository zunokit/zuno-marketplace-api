# Development Guide

Complete guide for developing the Zuno NFT Marketplace API.

## 📋 Table of Contents

- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Development Tools](#development-tools)
- [Testing](#testing)
- [Code Quality](#code-quality)
- [CI/CD](#cicd)
- [Troubleshooting](#troubleshooting)

## 🚀 Getting Started

### Prerequisites

**Required:**
- Go 1.21+ ([Download](https://go.dev/dl/))
- Docker Desktop ([Download](https://www.docker.com/products/docker-desktop))
- Git

**Optional but Recommended:**
- Tilt ([Install](https://docs.tilt.dev/install.html)) - For hot reload
- Make - For build automation
- golangci-lint - For linting
- VS Code with Go extension

### Initial Setup

```bash
# 1. Clone repository
git clone https://github.com/ZunoKit/zuno-marketplace-api.git
cd zuno-marketplace-api

# 2. Copy environment file
cp .env.example .env

# 3. Install development tools
make install-tools

# 4. Download dependencies
make deps

# 5. Start infrastructure
make dev
```

## 🔄 Development Workflow

### Option 1: Docker Compose (Recommended for Backend Only)

**Start development environment:**
```bash
make dev
# or
docker compose up -d
```

**Services available:**
- PostgreSQL: `localhost:5432`
- Redis: `localhost:6379`
- RabbitMQ: `localhost:5672`, UI: `http://localhost:15672`

**Stop environment:**
```bash
make dev-stop
```

### Option 2: Tilt (Recommended for Full Stack)

Tilt provides hot reload for all services with Kubernetes.

**Prerequisites:**
- Docker Desktop with Kubernetes enabled
- Tilt installed

**Start Tilt:**
```bash
make tilt-up
# or
tilt up
```

**Tilt UI:** http://localhost:10350

**Features:**
- 🔥 Hot reload on code changes
- 📊 Real-time logs for all services
- 🎯 Resource status dashboard
- 🔧 Manual trigger for builds

**Stop Tilt:**
```bash
make tilt-down
```

## 🛠️ Development Tools

### Makefile Commands

```bash
# Show all available commands
make help

# Development
make dev              # Start docker compose
make dev-stop         # Stop docker compose
make dev-logs         # Follow logs
make dev-clean        # Clean volumes

# Testing
make test             # Run all tests
make test-coverage    # Run with coverage report
make test-service SERVICE=auth  # Test specific service

# Building
make build            # Build all services
make build-service SERVICE=auth  # Build specific service
make clean            # Clean build artifacts

# Code Generation
make proto            # Generate protobuf code

# Code Quality
make lint             # Run linter
make format           # Format code
make vet              # Run go vet
make check            # Run lint + test

# Docker
make docker-build     # Build all Docker images
make docker-up        # Start Docker services
make docker-down      # Stop Docker services

# Database
make db-reset         # Reset database

# CI/CD
make ci               # Run CI pipeline locally
```

### VS Code Setup

**Recommended extensions:**
```json
{
  "recommendations": [
    "golang.go",
    "ms-vscode-remote.remote-containers",
    "ms-azuretools.vscode-docker",
    "ms-kubernetes-tools.vscode-kubernetes-tools",
    "GraphQL.vscode-graphql"
  ]
}
```

**Settings (`.vscode/settings.json`):**
```json
{
  "go.testFlags": ["-v"],
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "workspace",
  "editor.formatOnSave": true,
  "[go]": {
    "editor.defaultFormatter": "golang.go"
  }
}
```

## 🧪 Testing

### Test Structure

```
services/auth-service/
├── test/
│   ├── unit/          # Unit tests (fast, no external deps)
│   └── integration/   # Integration tests (with DB, Redis)
```

### Running Tests

```bash
# All tests
go test ./...

# Specific service
cd services/auth-service
go test ./...

# With coverage
make test-coverage

# Verbose
go test -v ./...

# Watch mode (requires gotestsum)
make test-watch
```

### Writing Tests

**Unit Test Example:**
```go
package service_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestAuthService_ValidateNonce(t *testing.T) {
    // Arrange
    nonce := "test-nonce-123"

    // Act
    result := ValidateNonce(nonce)

    // Assert
    assert.True(t, result)
}
```

**Integration Test Example:**
```go
package integration_test

import (
    "testing"
    "github.com/testcontainers/testcontainers-go"
)

func TestAuthRepository_CreateSession(t *testing.T) {
    // Setup testcontainer
    postgres := setupPostgresContainer(t)
    defer postgres.Terminate(context.Background())

    // Test with real database
    repo := NewAuthRepository(postgres.ConnectionString())
    session, err := repo.CreateSession(ctx, userID)

    assert.NoError(t, err)
    assert.NotEmpty(t, session.ID)
}
```

### Test Coverage Requirements

- **Minimum**: 80% for new code
- **Unit tests**: Core business logic
- **Integration tests**: Database operations
- **E2E tests**: Complete user flows

## ✨ Code Quality

### Pre-commit Checklist

Before committing:
```bash
# 1. Format code
make format

# 2. Run linter
make lint

# 3. Run tests
make test

# 4. Check go vet
make vet
```

### Code Style

**Follow Go best practices:**
- Use `gofmt` for formatting
- Maximum line length: 120 characters
- Maximum function complexity: 15 (gocyclo)
- Use meaningful variable names
- Add comments for exported functions

**Example:**
```go
// CreateSession creates a new user session with the given parameters.
// Returns an error if the session already exists or database operation fails.
func CreateSession(ctx context.Context, userID string, deviceID string) (*Session, error) {
    // Implementation
}
```

### Commit Message Format

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

<body> (minimum 100 characters)

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `style`: Code style (formatting, etc.)
- `refactor`: Code refactoring
- `perf`: Performance improvement
- `test`: Adding tests
- `chore`: Maintenance tasks
- `ci`: CI/CD changes

**Scopes:**
- `auth`, `user`, `wallet`, `gateway`
- `database`, `infra`, `ci`, `docs`

**Example:**
```
feat(auth): implement SIWE authentication flow

Added Sign-In With Ethereum (SIWE) authentication including:
- Nonce generation and validation
- Message signing verification
- Session creation with JWT tokens
- Device fingerprinting for security

Closes #123
```

## 🔄 CI/CD

### GitHub Actions Workflows

**CI Pipeline** (`.github/workflows/ci.yml`):
- Runs on: Push to `develop-claude`, `main`
- Runs on: Pull requests
- Jobs:
  - **Lint**: golangci-lint
  - **Test**: Unit + Integration tests
  - **Build**: Build all services
  - **Docker Build**: Build Docker images
  - **Security Scan**: gosec vulnerability scan

**PR Checks** (`.github/workflows/pr.yml`):
- PR title validation (conventional commits)
- Commit message validation
- Code formatting check
- Test coverage check (80% threshold)
- Dependency review
- PR size warnings

### Running CI Locally

```bash
# Run full CI pipeline
make ci

# This runs:
# 1. make lint
# 2. make test
# 3. make build
```

## 🐛 Troubleshooting

### Common Issues

**1. Port Already in Use**
```bash
# Check what's using the port
netstat -ano | findstr :5432

# Stop conflicting services
make dev-stop
```

**2. Docker Build Fails**
```bash
# Clean Docker cache
docker system prune -a

# Rebuild
make docker-build
```

**3. Tests Failing**
```bash
# Check if infrastructure is running
docker compose ps

# Reset database
make db-reset

# Run tests with verbose output
go test -v ./...
```

**4. Tilt Issues**
```bash
# Check Kubernetes
kubectl cluster-info

# Restart Tilt
make tilt-down
make tilt-up
```

**5. Go Module Issues**
```bash
# Clean cache
go clean -modcache

# Re-download
make deps

# Tidy modules
make deps-tidy
```

### Debug Mode

**Enable debug logging:**
```bash
# In .env
LOG_LEVEL=debug

# Restart services
make dev-stop && make dev
```

**View service logs:**
```bash
# Docker Compose
make dev-logs

# Specific service
docker logs -f nft-auth-service

# Tilt
tilt logs auth-service
```

## 📚 Additional Resources

- [Go Documentation](https://go.dev/doc/)
- [gRPC Go Tutorial](https://grpc.io/docs/languages/go/quickstart/)
- [Testcontainers Go](https://golang.testcontainers.org/)
- [Tilt Documentation](https://docs.tilt.dev/)
- [GitHub Actions](https://docs.github.com/en/actions)

## 🆘 Getting Help

- **Issues**: Create a GitHub issue
- **Discussions**: Use GitHub Discussions
- **Documentation**: Check README.md and BACKEND_TEST_REPORT.md

---

**Happy Coding!** 🚀
