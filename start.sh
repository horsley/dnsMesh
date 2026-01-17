#!/bin/bash

echo "🚀 启动 DNSMesh 系统..."

# 准备 SQLite 数据库（如需要则从 Postgres 迁移）
ROOT_DIR="$(pwd)"
SQLITE_PATH="${SQLITE_PATH:-$ROOT_DIR/backend/data/dnsmesh.db}"
if [ ! -f "$SQLITE_PATH" ]; then
    echo "🗄️  未找到 SQLite 数据库，尝试从 Postgres 迁移..."
    if command -v go &> /dev/null; then
        if ! (cd backend && SQLITE_PATH="$SQLITE_PATH" go run cmd/migrate_sqlite/main.go); then
            echo "❌ 迁移失败，请检查 Postgres 连接配置"
            exit 1
        fi
    else
        echo "❌ Go 未安装，无法执行迁移"
        exit 1
    fi
fi

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
    SQLITE_PATH="$SQLITE_PATH" go run cmd/main.go &
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
