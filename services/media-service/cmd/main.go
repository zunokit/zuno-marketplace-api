package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/quangdang46/NFT-Marketplace/services/media-service/internal/client"
	"github.com/quangdang46/NFT-Marketplace/services/media-service/internal/config"
	"github.com/quangdang46/NFT-Marketplace/services/media-service/internal/server"
	"github.com/quangdang46/NFT-Marketplace/services/media-service/internal/service"
	"github.com/quangdang46/NFT-Marketplace/shared/logger"
	pb "github.com/quangdang46/NFT-Marketplace/shared/proto/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Initialize logger
	log := logger.New(&logger.Config{
		Level:       logger.LevelInfo,
		ServiceName: "media-service",
		Pretty:      false,
	})

	log.Info("Starting Media Service...")

	// Load configuration
	cfg := config.Load()

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.FatalWithErr(err, "Configuration validation failed")
	}

	// Initialize Metadata Service client
	log.Infof("Connecting to Metadata Service: %s", cfg.MetadataService.URL)
	metadataClient := client.NewMetadataServiceClient(cfg.MetadataService.URL, cfg.MetadataService.APIKey)

	// Initialize service layer
	mediaService := service.NewMediaService(metadataClient)

	// Create gRPC server
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(50*1024*1024), // 50MB for batch uploads
		grpc.MaxSendMsgSize(10*1024*1024), // 10MB
	)

	// Register services
	mediaServer := server.NewMediaServer(mediaService, log)
	pb.RegisterMediaServiceServer(grpcServer, mediaServer)

	// Register health check
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	// Register reflection for development
	reflection.Register(grpcServer)

	// Start gRPC server
	listener, err := net.Listen("tcp", cfg.Server.GRPCPort)
	if err != nil {
		log.FatalWithErr(err, "Failed to listen on "+cfg.Server.GRPCPort)
	}

	log.Infof("Media Service listening on %s", cfg.Server.GRPCPort)

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		log.Info("Shutting down Media Service...")
		grpcServer.GracefulStop()
		log.Info("Media Service stopped")
	}()

	// Start serving
	if err := grpcServer.Serve(listener); err != nil {
		log.FatalWithErr(err, "Failed to serve")
	}
}
