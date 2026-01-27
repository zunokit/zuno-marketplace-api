#!/bin/bash
# Validate .env.development configuration

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="$PROJECT_ROOT/.env.development"

echo -e "${YELLOW}============================================================${NC}"
echo -e "${YELLOW}  Environment Validation${NC}"
echo -e "${YELLOW}============================================================${NC}"
echo

# Check if .env exists
if [ ! -f "$ENV_FILE" ]; then
  echo -e "${RED}✗${NC} .env.development not found"
  echo
  echo "Create with: cp .env.development.example .env.development"
  exit 1
fi

# Load environment
set -a
source "$ENV_FILE"
set +a

# Required variables
errors=0

check_var() {
  local var_name=$1
  local var_value=${!var_name}
  local is_required=${2:-true}

  if [ -z "$var_value" ]; then
    if [ "$is_required" = "true" ]; then
      echo -e "${RED}✗${NC} $var_name is not set"
      ((errors++))
    else
      echo -e "${YELLOW}○${NC} $var_name is optional (not set)"
    fi
  elif [[ "$var_value" == *"["* ]]; then
    echo -e "${YELLOW}○${NC} $var_name contains placeholder (needs real value)"
    ((errors++))
  else
    echo -e "${GREEN}✓${NC} $var_name"
  fi
}

echo "Required Variables:"
check_var "INFRA_MODE"
check_var "ENVIRONMENT"
check_var "AUTH_GRPC_PORT"
check_var "USER_GRPC_PORT"
check_var "WALLET_GRPC_PORT"
check_var "GATEWAY_HTTP_ADDR"
check_var "JWT_SECRET"
check_var "REFRESH_SECRET"

echo
echo "Infrastructure (serverless mode):"

if [ "$INFRA_MODE" = "serverless" ]; then
  check_var "DATABASE_URL" true
  check_var "REDIS_URL" true
  check_var "CLOUDAMQP_URL" true
fi

echo
echo "Optional Variables:"
check_var "ETHEREUM_RPC_URL" false
check_var "POLYGON_RPC_URL" false
check_var "IPFS_API_URL" false
check_var "IPFS_API_KEY" false

echo
if [ $errors -gt 0 ]; then
  echo -e "${RED}✗${NC} Validation failed with $errors error(s)"
  echo
  echo "Please fix the issues above in $ENV_FILE"
  exit 1
else
  echo -e "${GREEN}✓${NC} All required variables are set"
  echo
  exit 0
fi
