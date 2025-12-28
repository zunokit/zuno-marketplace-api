# Phase 03: Service Integration

**Context**: `plan.md` | **Priority**: P1 | **Effort**: 1h | **Depends**: Phase 01, Phase 02

---

## Overview

Integrate Sentry into all 4 services (Auth, User, Wallet, GraphQL Gateway) by adding initialization, middleware, and configuration. Each service gets error capture and tracing.

**Status**: Pending

---

## Related Files

- Core Package: `phase-01-core-package.md`
- Middleware: `phase-02-middleware.md`
- Service Main Files: `services/*/cmd/main.go`
- Config Files: `services/*/internal/config/config.go`

---

## Requirements

### Functional
- Add Sentry config to all 4 services
- Initialize Sentry at service startup
- Add appropriate middleware to each service
- Capture panics and send to Sentry
- Flush events on graceful shutdown

### Non-Functional
- Services continue working if Sentry init fails
- No blocking on Sentry operations

---

## Architecture

```
Service Startup Flow:
1. Load config (including Sentry DSN)
2. Init Sentry (non-blocking if fails)
3. Setup database, gRPC server
4. Add Sentry middleware/interceptors
5. Start server
6. On shutdown: Flush Sentry (2s timeout)
```

---

## Implementation Steps

### Step 1: Update Config for All Services

**Files to modify**:
- `services/auth-service/internal/config/config.go`
- `services/user-service/internal/config/config.go`
- `services/wallet-service/internal/config/config.go`
- `services/graphql-gateway/internal/config/config.go`

Add SentryConfig struct to each:

```go
// Add to existing Config struct
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Sentry   SentryConfig  // NEW
	JWT      JWTConfig      // existing
	Services ServicesConfig // existing
}

// Add new struct
type SentryConfig struct {
	DSN         string
	Environment string
}

// Update Load() function to add:
Sentry: SentryConfig{
	DSN:         env.GetString("SENTRY_DSN", ""),
	Environment: env.GetString("SENTRY_ENVIRONMENT", "development"),
},
```

### Step 2: Update Auth Service

**File**: `services/auth-service/cmd/main.go`

```go
package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpcMiddleware "github.com/quangdang46/NFT-Marketplace/shared/observability/middleware"
	obs "github.com/quangdang46/NFT-Marketplace/shared/observability/sentry"

	// ... existing imports
	"github.com/getsentry/sentry-go"
	"google.golang.org/grpc"
)

func main() {
	log.Println("Starting Auth Service...")

	// Load configuration
	cfg := config.Load()

	// Initialize Sentry (non-blocking on failure)
	if cfg.Sentry.DSN != "" {
		if err := obs.Init(
			cfg.Sentry.DSN,
			cfg.Sentry.Environment,
			"auth-service",
			getBuildVersion(), // Helper function
			0.2, // 20% sampling for staging
		); err != nil {
			log.Printf("Sentry init failed (continuing): %v", err)
		} else {
			log.Println("Sentry initialized")
			defer obs.Flush(2 * time.Second)
		}
	} else {
		log.Println("Sentry DSN not configured, skipping")
	}

	// Initialize database
	dsn := cfg.Database.GetDSN()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		// Capture to Sentry before fatal
		sentry.CaptureException(err)
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Database connected successfully")

	// ... existing repository/service setup ...

	// Create gRPC server with Sentry interceptor
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(10*1024*1024),
		grpc.MaxSendMsgSize(10*1024*1024),
		grpc.ChainUnaryInterceptor(
			grpcMiddleware.UnaryServerInterceptor(), // NEW
		),
	)

	// Register auth service
	authServer := server.NewAuthServer(nonceRepo, sessionRepo, loginEventRepo, siweService, jwtService, clients)
	pb.RegisterAuthServiceServer(grpcServer, authServer)

	// ... existing health/reflection setup ...

	// Start gRPC server
	listener, err := net.Listen("tcp", cfg.Server.GRPCPort)
	if err != nil {
		sentry.CaptureException(err)
		log.Fatalf("Failed to listen on %s: %v", cfg.Server.GRPCPort, err)
	}

	log.Printf("Auth Service listening on %s", cfg.Server.GRPCPort)

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down Auth Service...")

		// Flush Sentry before shutdown
		if cfg.Sentry.DSN != "" {
			log.Println("Flushing Sentry events...")
			obs.Flush(2 * time.Second)
		}

		grpcServer.GracefulStop()
		log.Println("Auth Service stopped")
	}()

	if err := grpcServer.Serve(listener); err != nil {
		sentry.CaptureException(err)
		log.Fatalf("Failed to serve: %v", err)
	}
}

// getBuildVersion returns the version from build info or git
func getBuildVersion() string {
	// Can use ldflags or git describe
	return "v0.1.0" // Placeholder
}
```

