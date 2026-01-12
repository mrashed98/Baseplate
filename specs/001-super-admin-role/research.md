# Research Findings: Super Admin Role Implementation

**Date**: 2026-01-12
**Feature**: Super Admin Role
**Phase**: 0 - Technical Research

---

## 1. JWT Token Extension Strategy

### Decision
**Add `is_super_admin` claim to new tokens with graceful degradation (Option A)**

### Rationale
- **Zero Breaking Changes**: Existing tokens (24h lifetime) remain valid with missing claim defaulting to `false`
- **No Forced Re-login**: Users continue working without interruption
- **Natural Migration**: All tokens automatically include new claim within 24 hours
- **Standard JWT Practice**: Optional claims via pointer/omitempty is standard JWT pattern
- **golang-jwt/v5 Support**: Library handles missing claims gracefully during unmarshal

### Implementation
```go
// Update JWTClaims struct
type JWTClaims struct {
    UserID       uuid.UUID `json:"user_id"`
    Email        string    `json:"email"`
    IsSuperAdmin *bool     `json:"is_super_admin,omitempty"` // Pointer for optional
    jwt.RegisteredClaims
}

// Graceful degradation in middleware
isSuperAdmin := claims.IsSuperAdmin != nil && *claims.IsSuperAdmin
c.Set(ContextIsSuperAdmin, isSuperAdmin)
```

### Alternatives Considered
- **Force re-login**: Rejected - poor UX, disrupts active users unnecessarily
- **Grace period with fallback**: Rejected - over-engineered for 24h token lifetime

---

## 2. Audit Log Storage Strategy

### Decision
**Use single `audit_logs` table with `actor_type` discriminator (Option A)**

### Rationale
- **Low Volume**: Super admin actions <5% of total audit entries
- **Unified Trail**: Single table enables cross-comparison queries without UNION
- **Existing Pattern**: Leverages current audit infrastructure, no duplication
- **Maintenance**: Single schema to evolve, no sync requirements
- **Performance**: Partial index on `actor_type = 'super_admin'` optimizes queries without bloating index

### Schema Extensions
```sql
ALTER TABLE audit_logs ADD COLUMN actor_type VARCHAR(20) DEFAULT 'team_member'
  CHECK (actor_type IN ('team_member', 'super_admin', 'api_key'));
ALTER TABLE audit_logs ADD COLUMN ip_address INET;
ALTER TABLE audit_logs ADD COLUMN user_agent TEXT;
ALTER TABLE audit_logs ADD COLUMN result_status VARCHAR(20)
  CHECK (result_status IN ('success', 'failure', 'partial'));
ALTER TABLE audit_logs ADD COLUMN request_context JSONB DEFAULT '{}';

CREATE INDEX idx_audit_logs_actor_type ON audit_logs(actor_type)
  WHERE actor_type = 'super_admin';
```

### Alternatives Considered
- **Separate table**: Rejected - overkill for <5% volume, creates maintenance burden
- **External audit service**: Out of scope - adds infrastructure complexity unnecessarily

---

## 3. Last Super Admin Check Implementation

### Decision
**Use transaction with SELECT FOR UPDATE (Option D)**

### Rationale
- **Eliminates Race Conditions**: `FOR UPDATE` provides row-level locks preventing concurrent demotions
- **Correctness Priority**: FR-012/FR-014 are critical safety requirements to prevent platform lockout
- **Acceptable Performance**: 10-20ms transaction time meets <100ms p95 requirement with headroom
- **Low Contention**: Super admin operations are rare (<1/min), lock contention negligible
- **Standard Pattern**: Database transaction management is well-understood, testable pattern

### Implementation
```go
func (s *Service) DemoteFromSuperAdmin(ctx context.Context, requestingUserID, targetUserID uuid.UUID) error {
    tx, err := s.repo.db.DB.BeginTx(ctx, &sql.TxOptions{
        Isolation: sql.LevelReadCommitted,
    })
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // Lock super admin users for counting
    // Note: PostgreSQL does not allow FOR UPDATE with aggregate functions (COUNT)
    // We must select IDs first, then count programmatically
    query := `SELECT id FROM users WHERE is_super_admin = true FOR UPDATE`
    rows, err := tx.QueryContext(ctx, query)
    if err != nil {
        return err
    }
    defer rows.Close()

    count := 0
    for rows.Next() {
        var id uuid.UUID
        if err := rows.Scan(&id); err != nil {
            return err
        }
        count++
    }

    // Check if last super admin attempting self-demotion
    if count <= 1 && requestingUserID == targetUserID {
        return ErrLastSuperAdmin
    }

    // Proceed with demotion
    updateQuery := `UPDATE users SET is_super_admin = false WHERE id = $1`
    if _, err := tx.ExecContext(ctx, updateQuery, targetUserID); err != nil {
        return err
    }

    return tx.Commit()
}
```

### Index Required
```sql
CREATE INDEX idx_users_super_admin ON users(is_super_admin)
  WHERE is_super_admin = true;
```

### Alternatives Considered
- **Simple COUNT query**: Rejected - race condition vulnerability unacceptable
- **Cached count**: Rejected - same race condition, adds cache complexity
- **Database trigger**: Rejected - adds overhead on all user operations, harder to test

---

## 4. Database Migration Initial Super Admin

### Decision
**Use post-migration Go initialization script (Option D)**

### Rationale
- **Security**: Environment variables passed at runtime, never committed to source control
- **Proper Password Hashing**: Uses existing Go auth service with bcrypt
- **Flexibility**: Validates email format, checks if super admin exists (idempotent)
- **Separation of Concerns**: Migrations define schema; operational data seeded separately
- **Consistency**: Aligns with existing architecture where Go handles all business logic

