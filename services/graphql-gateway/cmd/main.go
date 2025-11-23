package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/quangdang46/NFT-Marketplace/services/graphql-gateway/graph"
	"github.com/quangdang46/NFT-Marketplace/services/graphql-gateway/internal/config"
	appcontext "github.com/quangdang46/NFT-Marketplace/services/graphql-gateway/internal/context"
	"github.com/quangdang46/NFT-Marketplace/services/graphql-gateway/internal/handlers"
	"github.com/quangdang46/NFT-Marketplace/services/graphql-gateway/internal/health"
	authmiddleware "github.com/quangdang46/NFT-Marketplace/services/graphql-gateway/internal/middleware"
	pb "github.com/quangdang46/NFT-Marketplace/shared/proto/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	log.Println("Starting GraphQL Gateway...")

	// Load configuration
	cfg := config.Load()

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}

	// Connect to gRPC services
	log.Println("Connecting to gRPC services...")
	authConn, err := grpc.Dial(cfg.Services.AuthServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to auth service: %v", err)
	}
	defer authConn.Close()

	userConn, err := grpc.Dial(cfg.Services.UserServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to user service: %v", err)
	}
	defer userConn.Close()

	walletConn, err := grpc.Dial(cfg.Services.WalletServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
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
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)

	// CORS middleware - allow frontend to access GraphQL API
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:3001", "http://127.0.0.1:3000", "http://127.0.0.1:3001"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// JWT authentication middleware
	router.Use(authmiddleware.AuthMiddleware(cfg.JWT.Secret))

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

	// Start server
	log.Printf("GraphQL Gateway listening on %s", cfg.Server.HTTPAddr)
	log.Printf("GraphQL endpoint: http://localhost%s/graphql", cfg.Server.HTTPAddr)
	log.Printf("Upload Media endpoint: http://localhost%s/api/upload/media", cfg.Server.HTTPAddr)
	log.Printf("Batch Upload endpoint: http://localhost%s/api/upload/batch", cfg.Server.HTTPAddr)
	log.Printf("Indexer Webhook endpoint: http://localhost%s/api/webhooks/indexer", cfg.Server.HTTPAddr)

	if err := http.ListenAndServe(cfg.Server.HTTPAddr, router); err != nil {
		log.Fatal(err)
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
