# Super Admin Role

## Overview

The Super Admin role provides platform-level administrative access to all teams, users, and resources in the Baseplate system. Super admins bypass team-level permission checks and can manage the entire platform.

## Capabilities

### 1. Team Management
- **List all teams**: `GET /api/admin/teams` - View all teams in the system regardless of membership
- **View team details**: `GET /api/admin/teams/:teamId` - Access any team's information
- Super admins bypass team membership checks

### 2. User Management
- **List all users**: `GET /api/admin/users?limit=50&offset=0` - View all platform users with pagination
- **View user details**: `GET /api/admin/users/:userId` - See user information and team memberships
- **Update user**: `PUT /api/admin/users/:userId` - Modify user name and status
- Super admins can manage any user's account

### 3. Super Admin Delegation
- **Promote user**: `POST /api/admin/users/:userId/promote` - Grant super admin status
- **Demote super admin**: `POST /api/admin/users/:userId/demote` - Revoke super admin status
- Last super admin protection prevents lockout
- Promotion timestamp and actor tracked for audit trail

### 4. Cross-Team Resource Access
- Super admins can access blueprints, entities, and other resources in any team
- No team membership required
- All resource operations bypass permission checks

### 5. Audit Trail
- **Query audit logs**: `GET /api/admin/audit-logs?limit=50&offset=0` - View all super admin actions
- All super admin actions logged with:
  - Actor information (user ID)
  - IP address and user agent
  - Timestamp
  - Request context
  - Before/after data snapshots

## Setup

### Initialize First Super Admin

```bash
# Set environment variables
export SUPER_ADMIN_EMAIL="admin@example.com"
export SUPER_ADMIN_PASSWORD="secure-password"

# Run initialization tool
make init-superadmin
```

The `init-superadmin` tool:
- Creates a new super admin user with the provided email/password
- Promotes an existing user if email already exists
- Sets super admin flags and promotion metadata

### Environment Variables

```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=user
DB_PASSWORD=password
DB_NAME=baseplate
JWT_SECRET=your-secret-key
SUPER_ADMIN_EMAIL=admin@example.com
SUPER_ADMIN_PASSWORD=secure-password
```

## API Endpoints

All endpoints require:
- **Authentication**: Valid JWT token with `is_super_admin: true` claim
- **Method**: Super admin middleware enforces authorization

### Teams
```
GET  /api/admin/teams                    # List all teams
GET  /api/admin/teams/:teamId            # Get team details
```

### Users
```
GET  /api/admin/users                    # List all users
GET  /api/admin/users/:userId            # Get user details and memberships
PUT  /api/admin/users/:userId            # Update user (name, status)
POST /api/admin/users/:userId/promote    # Promote to super admin
POST /api/admin/users/:userId/demote     # Demote from super admin
```

### Audit Logs
```
GET  /api/admin/audit-logs               # Query super admin action logs
```

## Error Handling

### Common Error Codes

| Status | Error | Meaning |
|--------|-------|---------|
| 401 | Unauthorized | Missing or invalid JWT token |
| 403 | Forbidden | User is not a super admin |
| 404 | Not Found | Resource not found |
| 409 | Conflict | Cannot demote last super admin |
| 400 | Bad Request | User already super admin, or not a super admin |

### Example Error Response

```json
{
  "error": "cannot demote the last super admin"
}
```

## Security Considerations

### Authentication
- Super admin status embedded in JWT claims
- JWT signature validation prevents claim forgery
- Claims gracefully degrade if missing (old tokens default to non-super-admin)

### Authorization
- `RequireSuperAdmin()` middleware enforces super admin status
- Super admin check before any privileged operation
- Privilege escalation prevented via authorization layer

### Data Protection
- Last super admin protection via database transaction with SELECT FOR UPDATE lock
- Prevents race conditions during demotion
- Audit trail captures all modifications

### Audit Logging
- All super admin actions logged to audit_logs table
- IP address and user agent captured for forensics
- Request context and data snapshots stored for compliance

## Database Schema

### Users Table Extensions
```sql
ALTER TABLE users ADD COLUMN is_super_admin BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN super_admin_promoted_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE users ADD COLUMN super_admin_promoted_by UUID REFERENCES users(id);

-- Performance index
CREATE INDEX idx_users_super_admin ON users(is_super_admin) WHERE is_super_admin = true;
```

