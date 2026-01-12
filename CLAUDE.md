# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Development Commands

```bash
make db-up          # Start PostgreSQL (required first)
make run            # Run server on :8080
make build          # Build to bin/server
make test           # Run all tests
make db-reset       # Drop and recreate database
make fmt            # Format code
make tidy           # go mod tidy
```

Run single test: `go test -v -run TestName ./internal/core/entity/...`

## Architecture Overview

Baseplate is a headless backend engine (Port.ai clone) with dynamic schema management via Blueprints.

### Core Concepts

- **Blueprints**: JSON Schema definitions for entity types (e.g., "Service", "Cluster"). Created at runtime, stored in `blueprints` table.
- **Entities**: Instances of Blueprints. Data validated against Blueprint's JSON Schema before persistence.
- **Teams**: Multi-tenant isolation. All resources scoped to team_id.
- **RBAC**: Roles with permission arrays (e.g., `blueprint:read`, `entity:write`). Three default roles: admin, editor, viewer.

### Request Flow

```
HTTP Request → Router → Auth Middleware → Permission Middleware → Handler → Service → Repository → PostgreSQL
```

### Layer Responsibilities

| Layer | Location | Purpose |
|-------|----------|---------|
| Handlers | `internal/api/handlers/` | HTTP binding, response formatting |
| Services | `internal/core/*/service.go` | Business logic, validation orchestration |
| Repositories | `internal/core/*/repository.go` | SQL queries, data mapping |
| Middleware | `internal/api/middleware/` | JWT/API key auth, RBAC checks |

### Key Patterns

**Authentication**: Two modes supported via `Authorization` header:
- `Bearer <jwt>` - User JWT token
- `ApiKey <key>` - Team API key (prefix: `bp_`)

**Team Context**: Required for most endpoints. Set via:
- URL param: `/teams/:teamId/...`
- Header: `X-Team-ID`
- Automatically from API key

**Entity Search**: POST `/api/blueprints/:blueprintId/entities/search` with filter operators: `eq`, `neq`, `gt`, `lt`, `gte`, `lte`, `contains`, `exists`, `in`. Filters query JSONB `data` column.

### Database

PostgreSQL with JSONB for flexible schema storage. Key tables:
- `users`, `teams`, `roles`, `team_memberships`, `api_keys` - Auth/RBAC
- `blueprints` - Schema definitions with `schema JSONB`
- `entities` - Entity instances with `data JSONB`

Migrations auto-run via Docker init scripts from `migrations/001_initial.sql`.

### Environment Variables

```
DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME
JWT_SECRET, JWT_EXPIRATION_HOURS
SERVER_PORT, GIN_MODE
```

## Development Reuqirements

ALWAYS Update the docs at @docs directory with every change

## Active Technologies
- Go 1.25.1 + Gin (HTTP framework), golang-jwt/jwt/v5, lib/pq (PostgreSQL driver), golang.org/x/crypto (001-super-admin-role)
- PostgreSQL 14+ with JSONB suppor (001-super-admin-role)

## Recent Changes
- 001-super-admin-role: Added Go 1.25.1 + Gin (HTTP framework), golang-jwt/jwt/v5, lib/pq (PostgreSQL driver), golang.org/x/crypto
