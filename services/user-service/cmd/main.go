package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/quangdang46/NFT-Marketplace/services/user-service/internal/config"
	"github.com/quangdang46/NFT-Marketplace/services/user-service/internal/repository"
	"github.com/quangdang46/NFT-Marketplace/services/user-service/internal/server"
	"github.com/quangdang46/NFT-Marketplace/shared/database"
	"github.com/quangdang46/NFT-Marketplace/shared/logger"
	pb "github.com/quangdang46/NFT-Marketplace/shared/proto/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	gormlogger "gorm.io/gorm/logger"
)

func main() {
	// Initialize logger
	log := logger.New(&logger.Config{
		Level:       logger.LevelInfo,
		ServiceName: "user-service",
		Pretty:      false,
	})

	log.Info("Starting User Service...")

	// Load configuration
	cfg := config.Load()

	// Initialize database using shared package
	dbConfig := &database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.Database,
		SSLMode:  cfg.Database.SSLMode,
		LogLevel: gormlogger.Info,
	}
	db := database.MustConnect(dbConfig)

	// Initialize repository
	userRepo := repository.NewUserRepository(db)

	// Create gRPC server
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(10*1024*1024), // 10MB
		grpc.MaxSendMsgSize(10*1024*1024), // 10MB
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
		log.FatalWithErr(err, "Failed to listen on "+cfg.Server.GRPCPort)
	}

	log.Infof("User Service listening on %s", cfg.Server.GRPCPort)

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		log.Info("Shutting down User Service...")
		grpcServer.GracefulStop()
		log.Info("User Service stopped")
	}()

	// Start serving
	if err := grpcServer.Serve(listener); err != nil {
		log.FatalWithErr(err, "Failed to serve")
	}
}
