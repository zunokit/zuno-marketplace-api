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

	"github.com/zunokit/zuno-marketplace-api/services/wallet-service/internal/config"
	"github.com/zunokit/zuno-marketplace-api/services/wallet-service/internal/repository"
	"github.com/zunokit/zuno-marketplace-api/services/wallet-service/internal/server"
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
		ServiceName: "wallet-service",
		Pretty:      false,
	})

	log.Info("Starting Wallet Service...")

	// Load configuration
	cfg := config.Load()

	// Initialize Sentry
	if cfg.Sentry.DSN != "" {
		if err := obs.Init(
			cfg.Sentry.DSN,
			cfg.Sentry.Environment,
			"wallet-service",
			getBuildVersion(),
			obsTrace.GetTracesSampleRate(cfg.Sentry.Environment),
		); err != nil {
			log.Infof("Sentry init failed (continuing): %v", err)
		} else {
			log.Info("Sentry initialized")
			defer obs.Flush(2 * time.Second)
		}
	} else {
		log.Info("Sentry DSN not configured, skipping")
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
		log.Infof("Redis init failed (continuing without cache): %v", err)
	} else {
		log.Info("Redis connected")
		defer sharedredis.Close()
	}

	// Initialize RabbitMQ (non-blocking)
	if err := sharedrabbitmq.Init(cfg.RabbitMQ.GetURL()); err != nil {
		log.Infof("RabbitMQ init failed (continuing without events): %v", err)
	} else {
		log.Info("RabbitMQ connected")
		defer sharedrabbitmq.Close()
	}

	// Initialize repository
	walletRepo := repository.NewWalletRepository(db)

	// Create gRPC server with Sentry interceptor
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(10*1024*1024), // 10MB
		grpc.MaxSendMsgSize(10*1024*1024), // 10MB
		grpc.ChainUnaryInterceptor(
			grpcMiddleware.UnaryServerInterceptor(),
		),
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

		// Flush Sentry before shutdown
		if cfg.Sentry.DSN != "" {
			log.Info("Flushing Sentry events...")
			obs.Flush(2 * time.Second)
		}

		grpcServer.GracefulStop()
		log.Info("Wallet Service stopped")
	}()

	// Start serving
	if err := grpcServer.Serve(listener); err != nil {
		log.FatalWithErr(err, "Failed to serve")
	}
}

// getBuildVersion returns the version injected by build ldflags
func getBuildVersion() string {
	return Version
}
