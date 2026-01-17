# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

DNSMesh is a DNS management system that unifies DNS records from multiple providers (Cloudflare and Tencent Cloud DNSPod). It features intelligent analysis to identify server-application domain relationships and group records by server.

## Development Commands

### Backend (Go 1.21+)

The backend uses Go modules with module name `dnsmesh`. All commands should be run from the `backend/` directory.

**Running the backend:**
```bash
cd backend
go run cmd/main.go
```

The backend automatically loads environment variables from `.env` file (via godotenv). If `.env` doesn't exist, it falls back to system environment variables.

**Building:**
```bash
cd backend
go build -o bin/dnsmesh cmd/main.go
```

**Dependencies:**
```bash
cd backend
go mod tidy
```

### Frontend (Mithril.js + Vite)

**Development server:**
```bash
cd frontend
npm run dev
```

**Production build:**
```bash
cd frontend
npm run build
```

The build output goes to `frontend/dist/` and should be copied to `backend/public/` for production deployment.

### Database

The backend uses SQLite by default. To migrate from an existing Postgres instance, run:
```bash
cd backend
go run cmd/migrate_sqlite/main.go
```

**Full stack startup:**
```bash
./start.sh
```

### Testing

The project includes an API testing script:
```bash
./test.sh
```

This script tests authentication, provider management, DNS records, and audit logs APIs.

## Architecture

### Backend Structure

- **Entry point:** `backend/cmd/main.go` - Initializes crypto, database, Gin router, sessions, and routes
- **Database:** `backend/internal/database/database.go` - GORM with SQLite, handles migrations and cleanup
- **Models:** `backend/internal/models/` - GORM models for User, Provider, DNSRecord, AuditLog
- **Handlers:** `backend/internal/handlers/` - Gin HTTP handlers for auth, providers, records, audit logs
- **Services:** `backend/internal/services/` - Business logic including:
  - `analyzer.go` - Intelligent DNS record analysis algorithm
  - `cloudflare.go` - Cloudflare API integration
  - `tencentcloud.go` - Tencent Cloud DNSPod integration
  - `types.go` - Service interfaces and data structures
- **Middleware:** `backend/internal/middleware/auth.go` - Session-based authentication
- **Crypto:** `backend/pkg/crypto/crypto.go` - AES-256 encryption for provider credentials

### Frontend Structure

- **Framework:** Mithril.js (lightweight MVC)
- **Entry:** `frontend/src/main.js`
- **Views:** `frontend/src/views/` - Page components
- **Components:** `frontend/src/components/` - Reusable UI components
- **Services:** `frontend/src/services/` - API client logic
- **Styles:** `frontend/src/styles/` - CSS files

### Key Architectural Patterns

**DNS Record Analysis Algorithm** (`backend/internal/services/analyzer.go`):
- Pattern-based server detection using region codes (e.g., `hk-1.example.com`, `us-2.example.com`)
- Supports multiple formats: `region-number.domain`, `regionnumber.domain`, wildcard domains
- CNAME chain analysis to group application domains under their target servers
- IP-based grouping for domains pointing to the same server
- Confidence scoring (high/medium/low) based on pattern strength and references

**Authentication Flow**:
- Session-based auth using gin-contrib/sessions with cookie store
- Credentials stored in sessions, 7-day expiry
- AuthRequired middleware protects all `/api/*` routes except `/api/auth/login`
- Default admin user created on first startup (configurable via env vars)

**Provider Credential Security**:
- All provider API tokens/secrets encrypted using AES-256 before database storage
- Encryption key loaded from `ENCRYPTION_KEY` environment variable (must be 32 bytes)
- Decrypted only when making provider API calls

**Database Initialization**:
- Auto-migration on startup creates/updates tables
- Default admin user created if no users exist
- Cleanup routine removes providers with no associated DNS records

## Environment Configuration

Required environment variables (see `backend/.env.example`):

**Database:**
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`

**Server:**
- `PORT` (default: 8080)
- `GIN_MODE` (development/release)

**Security:**
- `SESSION_SECRET` - Session encryption key
- `ENCRYPTION_KEY` - Must be exactly 32 characters for AES-256
- `ADMIN_USERNAME`, `ADMIN_PASSWORD` - Default admin credentials

## API Routes

### Public Routes
- `POST /api/auth/login` - User login

### Protected Routes (require authentication)

**Auth:**
- `POST /api/auth/logout`
- `GET /api/auth/user`
- `POST /api/auth/change-password`

**Providers:**
- `GET /api/providers` - List all providers
- `POST /api/providers` - Create provider
- `PUT /api/providers/:id` - Update provider
- `DELETE /api/providers/:id` - Delete provider
- `POST /api/providers/:id/sync` - Sync records from provider

**DNS Records:**
- `GET /api/records` - List all records
- `POST /api/records` - Create record
- `PUT /api/records/:id` - Update record
- `POST /api/records/:id/hide` - Hide record from view
- `DELETE /api/records/:id` - Delete record
- `POST /api/records/import` - Bulk import records
- `POST /api/records/reanalyze` - Re-run analysis algorithm

**Audit:**
- `GET /api/audit-logs` - View operation history

## Important Notes

- The backend serves the frontend from `backend/public/` in production
- All API routes use `/api` prefix to avoid conflicts with frontend routing
- Provider credentials are never returned in API responses (security measure)
- DNS record analysis runs automatically after syncing from providers
- The analyzer supports region codes for both international (us, hk, sg, etc.) and Chinese cities (bj, sh, gz, etc.)
