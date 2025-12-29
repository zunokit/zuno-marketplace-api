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

	"github.com/quangdang46/NFT-Marketplace/services/auth-service/internal/client"
	"github.com/quangdang46/NFT-Marketplace/services/auth-service/internal/config"
	"github.com/quangdang46/NFT-Marketplace/services/auth-service/internal/repository"
	"github.com/quangdang46/NFT-Marketplace/services/auth-service/internal/server"
	"github.com/quangdang46/NFT-Marketplace/services/auth-service/internal/service"
	pb "github.com/quangdang46/NFT-Marketplace/shared/proto/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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
			getBuildVersion(),
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
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Database connected successfully")

	// Initialize repositories
	nonceRepo := repository.NewNonceRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	loginEventRepo := repository.NewLoginEventRepository(db)

	// Initialize services
	siweService := service.NewSIWEService()
	jwtService := service.NewJWTService(
		cfg.JWT.Secret,
		cfg.JWT.RefreshSecret,
		cfg.JWT.AccessExpiration,
		cfg.JWT.RefreshExpiration,
	)

	// Initialize gRPC clients
	clients, err := client.NewServiceClients(cfg.Services.UserServiceURL, cfg.Services.WalletServiceURL)
	if err != nil {
		log.Fatalf("Failed to initialize gRPC clients: %v", err)
	}
	log.Println("gRPC clients initialized")

	// Create gRPC server with Sentry interceptor
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(10*1024*1024),
		grpc.MaxSendMsgSize(10*1024*1024),
		grpc.ChainUnaryInterceptor(
			grpcMiddleware.UnaryServerInterceptor(),
		),
	)

	// Register auth service
	authServer := server.NewAuthServer(nonceRepo, sessionRepo, loginEventRepo, siweService, jwtService, clients)
	pb.RegisterAuthServiceServer(grpcServer, authServer)

	// Register health check
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	// Register reflection
	reflection.Register(grpcServer)

	// Start gRPC server
	listener, err := net.Listen("tcp", cfg.Server.GRPCPort)
	if err != nil {
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
		log.Fatalf("Failed to serve: %v", err)
	}
}

// getBuildVersion returns the version from build info or git
func getBuildVersion() string {
	return "v0.1.0" // Placeholder
}
