package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	serviceName := flag.String("name", "", "Name of the service (e.g., user, payment)")
	flag.Parse()

	if *serviceName == "" {
		fmt.Println("Please provide a service name using -name flag")
		fmt.Println("Usage: go run tools/create_service.go -name=<service-name>")
		os.Exit(1)
	}

	// Create service directory structure
	basePath := filepath.Join("services", *serviceName+"-service")
	dirs := []string{
		"cmd",
		"internal/config",
		"internal/server",
		"internal/service",
		"internal/repository",
		"internal/models",
	}

	fmt.Printf("Creating %s service structure...\n", *serviceName)

	// Create directories
	for _, dir := range dirs {
		fullPath := filepath.Join(basePath, dir)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			fmt.Printf("Error creating directory %s: %v\n", dir, err)
			os.Exit(1)
		}
	}

	// Create files
	files := map[string]string{
		"README.md":                   generateReadme(*serviceName),
		".air.toml":                   generateAirToml(),
		"cmd/main.go":                 generateMainGo(*serviceName),
		"internal/config/config.go":   generateConfigGo(*serviceName),
		"internal/config/errors.go":   generateConfigErrors(),
		"internal/server/server.go":   generateServerGo(*serviceName),
		"internal/service/service.go": generateServiceGo(*serviceName),
	}

	for filename, content := range files {
		filePath := filepath.Join(basePath, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			fmt.Printf("Error creating file %s: %v\n", filename, err)
			os.Exit(1)
		}
	}

	fmt.Printf("\n✅ Successfully created %s service!\n\n", *serviceName)
	fmt.Println("📁 Directory structure:")
	fmt.Printf(`
services/%s-service/
├── cmd/
│   └── main.go              # Service entry point
├── internal/
│   ├── config/             # Configuration management
│   │   ├── config.go
│   │   └── errors.go
│   ├── server/             # gRPC server implementation
│   │   └── server.go
│   ├── service/            # Business logic
│   │   └── service.go
│   ├── repository/         # Data access layer
│   └── models/             # Domain models
├── .air.toml               # Hot reload configuration
└── README.md               # Setup documentation

`, *serviceName)

	fmt.Println("🚀 Next steps:")
	fmt.Printf(`
1. Read the setup guide:
   services/%s-service/README.md

2. Add environment variables to root .env.development
   (See README.md section "1. Environment Variables Setup")

3. Add build commands to root Makefile
   (See README.md section "2. Makefile Setup")

4. Update proto files if needed:
   proto/%s.proto

5. Generate protobuf code:
   make proto

6. Start development:
   ./scripts/dev.sh %s

`, *serviceName, *serviceName, *serviceName)
}

