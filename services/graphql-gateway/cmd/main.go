package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	obs "github.com/zunokit/zuno-marketplace-api/shared/observability/middleware"
	obsentry "github.com/zunokit/zuno-marketplace-api/shared/observability/sentry"
	obsTrace "github.com/zunokit/zuno-marketplace-api/shared/observability/tracing"
	sharedrabbitmq "github.com/zunokit/zuno-marketplace-api/shared/rabbitmq"
	sharedredis "github.com/zunokit/zuno-marketplace-api/shared/redis"

	"github.com/zunokit/zuno-marketplace-api/services/graphql-gateway/graph"
	"github.com/zunokit/zuno-marketplace-api/services/graphql-gateway/internal/config"
	appcontext "github.com/zunokit/zuno-marketplace-api/services/graphql-gateway/internal/context"
	"github.com/zunokit/zuno-marketplace-api/services/graphql-gateway/internal/handlers"
	"github.com/zunokit/zuno-marketplace-api/services/graphql-gateway/internal/health"
	authmiddleware "github.com/zunokit/zuno-marketplace-api/services/graphql-gateway/internal/middleware"
	pb "github.com/zunokit/zuno-marketplace-api/shared/proto/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Version and BuildTime are injected via ldflags during build
var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	log.Println("Starting GraphQL Gateway...")

	// Load configuration
	cfg := config.Load()

	// Initialize Sentry
	if cfg.Sentry.DSN != "" {
		if err := obsentry.Init(
			cfg.Sentry.DSN,
			cfg.Sentry.Environment,
			"graphql-gateway",
			getBuildVersion(),
			obsTrace.GetTracesSampleRate(cfg.Sentry.Environment),
		); err != nil {
			log.Printf("Sentry init failed (continuing): %v", err)
		} else {
			log.Println("Sentry initialized")
			defer obsentry.Flush(2 * time.Second)
		}
	} else {
		log.Println("Sentry DSN not configured, skipping")
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}

	// Initialize Redis (non-blocking)
	if err := sharedredis.Init(cfg.Redis.GetAddr()); err != nil {
		log.Printf("Redis init failed (continuing without cache): %v", err)
	} else {
		log.Println("Redis connected")
		defer sharedredis.Close()
	}

	// Initialize RabbitMQ (non-blocking)
	if err := sharedrabbitmq.Init(cfg.RabbitMQ.GetURL()); err != nil {
		log.Printf("RabbitMQ init failed (continuing without events): %v", err)
	} else {
		log.Println("RabbitMQ connected")
		defer sharedrabbitmq.Close()
	}

	// Connect to gRPC services with Sentry client interceptors
	log.Println("Connecting to gRPC services...")

	authConn, err := grpc.Dial(
		cfg.Services.AuthServiceURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			obs.UnaryClientInterceptor(),
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
			obs.UnaryClientInterceptor(),
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
			obs.UnaryClientInterceptor(),
		),
	)
	if err != nil {
		log.Fatalf("Failed to connect to wallet service: %v", err)
	}
	defer walletConn.Close()

	collectionConn, err := grpc.Dial(cfg.Services.CollectionServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to collection service: %v", err)
	}
	defer collectionConn.Close()

	mediaConn, err := grpc.Dial(cfg.Services.MediaServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to media service: %v", err)
	}
	defer mediaConn.Close()

	// Create GraphQL resolver with gRPC clients
	resolver := &graph.Resolver{
		AuthClient:       pb.NewAuthServiceClient(authConn),
		UserClient:       pb.NewUserServiceClient(userConn),
		WalletClient:     pb.NewWalletServiceClient(walletConn),
		CollectionClient: pb.NewCollectionServiceClient(collectionConn),
	}

	// Create GraphQL server
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	// Initialize upload handler with Media Service client
	logger := log.New(os.Stdout, "[GraphQL-Gateway] ", log.LstdFlags)
	mediaClient := pb.NewMediaServiceClient(mediaConn)
	uploadHandler := handlers.NewUploadHandler(mediaClient, logger)

	// Initialize webhook handler with Collection Service client
	webhookHandler := handlers.NewWebhookHandler(collectionConn, cfg.WebhookSecret)

	// Setup HTTP router
	router := chi.NewRouter()

	// IMPORTANT: Sentry middleware FIRST (captures entire request)
	router.Use(obs.SentryHTTP)

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)

	// CORS middleware - allow frontend to access GraphQL API
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:3001", "http://127.0.0.1:3000", "http://127.0.0.1:3001", "http://localhost:42069"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// JWT authentication middleware
	router.Use(authmiddleware.AuthMiddleware(cfg.JWT.Secret))

	// Rate limiting middleware
	router.Use(authmiddleware.RateLimit)

	// Health check endpoint
	healthRegistry := health.NewRegistry()
	healthRegistry.Register("auth_service", health.NewServiceHealthChecker(authConn))
	healthRegistry.Register("user_service", health.NewServiceHealthChecker(userConn))
	healthRegistry.Register("wallet_service", health.NewServiceHealthChecker(walletConn))
	healthRegistry.Register("collection_service", health.NewServiceHealthChecker(collectionConn))
	healthRegistry.Register("media_service", health.NewServiceHealthChecker(mediaConn))

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

	// GraphQL Playground (development only)
	if cfg.Features.PlaygroundEnabled {
		router.Handle("/playground", playground.Handler("GraphQL Playground", "/graphql"))
		log.Println("GraphQL Playground enabled at http://localhost" + cfg.Server.HTTPAddr + "/playground")
	}

	// Upload endpoints
	router.Post("/api/upload/media", uploadHandler.UploadMedia)
	router.Post("/api/upload/batch", uploadHandler.BatchUploadMedia)

	// Webhook endpoints
	router.Post("/api/webhooks/indexer", webhookHandler.HandleIndexerWebhook)

	// Start server with graceful shutdown
	log.Printf("GraphQL Gateway listening on %s", cfg.Server.HTTPAddr)
	log.Printf("GraphQL endpoint: http://localhost%s/graphql", cfg.Server.HTTPAddr)
	log.Printf("Upload Media endpoint: http://localhost%s/api/upload/media", cfg.Server.HTTPAddr)
	log.Printf("Batch Upload endpoint: http://localhost%s/api/upload/batch", cfg.Server.HTTPAddr)
	log.Printf("Indexer Webhook endpoint: http://localhost%s/api/webhooks/indexer", cfg.Server.HTTPAddr)

	// Create server with timeout context
	server := &http.Server{
		Addr:    cfg.Server.HTTPAddr,
		Handler: router,
	}

	// Graceful shutdown in goroutine
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down GraphQL Gateway...")

		// Flush Sentry before shutdown
		if cfg.Sentry.DSN != "" {
			log.Println("Flushing Sentry events...")
			obsentry.Flush(2 * time.Second)
		}

		// Graceful shutdown with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
		log.Println("GraphQL Gateway stopped")
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to serve: %v", err)
	}
}

// contextMiddleware adds HTTP request and response to GraphQL context
func contextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = appcontext.WithHTTPRequest(ctx, r)
		ctx = appcontext.WithHTTPResponse(ctx, w)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// getBuildVersion returns the version injected by build ldflags
func getBuildVersion() string {
	return Version
}