### Step 3: Update User Service

**File**: `services/user-service/cmd/main.go`

Same pattern as Auth Service:

```go
package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpcMiddleware "github.com/quangdang46/NFT-Marketplace/shared/observability/middleware"
	obs "github.com/quangdang46/NFT-Marketplace/shared/observability/sentry"
	"github.com/getsentry/sentry-go"
)

func main() {
	log.Println("Starting User Service...")

	cfg := config.Load()

	// Initialize Sentry
	if cfg.Sentry.DSN != "" {
		if err := obs.Init(
			cfg.Sentry.DSN,
			cfg.Sentry.Environment,
			"user-service",
			getBuildVersion(),
			0.2,
		); err != nil {
			log.Printf("Sentry init failed: %v", err)
		} else {
			defer obs.Flush(2 * time.Second)
		}
	}

	// ... existing database/repository setup ...

	// Create gRPC server with Sentry interceptor
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpcMiddleware.UnaryServerInterceptor(),
		),
	)

	// ... rest of setup ...

	// Graceful shutdown with Sentry flush
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		if cfg.Sentry.DSN != "" {
			obs.Flush(2 * time.Second)
		}
		grpcServer.GracefulStop()
	}()

	// ... serve
}
```

### Step 4: Update Wallet Service

**File**: `services/wallet-service/cmd/main.go`

Same pattern as User Service.

### Step 5: Update GraphQL Gateway

**File**: `services/graphql-gateway/cmd/main.go`

