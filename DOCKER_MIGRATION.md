# Docker Migration Guide

## 从 Nixpacks 迁移到 Docker

本项目已从使用 `nixpacks` 构建迁移到使用 `Dockerfile` 构建，以降低资源占用并提高构建效率。

## 主要变更

### 1. 新增文件
- `Dockerfile` - 多阶段构建配置，包含前端和后端构建
- `.dockerignore` - 优化构建性能，排除不必要的文件
- `build.sh` - 简化构建过程的脚本
- `DOCKER_MIGRATION.md` - 本迁移文档

### 2. 修改文件
- `docker-compose.yml` - 更新构建上下文为根目录

### 3. 移除依赖
- `nixpacks.toml` - 可以安全删除此文件

## 构建方式对比

### 之前 (Nixpacks)
```bash
nixpacks build
```

### 现在 (Docker)
```bash
# 方式1: 使用构建脚本
./build.sh

# 方式2: 直接使用 Docker
docker build -t dnsmesh:latest .

# 方式3: 使用 docker-compose (推荐)
docker-compose up --build
```

## 优势

1. **资源占用更低** - Docker 构建比 Nixpacks 更轻量
2. **构建速度更快** - 利用 Docker 层缓存机制和腾讯云镜像源
3. **更好的控制** - 完全控制构建过程和环境
4. **标准化** - 使用业界标准的容器化方案
5. **安全性** - 非 root 用户运行应用
6. **网络优化** - 使用腾讯云 Alpine 镜像源，大幅提升包安装速度

## 构建阶段说明

新的 Dockerfile 使用多阶段构建：

1. **frontend-builder** - 构建前端应用（使用腾讯云镜像源）
2. **backend-builder** - 构建后端应用并集成前端文件（使用腾讯云镜像源）
3. **runtime** - 最小化的运行时镜像（使用腾讯云镜像源）

每个构建阶段都会自动将 Alpine 包源替换为腾讯云镜像源，显著提升包下载速度。

## Go 模块代理优化

在 Go 构建阶段配置了以下环境变量：
- `GOPROXY=https://goproxy.cn,direct` - 使用国内 Go 模块代理
- `GOSUMDB=sum.golang.google.cn` - 使用国内校验和数据库

这将大幅提升 Go 依赖包的下载速度，特别是在国内网络环境下。

## 运行应用

```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

## 清理旧文件

迁移完成后，可以安全删除以下文件：
```bash
rm nixpacks.toml
```

## 故障排除

如果遇到构建问题，可以尝试：
```bash
# 清理 Docker 缓存
docker system prune -f

# 重新构建
docker-compose build --no-cache
```