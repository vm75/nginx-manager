#!/bin/bash

# Development script - runs backend and frontend separately

echo "🚀 Starting development servers..."

# Create test config directory if it doesn't exist
if [ ! -d "test-config" ]; then
    echo "📁 Creating test-config directory..."
    mkdir -p test-config
    echo "# Test nginx configuration" > test-config/nginx.conf
    echo "server {" >> test-config/nginx.conf
    echo "    listen 80;" >> test-config/nginx.conf
    echo "    server_name localhost;" >> test-config/nginx.conf
    echo "    root /var/www/html;" >> test-config/nginx.conf
    echo "    index index.html;" >> test-config/nginx.conf
    echo "}" >> test-config/nginx.conf
fi

# Start backend in background
echo "🔧 Starting Go backend on :8080..."
go run main.go -config ./test-config -port 8080 &
BACKEND_PID=$!

# Wait a moment for backend to start
sleep 2

# Start frontend dev server
echo "📦 Starting Vite dev server on :5173..."
cd frontend
npm run dev

# Cleanup on exit
trap "kill $BACKEND_PID" EXIT
