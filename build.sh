#!/bin/bash

set -e

echo "🔨 Building Nginx Config Editor..."

# Build frontend
echo "📦 Building frontend..."
cd frontend
npm install
npm run build
cd ..

# Build Go binary
echo "🔧 Building Go binary..."
go build -o nginx-editor main.go

echo "✅ Build complete!"
echo ""
echo "Run the server with:"
echo "  ./nginx-editor -config /etc/nginx"
echo ""
echo "Or for development:"
echo "  ./nginx-editor -config ./test-config -port 8080"
