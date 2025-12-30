# Code Standards & Codebase Structure

## Go Code Organization

### Clean Architecture Layers

All services follow consistent architectural layering:

```
cmd/main.go                 # Entry point, dependency wiring
  ↓
internal/config/           # Configuration from environment
  ↓
internal/models/           # Domain models, business logic types
  ↓
internal/repository/       # Data access layer (interface-based)
  ↓
internal/service/          # Business logic, orchestration
  ↓
internal/server/           # gRPC handlers, HTTP middleware
  ↓
internal/client/           # gRPC client connections
  ↓
Database                    # PostgreSQL
```

### Directory Structure Per Service

```
service-name/
├── cmd/
│   └── main.go           # Entry point, bootstrap
├── internal/
│   ├── config/
│   │   └── config.go     # Load config from environment
│   ├── models/
│   │   ├── model1.go     # Domain model 1
│   │   └── model2.go     # Domain model 2
│   ├── repository/
│   │   ├── interface.go  # Repository interface definitions
│   │   └── impl.go       # PostgreSQL implementations
│   ├── service/
│   │   ├── service.go    # Business logic
│   │   └── service_test.go
│   ├── server/
│   │   └── grpc_server.go
│   ├── client/
│   │   └── grpc_clients.go
│   └── middleware/       # Optional: HTTP/gRPC middleware
├── db/
│   └── up.sql            # Schema definition
├── migrations/           # Migration files
├── test/
│   └── .gitkeep          # Test data, fixtures
└── README.md             # Service documentation
```

## Repository Interface Pattern

### Definition

Always define repository interfaces explicitly:

```go
// User repository operations
type UserRepository interface {
  GetUser(ctx context.Context, id string) (*User, error)
  CreateUser(ctx context.Context, user *User) error
  UpdateUser(ctx context.Context, user *User) error
  DeleteUser(ctx context.Context, id string) error
  List(ctx context.Context, limit int, offset int) ([]*User, error)
}
```

### Implementation

PostgreSQL implementation in same package:

```go
type pgUserRepository struct {
  db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
  return &pgUserRepository{db: db}
}

func (r *pgUserRepository) GetUser(ctx context.Context, id string) (*User, error) {
  // Implementation
}
```

### Benefits

- **Testability**: Easy to mock in tests
- **Flexibility**: Swap implementations without changing callers
- **Isolation**: Data logic contained in repository
- **Consistency**: All services follow same pattern

## Dependency Injection

### Constructor-Based Injection

Pass dependencies explicitly in constructors:

```go
type AuthService struct {
  userRepo   UserRepository
  walletRepo WalletRepository
  jwtSvc     *JWTService
}

func NewAuthService(
  userRepo UserRepository,
  walletRepo WalletRepository,
  jwtSvc *JWTService,
) *AuthService {
  return &AuthService{
    userRepo:   userRepo,
    walletRepo: walletRepo,
    jwtSvc:     jwtSvc,
  }
}
```

### Bootstrap Pattern (main.go)

Wire all dependencies at startup:

```go
func main() {
  // Load config
  cfg := config.LoadConfig()

  // Initialize database
  db, err := sql.Open("postgres", cfg.DatabaseURL)

  // Create repositories
  userRepo := repository.NewUserRepository(db)
  walletRepo := repository.NewWalletRepository(db)

  // Create services
  jwtSvc := service.NewJWTService(cfg.JWTSecret)
  authSvc := service.NewAuthService(userRepo, walletRepo, jwtSvc)

  // Create gRPC server
  server := grpc.NewAuthServer(authSvc)

  // Start listening
  listener, _ := net.Listen("tcp", ":50051")
  grpcServer := grpc.NewServer()
  pb.RegisterAuthServer(grpcServer, server)
  grpcServer.Serve(listener)
}
```

## Configuration Loading

### Pattern

Type-safe environment loading in `internal/config/config.go`:

```go
package config

import "github.com/yourusername/shared/env"

type Config struct {
  // Database
  DatabaseURL  string
  DatabaseHost string
  DatabasePort int
  DatabaseUser string

  // JWT
  JWTSecret    string
  RefreshSecret string

  // Service
  Port int
  Env  string
}

func LoadConfig() *Config {
  return &Config{
    DatabaseURL: env.GetString("DATABASE_URL", ""),
    DatabasePort: env.GetInt("POSTGRES_PORT", 5432),
    JWTSecret: env.GetString("JWT_SECRET", ""),
    Port: env.GetInt("PORT", 50051),
    Env: env.GetString("ENV", "development"),
  }
}
```

### Best Practices

- Use `env` package for type safety
- Provide sensible defaults
- Document required vs optional
- Validate critical values at startup
- Never commit secrets; use `.env.example`

