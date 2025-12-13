# Multi-stage build for DNSMesh application
FROM node:18-alpine AS frontend-builder

WORKDIR /app/frontend

# Replace Alpine repositories with Tencent Cloud mirror for faster package installation
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.cloud.tencent.com/g' /etc/apk/repositories

# Copy frontend package files
COPY frontend/package*.json ./
RUN npm ci

# Copy frontend source code
COPY frontend/ ./

# Build frontend
RUN npm run build

# Backend build stage
FROM golang:1.21-alpine AS backend-builder

WORKDIR /app/backend

# Replace Alpine repositories with Tencent Cloud mirror for faster package installation
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.cloud.tencent.com/g' /etc/apk/repositories

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy backend source code
COPY backend/ ./

# Copy frontend build from previous stage
COPY --from=frontend-builder /app/backend/public ./public

# Build the backend application
RUN CGO_ENABLED=0 GOOS=linux go build -o dnsmesh ./cmd/main.go

# Final runtime stage
FROM alpine:latest

WORKDIR /app

# Replace Alpine repositories with Tencent Cloud mirror for faster package installation
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.cloud.tencent.com/g' /etc/apk/repositories

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S dnsmesh && \
    adduser -S dnsmesh -u 1001 -G dnsmesh

# Copy binary from builder
COPY --from=backend-builder /app/backend/dnsmesh .

# Copy public files (frontend build)
COPY --from=backend-builder /app/backend/public ./public

# Change ownership to non-root user
RUN chown -R dnsmesh:dnsmesh /app

USER dnsmesh

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./dnsmesh"]