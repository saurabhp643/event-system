#!/bin/sh

# entrypoint.sh - Manages both Nginx and Go backend processes

set -e

echo "Starting Event Ingestion System..."

# Function to handle shutdown
shutdown() {
    echo "Shutting down..."
    kill $BACKEND_PID 2>/dev/null || true
    nginx -s quit 2>/dev/null || true
    exit 0
}

# Trap signals for graceful shutdown
trap shutdown SIGTERM SIGINT SIGHUP

# Start the Go backend server in background
echo "Starting Go backend server..."
cd /app/backend
./server &
BACKEND_PID=$!

# Wait a moment for backend to start
sleep 2

# Check if backend started successfully
if ! kill -0 $BACKEND_PID 2>/dev/null; then
    echo "ERROR: Backend server failed to start"
    exit 1
fi

echo "Backend server started (PID: $BACKEND_PID)"

# Start Nginx in foreground
echo "Starting Nginx..."
nginx -g "daemon off;" &

echo "All services started. Application is ready."

# Wait for any process to exit
wait