### Audit Logs Table Extensions
```sql
ALTER TABLE audit_logs ADD COLUMN actor_type VARCHAR(20) CHECK (actor_type IN ('team_member', 'super_admin', 'api_key'));
ALTER TABLE audit_logs ADD COLUMN ip_address INET;
ALTER TABLE audit_logs ADD COLUMN user_agent TEXT;
ALTER TABLE audit_logs ADD COLUMN result_status VARCHAR(20) CHECK (result_status IN ('success', 'failure', 'partial'));
ALTER TABLE audit_logs ADD COLUMN request_context JSONB;

-- Performance index for super admin audit queries
CREATE INDEX idx_audit_logs_actor_type ON audit_logs(actor_type) WHERE actor_type = 'super_admin';
```

## Examples

### Promote a User to Super Admin

```bash
curl -X POST http://localhost:8080/api/admin/users/{userId}/promote \
  -H "Authorization: Bearer {jwt-token}" \
  -H "Content-Type: application/json"
```

Response:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "name": "New Admin",
  "is_super_admin": true,
  "super_admin_promoted_at": "2026-01-12T10:30:00Z",
  "super_admin_promoted_by": "admin-user-id"
}
```

### List All Users

```bash
curl -X GET "http://localhost:8080/api/admin/users?limit=10&offset=0" \
  -H "Authorization: Bearer {jwt-token}"
```

Response:
```json
{
  "users": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "user@example.com",
      "name": "User Name",
      "is_super_admin": false,
      "created_at": "2026-01-12T10:00:00Z"
    }
  ],
  "limit": 10,
  "offset": 0
}
```

### Query Audit Logs

```bash
curl -X GET "http://localhost:8080/api/admin/audit-logs?limit=20&offset=0" \
  -H "Authorization: Bearer {jwt-token}"
```

Response:
```json
{
  "logs": [
    {
      "id": "log-id",
      "user_id": "actor-id",
      "actor_type": "super_admin",
      "entity_type": "user",
      "entity_id": "target-user-id",
      "action": "promote",
      "ip_address": "192.168.1.100",
      "user_agent": "curl/7.68.0",
      "result_status": "success",
      "created_at": "2026-01-12T10:30:00Z"
    }
  ],
  "limit": 20,
  "offset": 0
}
```

## Best Practices

1. **Minimize Super Admin Accounts**: Keep the number of super admins minimal to reduce security risk
2. **Monitor Audit Logs**: Regularly review super admin actions in audit logs
3. **Secure Initial Setup**: Use strong passwords for initial super admin account
4. **Rotation**: Periodically rotate super admin privileges
5. **Team Membership**: Consider making super admins team members when focused on specific teams
6. **API Keys**: Use separate API keys for different applications/services

## Troubleshooting

### Cannot Demote User
**Error**: "cannot demote the last super admin"
- **Cause**: Attempting to demote the only super admin
- **Solution**: Promote another user to super admin first, then demote

### Unauthorized Access
**Error**: "super admin privileges required"
- **Cause**: User JWT token does not have `is_super_admin: true` claim
- **Solution**: Ensure user was promoted to super admin and token was refreshed

### User Not Found
**Error**: "user not found"
- **Cause**: Invalid user ID in request
- **Solution**: Verify the user ID is correct and user exists

## Implementation Details

### JWT Token Extension
- Super admin status included in JWT claims when token generated
- Graceful degradation: old tokens without claim default to non-super-admin
- Claim set by `generateToken()` service method

### Middleware Authorization
- `RequireSuperAdmin()` middleware enforces super admin check
- `RequirePermission()` middleware bypasses checks for super admins
- Super admin flag set in context during `Authenticate()` middleware

### Transaction Safety
- Demotion uses database transaction with SELECT FOR UPDATE lock
- Prevents race condition where last super admin could be demoted
- Atomic operation ensures consistency

## Limitations

- Super admin cannot be deleted if they are the only super admin
- Super admin cannot self-demote to prevent lockout
- Super admin actions still go through JWT expiration
- API key authentication does not support super admin privileges

## Future Enhancements

- Role-based super admin capabilities (e.g., team-only super admins)
- Super admin session management and concurrent login limits
- Temporary super admin elevation for emergency access
- Super admin approval workflows for critical operations