## Error Handling

### Standard Error Pattern

Return errors explicitly, don't hide:

```go
func (r *pgUserRepository) GetUser(ctx context.Context, id string) (*User, error) {
  user := &User{}
  err := r.db.QueryRowContext(ctx, "SELECT id, email FROM users WHERE id = $1", id).
    Scan(&user.ID, &user.Email)

  if err != nil {
    if err == sql.ErrNoRows {
      return nil, ErrUserNotFound
    }
    return nil, fmt.Errorf("get user: %w", err)
  }

  return user, nil
}
```

### Error Constants

Define domain-specific errors:

```go
var (
  ErrUserNotFound    = errors.New("user not found")
  ErrInvalidToken    = errors.New("invalid token")
  ErrNonceExpired    = errors.New("nonce expired")
  ErrWalletNotLinked = errors.New("wallet not linked to user")
)
```

### Wrapping Errors

Wrap errors with context using `%w`:

```go
if err != nil {
  return fmt.Errorf("create user: %w", err)  // Good
  // return fmt.Errorf("create user: %s", err)  // Bad - loses error chain
}
```

## Testing Patterns

### Table-Driven Tests

Standard approach for multiple test cases:

```go
func TestJWTValidation(t *testing.T) {
  tests := []struct {
    name      string
    token     string
    secret    string
    wantValid bool
    wantErr   string
  }{
    {
      name:      "valid token",
      token:     validToken,
      secret:    testSecret,
      wantValid: true,
    },
    {
      name:      "expired token",
      token:     expiredToken,
      secret:    testSecret,
      wantValid: false,
      wantErr:   "token expired",
    },
    {
      name:      "invalid signature",
      token:     validToken,
      secret:    wrongSecret,
      wantValid: false,
      wantErr:   "invalid signature",
    },
  }

  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      valid, err := ValidateToken(tt.token, tt.secret)

      if valid != tt.wantValid {
        t.Errorf("got %v, want %v", valid, tt.wantValid)
      }
      if tt.wantErr != "" && !strings.Contains(err.Error(), tt.wantErr) {
        t.Errorf("got error %q, want %q", err, tt.wantErr)
      }
    })
  }
}
```

### In-Memory Repository for Tests

Mock repository interface in tests:

```go
type mockUserRepository struct {
  users map[string]*User
}

func (m *mockUserRepository) GetUser(ctx context.Context, id string) (*User, error) {
  user, ok := m.users[id]
  if !ok {
    return nil, ErrUserNotFound
  }
  return user, nil
}

func TestAuthServiceLogin(t *testing.T) {
  repo := &mockUserRepository{
    users: map[string]*User{
      "user1": {ID: "user1", Email: "test@example.com"},
    },
  }

  svc := NewAuthService(repo)

  user, err := svc.Login("user1")
  if err != nil {
    t.Fatalf("login failed: %v", err)
  }
  if user.ID != "user1" {
    t.Errorf("got %q, want user1", user.ID)
  }
}
```

### Test File Naming

- `service_test.go` - Tests for `service.go`
- `repository_test.go` - Tests for `repository.go`
- `integration_test.go` - Integration tests
- `-i` flag: `go test -run IntegrationTests` for integration tests

### Coverage Requirements

- Minimum 80% on new code
- Focus on critical paths (auth, data access)
- Test error cases, not just happy paths
- Unit tests for business logic
- Integration tests for database operations

## Commit Message Format

### Conventional Commits

All commits follow conventional commit format:

```
<type>(<scope>): <description>

<body (optional, min 100 chars)>

<footer (optional)>
```

### Types

- **feat**: New feature
- **fix**: Bug fix
- **docs**: Documentation only
- **style**: Code style changes (formatting, missing semicolons)
- **refactor**: Code refactoring without feature changes
- **perf**: Performance improvements
- **test**: Adding or updating tests
- **chore**: Build process, dependencies, tooling

### Scopes

- **auth**: Authentication service changes
- **user**: User service changes
- **wallet**: Wallet service changes
- **gateway**: GraphQL gateway changes
- **database**: Schema or migration changes
- **infra**: Infrastructure/deployment changes
- **shared**: Shared package changes

### Examples

```
feat(auth): implement SIWE signature verification

Add atomic nonce consumption to prevent replay attacks.
Implement JWT token generation with refresh rotation.

Closes #42
```

```
fix(user): handle concurrent user creation

Use UPSERT pattern to prevent race conditions
when multiple requests create same user.
```

```
docs(gateway): update GraphQL schema documentation

Add field descriptions and query examples.
```

## Code Quality Gates

### Linting

**Tool**: golangci-lint
**Config**: `.golangci.yml` (project root)

