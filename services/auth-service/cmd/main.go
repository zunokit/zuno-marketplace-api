package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpcMiddleware "github.com/zunokit/zuno-marketplace-api/shared/observability/middleware"
	obs "github.com/zunokit/zuno-marketplace-api/shared/observability/sentry"
	obsTrace "github.com/zunokit/zuno-marketplace-api/shared/observability/tracing"
	sharedrabbitmq "github.com/zunokit/zuno-marketplace-api/shared/rabbitmq"
	sharedredis "github.com/zunokit/zuno-marketplace-api/shared/redis"

	"github.com/zunokit/zuno-marketplace-api/services/auth-service/internal/client"
	"github.com/zunokit/zuno-marketplace-api/services/auth-service/internal/config"
	"github.com/zunokit/zuno-marketplace-api/services/auth-service/internal/repository"
	"github.com/zunokit/zuno-marketplace-api/services/auth-service/internal/server"
	"github.com/zunokit/zuno-marketplace-api/services/auth-service/internal/service"
	"github.com/zunokit/zuno-marketplace-api/shared/logger"
	pb "github.com/zunokit/zuno-marketplace-api/shared/proto/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// Version and BuildTime are injected via ldflags during build
var (
	Version   = "dev"
	BuildTime = "unknown"
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

	// Initialize Sentry (non-blocking on failure)
	if cfg.Sentry.DSN != "" {
		if err := obs.Init(
			cfg.Sentry.DSN,
			cfg.Sentry.Environment,
			"auth-service",
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
		Logger: gormlogger.Default.LogMode(gormlogger.Info),
	})
	if err != nil {
		log.FatalWithErr(err, "Failed to connect to database")
	}

	// Initialize Redis (non-blocking)
	if err := sharedredis.Init(cfg.Redis.GetAddr()); err != nil {
		log.Printf("Redis init failed (continuing without cache): %v", err)
	} else {
		log.Println("Redis connected")
		defer sharedredis.Close()
	}

	// Initialize RabbitMQ (non-blocking)
	if err := sharedrabbitmq.Init(cfg.RabbitMQ.GetURL()); err != nil {
		log.Printf("RabbitMQ init failed (continuing without events): %v", err)
	} else {
		log.Println("RabbitMQ connected")
		defer sharedrabbitmq.Close()
	}

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
		log.FatalWithErr(err, "Failed to listen on "+cfg.Server.GRPCPort)
	}

	log.Infof("Auth Service listening on %s", cfg.Server.GRPCPort)

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
		log.Info("Auth Service stopped")
	}()

	if err := grpcServer.Serve(listener); err != nil {
		log.FatalWithErr(err, "Failed to serve")
	}
}

// getBuildVersion returns the version injected by build ldflags
func getBuildVersion() string {
	return Version
}
