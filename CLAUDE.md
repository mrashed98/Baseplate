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

## Super Admin Implementation Patterns

### Authorization Model
The super admin role provides platform-level access that bypasses team-scoped permission checks:
- **RequireSuperAdmin()** middleware: Enforces super admin status for sensitive endpoints
- **RequireTeam()** middleware: Grants AllPermissions to super admins instead of querying permissions
- **RequirePermission()** middleware: Bypasses permission checks for super admins

### Service Layer Pattern
Super admin operations follow a consistent pattern:
```go
// 1. Verify actor is super admin
// 2. Validate target resource exists
// 3. Check specific business rule (e.g., not already super admin)
// 4. Use transaction if modifying multiple records
// 5. Return updated resource or error
```

### Examples from Codebase
- **PromoteToSuperAdmin()**: Actor validation → target user validation → idempotency check → update
- **DemoteFromSuperAdmin()**: Actor validation → target validation → last admin check (with SELECT FOR UPDATE) → transaction → update
- **GetAllTeams()**: Simple delegation to repository (no auth validation - handled by middleware)

### Audit Logging
Super admin actions can be logged via:
1. Create AuditLog struct with actor_type = "super_admin"
2. Set ip_address and user_agent from middleware context
3. Capture request_context for forensics
4. Call repo.CreateAuditLog() in background or post-request

### Transaction Safety
For operations that prevent race conditions:
1. Use `client.BeginTx(ctx, nil)` to start transaction
2. Call `repo.CountSuperAdminsForUpdate(ctx, tx)` for pessimistic lock (SELECT FOR UPDATE)
3. Check business rule (e.g., count > 1)
4. Call update method with tx parameter
5. `tx.Commit()` if successful, otherwise `tx.Rollback()`

### JWT Token Extension
Super admin status added to JWTClaims with graceful degradation:
- Claims struct: `IsSuperAdmin *bool` (pointer for backward compatibility)
- Generation: Always set by generateToken() from User.IsSuperAdmin
- Validation: ValidateToken() defaults missing claim to false
- Middleware: handleJWT() extracts and sets ContextIsSuperAdmin in context

### Error Handling
Standard error types for super admin operations:
- `ErrUnauthorized`: User is not a super admin
- `ErrLastSuperAdmin`: Cannot demote the only super admin
- `ErrAlreadySuperAdmin`: User already has super admin status
- `ErrNotSuperAdmin`: User is not a super admin (for demotion)

## Recent Changes
- 001-super-admin-role: Complete super admin implementation with 7 phases
  - Phase 1: Database migration and CLI tool
  - Phase 2: JWT extension and foundational auth
  - Phase 3: Team management endpoints
  - Phase 4: User management endpoints
  - Phase 5: Promotion/demotion with last admin protection
  - Phase 6: Permission bypass for cross-team access
  - Phase 7: Audit logging with IP/user agent tracking
