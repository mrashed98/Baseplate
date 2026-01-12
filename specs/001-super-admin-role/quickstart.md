# Quickstart Guide: Super Admin Role

**Feature**: Super Admin Role
**Target**: Local Development & Testing
**Date**: 2026-01-12

---

## Prerequisites

- Docker and Docker Compose installed
- Go 1.25.1+ installed
- PostgreSQL client (for manual database inspection)
- curl or similar HTTP client

---

## Setup Instructions

### 1. Start the Database

```bash
cd /Users/rashed/instabug/code-days/Baseplate
make db-up
```

This starts PostgreSQL and runs migrations including `002_super_admin.sql`.

### 2. Create Initial Super Admin

Set environment variables:

```bash
export SUPER_ADMIN_EMAIL="admin@example.com"
export SUPER_ADMIN_PASSWORD="SuperSecure123!"
```

Run the initialization script:

```bash
make init-superadmin
```

**Expected output**:
```
Super admin user created successfully
Email: admin@example.com
User ID: <uuid>
```

### 3. Start the Server

```bash
make run
```

Server starts on `http://localhost:8080`

---

## Testing Scenarios

### Scenario 1: Super Admin Login

**Login as super admin:**

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "SuperSecure123!"
  }'
```

**Response**:
```json
{
  "token": "eyJhbGc...",
  "user": {
    "id": "<uuid>",
    "email": "admin@example.com",
    "name": "Super Admin",
    "status": "active",
    "is_super_admin": true,
    "created_at": "2026-01-12T..."
  }
}
```

**Save the token** for subsequent requests:
```bash
export SUPER_ADMIN_TOKEN="<token-from-response>"
```

---

### Scenario 2: List All Users (Platform-Wide)

**Request:**

```bash
curl -X GET http://localhost:8080/api/admin/users \
  -H "Authorization: Bearer $SUPER_ADMIN_TOKEN"
```

**Expected**: List of all users across all teams

**Verify non-super-admin cannot access:**

```bash
# First, create and login as regular user
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "name": "Regular User"
  }'

# Try to access admin endpoint (should fail with 403)
curl -X GET http://localhost:8080/api/admin/users \
  -H "Authorization: Bearer <regular-user-token>"
```

**Expected**: `403 Forbidden - Super admin access required`

---

### Scenario 3: Promote User to Super Admin

**Create a regular user first:**

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newadmin@example.com",
    "password": "password123",
    "name": "Future Admin"
  }'

# Save user ID from response
export NEW_USER_ID="<user-id-from-response>"
```

**Promote to super admin:**

```bash
curl -X POST "http://localhost:8080/api/admin/users/$NEW_USER_ID/promote" \
  -H "Authorization: Bearer $SUPER_ADMIN_TOKEN"
```

**Expected response:**
```json
{
  "message": "User promoted to super admin",
  "user": {
    "id": "<uuid>",
    "email": "newadmin@example.com",
    "is_super_admin": true,
    "super_admin_promoted_at": "2026-01-12T...",
    "super_admin_promoted_by": "<super-admin-uuid>"
  }
}
```

**Verify promotion:**

```bash
# Login as newly promoted user
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newadmin@example.com",
    "password": "password123"
  }'

# Token should include is_super_admin claim
# Decode JWT at https://jwt.io to verify
```

---

### Scenario 4: Attempt to Demote Last Super Admin (Should Fail)

**If there's only one super admin, try self-demotion:**

```bash
# Get your own user ID from /api/auth/me
export MY_USER_ID=$(curl -s http://localhost:8080/api/auth/me \
  -H "Authorization: Bearer $SUPER_ADMIN_TOKEN" \
  | jq -r '.id')

# Attempt self-demotion
curl -X POST "http://localhost:8080/api/admin/users/$MY_USER_ID/demote" \
  -H "Authorization: Bearer $SUPER_ADMIN_TOKEN"
```

**Expected response:**
```json
{
  "error": {
    "code": "LAST_SUPER_ADMIN",
    "message": "Cannot demote the last super admin"
  }
}
```

**Status code**: `409 Conflict`

---

### Scenario 5: Cross-Team Access (Super Admin Bypass)

**Setup: Create two teams and a regular admin:**

```bash
# Login as super admin
export SUPER_ADMIN_TOKEN="<token>"

# Create Team A
curl -X POST http://localhost:8080/api/teams \
  -H "Authorization: Bearer $SUPER_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Team Alpha",
    "slug": "team-alpha"
  }'

export TEAM_A_ID="<team-id-from-response>"

# Create Team B (as super admin or another user)
# ... similar to above
export TEAM_B_ID="<team-b-id>"

# Create a blueprint in Team A
curl -X POST http://localhost:8080/api/blueprints \
  -H "Authorization: Bearer $SUPER_ADMIN_TOKEN" \
  -H "X-Team-ID: $TEAM_A_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "service",
    "title": "Service",
    "description": "Microservice definition",
    "schema": {
      "type": "object",
      "properties": {
        "name": {"type": "string"}
      }
    }
  }'
```

**Test: Super admin can access Team A resources without membership:**

