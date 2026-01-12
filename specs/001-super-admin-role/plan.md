# Implementation Plan: Super Admin Role

**Branch**: `001-super-admin-role` | **Date**: 2026-01-12 | **Spec**: [spec.md](./spec.md)

## Summary

Add platform-level super admin role to enable cross-team management and oversight. Super admins can view/manage all teams, users, and resources without team membership restrictions. Implementation extends existing RBAC system with `is_super_admin` boolean flag on users table, bypasses team-scoped permission checks in authorization middleware, and enhances audit logging to capture detailed super admin actions.

## Technical Context

**Language/Version**: Go 1.25.1
**Primary Dependencies**: Gin (HTTP framework), golang-jwt/jwt/v5, lib/pq (PostgreSQL driver), golang.org/x/crypto
**Storage**: PostgreSQL 14+ with JSONB support
**Testing**: Go testing package (`go test`), table-driven tests
**Target Platform**: Linux server (containerized via Docker)
**Project Type**: Single backend service (REST API)
**Performance Goals**:
- Super admin permission checks: <100ms p95
- List all users/teams queries: <200ms p95
- Audit log writes: <50ms p95 (async acceptable)
**Constraints**:
- Must maintain existing team-scoped RBAC behavior
- Must prevent privilege escalation attacks
- Must ensure at least one super admin exists at all times
**Scale/Scope**:
- Expected super admin count: <50
- Total user base: 1000-10000 users
- Total teams: 100-1000 teams

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Code Quality & Maintainability ✅
- Single Responsibility Principle: auth middleware handles authorization only
- Clear naming: `IsSuperAdmin()`, `RequireSuperAdmin()` middleware functions
- Proper error handling: return structured errors for unauthorized access
- Go idioms: follow existing patterns in `internal/api/middleware/auth.go`

### II. Testing Standards ✅
- Unit tests required for:
  - `auth.Service` methods (PromoteToSuperAdmin, DemoteFromSuperAdmin, IsSuperAdmin)
  - Middleware permission bypass logic
  - Last super admin protection logic
- Integration tests required for:
  - Super admin endpoints (GET /admin/users, GET /admin/teams, POST /admin/users/:id/promote)
  - Cross-team resource access
  - Audit log writes
- Target coverage: 70% for services, 60% for handlers
- Tests must pass via `make test`

### III. User Experience Consistency ✅
- Follow existing API response patterns:
  - Success: `200 OK` with data
  - Created: `201 Created`
  - Validation errors: `400 Bad Request`
  - Auth failures: `403 Forbidden` (for non-super-admin attempts)
- Error format: `{"error": {"code": "string", "message": "string", "details": object}}`
- Timestamps in ISO 8601 format
- Pagination for GET /admin/users and GET /admin/teams

### IV. Performance Requirements ✅
- Permission checks complete in <100ms p95 (SC-004)
- Database queries use indexes on `users.is_super_admin` and team_id filters
- Audit logging async to avoid blocking request path
- Connection pool already configured (10-100 connections)

### V. Security & Data Integrity ✅
- JWT tokens extended to include `is_super_admin` claim
- Authorization middleware enforces super admin checks at service layer
- Privilege escalation prevention: only existing super admins can promote (FR-008)
- Last super admin protection prevents lockout (FR-012, FR-014)
- Audit logs capture full request context (FR-010)
- Database migration creates initial super admin securely

**Gate Status**: ✅ PASS - All constitution requirements met

## Project Structure

### Documentation (this feature)

