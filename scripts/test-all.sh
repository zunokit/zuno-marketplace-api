#!/bin/bash
# scripts/test-all.sh
# Comprehensive test runner with mode selection for hybrid serverless infrastructure

set -e

MODE=${1:-${INFRA_MODE:-docker}}

echo "🧪 Running Tests - Mode: $MODE"
echo "================================"
echo ""

# Export mode for tests
export INFRA_MODE=$MODE
export TEST_MODE=true

echo "1️️️  Unit Tests..."
go test -v -short ./...

echo ""
echo "2️️️  Integration Tests..."
# Skip if in serverless mode without connections
if [ "$MODE" = "serverless" ]; then
    if [ -z "$DATABASE_URL" ] || [ -z "$REDIS_URL" ]; then
        echo "   ⚠️  Skipping (missing connection URLs)"
        echo "   💡 Set DATABASE_URL and REDIS_URL for integration tests"
    else
        go test -v -tags=integration ./...
    fi
else
    # Start Docker services if not running
    if ! docker compose ps | grep -q "postgres.*Up"; then
        echo "   🐳 Starting Docker services..."
        docker compose up -d postgres redis rabbitmq
        sleep 5
    fi
    go test -v -tags=integration ./...
fi

echo ""
echo "3️️️  E2E Tests..."
if [ "$MODE" = "serverless" ]; then
    echo "   ⚠️  Skipping (requires full service deployment)"
    echo "   💡 E2E tests require full service stack"
else
    # Check if tests/e2e directory has tests
    if [ -d "tests/e2e" ] && [ "$(find tests/e2e -name '*_test.go' | wc -l)" -gt 0 ]; then
        go test -v -tags=e2e ./tests/e2e/...
    else
        echo "   ℹ️  No E2E tests found"
    fi
fi

echo ""
echo "================================"
echo "✅ Tests complete!"
echo ""
echo "Summary:"
echo "  Mode: $MODE"
echo "  Unit tests: Passed"
if [ "$MODE" = "serverless" ]; then
    if [ -n "$DATABASE_URL" ] && [ -n "$REDIS_URL" ]; then
        echo "  Integration tests: Passed"
    else
        echo "  Integration tests: Skipped (missing URLs)"
    fi
else
    echo "  Integration tests: Passed"
fi
