package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zunokit/zuno-marketplace-api/services/collection-service/internal/config"
	"github.com/zunokit/zuno-marketplace-api/services/collection-service/internal/repository"
	"github.com/zunokit/zuno-marketplace-api/services/collection-service/internal/server"
	"github.com/zunokit/zuno-marketplace-api/services/collection-service/internal/service"
	"github.com/zunokit/zuno-marketplace-api/shared/database"
	"github.com/zunokit/zuno-marketplace-api/shared/logger"
	grpcMiddleware "github.com/zunokit/zuno-marketplace-api/shared/observability/middleware"
	obs "github.com/zunokit/zuno-marketplace-api/shared/observability/sentry"
	obsTrace "github.com/zunokit/zuno-marketplace-api/shared/observability/tracing"
	"github.com/zunokit/zuno-marketplace-api/shared/proto/pb"
	sharedredis "github.com/zunokit/zuno-marketplace-api/shared/redis"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
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
		ServiceName: "collection-service",
		Pretty:      false,
	})

	log.Info("Starting Collection Service...")

	// Load configuration
	cfg := config.Load()

	// Initialize Sentry (non-blocking)
	if cfg.Sentry.DSN != "" {
		if err := obs.Init(
			cfg.Sentry.DSN,
			cfg.Sentry.Environment,
			"collection-service",
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

	// Initialize database using shared package
	dbConfig := &database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.Database,
		SSLMode:  cfg.Database.SSLMode,
		DSN:      cfg.DatabaseDSN, // Use full DSN for serverless mode
		LogLevel: gormlogger.Info,
	}
	db := database.MustConnect(dbConfig)

	// Initialize Redis (non-blocking - continues without cache if Redis fails)
	redisConfig := config.GetRedisConfig()
	if err := sharedredis.Init(redisConfig.GetAddr()); err != nil {
		log.Infof("Redis init failed (continuing without cache): %v", err)
	} else {
		log.Info("Redis connected")
		defer sharedredis.Close()
	}

	// Create cache instance
	cache := sharedredis.NewCache()

	// Initialize repositories
	baseRepo := repository.NewCollectionRepository(db)
	collectionRepo := repository.NewCachedCollectionRepository(baseRepo, cache)
	allowlistRepo := repository.NewAllowlistRepository(db)
	metadataRepo := repository.NewMetadataRepository(db)
	processedEventRepo := repository.NewProcessedEventRepository(db)

	// Initialize service (uses base repo for writes, cached repo for reads)
	collectionSvc := service.NewCollectionService(baseRepo, allowlistRepo, metadataRepo)

	// Initialize zap logger
	zapLogger, err := zap.NewProduction()
	if err != nil {
		log.FatalWithErr(err, "Failed to initialize zap logger")
	}
	defer zapLogger.Sync()

	// Create gRPC server
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(10*1024*1024), // 10MB
		grpc.MaxSendMsgSize(10*1024*1024), // 10MB
		grpc.ChainUnaryInterceptor(
			grpcMiddleware.UnaryServerInterceptor(),
		),
	)

	// Register services
	collectionServer := server.NewCollectionServer(collectionSvc, processedEventRepo, collectionRepo, zapLogger)
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
		log.FatalWithErr(err, "Failed to listen on "+cfg.Server.GRPCPort)
	}

	log.Infof("Collection Service listening on %s", cfg.Server.GRPCPort)

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		log.Info("Shutting down Collection Service...")

		// Flush Sentry before shutdown
		if cfg.Sentry.DSN != "" {
			log.Info("Flushing Sentry events...")
			obs.Flush(2 * time.Second)
		}

		grpcServer.GracefulStop()
		log.Info("Collection Service stopped")
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
