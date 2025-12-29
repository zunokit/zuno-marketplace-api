#!/bin/bash
# scripts/health-check.sh

set -e

echo "🔍 Zuno Marketplace - Infrastructure Health Check"
echo "================================================"
echo ""

# Load environment
if [ ! -f .env ]; then
    echo "❌ .env file not found"
    echo "   Run: ./scripts/setup-env.sh"
    exit 1
fi

source .env

echo "📋 Configuration:"
echo "   INFRA_MODE: ${INFRA_MODE:-not set}"
echo ""

# Check PostgreSQL
echo "🐘 PostgreSQL:"
if [ "$INFRA_MODE" = "serverless" ]; then
    if [ -n "$DATABASE_URL" ]; then
        echo "   ✅ DATABASE_URL set"
        # Extract host from URL
        HOST=$(echo $DATABASE_URL | awk -F'@' '{print $2}' | cut -d':' -f1)
        echo "   Host: $HOST"
    else
        echo "   ❌ DATABASE_URL not set"
    fi
else
    echo "   Host: ${POSTGRES_HOST:-localhost}:${POSTGRES_PORT:-5432}"
fi

# Check Redis
echo ""
echo "🔴 Redis:"
if [ "$INFRA_MODE" = "serverless" ]; then
    if [ -n "$REDIS_URL" ]; then
        echo "   ✅ REDIS_URL set"
        HOST=$(echo $REDIS_URL | awk -F'@' '{print $2}' | cut -d':' -f1)
        echo "   Host: $HOST"
    else
        echo "   ❌ REDIS_URL not set"
    fi
else
    echo "   Host: ${REDIS_HOST:-localhost}:${REDIS_PORT:-6379}"
fi

# Check RabbitMQ
echo ""
echo "🐰 RabbitMQ:"
if [ "$INFRA_MODE" = "serverless" ]; then
    if [ -n "$CLOUDAMQP_URL" ]; then
        echo "   ✅ CLOUDAMQP_URL set"
        HOST=$(echo $CLOUDAMQP_URL | awk -F'@' '{print $2}' | cut -d':' -f1)
        echo "   Host: $HOST"
    else
        echo "   ❌ CLOUDAMQP_URL not set"
    fi
else
    echo "   Host: ${RABBITMQ_HOST:-localhost}:${RABBITMQ_PORT:-5672}"
fi

echo ""
echo "================================================"
echo "✅ Configuration check complete"
echo ""
echo "Next steps:"
if [ "$INFRA_MODE" = "serverless" ]; then
    echo "   1. Run services directly: go run ./services/auth-service/cmd/main.go"
    echo "   2. Or use make build && ./bin/auth-service"
else
    echo "   1. Start Docker: docker compose up -d"
    echo "   2. Check status: docker compose ps"
fi
