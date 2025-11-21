# Tiltfile for Zuno NFT Marketplace API
# Hot reload development environment with Kubernetes

# Load extensions
load('ext://restart_process', 'docker_build_with_restart')
load('ext://helm_remote', 'helm_remote')

# Configuration
allow_k8s_contexts('docker-desktop')

# Set default namespace
k8s_namespace = 'dev'

# Environment config
config.define_string('namespace', args=False, usage='Kubernetes namespace to use')
cfg = config.parse()
if cfg.get('namespace'):
    k8s_namespace = cfg.get('namespace')

# Ensure namespace exists
local('kubectl create namespace {} --dry-run=client -o yaml | kubectl apply -f -'.format(k8s_namespace))

# Apply secrets (if file exists)
local('kubectl apply -f infra/development/k8s/secrets.yaml --namespace={} || true'.format(k8s_namespace))

# Set kubectl context
k8s_yaml('infra/development/k8s/app-config.yaml', allow_duplicates=True)

# ===================================
# Infrastructure Services
# ===================================

# PostgreSQL
k8s_yaml('infra/development/k8s/postgres.yaml')
k8s_resource(
    'postgres',
    port_forwards=['5433:5432'],
    labels=['infrastructure'],
    resource_deps=[]
)

# Redis
k8s_yaml('infra/development/k8s/redis.yaml')
k8s_resource(
    'redis',
    port_forwards=['6379:6379'],
    labels=['infrastructure'],
    resource_deps=[]
)

# RabbitMQ
k8s_yaml('infra/development/k8s/rabbitmq.yaml')
k8s_resource(
    'rabbitmq',
    port_forwards=[
        '5672:5672',   # AMQP
        '15672:15672'  # Management UI
    ],
    labels=['infrastructure'],
    resource_deps=[]
)

# ===================================
# Application Services
# ===================================

# Helper function to build service
def build_service(name, path, port, deps=['postgres', 'redis', 'rabbitmq']):
    # Build Docker image with live update
    docker_build(
        '{}'.format(name),
        '.',
        dockerfile='infra/development/docker/{}.Dockerfile'.format(name),
        only=[
            './shared',
            './services/{}'.format(path),
            './proto',
            './go.mod',
            './go.sum',
        ],
        live_update=[
            sync('./shared', '/app/shared'),
            sync('./services/{}'.format(path), '/app/services/{}'.format(path)),
            sync('./proto', '/app/proto'),
            sync('./go.mod', '/app/go.mod'),
            sync('./go.sum', '/app/go.sum'),
            run('cd /app && go mod download', trigger=['./go.mod', './go.sum']),
        ],
    )

    # Deploy to k8s
    k8s_yaml('infra/development/k8s/{}-deployment.yaml'.format(name))

    # Configure resource
    k8s_resource(
        name,
        port_forwards=[port] if port else [],
        labels=['services'],
        resource_deps=deps,
        auto_init=True,
        trigger_mode=TRIGGER_MODE_AUTO
    )

# Auth Service
build_service(
    'auth-service',
    'auth-service',
    '50051:50051',
    deps=['postgres', 'redis', 'rabbitmq', 'user-service', 'wallet-service']
)

# User Service
build_service(
    'user-service',
    'user-service',
    '50052:50052',
    deps=['postgres', 'redis', 'rabbitmq']
)

# Wallet Service
build_service(
    'wallet-service',
    'wallet-service',
    '50053:50053',
    deps=['postgres', 'redis', 'rabbitmq']
)

# Collection Service
build_service(
    'collection-service',
    'collection-service',
    '50054:50054',
    deps=['postgres', 'redis', 'rabbitmq']
)

# GraphQL Gateway
docker_build(
    'graphql-gateway',
    '.',
    dockerfile='infra/development/docker/graphql-gateway.Dockerfile',
    only=[
        './shared',
        './services/graphql-gateway',
        './proto',
        './go.mod',
        './go.sum',
    ],
    live_update=[
        sync('./shared', '/app/shared'),
        sync('./services/graphql-gateway', '/app/services/graphql-gateway'),
        sync('./proto', '/app/proto'),
        sync('./go.mod', '/app/go.mod'),
        sync('./go.sum', '/app/go.sum'),
        run('cd /app && go mod download', trigger=['./go.mod', './go.sum']),
    ],
)

k8s_yaml('infra/development/k8s/graphql-gateway-deployment.yaml')
k8s_resource(
    'graphql-gateway',
    port_forwards=['8081:8081'],
    labels=['services'],
    resource_deps=['auth-service', 'user-service', 'wallet-service'],
    auto_init=True,
    trigger_mode=TRIGGER_MODE_AUTO
)

# ===================================
# Development Helpers
# ===================================

# Add local scripts for common tasks
local_resource(
    'proto-gen',
    'make generate-proto',
    deps=['proto'],
    labels=['helpers'],
    auto_init=False,
    trigger_mode=TRIGGER_MODE_MANUAL
)

local_resource(
    'test-all',
    'go test ./...',
    deps=['services', 'shared'],
    labels=['helpers'],
    auto_init=False,
    trigger_mode=TRIGGER_MODE_MANUAL
)

local_resource(
    'db-migrate',
    'echo "Run migrations manually"',
    labels=['helpers'],
    auto_init=False,
    trigger_mode=TRIGGER_MODE_MANUAL
)

# ===================================
# UI Configuration
# ===================================

# Group services by type
update_settings(
    k8s_upsert_timeout_secs=60,
    suppress_unused_image_warnings=['auth-service', 'user-service', 'wallet-service', 'graphql-gateway']
)

print("""
╔══════════════════════════════════════════════════════════════╗
║  Zuno NFT Marketplace - Development Environment             ║
╚══════════════════════════════════════════════════════════════╝

🚀 Services Starting...

📦 Infrastructure:
   - PostgreSQL:     localhost:5432
   - Redis:          localhost:6379
   - RabbitMQ:       localhost:5672
   - RabbitMQ UI:    http://localhost:15672 (guest/guest)

🔧 Application Services:
   - Auth Service:       localhost:50051 (gRPC)
   - User Service:       localhost:50052 (gRPC)
   - Wallet Service:     localhost:50053 (gRPC)
   - GraphQL Gateway:    http://localhost:8081 (HTTP/WS)

🛠️  Helper Commands:
   - proto-gen:     Generate protobuf code
   - test-all:      Run all tests
   - db-migrate:    Database migrations

📝 Quick Commands:
   - tilt up          Start all services
   - tilt down        Stop all services
   - tilt logs [svc]  View service logs
   - tilt trigger     Manual trigger builds

🔥 Hot Reload Active - Edit code and see changes instantly!
""")