### Implementation
Create `/cmd/init-superadmin/main.go`:
```go
// Reads SUPER_ADMIN_EMAIL and SUPER_ADMIN_PASSWORD from env
// Uses existing auth.Service to hash password
// Creates super admin user if not exists (idempotent)
// Sets is_super_admin = true
```

### Deployment Flow
```bash
1. docker-compose up -d db      # Runs migrations
2. make init-superadmin         # Creates super admin
3. make run                     # Starts server
```

### Environment Variables
```bash
SUPER_ADMIN_EMAIL=admin@company.com
SUPER_ADMIN_PASSWORD=secure-initial-password  # Force reset on first login
```

### Alternatives Considered
- **Environment variable in migration**: Rejected - SQL files cannot read env vars directly
- **Migration file parameter**: Rejected - breaks simple Docker entrypoint pattern
- **Seed data file**: Rejected - security risk (hardcoded passwords in version control)

---

## 5. Permission Bypass Implementation

### Decision
**Hybrid approach: Conditional bypass in both RequireTeam() and RequirePermission() (Option D)**

### Rationale
- **Meets FR-005**: Bypasses team-level permission checks at exact enforcement point
- **Meets FR-004**: Allows super admins to access teams without membership
- **Meets FR-011**: Leverages super admin status in JWT claims for performance
- **Security**: Super admin check is explicit and visible at each authorization point
- **Performance**: No additional database queries per request (status in JWT)
- **Maintainability**: Clear bypass logic in authorization middleware, not scattered

### Implementation Points

**1. In Authenticate() - handleJWT:**
```go
c.Set(ContextUserID, claims.UserID)
c.Set(ContextIsSuperAdmin, claims.IsSuperAdmin != nil && *claims.IsSuperAdmin)
c.Next()
```

**2. In RequireTeam():**
```go
if isSuperAdmin, ok := c.Get(ContextIsSuperAdmin); ok && isSuperAdmin.(bool) {
    c.Set(ContextPermissions, auth.AllPermissions) // Grant all permissions
    c.Set(ContextTeamID, teamID)
    c.Next()
    return
}
// Existing membership check for regular users...
```

**3. In RequirePermission():**
```go
if isSuperAdmin, exists := c.Get(ContextIsSuperAdmin); exists && isSuperAdmin.(bool) {
    c.Next()
    return
}
// Existing permission check...
```

### Security Considerations
- Super admin status cryptographically verified via JWT signature
- Cannot be set by client (JWT signed by server)
- Middleware checks fail-safe (defaults to deny if flag not present)
- Privilege escalation prevention enforced in service layer (separate from middleware)

### Alternatives Considered
- **Early bypass in Authenticate()**: Rejected - violates separation of concerns
- **Only RequirePermission() bypass**: Rejected - doesn't handle RequireTeam() membership requirement
- **Separate super admin middleware**: Rejected - doesn't address bypass requirement for existing routes

---

## 6. Go Error Handling for Authorization

### Best Practice
- **401 Unauthorized**: Authentication failed (invalid/missing JWT)
- **403 Forbidden**: Authenticated but lacks permission
- **404 Not Found**: Resource doesn't exist (check permissions first to avoid info leakage)

### Implementation
```go
// In RequirePermission middleware
c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
    "error": "insufficient permissions",
})

// In handlers when non-super-admin attempts super admin action
c.JSON(http.StatusForbidden, gin.H{
    "error": "super admin access required",
})
```

---

## 7. PostgreSQL Boolean Column Performance

### Decision
**Use partial index on `is_super_admin` column**

### Rationale
- **Sparse Data**: Only ~50 super admins out of 1000-10000 users (<1%)
- **Partial Index Advantage**: Index only `WHERE is_super_admin = true` reduces index size by 99%
- **Query Performance**: COUNT query on partial index <5ms for ~50 rows
- **Storage Efficiency**: Minimal overhead compared to full index

### Index DDL
```sql
CREATE INDEX idx_users_super_admin ON users(is_super_admin)
  WHERE is_super_admin = true;
```

---

## 8. Gin Middleware Context Patterns

### Best Practice
Follow existing patterns in `internal/api/middleware/auth.go`:

```go
// Define context key constants (line 13-17)
const (
    ContextUserID       = "user_id"
    ContextTeamID       = "team_id"
    ContextPermissions  = "permissions"
    ContextIsSuperAdmin = "is_super_admin"  // NEW
)

// Set in middleware
c.Set(ContextIsSuperAdmin, isSuperAdmin)

// Retrieve in handlers
if isSuperAdmin, exists := c.Get(ContextIsSuperAdmin); exists {
    // Use isSuperAdmin.(bool)
}
```

---

## Summary of Key Decisions

| Decision Area | Choice | Primary Benefit |
|---------------|--------|-----------------|
| JWT Extension | Graceful degradation | No user disruption |
| Audit Storage | Single table with discriminator | Unified trail, simple maintenance |
| Last Admin Check | Transaction with FOR UPDATE | Eliminates race conditions |
| Initial Admin | Post-migration Go script | Security + proper password hashing |
| Permission Bypass | Hybrid middleware approach | Clean separation, explicit behavior |
| Database Index | Partial index on boolean | 99% storage reduction |

---

## Next Steps

Proceed to Phase 1:
1. Generate `data-model.md` with extended User entity and enhanced AuditLog schema
2. Create OpenAPI contract in `contracts/super-admin-api.yaml`
3. Write `quickstart.md` for local development/testing
4. Update agent context with new technology decisions
