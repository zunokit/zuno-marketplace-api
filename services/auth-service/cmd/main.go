package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/zunokit/zuno-marketplace-api/services/auth-service/internal/client"
	"github.com/zunokit/zuno-marketplace-api/services/auth-service/internal/config"
	"github.com/zunokit/zuno-marketplace-api/services/auth-service/internal/repository"
	"github.com/zunokit/zuno-marketplace-api/services/auth-service/internal/server"
	"github.com/zunokit/zuno-marketplace-api/services/auth-service/internal/service"
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
		ServiceName: "auth-service",
		Pretty:      false,
	})

	log.Info("Starting Auth Service...")

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
		log.FatalWithErr(err, "Failed to initialize gRPC clients")
	}
	log.Info("gRPC clients initialized")

	// Create gRPC server
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(10*1024*1024),
		grpc.MaxSendMsgSize(10*1024*1024),
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
		log.FatalWithErr(err, "Failed to listen on "+cfg.Server.GRPCPort)
	}

	log.Infof("Auth Service listening on %s", cfg.Server.GRPCPort)

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Info("Shutting down Auth Service...")
		grpcServer.GracefulStop()
		log.Info("Auth Service stopped")
	}()

	if err := grpcServer.Serve(listener); err != nil {
		log.FatalWithErr(err, "Failed to serve")
	}
}
