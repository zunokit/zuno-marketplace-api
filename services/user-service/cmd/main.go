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
	obsTrace "github.com/quangdang46/NFT-Marketplace/shared/observability/tracing"

	"github.com/quangdang46/NFT-Marketplace/services/user-service/internal/config"
	"github.com/quangdang46/NFT-Marketplace/services/user-service/internal/repository"
	"github.com/quangdang46/NFT-Marketplace/services/user-service/internal/server"
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
	log.Println("Starting User Service...")

	// Load configuration
	cfg := config.Load()

	// Initialize Sentry
	if cfg.Sentry.DSN != "" {
		if err := obs.Init(
			cfg.Sentry.DSN,
			cfg.Sentry.Environment,
			"user-service",
			getBuildVersion(),
			obsTrace.GetTracesSampleRate(cfg.Sentry.Environment),
		); err != nil {
			log.Printf("Sentry init failed (continuing): %v", err)
		} else {
			log.Println("Sentry initialized")
			defer obs.Flush(2 * time.Second)
		}
	} else {
		log.Println("Sentry DSN not configured, skipping")
	}

	// Initialize database connection
	dsn := cfg.Database.GetDSN()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Database connected successfully")

	// Initialize repository
	userRepo := repository.NewUserRepository(db)

	// Create gRPC server with Sentry interceptor
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(10*1024*1024), // 10MB
		grpc.MaxSendMsgSize(10*1024*1024), // 10MB
		grpc.ChainUnaryInterceptor(
			grpcMiddleware.UnaryServerInterceptor(),
		),
	)

	// Register services
	userServer := server.NewUserServer(userRepo)
	pb.RegisterUserServiceServer(grpcServer, userServer)

	// Register health check
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	// Register reflection for development
	reflection.Register(grpcServer)

	// Start gRPC server
	listener, err := net.Listen("tcp", cfg.Server.GRPCPort)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", cfg.Server.GRPCPort, err)
	}

	log.Printf("User Service listening on %s", cfg.Server.GRPCPort)

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down User Service...")

		// Flush Sentry before shutdown
		if cfg.Sentry.DSN != "" {
			log.Println("Flushing Sentry events...")
			obs.Flush(2 * time.Second)
		}

		grpcServer.GracefulStop()
		log.Println("User Service stopped")
	}()

	// Start serving
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

// getBuildVersion returns the version from build info or git
func getBuildVersion() string {
	return "v0.1.0" // Placeholder
}
