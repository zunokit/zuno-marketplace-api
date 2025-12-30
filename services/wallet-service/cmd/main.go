package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/zunokit/zuno-marketplace-api/services/wallet-service/internal/config"
	"github.com/zunokit/zuno-marketplace-api/services/wallet-service/internal/repository"
	"github.com/zunokit/zuno-marketplace-api/services/wallet-service/internal/server"
	"github.com/zunokit/zuno-marketplace-api/shared/database"
	"github.com/zunokit/zuno-marketplace-api/shared/logger"
	pb "github.com/zunokit/zuno-marketplace-api/shared/proto/pb"
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
		ServiceName: "wallet-service",
		Pretty:      false,
	})

	log.Info("Starting Wallet Service...")

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
	walletRepo := repository.NewWalletRepository(db)

	// Create gRPC server
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(10*1024*1024), // 10MB
		grpc.MaxSendMsgSize(10*1024*1024), // 10MB
	)

	// Register services
	walletServer := server.NewWalletServer(walletRepo)
	pb.RegisterWalletServiceServer(grpcServer, walletServer)

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

	log.Infof("Wallet Service listening on %s", cfg.Server.GRPCPort)

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		log.Info("Shutting down Wallet Service...")
		grpcServer.GracefulStop()
		log.Info("Wallet Service stopped")
	}()

	// Start serving
	if err := grpcServer.Serve(listener); err != nil {
		log.FatalWithErr(err, "Failed to serve")
	}
}
