# Development Guide

This guide covers the development workflow, testing practices, and environment setup for the Zuno NFT Marketplace API.

## Development Environment Modes

### Serverless Mode (Recommended for Development)

- Zero local infrastructure overhead
- Free-tier cloud services (Supabase, Upstash, CloudAMQP)
- Faster iteration (no container rebuilds)
- Best for feature development

### Docker Mode

- Full production parity
- Offline development
- Integration testing
- Best for final testing before deployment

Switch modes via `INFRA_MODE` in `.env`.

See [Serverless Development Guide](./docs/serverless-development.md) for complete setup instructions.

## Getting Started

1. **Clone and setup:**
   ```bash
   git clone <repo-url>
   cd zuno-marketplace-api
   ./scripts/setup-env.sh
   ```

2. **Choose your mode:**
   - Serverless: No Docker, use cloud services
   - Docker: Local containers with `docker compose up -d`

3. **Verify setup:**
   ```bash
   ./scripts/health-check.sh
   ```

## Test-Driven Development (TDD)

Follow the Red-Green-Refactor cycle:

### 🔴 RED - Write Failing Test

```go
// services/auth-service/internal/auth/usecase/nonce_test.go
func TestGenerateNonce(t *testing.T) {
    // Arrange
    uc := NewNonceUseCase(mockRepo)

    // Act
    nonce, err := uc.Generate(context.Background(), "0x123...")

    // Assert
    assert.NoError(t, err)
    assert.NotEmpty(t, nonce.Value)
    assert.Equal(t, "0x123...", nonce.WalletAddress)
}
```

### 🟢 GREEN - Make Test Pass

Write minimal code to pass the test:

```go
func (uc *NonceUseCase) Generate(ctx context.Context, wallet string) (*Nonce, error) {
    nonce := &Nonce{
        Value:          generateRandomString(),
        WalletAddress:  wallet,
        ExpirationTime: time.Now().Add(5 * time.Minute),
    }
    return nonce, uc.repo.Create(ctx, nonce)
}
```

### 🔵 REFACTOR - Improve Code

Clean up while keeping tests green:
- Extract common logic
- Improve naming
- Add missing error handling
- Apply SOLID principles

## Testing Strategy

### Unit Tests
- Test individual functions/methods
- Mock external dependencies
- Fast execution (< 1ms per test)

### Integration Tests
- Test with real databases
- Use testcontainers for isolated tests
- Verify contract between layers

### E2E Tests
- Test complete user flows
- Real infrastructure or testcontainers
- Slow but comprehensive

### Coverage Targets
- Minimum: 80% for new code
- Aim for: 90%+ for core business logic

## Common Commands

```bash
# Run all tests
make test

# Run specific service tests
go test ./services/auth-service/...

# Run with coverage
go test -cover ./...

# Run with race detection
go test -race ./...

# Build all services
make build

# Run linter
make lint

# Format code
make format

# Run CI locally
make ci
```

## Project Structure

```
services/
├── auth-service/
│   ├── cmd/              # Application entry points
│   ├── internal/         # Private application code
│   │   ├── auth/         # Auth domain logic
│   │   │   ├── domain/   # Entities, value objects
│   │   │   ├── usecase/  # Business logic (use cases)
│   │   │   ├── repository/# Database interfaces
│   │   │   └── transport/# gRPC/HTTP handlers
│   │   └── infra/        # Infrastructure (config, logging)
│   ├── db/               # Database migrations
│   └── test/             # Test utilities
```

## Code Standards

- Follow [Code Standards](./docs/code-standards.md)
- Use conventional commits
- Write self-documenting code
- Keep functions small (< 50 lines)
- Apply SOLID principles

## Troubleshooting

See [Troubleshooting Guide](./docs/troubleshooting.md) for common issues.

## Related Documentation

- [Serverless Development Guide](./docs/serverless-development.md)
- [System Architecture](./docs/system-architecture.md)
- [Code Standards](./docs/code-standards.md)
- [Project Overview](./docs/project-overview-pdr.md)
