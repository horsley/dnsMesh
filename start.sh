#!/bin/bash

echo "🚀 启动 DNSMesh 系统..."

# 检查 Docker
if ! command -v podman &> /dev/null; then
    echo "❌ Docker 未安装，请先安装 Docker"
    exit 1
fi

# 启动数据库
echo "📦 启动 PostgreSQL 数据库..."
docker-compose up -d postgres

# 等待数据库就绪
echo "⏳ 等待数据库就绪..."
sleep 5

# 安装后端依赖
echo "📥 安装后端依赖..."
cd backend
if command -v go &> /dev/null; then
    go mod tidy
else
    echo "⚠️  Go 未安装，跳过后端依赖安装"
fi

# 启动后端
echo "🔧 启动后端服务..."
if command -v go &> /dev/null; then
    go run cmd/main.go &
    BACKEND_PID=$!
    echo "✅ 后端服务已启动 (PID: $BACKEND_PID)"
else
    echo "❌ Go 未安装，无法启动后端"
    exit 1
fi

# 安装前端依赖并启动
echo "🎨 安装前端依赖..."
cd ../frontend
if command -v npm &> /dev/null; then
    npm install
    echo "🌐 启动前端开发服务器..."
    npm run dev
else
    echo "❌ Node.js/npm 未安装，无法启动前端"
    kill $BACKEND_PID
    exit 1
fi
