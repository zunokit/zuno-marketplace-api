#!/bin/bash
# Zuno Marketplace - Environment Setup Script
# This script helps you set up your .env file for either serverless or Docker mode

set -e

echo "=================================================="
echo "  Zuno Marketplace - Environment Setup"
echo "=================================================="
echo ""

# Check if .env exists
if [ -f .env ]; then
    echo "⚠️  .env file already exists"
    read -p "Overwrite? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "❌ Aborted"
        exit 1
    fi
    rm .env
    echo "🗑️  Removed existing .env file"
fi

# Ask for infrastructure mode
echo ""
echo "Select infrastructure mode:"
echo "  1) serverless - Supabase, Upstash, CloudAMQP (free, no Docker required)"
echo "  2) docker     - Local containers (resource intensive)"
echo ""
read -p "Choice [1-2]: " mode_choice

case $mode_choice in
    1)
        echo ""
        echo "📋 Copying .env.development.example → .env"
        cp .env.development.example .env
        echo "✅ Serverless environment configured"
        echo ""
        echo "⚠️  IMPORTANT: Edit .env and add your connection strings:"
        echo "   - DATABASE_URL (from Supabase Project → Settings → Database)"
        echo "   - REDIS_URL (from Upstash Dashboard → Database → Details)"
        echo "   - CLOUDAMQP_URL (from CloudAMQP Dashboard → Instance → Details)"
        echo ""
        echo "📖 Setup guides:"
        echo "   - Supabase: https://supabase.com/docs/guides/getting-started"
        echo "   - Upstash: https://upstash.com/docs/redis/quickstart/redis"
        echo "   - CloudAMQP: https://www.cloudamqp.com/docs/how-to-connection-url.html"
        ;;
    2)
        echo ""
        echo "📋 Copying .env.production.example → .env"
        cp .env.production.example .env
        echo "✅ Docker environment configured"
        echo ""
        echo "🐳 Start services with: docker compose up -d"
        echo "📊 Check status with: docker compose ps"
        ;;
    *)
        echo "❌ Invalid choice"
        exit 1
        ;;
esac

echo ""
echo "✨ Setup complete!"
echo ""
echo "Next steps:"
echo "  1. Edit .env file with your actual values"
echo "  2. Run services (depending on mode):"
echo "     - Serverless: go run ./services/..."
echo "     - Docker: docker compose up -d"
echo ""