```bash
# List blueprints in Team A (super admin not a member, but should work)
curl -X GET http://localhost:8080/api/blueprints \
  -H "Authorization: Bearer $SUPER_ADMIN_TOKEN" \
  -H "X-Team-ID: $TEAM_A_ID"
```

**Expected**: Successfully lists blueprints despite not being a team member

**Verify regular user cannot:**

```bash
# Create regular user, have them try to access Team A
# Should get 403 Forbidden if not a member
```

---

### Scenario 6: Audit Log Verification

**Perform actions as super admin, then query audit logs:**

```bash
# Query super admin audit logs
curl -X GET "http://localhost:8080/api/admin/audit-logs?limit=10" \
  -H "Authorization: Bearer $SUPER_ADMIN_TOKEN"
```

**Expected response:**
```json
{
  "logs": [
    {
      "id": "<uuid>",
      "user_id": "<super-admin-uuid>",
      "actor_type": "super_admin",
      "entity_type": "user",
      "entity_id": "<promoted-user-uuid>",
      "action": "promote",
      "ip_address": "127.0.0.1",
      "user_agent": "curl/7.x",
      "result_status": "success",
      "created_at": "2026-01-12T..."
    }
  ],
  "total": 5,
  "limit": 10,
  "offset": 0
}
```

**Verify before/after snapshots:**

```bash
# Query specific log entry with details
curl -X GET "http://localhost:8080/api/admin/audit-logs?action=promote" \
  -H "Authorization: Bearer $SUPER_ADMIN_TOKEN"
```

Should see `old_data` (user before promotion) and `new_data` (user after promotion).

---

## Database Inspection

**Check super admin status directly:**

```bash
psql -h localhost -p 5432 -U baseplate -d baseplate

-- List all super admins
SELECT id, email, name, is_super_admin, super_admin_promoted_at
FROM users
WHERE is_super_admin = true;

-- Check audit logs for super admin actions
SELECT user_id, actor_type, action, entity_type, created_at
FROM audit_logs
WHERE actor_type = 'super_admin'
ORDER BY created_at DESC
LIMIT 10;
```

---

## Testing Permission Bypass

### Test Team-Level Permission Bypass

**Setup:**
1. Create a team as regular user
2. Add blueprint/entity to that team
3. Login as super admin (not member of the team)
4. Attempt to modify resources in that team

**Expected**: Super admin can CRUD resources despite not being a team member

### Test Regular User Cannot Self-Promote

**Attempt:**

```bash
# Login as regular user
export REGULAR_TOKEN="<regular-user-token>"
export REGULAR_USER_ID="<regular-user-id>"

# Try to promote self
curl -X POST "http://localhost:8080/api/admin/users/$REGULAR_USER_ID/promote" \
  -H "Authorization: Bearer $REGULAR_TOKEN"
```

**Expected**: `403 Forbidden - Super admin access required`

---

## Performance Testing

### Test Permission Check Latency

**Measure super admin permission check time:**

```bash
# Time multiple requests
for i in {1..100}; do
  curl -w "@curl-format.txt" -o /dev/null -s \
    "http://localhost:8080/api/admin/users" \
    -H "Authorization: Bearer $SUPER_ADMIN_TOKEN"
done | awk '{sum+=$1; count++} END {print "Average: " sum/count "ms"}'
```

**Expected**: Average <50ms (well under 100ms p95 requirement)

### Test Last Super Admin Check Performance

**Measure demotion check time:**

```bash
# Attempt demotion (will fail as last admin, but measures check time)
time curl -X POST "http://localhost:8080/api/admin/users/$MY_USER_ID/demote" \
  -H "Authorization: Bearer $SUPER_ADMIN_TOKEN"
```

**Expected**: <100ms including SELECT FOR UPDATE transaction

---

## Troubleshooting

### Issue: "Super admin access required" for valid super admin

**Diagnosis**:
1. Check JWT token contains `is_super_admin` claim:
   - Decode token at https://jwt.io
   - Look for `"is_super_admin": true` in payload

2. Check database:
   ```sql
   SELECT is_super_admin FROM users WHERE email = 'admin@example.com';
   ```

**Solution**: If JWT is old (generated before promotion), re-login to get new token.

### Issue: Cannot create initial super admin

**Diagnosis**: Check migration ran successfully:
```bash
docker-compose logs db | grep "002_super_admin"
```

**Solution**: Run migration manually:
```bash
psql -h localhost -U baseplate -d baseplate < migrations/002_super_admin.sql
```

### Issue: "Cannot demote last super admin" when multiple exist

**Diagnosis**: Transaction race condition or count query issue.

**Check count**:
```sql
SELECT COUNT(*) FROM users WHERE is_super_admin = true;
```

**Solution**: Ensure `idx_users_super_admin` index exists:
```sql
SELECT indexname FROM pg_indexes WHERE tablename = 'users' AND indexname = 'idx_users_super_admin';
```

---

## Clean Up

**Reset database to clean state:**

```bash
make db-reset
make init-superadmin
```

**Remove all data:**

```bash
docker-compose down -v
docker-compose up -d db
make init-superadmin
```

---

## Next Steps

1. Run integration tests: `make test`
2. Review audit logs for anomalies
3. Test with production-like data volumes
4. Proceed to `/speckit.tasks` for task decomposition