```text
specs/001-super-admin-role/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (OpenAPI specs)
│   └── super-admin-api.yaml
└── tasks.md             # Phase 2 output (NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
internal/
├── core/
│   └── auth/
│       ├── models.go              # Add IsSuperAdmin field, SuperAdminPromotion methods
│       ├── service.go             # Add PromoteToSuperAdmin, DemoteFromSuperAdmin, GetAllUsers, GetAllTeams
│       ├── repository.go          # Add SQL queries for super admin operations
│       └── service_test.go        # Add unit tests
├── api/
│   ├── handlers/
│   │   ├── admin.go (NEW)         # Super admin endpoints handler
│   │   └── admin_test.go (NEW)    # Handler integration tests
│   ├── middleware/
│   │   └── auth.go                # Modify: add RequireSuperAdmin(), update Authenticate() to set super admin in context
│   └── router.go                  # Add /admin/* routes
└── storage/
    └── postgres/
        └── migrations/
            └── 002_super_admin.sql (NEW)  # Add is_super_admin column, create initial super admin

migrations/
└── 002_super_admin.sql (NEW)      # Database migration

cmd/
└── server/
    └── main.go                    # No changes required

tests/
├── integration/
│   ├── super_admin_test.go (NEW)  # End-to-end super admin tests
│   └── audit_log_test.go (NEW)    # Audit logging tests
└── contract/
    └── admin_api_test.go (NEW)    # API contract tests
```

**Structure Decision**: Single backend service pattern. No frontend changes required (super admin features accessible via existing API patterns). New `/admin/*` route group for platform-level operations. Extends existing `internal/core/auth` module to minimize complexity.

## Complexity Tracking

> No Constitution violations - this section intentionally left empty.

---

## Phase 0: Research & Technical Decisions

### Research Tasks

1. **JWT Token Extension Strategy**
   - **Question**: How to add `is_super_admin` claim to JWT without breaking existing tokens?
   - **Options**: (a) Add claim to new tokens only, (b) Force re-login for all users, (c) Grace period with fallback
   - **Research**: Review `internal/core/auth/service.go` generateToken() method, check JWT token expiration settings

2. **Audit Log Storage Strategy**
   - **Question**: Use existing audit_logs table or create separate super_admin_audit table?
   - **Options**: (a) Single table with type discriminator, (b) Separate table, (c) External audit service
   - **Research**: Review existing audit_logs table schema, assess query performance implications

3. **Last Super Admin Check Implementation**
   - **Question**: How to efficiently check if user is last super admin during demotion/deletion?
   - **Options**: (a) Count query on every operation, (b) Cached super admin count, (c) Database constraint
   - **Research**: Evaluate performance of `SELECT COUNT(*) FROM users WHERE is_super_admin = true`

4. **Database Migration Initial Super Admin**
   - **Question**: How to specify initial super admin email during deployment?
   - **Options**: (a) Environment variable, (b) Migration file parameter, (c) Seed data file
   - **Research**: Review existing migration execution in Docker init scripts

5. **Permission Bypass Implementation**
   - **Question**: Where in middleware stack should super admin bypass occur?
   - **Options**: (a) Early bypass in Authenticate(), (b) Conditional bypass in RequirePermission(), (c) Separate middleware
   - **Research**: Review existing `internal/api/middleware/auth.go` flow

### Technology Best Practices

6. **Go Error Handling for Authorization**
   - Research: Review Go error patterns for permission denied vs not found scenarios
   - Context: Ensure clear distinction between "forbidden" (403) and "unauthorized" (401)

7. **PostgreSQL Boolean Column Performance**
   - Research: Index strategy for `is_super_admin` boolean column
   - Context: Determine if partial index `WHERE is_super_admin = true` improves lookup performance

8. **Gin Middleware Context Patterns**
   - Research: Best practices for storing super admin flag in Gin context
   - Context: Review existing team_id and user_id context handling in codebase

---

## Phase 1: Design Artifacts

*Generated after Phase 0 research completes. Will include:*
- `data-model.md` - Extended User entity with super admin attributes, enhanced AuditLog schema
- `contracts/super-admin-api.yaml` - OpenAPI spec for new admin endpoints
- `quickstart.md` - Developer guide for testing super admin features locally
- Updated `.claude/context.md` or `.specify/memory/technology-context.md`

---

## Next Steps

1. Run Phase 0 research (automated via planning workflow)
2. Generate Phase 1 design artifacts based on research findings
3. Proceed to `/speckit.tasks` for task decomposition after design approval