```go
package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	obshttp "github.com/quangdang46/NFT-Marketplace/shared/observability/middleware"
	obsgrpc "github.com/quangdang46/NFT-Marketplace/shared/observability/middleware"
	obs "github.com/quangdang46/NFT-Marketplace/shared/observability/sentry"

	pb "github.com/quangdang46/NFT-Marketplace/shared/proto/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	log.Println("Starting GraphQL Gateway...")

	cfg := config.Load()

	// Initialize Sentry
	if cfg.Sentry.DSN != "" {
		if err := obs.Init(
			cfg.Sentry.DSN,
			cfg.Sentry.Environment,
			"graphql-gateway",
			getBuildVersion(),
			0.2,
		); err != nil {
			log.Printf("Sentry init failed: %v", err)
		} else {
			defer obs.Flush(2 * time.Second)
		}
	}

	// Validate config
	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}

	// Connect to gRPC services with Sentry client interceptors
	log.Println("Connecting to gRPC services...")

	authConn, err := grpc.Dial(
		cfg.Services.AuthServiceURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			obsgrpc.UnaryClientInterceptor(), // NEW
		),
	)
	if err != nil {
		log.Fatalf("Failed to connect to auth service: %v", err)
	}
	defer authConn.Close()

	userConn, err := grpc.Dial(
		cfg.Services.UserServiceURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			obsgrpc.UnaryClientInterceptor(), // NEW
		),
	)
	if err != nil {
		log.Fatalf("Failed to connect to user service: %v", err)
	}
	defer userConn.Close()

	walletConn, err := grpc.Dial(
		cfg.Services.WalletServiceURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			obsgrpc.UnaryClientInterceptor(), // NEW
		),
	)
	if err != nil {
		log.Fatalf("Failed to connect to wallet service: %v", err)
	}
	defer walletConn.Close()

	// Create GraphQL resolver with gRPC clients
	resolver := &graph.Resolver{
		AuthClient:   pb.NewAuthServiceClient(authConn),
		UserClient:   pb.NewUserServiceClient(userConn),
		WalletClient: pb.NewWalletServiceClient(walletConn),
	}

	// Create GraphQL server with Sentry middleware
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	// Add GraphQL field middleware for Sentry
	srv.Use(obsgraphql.GraphQLFieldMiddleware())     // NEW
	srv.Use(obsgraphql.GraphQLResponseMiddleware())  // NEW

	// Setup HTTP router
	router := chi.NewRouter()

	// IMPORTANT: Sentry middleware FIRST (captures entire request)
	router.Use(obshttp.SentryHTTP) // NEW - add before other middleware

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)

	// CORS setup (existing)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:3001"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// JWT auth middleware (existing)
	router.Use(authmiddleware.AuthMiddleware(cfg.JWT.AccessSecret))

	// Health check (existing - Sentry middleware skips /health)
	healthRegistry := health.NewRegistry()
	healthRegistry.Register("auth_service", health.NewServiceHealthChecker(authConn))
	healthRegistry.Register("user_service", health.NewServiceHealthChecker(userConn))
	healthRegistry.Register("wallet_service", health.NewServiceHealthChecker(walletConn))

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		healthStatus := healthRegistry.CheckAll(r.Context())
		overallStatus := healthStatus["status"]

		w.Header().Set("Content-Type", "application/json")
		if overallStatus == "healthy" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		json.NewEncoder(w).Encode(healthStatus)
	})

	// GraphQL endpoint with context middleware
	router.Handle("/graphql", contextMiddleware(srv))

	// GraphQL Playground (existing)
	if cfg.Features.PlaygroundEnabled {
		router.Handle("/playground", playground.Handler("GraphQL Playground", "/graphql"))
		log.Println("GraphQL Playground enabled")
	}

	// Start server
	log.Printf("GraphQL Gateway listening on %s", cfg.Server.HTTPAddr)

	if err := http.ListenAndServe(cfg.Server.HTTPAddr, router); err != nil {
		log.Fatal(err)
	}
}

// contextMiddleware (existing - keep as is)
func contextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = appcontext.WithHTTPRequest(ctx, r)
		ctx = appcontext.WithHTTPResponse(ctx, w)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
```

### Step 6: Update .env.example

**File**: `.env.example`

Add to existing file:

```env
# Sentry Monitoring
SENTRY_DSN=https://xxxxxxxxxxxx@xxxxx.ingest.sentry.io/xxxxxxx
SENTRY_ENVIRONMENT=development
```

---

## Todo List

### Config
- [ ] Update auth-service config.go with SentryConfig
- [ ] Update user-service config.go with SentryConfig
- [ ] Update wallet-service config.go with SentryConfig
- [ ] Update graphql-gateway config.go with SentryConfig

### Auth Service
- [ ] Add Sentry initialization in main.go
- [ ] Add gRPC server interceptor
- [ ] Add panic capture
- [ ] Add graceful shutdown flush

### User Service
- [ ] Add Sentry initialization in main.go
- [ ] Add gRPC server interceptor
- [ ] Add panic capture
- [ ] Add graceful shutdown flush

### Wallet Service
- [ ] Add Sentry initialization in main.go
- [ ] Add gRPC server interceptor
- [ ] Add panic capture
- [ ] Add graceful shutdown flush

### GraphQL Gateway
- [ ] Add Sentry initialization in main.go
- [ ] Add HTTP middleware to Chi router
- [ ] Add gRPC client interceptors
- [ ] Add GraphQL field middleware
- [ ] Add graceful shutdown flush

### Environment
- [ ] Update .env.example with Sentry variables

---

## Success Criteria

- [ ] All 4 services initialize Sentry on startup
- [ ] GraphQL Gateway traces HTTP requests
- [ ] gRPC services trace incoming calls
- [ ] Client interceptors inject trace headers
- [ ] Panics captured and sent to Sentry
- [ ] Graceful shutdown flushes events

---

## Risk Assessment

| Risk | Mitigation |
|------|------------|
| Service fails if Sentry down | Non-blocking init, log error only |
| Missing trace context | E2E test to verify waterfalls |
| Config errors | Validate DSN format before init |

---

## Next Steps

→ Phase 04: Distributed Tracing
