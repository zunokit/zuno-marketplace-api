#!/bin/bash
# scripts/test-smoke.sh
# Quick validation connection test for hybrid serverless infrastructure

set -e

echo "🔥 Smoke Test - Quick Validation"
echo "================================="
echo ""

# Load environment
if [ ! -f .env ]; then
    echo "❌ .env not found"
    echo "   Run: ./scripts/setup-env.sh"
    exit 1
fi

source .env

echo "📋 Mode: ${INFRA_MODE:-docker}"
echo ""

# Test PostgreSQL
echo "🐘 Testing PostgreSQL..."
if [ "$INFRA_MODE" = "serverless" ]; then
    # Test with psql if available
    if command -v psql &> /dev/null; then
        psql "$DATABASE_URL" -c "SELECT 1" > /dev/null 2>&1
        echo "   ✅ Connection successful"
    else
        echo "   ⚠️  psql not installed, skipping connection test"
        echo "   💡 DATABASE_URL is set: ${DATABASE_URL:0:20}..."
    fi
else
    # Test Docker container
    docker exec postgres pg_isready -U postgres > /dev/null 2>&1
    echo "   ✅ Container ready"
fi

# Test Redis
echo ""
echo "🔴 Testing Redis..."
if [ "$INFRA_MODE" = "serverless" ]; then
    if command -v redis-cli &> /dev/null; then
        redis-cli -u "$REDIS_URL" PING > /dev/null 2>&1
        echo "   ✅ Connection successful"
    else
        echo "   ⚠️  redis-cli not installed, skipping connection test"
        echo "   💡 REDIS_URL is set: ${REDIS_URL:0:20}..."
    fi
else
    docker exec redis redis-cli PING > /dev/null 2>&1
    echo "   ✅ Container ready"
fi

# Test RabbitMQ
echo ""
echo "🐰 Testing RabbitMQ..."
if [ "$INFRA_MODE" = "serverless" ]; then
    # Can't easily test without amqp tools
    echo "   ⚠️  Skip (requires app to test)"
    echo "   💡 CLOUDAMQP_URL is set: ${CLOUDAMQP_URL:0:20}..."
else
    curl -f http://localhost:15672 > /dev/null 2>&1
    echo "   ✅ Management UI accessible"
fi

echo ""
echo "================================="
echo "✅ Smoke test passed!"
echo ""
echo "Ready to run full test suite:"
echo "  make test"
echo "  ./scripts/test-all.sh"