```bash
make lint  # Run locally before commit
```

**Checks**:
- `gofmt` - Code formatting
- `govet` - Common Go mistakes
- `errcheck` - Unchecked errors
- `ineffassign` - Unused assignments
- `staticcheck` - Static analysis
- And 20+ other checks

### Code Formatting

**Tool**: gofmt (standard Go formatter)

```bash
make format  # Format all Go files

# Or manually
gofmt -w ./services/auth-service
```

### Vet

**Tool**: go vet (included with Go)

```bash
go vet ./...  # Run in repo root
```

**Checks**:
- Unreachable code
- Type mismatches
- Build constraints
- Unused variables

## Development Workflow: TDD

### The Three Phases

#### 1. RED Phase: Write Failing Test

Write test before implementation:

```go
func TestSIWEVerification(t *testing.T) {
  svc := NewSIWEService()

  // This test fails because function doesn't exist yet
  valid, err := svc.VerifyMessage(testMessage, testSignature)

  if !valid {
    t.Fatalf("verification failed: %v", err)
  }
}
```

#### 2. GREEN Phase: Write Minimal Code

Write minimal code to pass test:

```go
func (s *SIWEService) VerifyMessage(msg, sig string) (bool, error) {
  // Minimal implementation to pass test
  return true, nil
}
```

#### 3. REFACTOR Phase: Improve Code

Improve implementation while keeping tests green:

```go
func (s *SIWEService) VerifyMessage(msg, sig string) (bool, error) {
  // Proper implementation with error handling
  pubKey, err := RecoverPublicKey(msg, sig)
  if err != nil {
    return false, fmt.Errorf("recover pubkey: %w", err)
  }

  address := PubKeyToAddress(pubKey)
  // ... verification logic

  return true, nil
}
```

### Workflow Steps

1. **Create feature branch**: `git checkout -b feature/auth-siwe`
2. **Write test**: Add test case that fails
3. **Run test**: `go test ./...` - verify it fails
4. **Write code**: Implement feature
5. **Run test**: Verify test passes
6. **Refactor**: Improve code quality
7. **Run all tests**: `make test` - ensure nothing broke
8. **Commit**: `git commit -m "feat(auth): implement SIWE verification"`
9. **Push & PR**: `git push origin feature/auth-siwe`

## Interfaces & Abstractions

### When to Use Interfaces

Use interfaces for:

1. **Repository/Data Access**: Always abstract database
2. **External Services**: gRPC clients, third-party APIs
3. **Testability**: Components with side effects
4. **Plugin Points**: Extensible behavior

Don't use interfaces for:

1. **Simple implementations**: Unnecessary complexity
2. **Single use cases**: Only one implementation needed
3. **Trivial types**: Primitive wrappers

### Service-to-Service Communication

All inter-service calls go through gRPC:

```go
type AuthService struct {
  userClient pb.UserClient  // gRPC stub
}

func (s *AuthService) ensureUser(ctx context.Context, address string) error {
  _, err := s.userClient.CreateUser(ctx, &pb.CreateUserRequest{
    Address: address,
  })
  return err
}
```

## Database Patterns

### Connection Pooling

Always use connection pools:

```go
db, err := sql.Open("postgres", config.DatabaseURL)
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
```

### Prepared Statements

Use parameterized queries (always):

```go
// Good - prevents SQL injection
err := db.QueryRowContext(ctx,
  "SELECT id FROM users WHERE email = $1",
  email).Scan(&id)

// Bad - vulnerable to SQL injection
query := fmt.Sprintf("SELECT id FROM users WHERE email = '%s'", email)
```

### Context Usage

Always pass context through call chain:

```go
func (s *AuthService) Login(ctx context.Context, address string) (*User, error) {
  user, err := s.userRepo.GetUser(ctx, address)  // Pass context
  if err != nil {
    return nil, err
  }
  return user, nil
}
```

### Transaction Handling

Use transactions for multi-step operations:

```go
tx, err := s.db.BeginTx(ctx, nil)
if err != nil {
  return err
}
defer tx.Rollback()

// Multiple operations
_, err = tx.ExecContext(ctx, "UPDATE users SET ...")
// ...
_, err = tx.ExecContext(ctx, "INSERT INTO audit_log ...")

if err != nil {
  return err
}

return tx.Commit().Error
```

## Atomic Operations Pattern

### Atomic Nonce Consumption

Database-level atomicity for critical operations:

```sql
-- Consume nonce atomically
CREATE OR REPLACE FUNCTION consume_nonce(
  p_address VARCHAR,
  p_nonce VARCHAR
) RETURNS BOOLEAN AS $$
BEGIN
  UPDATE auth_nonces
  SET used_at = NOW()
  WHERE address = p_address
    AND nonce = p_nonce
    AND used_at IS NULL;

  RETURN FOUND;
END;
$$ LANGUAGE plpgsql;
```