func generateReadme(serviceName string) string {
	serviceTitle := strings.Title(serviceName)
	serviceUpper := strings.ToUpper(serviceName)
	return fmt.Sprintf(`# %s Service

This service handles all %s-related operations in the Zuno Marketplace API.

## Architecture

The service follows Clean Architecture principles with clear separation of concerns:

`+"```"+`
services/%s-service/
├── cmd/                      # Application entry point
│   └── main.go              # Main application setup
├── internal/                # Private application code
│   ├── config/             # Configuration management
│   │   ├── config.go       # Config struct and loader
│   │   └── errors.go       # Config validation errors
│   ├── server/             # gRPC server implementation
│   │   └── server.go       # gRPC handlers
│   ├── service/            # Business logic layer
│   │   └── service.go      # Service implementations
│   ├── repository/         # Data access layer
│   │   └── repository.go   # Database operations
│   └── models/             # Domain models
│       └── models.go       # Data structures
├── .air.toml               # Hot reload configuration
├── .env.example            # Environment template
├── Makefile                # Common commands
└── README.md               # This file
`+"```"+`

## Layer Responsibilities

### 1. **Config Layer** (`+"`internal/config/`"+`)
- Loads configuration from environment variables
- Validates required settings
- Provides typed configuration structs
- Handles configuration errors

### 2. **Server Layer** (`+"`internal/server/`"+`)
- Implements gRPC service handlers
- Request validation
- Response formatting
- Error handling
- Converts between gRPC types and domain models

### 3. **Service Layer** (`+"`internal/service/`"+`)
- Contains business logic
- Coordinates between repositories
- Implements use cases
- Independent of external frameworks

### 4. **Repository Layer** (`+"`internal/repository/`"+`)
- Database operations
- Data persistence
- Query implementation
- Transaction management

### 5. **Models Layer** (`+"`internal/models/`"+`)
- Domain entities
- Database models (GORM)
- Business objects

## Setup

### Prerequisites

- Go 1.25.1 or higher
- PostgreSQL (via Neon serverless)
- Redis (via Upstash)
- RabbitMQ (via CloudAMQP) - if using events
- Air (for hot reload): `+"`go install github.com/air-verse/air@latest`"+`

### 1. Environment Variables Setup

Add these variables to **root `+"`.env.development`"+`** file:

`+"```bash"+`
# ============================================================
# %s Service
# ============================================================
%s_GRPC_PORT=:4XXX              # Choose available port (e.g., :4006)
%s_SERVICE_URL=localhost:4XXX    # Same port as above
`+"```"+`

**Note:** Update the port number (4XXX) to an available port in the 4xxx range.

### 2. Makefile Setup

Add these commands to **root `+"`Makefile`"+`**:

#### In the `+"`build-%s`"+` section (around line 190):

`+"```makefile"+`
build-%s: ## Build %s service
	@echo Building %s-service...
	@$(MKDIR_CMD)
	cd services/%s-service/cmd && go build $(LDFLAGS) -o ../../../build/%s-service$(if $(filter $(OS),Windows_NT),.exe,) .
`+"```"+`

#### Update the `+"`build`"+` target to include your service (around line 196):

`+"```makefile"+`
build: build-auth build-user build-wallet build-collection build-media build-gateway build-%s ## Build all services
	@echo Build complete! Binaries in ./build/
	@echo Version: $(VERSION) BuildTime: $(BUILD_TIME)
`+"```"+`

### 3. Protocol Buffers Setup (Optional)

If this service needs gRPC communication:

1. Create `+"`proto/%s.proto`"+` in the root:

`+"```protobuf"+`
syntax = "proto3";

package %s;

option go_package = "github.com/zunokit/zuno-marketplace-api/shared/proto/pb";

// %s service definition
service %sService {
  // Define your RPC methods here
  // rpc GetExample(GetExampleRequest) returns (GetExampleResponse);
}

// Define your messages here
// message GetExampleRequest {
//   string id = 1;
// }
// 
// message GetExampleResponse {
//   string data = 1;
// }
`+"```"+`

2. Generate protobuf code:
   `+"```bash"+`
   make proto
   `+"```"+`

### 4. Development Scripts

Update `+"`scripts/dev.sh`"+` to include your service in the Air hot-reload script.

Add to the services array:

`+"```bash"+`
# Around line 20-30, add:
"%s")
    start_service "services/%s-service" "%s-service" "$LOG_DIR/%s.log"
    ;;
`+"```"+`

### Development

Start the development server:

`+"```bash"+`
# Start only this service
./scripts/dev.sh %s

# Or start all services
./scripts/dev.sh all
`+"```"+`

### Production Build

Build the binary:

`+"```bash"+`
make build-%s
`+"```"+`

Run the binary:

`+"```bash"+`
./build/%s-service
`+"```"+`

## Testing

Run all tests:

`+"```bash"+`
make test
`+"```"+`

Run tests with coverage:

`+"```bash"+`
go test -v -cover ./...
`+"```"+`

## Integration with Other Services

This service communicates with:

- **Shared Database**: PostgreSQL for data persistence
- **Shared Cache**: Redis for caching
- **Message Queue**: RabbitMQ for event-driven communication
- **gRPC**: For inter-service communication

## Key Benefits

1. **Clean Architecture**: Clear separation between layers
2. **Testability**: Easy to mock dependencies
3. **Maintainability**: Well-organized code structure
4. **Scalability**: Independent deployment and scaling
5. **Type Safety**: Strong typing with Go and Protocol Buffers
6. **Hot Reload**: Fast development with Air

## Configuration Management

The service uses a centralized configuration pattern:

1. Environment variables are loaded via `+"`shared/env`"+` package
2. Configuration is validated on startup
3. Missing required variables cause startup failure
4. Type-safe configuration structs

## Error Handling

- Configuration errors: Fail fast on startup
- Runtime errors: Logged with context
- gRPC errors: Proper status codes returned
- Database errors: Wrapped with meaningful messages

## Logging

Uses the shared logger (`+"`github.com/zunokit/zuno-marketplace-api/shared/logger`"+`):

- Structured logging with zerolog
- Service name tagging
- Log levels: Debug, Info, Warn, Error, Fatal
- Pretty printing in development

## Health Checks

gRPC health check endpoint is automatically registered:

`+"```bash"+`
grpcurl -plaintext localhost:PORT grpc.health.v1.Health/Check
`+"```"+`

## Contributing

1. Follow the existing code structure
2. Write tests for new features
3. Update documentation
4. Use conventional commits
5. Run tests before committing

## License

Private - Zuno Marketplace API
`, serviceTitle, serviceName, serviceName, serviceUpper, serviceUpper, serviceUpper, serviceName, serviceName, serviceName, serviceName, serviceName, serviceName, serviceName, serviceName, serviceName, serviceTitle, serviceTitle, serviceName, serviceName, serviceName, serviceName, serviceName, serviceName, serviceName)
}

