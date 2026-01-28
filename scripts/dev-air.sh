#!/bin/bash
# Start services with Air hot-reload
# Usage: ./scripts/dev-air.sh [all|auth|user|wallet|gateway]

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SERVICES_DIR="$PROJECT_ROOT/services"
LOGS_DIR="$PROJECT_ROOT/logs"
PID_DIR="$PROJECT_ROOT/logs"

# Ensure logs directory exists
mkdir -p "$LOGS_DIR"

# Load environment variables from .env.development
if [ -f "$PROJECT_ROOT/.env.development" ]; then
  set -a
  source "$PROJECT_ROOT/.env.development"
  set +a
else
  echo "ERROR: .env.development not found"
  exit 1
fi

# Service configurations (name:port)
declare -A SERVICES=(
  ["auth-service"]="4001"
  ["user-service"]="4002"
  ["wallet-service"]="4003"
  ["graphql-gateway"]="4080"
)

# Function to print colored output
print_info() {
  echo -e "${BLUE}ℹ${NC} $1"
}

print_success() {
  echo -e "${GREEN}✓${NC} $1"
}

print_error() {
  echo -e "${RED}✗${NC} $1"
}

print_header() {
  echo
  echo -e "${YELLOW}============================================================${NC}"
  echo -e "${YELLOW}  $1${NC}"
  echo -e "${YELLOW}============================================================${NC}"
  echo
}

# Function to start a single service
start_service() {
  local service=$1
  local port=$2
  local log_file="$LOGS_DIR/$service.log"
  local pid_file="$PID_DIR/$service.pid"

  # Check if already running
  if [ -f "$pid_file" ]; then
    local pid=$(cat "$pid_file")
    if ps -p "$pid" > /dev/null 2>&1; then
      print_info "$service already running (PID: $pid)"
      return 0
    else
      rm -f "$pid_file"
    fi
  fi

  print_info "Starting $service on port $port..."

  # Navigate to service directory
  cd "$SERVICES_DIR/$service"

  # Start Air in background with environment
  env $(cat "$PROJECT_ROOT/.env.development" | grep -v '^#' | grep -v '^$' | xargs) nohup air > "$log_file" 2>&1 &
  local pid=$!

  # Save PID
  echo $pid > "$pid_file"

  cd "$PROJECT_ROOT"

  # Wait briefly for startup
  sleep 2

  # Verify process started
  if ps -p $pid > /dev/null 2>&1; then
    print_success "$service started (PID: $pid)"
    return 0
  else
    print_error "$service failed to start. Check $log_file"
    return 1
  fi
}

# Parse arguments
SERVICE_ARG="${1:-all}"

print_header "Starting Air Development Environment"

# Validate Air is installed
if ! command -v air &> /dev/null; then
  print_error "Air is not installed"
  echo
  echo "Install with: go install github.com/air-verse/air@latest"
  exit 1
fi

# Validate .env file exists
if [ ! -f "$PROJECT_ROOT/.env.development" ]; then
  print_error ".env.development not found"
  echo
  echo "Create from example: cp .env.development.example .env.development"
  echo "Then edit with your credentials"
  exit 1
fi

# Start services based on argument
case "$SERVICE_ARG" in
  all)
    print_info "Starting all services..."
    for service in "${!SERVICES[@]}"; do
      start_service "$service" "${SERVICES[$service]}"
    done
    ;;
  auth|auth-service)
    start_service "auth-service" "${SERVICES[auth-service]}"
    ;;
  user|user-service)
    start_service "user-service" "${SERVICES[user-service]}"
    ;;
  wallet|wallet-service)
    start_service "wallet-service" "${SERVICES[wallet-service]}"
    ;;
  gateway|graphql-gateway)
    start_service "graphql-gateway" "${SERVICES[graphql-gateway]}"
    ;;
  *)
    print_error "Unknown service: $SERVICE_ARG"
    echo
    echo "Usage: $0 [all|auth|user|wallet|gateway]"
    exit 1
    ;;
esac

# Print summary
print_header "Services Running"

for service in "${!SERVICES[@]}"; do
  pid_file="$PID_DIR/$service.pid"
  if [ -f "$pid_file" ]; then
    pid=$(cat "$pid_file")
    if ps -p "$pid" > /dev/null 2>&1; then
      echo -e "  ${GREEN}●${NC} $service - Port ${SERVICES[$service]} (PID: $pid)"
    else
      echo -e "  ${RED}○${NC} $service - Stopped"
    fi
  else
    echo -e "  ${YELLOW}○${NC} $service - Not started"
  fi
done

echo
echo -e "GraphQL Playground: ${GREEN}http://localhost:4080/graphql${NC}"
echo
echo -e "Logs: ${BLUE}$LOGS_DIR${NC}"
echo -e "Stop: ${YELLOW}make dev-air-stop${NC} or ${YELLOW}./scripts/stop-air.sh${NC}"
echo