Go code:

```go
func (r *pgNonceRepository) ConsumeNonce(ctx context.Context, address, nonce string) (bool, error) {
  var consumed bool
  err := r.db.QueryRowContext(ctx,
    "SELECT consume_nonce($1, $2)",
    address, nonce).Scan(&consumed)
  return consumed, err
}
```

## Idempotency Pattern

### UPSERT Operations

Idempotent creates using UPSERT:

```sql
INSERT INTO users (id, address, created_at)
VALUES ($1, $2, NOW())
ON CONFLICT (id) DO NOTHING
RETURNING id;
```

Go implementation:

```go
func (r *pgUserRepository) EnsureUser(ctx context.Context, user *User) error {
  _, err := r.db.ExecContext(ctx, `
    INSERT INTO users (id, address, email, created_at)
    VALUES ($1, $2, $3, NOW())
    ON CONFLICT (id) DO NOTHING
  `, user.ID, user.Address, user.Email)
  return err
}
```

## Naming Conventions

### Go Conventions

- **Packages**: `lowercase`, `short` (e.g., `repository`, `service`)
- **Functions**: `CamelCase`, exported with verb (e.g., `GetUser`, `CreateUser`)
- **Variables**: `camelCase` (e.g., `userID`, `authToken`)
- **Constants**: `UPPERCASE` (e.g., `MAX_RETRIES`, `DEFAULT_TIMEOUT`)
- **Interfaces**: `CamelCase` ending in "er" (e.g., `Reader`, `Writer`, `Repository`)

### Database Naming

- **Tables**: `snake_case`, `plural` (e.g., `users`, `wallet_links`)
- **Columns**: `snake_case` (e.g., `user_id`, `created_at`, `email_verified`)
- **Constraints**: `{table}_{column}_key` (e.g., `users_email_key`)
- **Indexes**: `idx_{table}_{column}` (e.g., `idx_users_email`)

### File Naming

- **Source files**: `snake_case.go` (e.g., `auth_service.go`)
- **Test files**: `{source}_test.go` (e.g., `auth_service_test.go`)
- **Generated**: `{name}.pb.go`, `{name}_grpc.pb.go`

## Security Best Practices

### Environment Secrets

Never commit secrets:

```bash
# Bad - never do this
JWT_SECRET=my-secret-key-here

# Good - use .env.example template
JWT_SECRET=your-256-bit-secret-key-here
```

### SQL Injection Prevention

Always use parameterized queries:

```go
// ✓ Safe
db.Query("SELECT * FROM users WHERE id = $1", userID)

// ✗ Unsafe
db.Query(fmt.Sprintf("SELECT * FROM users WHERE id = %s", userID))
```

### Password Hashing

Use bcrypt for password hashing:

```go
import "golang.org/x/crypto/bcrypt"

hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
err = bcrypt.CompareHashAndPassword(hash, []byte(inputPassword))
```

### Token Secrets

- Minimum 256 bits for HMAC-SHA256
- Rotate on schedule
- Store in secure vault (not config files)
- Never log token values

## Documentation Standards

### Code Comments

- **Functions**: Document public functions with comment blocks
- **Complex logic**: Explain non-obvious implementation decisions
- **TODOs**: Mark incomplete work with `// TODO: description`
- **Deprecation**: Mark deprecated functions with `// Deprecated: ...`

### Example

```go
// AuthService handles SIWE authentication and JWT token management.
// It coordinates between the auth repository and external services.
type AuthService struct {
  repo  repository.AuthRepository
  siwe  *SIWEVerifier
}

// VerifySignature validates a SIWE message signature against the account address.
// It prevents replay attacks through atomic nonce consumption.
func (s *AuthService) VerifySignature(ctx context.Context, message, signature string) (bool, error) {
  // Implementation
}
```

## File Sizing

### Guideline

- **Packages**: Keep packages focused (one responsibility)
- **Files**: Aim for < 500 lines per file
- **Functions**: Aim for < 50 lines; split if longer
- **Structs**: Aim for < 10 fields; consider composition

### Rationale

- Easier to understand
- Simpler to test
- Fewer merge conflicts
- Better code organization

## Unresolved Standards Questions

1. Should we require 100% coverage for critical paths (auth, wallet)?
2. What's the preferred approach for logging (structured vs fmt)?
3. Should service-to-service calls use retries with backoff?
4. What metrics should be exposed for monitoring?
5. Should we enforce specific gRPC timeout values?

---

**Version**: 1.0
**Last Updated**: 2025-12-04
**Applies To**: All Go services in the project
