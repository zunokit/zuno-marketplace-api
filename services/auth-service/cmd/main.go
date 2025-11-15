package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/quangdang46/NFT-Marketplace/services/auth-service/internal/client"
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
	grpcPort := getEnv("AUTH_GRPC_PORT", ":50051")
	userServiceURL := getEnv("USER_SERVICE_URL", "localhost:50052")
	walletServiceURL := getEnv("WALLET_SERVICE_URL", "localhost:50053")

	dbHost := getEnv("POSTGRES_HOST", "localhost")
	dbPort := getEnv("POSTGRES_PORT", "5432")
	dbUser := getEnv("POSTGRES_USER", "postgres")
	dbPassword := getEnv("POSTGRES_PASSWORD", "postgres")
	dbName := getEnv("POSTGRES_DATABASE", "nft_marketplace")
	dbSSLMode := getEnv("POSTGRES_SSL_MODE", "disable")

	jwtSecret := getEnv("JWT_SECRET", "your-jwt-secret-key-change-in-production")
	refreshSecret := getEnv("REFRESH_SECRET", "your-refresh-secret-key-change-in-production")

	// Initialize database
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)

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

	// Initialize services
	siweService := service.NewSIWEService()
	jwtService := service.NewJWTService(jwtSecret, refreshSecret, 1*time.Hour, 7*24*time.Hour)

	// Initialize gRPC clients
	clients, err := client.NewServiceClients(userServiceURL, walletServiceURL)
	if err != nil {
		log.Fatalf("Failed to initialize gRPC clients: %v", err)
	}
	log.Println("gRPC clients initialized")

	// Create gRPC server
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(10*1024*1024),
		grpc.MaxSendMsgSize(10*1024*1024),
	)

	// Register auth service
	authServer := server.NewAuthServer(nonceRepo, sessionRepo, siweService, jwtService, clients)
	pb.RegisterAuthServiceServer(grpcServer, authServer)

	// Register health check
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	// Register reflection
	reflection.Register(grpcServer)

	// Start gRPC server
	listener, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", grpcPort, err)
	}

	log.Printf("Auth Service listening on %s", grpcPort)

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down Auth Service...")
		grpcServer.GracefulStop()
		log.Println("Auth Service stopped")
	}()

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
