package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/quangdang46/NFT-Marketplace/services/collection-service/internal/config"
	"github.com/quangdang46/NFT-Marketplace/services/collection-service/internal/repository"
	"github.com/quangdang46/NFT-Marketplace/services/collection-service/internal/server"
	"github.com/quangdang46/NFT-Marketplace/services/collection-service/internal/service"
	"github.com/quangdang46/NFT-Marketplace/shared/proto/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	log.Println("Starting Collection Service...")

	// Load configuration
	cfg := config.Load()

	// Initialize database connection
	db, err := gorm.Open(postgres.Open(cfg.Database.GetDSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Database connected successfully")

	// Initialize repositories
	collectionRepo := repository.NewCollectionRepository(db)
	allowlistRepo := repository.NewAllowlistRepository(db)
	metadataRepo := repository.NewMetadataRepository(db)

	// Initialize service
	collectionSvc := service.NewCollectionService(collectionRepo, allowlistRepo, metadataRepo)

	// Create gRPC server
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(10*1024*1024), // 10MB
		grpc.MaxSendMsgSize(10*1024*1024), // 10MB
	)

	// Register services
	collectionServer := server.NewCollectionServer(collectionSvc)
	pb.RegisterCollectionServiceServer(grpcServer, collectionServer)

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

	log.Printf("Collection Service listening on %s", cfg.Server.GRPCPort)

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down Collection Service...")
		grpcServer.GracefulStop()
		log.Println("Collection Service stopped")
	}()

	// Start serving
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
