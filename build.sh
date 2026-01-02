#!/bin/bash

set -e

echo "🔨 Building Nginx Manager..."

# Build frontend
echo "📦 Building frontend..."
cd frontend
npm install
npm run build
cd ..

# Build Go binary
echo "🔧 Building Go binary..."
go build -o nginx-manager main.go

echo "✅ Build complete!"
echo ""
echo "Run the server with:"
echo "  ./nginx-manager -config /etc/nginx"
echo ""
echo "Or for development:"
echo "  ./nginx-manager -config ./test-config -port 8080"