func generateAirToml() string {
	return `root = "."
tmp_dir = "tmp"

[build]
cmd = "go build -ldflags=\"-s -w\" -o ./tmp/main.exe ./cmd/"
bin = "tmp/main.exe"
include_ext = ["go", "tpl", "tmpl", "html"]
exclude_dir = ["tmp", "vendor", "build", "logs"]
include_dir = ["cmd", "internal", "../../shared/env", "../../shared/database", "../../shared/logger", "../../shared/config", "../../shared/proto"]
exclude_file = []
exclude_unchanged = false
follow_symlink = false
delay = 1000
stop_on_root = false
rerun = false
rerun_delay = 500

[log]
time = true
main_only = false

[color]
main = "magenta"
builder = "yellow"
runner = "green"

[misc]
clean_on_exit = false
`
}

func generateMainGo(serviceName string) string {
	serviceTitle := toTitle(serviceName)
	return fmt.Sprintf(`package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/zunokit/zuno-marketplace-api/services/%s-service/internal/config"
	"github.com/zunokit/zuno-marketplace-api/services/%s-service/internal/server"
	"github.com/zunokit/zuno-marketplace-api/services/%s-service/internal/service"
	"github.com/zunokit/zuno-marketplace-api/shared/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Initialize logger
	log := logger.New(&logger.Config{
		Level:       logger.LevelInfo,
		ServiceName: "%s-service",
		Pretty:      true, // Set to false in production
	})

	log.Info("Starting %s Service...")

	// Load configuration
	cfg := config.Load()

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.FatalWithErr(err, "Configuration validation failed")
	}

	// TODO: Initialize database connection if needed
	// db, err := database.NewPostgres(cfg.Database.URL)
	// if err != nil {
	//     log.FatalWithErr(err, "Failed to connect to database")
	// }
	// defer db.Close()

	// TODO: Initialize Redis client if needed
	// redisClient, err := redis.NewClient(cfg.Redis.URL)
	// if err != nil {
	//     log.FatalWithErr(err, "Failed to connect to Redis")
	// }
	// defer redisClient.Close()

	// Initialize service layer
	%sService := service.New%sService()

	// Create gRPC server
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(10*1024*1024), // 10MB
		grpc.MaxSendMsgSize(10*1024*1024), // 10MB
	)

	// Create server
	_ = server.New%sServer(%sService, *log)
	// TODO: Uncomment after creating proto file and running 'make proto'
	// import pb "github.com/zunokit/zuno-marketplace-api/shared/proto/pb"
	// pb.Register%sServiceServer(grpcServer, %sServer)

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

	log.Infof("%s Service listening on %%s", cfg.Server.GRPCPort)

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		log.Info("Shutting down %s Service...")
		grpcServer.GracefulStop()
		log.Info("%s Service stopped")
	}()

	// Start serving
	if err := grpcServer.Serve(listener); err != nil {
		log.FatalWithErr(err, "Failed to serve")
	}
}
`, serviceName, serviceName, serviceName, serviceName, serviceTitle, serviceName, serviceTitle, serviceTitle, serviceName, serviceTitle, serviceTitle, serviceTitle, serviceTitle, serviceTitle)
}

