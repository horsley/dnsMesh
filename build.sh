#!/bin/bash

# DNSMesh Docker Build Script
# This script replaces nixpacks with Docker-based builds
# Uses Tencent Cloud Alpine mirrors for faster package installation

set -e

echo "🏗️  Building DNSMesh application with Docker..."
echo "📡 Using Tencent Cloud mirrors for faster package installation..."
echo "🚀 Using goproxy.cn for faster Go module downloads..."

# Build the Docker image
echo "📦 Building Docker image..."
docker build -t dnsmesh:latest .

echo "✅ Build completed successfully!"
echo ""
echo "🚀 To run the application:"
echo "   docker-compose up -d"
echo ""
echo "🔍 To view logs:"
echo "   docker-compose logs -f"
echo ""
echo "🛑 To stop the application:"
echo "   docker-compose down"