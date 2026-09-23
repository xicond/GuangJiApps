#!/usr/bin/env bash

set -eo pipefail

echo "=================================================="
echo "         Docker Desktop Recovery Script           "
echo "=================================================="

echo "==> [1/4] Terminating hung Docker CLI commands..."
pkill -9 -f "docker compose" 2>/dev/null || true
pkill -9 -f "docker-compose" 2>/dev/null || true
pkill -9 -f "docker info" 2>/dev/null || true

echo "==> [2/4] Killing frozen Docker background daemons..."
killall -9 "Docker" 2>/dev/null || true
killall -9 "com.docker.backend" 2>/dev/null || true
killall -9 "com.docker.virtualization" 2>/dev/null || true
killall -9 "com.docker.build" 2>/dev/null || true
killall -9 "docker-agent" 2>/dev/null || true

sleep 2

echo "==> [3/4] Launching Docker Desktop application..."
open -a "/Applications/Docker.app"

echo "==> [4/4] Waiting for Docker daemon to become ready (up to 60s)..."
MAX_WAIT=60
ELAPSED=0
READY=false

while [ $ELAPSED -lt $MAX_WAIT ]; do
  if docker info >/dev/null 2>&1; then
    READY=true
    break
  fi
  printf "."
  sleep 2
  ELAPSED=$((ELAPSED + 2))
done
echo ""

if [ "$READY" = true ]; then
  echo "Docker daemon is ready!"
  echo ""
  docker info --format 'Engine Version: {{.ServerVersion}}'
  echo ""
  echo "Current containers:"
  docker ps --format "table {{.ID}}\t{{.Names}}\t{{.Status}}\t{{.Ports}}"
else
  echo "Timed out waiting for Docker daemon to respond."
  echo "Please check Docker Desktop application window or system logs."
  exit 1
fi