func generateConfigGo(serviceName string) string {
	serviceUpper := strings.ToUpper(serviceName)
	return fmt.Sprintf(`package config

import (
	sharedConfig "github.com/zunokit/zuno-marketplace-api/shared/config"
	"github.com/zunokit/zuno-marketplace-api/shared/env"
)

// Config holds all configuration for the %s Service
type Config struct {
	Server   sharedConfig.ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	// Add other configuration sections as needed
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	URL string
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	URL string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Server: sharedConfig.ServerConfig{
			GRPCPort: env.GetString("%s_GRPC_PORT", ":50051"),
		},
		Database: DatabaseConfig{
			URL: env.GetString("DATABASE_URL", ""),
		},
		Redis: RedisConfig{
			URL: env.GetString("REDIS_URL", ""),
		},
	}
}

// Validate checks if required configuration is present
func (c *Config) Validate() error {
	// Uncomment validation as needed
	// if c.Database.URL == "" {
	//     return ErrMissingDatabaseURL
	// }
	// if c.Redis.URL == "" {
	//     return ErrMissingRedisURL
	// }
	return nil
}
`, serviceUpper, serviceUpper)
}

func generateConfigErrors() string {
	return `package config

// Common configuration errors
var (
	ErrMissingDatabaseURL = NewConfigError("DATABASE_URL environment variable is required")
	ErrMissingRedisURL    = NewConfigError("REDIS_URL environment variable is required")
	// Add more configuration errors as needed
)

// ConfigError represents a configuration validation error
type ConfigError struct {
	message string
}

func NewConfigError(message string) error {
	return &ConfigError{message: message}
}

func (e *ConfigError) Error() string {
	return e.message
}
`
}

func generateServerGo(serviceName string) string {
	serviceTitle := toTitle(serviceName)
	return fmt.Sprintf(`package server

import (
	"github.com/zunokit/zuno-marketplace-api/services/%s-service/internal/service"
	"github.com/zunokit/zuno-marketplace-api/shared/logger"
)

// %sServer implements the gRPC %sService
type %sServer struct {
	// TODO: Uncomment after creating proto file and running 'make proto'
	// pb.Unimplemented%sServiceServer
	service *service.%sService
	logger  logger.Logger
}

// New%sServer creates a new instance of %sServer
func New%sServer(service *service.%sService, logger logger.Logger) *%sServer {
	return &%sServer{
		service: service,
		logger:  logger,
	}
}

// TODO: Implement gRPC methods here
// After creating proto file and running 'make proto', implement methods like:
// 
// import (
//     "context"
//     pb "github.com/zunokit/zuno-marketplace-api/shared/proto/pb"
//     "google.golang.org/grpc/codes"
//     "google.golang.org/grpc/status"
// )
//
// func (s *%sServer) GetExample(ctx context.Context, req *pb.GetExampleRequest) (*pb.GetExampleResponse, error) {
//     s.logger.Info("GetExample called")
//     
//     // Call service layer
//     result, err := s.service.GetExample(ctx, req.GetId())
//     if err != nil {
//         s.logger.ErrorWithErr(err, "Failed to get example")
//         return nil, status.Error(codes.Internal, "failed to get example")
//     }
//     
//     return &pb.GetExampleResponse{
//         Data: result,
//     }, nil
// }
`, serviceName, serviceTitle, serviceTitle, serviceTitle, serviceTitle, serviceTitle, serviceTitle, serviceTitle, serviceTitle, serviceTitle, serviceTitle, serviceTitle, serviceTitle)
}

func generateServiceGo(serviceName string) string {
	serviceTitle := toTitle(serviceName)
	return fmt.Sprintf(`package service

// %sService handles business logic for %s operations
type %sService struct {
	// Add dependencies here (repositories, other services, etc.)
}

// New%sService creates a new instance of %sService
func New%sService() *%sService {
	return &%sService{
		// Initialize dependencies
	}
}

// TODO: Implement business logic methods here
// Example:
// import "context"
// func (s *%sService) GetExample(ctx context.Context, id string) (interface{}, error) {
//     // Implement business logic
//     return nil, nil
// }
`, serviceTitle, serviceName, serviceTitle, serviceTitle, serviceTitle, serviceTitle, serviceTitle, serviceTitle, serviceTitle)
}

// Helper function to convert service name to title case
func toTitle(s string) string {
	parts := strings.Split(s, "-")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}
