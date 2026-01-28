#!/bin/bash
# Stop all Air services

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PID_DIR="$PROJECT_ROOT/logs"

echo -e "${YELLOW}============================================================${NC}"
echo -e "${YELLOW}  Stopping Air Services${NC}"
echo -e "${YELLOW}============================================================${NC}"
echo

# Stop services from PID files
if [ -d "$PID_DIR" ]; then
  for pid_file in "$PID_DIR"/*.pid; do
    if [ -f "$pid_file" ]; then
      service=$(basename "$pid_file" .pid)
      pid=$(cat "$pid_file")

      if ps -p "$pid" > /dev/null 2>&1; then
        echo -e "  Stopping ${GREEN}$service${NC} (PID: $pid)..."
        kill "$pid"
        rm -f "$pid_file"
      else
        echo -e "  ${YELLOW}$service${NC} not running (stale PID file)"
        rm -f "$pid_file"
      fi
    fi
  done
fi

# Also kill any remaining Air processes
pkill -f "air" || true

echo
echo -e "${GREEN}✓ All services stopped${NC}"
