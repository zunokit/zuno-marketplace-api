package main

import (
	"log"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/quangdang46/NFT-Marketplace/services/graphql-gateway/graph"
	"github.com/quangdang46/NFT-Marketplace/services/graphql-gateway/internal/config"
	appcontext "github.com/quangdang46/NFT-Marketplace/services/graphql-gateway/internal/context"
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

	// Create GraphQL resolver with gRPC clients
	resolver := &graph.Resolver{
		AuthClient:   pb.NewAuthServiceClient(authConn),
		UserClient:   pb.NewUserServiceClient(userConn),
		WalletClient: pb.NewWalletServiceClient(walletConn),
	}

	// Create GraphQL server
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

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
	router.Use(authmiddleware.AuthMiddleware(cfg.JWT.AccessSecret))

	// Health check endpoint
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	// GraphQL endpoint with context middleware
	router.Handle("/graphql", contextMiddleware(srv))

	// GraphQL Playground (development only)
	if cfg.Features.PlaygroundEnabled {
		router.Handle("/playground", playground.Handler("GraphQL Playground", "/graphql"))
		log.Println("GraphQL Playground enabled at http://localhost" + cfg.Server.HTTPAddr + "/playground")
	}

	// Start server
	log.Printf("GraphQL Gateway listening on %s", cfg.Server.HTTPAddr)
	log.Printf("GraphQL endpoint: http://localhost%s/graphql", cfg.Server.HTTPAddr)

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
