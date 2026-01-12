# Baseplate Security Documentation

Comprehensive security documentation covering authentication, authorization, data protection, and security best practices for Baseplate.

## Table of Contents

- [Security Overview](#security-overview)
- [Authentication](#authentication)
- [Authorization](#authorization)
- [Data Protection](#data-protection)
- [Input Validation](#input-validation)
- [Security Best Practices](#security-best-practices)
- [Security Checklist](#security-checklist)
- [Vulnerability Reporting](#vulnerability-reporting)

## Security Overview

Baseplate implements defense-in-depth security with multiple layers:

1. **Authentication** - JWT tokens and API keys
2. **Authorization** - RBAC with 13 permissions
3. **Data Isolation** - Multi-tenant team-based access control
4. **Input Validation** - JSON Schema validation and SQL injection prevention
5. **Secure Storage** - bcrypt for passwords, SHA-256 for API keys
6. **Transport Security** - HTTPS recommended for production

## Authentication

Baseplate supports two authentication mechanisms with different use cases.

### JWT Token Authentication

**Use Case**: User sessions, web applications, mobile apps

**Flow**:
1. User registers or logs in with email/password
2. Server validates credentials
3. Server generates JWT token (24-hour expiration)
4. Client includes token in `Authorization: Bearer <token>` header
5. Server validates token signature and expiration on each request

**Token Structure**:
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "john@example.com",
  "iat": 1705401600,
  "exp": 1705488000
}
```

**Security Properties**:
- **Algorithm**: HS256 (HMAC with SHA-256)
- **Secret**: 256-bit secret from `JWT_SECRET` environment variable
- **Expiration**: 24 hours (configurable via `JWT_EXPIRATION_HOURS`)
- **Signature**: Prevents tampering
- **Stateless**: No server-side session storage

**Implementation**: `internal/core/auth/service.go:103-137`

**Security Recommendations**:
```bash
# Generate secure JWT secret (32+ characters)
openssl rand -base64 32

# Set in environment
export JWT_SECRET="your-secure-secret-here"
```

**Never**:
- Use weak secrets (e.g., "secret", "password")
- Commit secrets to version control
- Share secrets across environments

---

### API Key Authentication

**Use Case**: Service-to-service, CI/CD, integrations, automated scripts

**Flow**:
1. Admin creates API key via web interface
2. Server generates 32-byte random key
3. Server hashes key with SHA-256 and stores hash
4. Raw key returned once (never retrievable again)
5. Client includes key in `Authorization: ApiKey <key>` header
6. Server hashes incoming key and looks up in database
7. Server validates expiration and permissions

**Key Format**: `bp_<64_hex_characters>`

**Example**: `bp_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6a7b8c9d0e1f2`

**Security Properties**:
- **Generation**: Cryptographically secure random (32 bytes)
- **Hashing**: SHA-256 (never store plain text)
- **Expiration**: Optional timestamp
- **Permissions**: Optional permission array (subset of role permissions)
- **Last Used Tracking**: Async update to avoid blocking
- **Revocation**: Immediate via DELETE endpoint

**Implementation**: `internal/core/auth/service.go:325-384`

**Security Recommendations**:
- Treat API keys like passwords
- Store securely (environment variables, secrets managers)
- Rotate regularly (every 90 days recommended)
- Use minimum required permissions
- Set expiration dates
- Monitor last_used_at for anomalies
- Revoke immediately if compromised

**Never**:
- Commit API keys to version control
- Log API keys in plain text
- Share API keys via insecure channels (email, Slack)
- Use same API key across environments

---

### Password Security

**Hashing Algorithm**: bcrypt with default cost factor (10)

**Properties**:
- **Salt**: Automatically generated per password (random)
- **Cost Factor**: 10 (2^10 = 1024 iterations)
- **Output**: 60-character string
- **Time**: ~100ms per hash (protects against brute force)

**Storage**: `users.password_hash` column

**Validation**: Constant-time comparison via `bcrypt.CompareHashAndPassword`

**Implementation**: `internal/core/auth/service.go:53-96`

**Password Requirements**:
```go
// Minimum length validation
if len(password) < 8 {
    return errors.New("password must be at least 8 characters")
}
```

**Security Recommendations**:
- Enforce strong password policies (min 12 characters, complexity)
- Implement password rotation policies
- Use rate limiting on login attempts
- Monitor failed login attempts
- Never log passwords (even hashed)

---

### Authentication Comparison

| Feature | JWT Token | API Key |
|---------|-----------|---------|
| **Use Case** | User sessions | Service-to-service |
| **Lifetime** | 24 hours | Until revoked or expired |
| **Storage** | Client-side | Server-side hash |
| **Revocation** | Wait for expiration | Immediate |
| **Team Context** | Via header/URL | Automatic |
| **Rotation** | Automatic (re-login) | Manual |
| **Permissions** | Via role membership | Direct assignment |

---

## Authorization

Baseplate implements Role-Based Access Control (RBAC) with team-scoped permissions.

### Permission Model

**13 Permissions across 6 resource types**:

```
team:manage           # Manage team settings, roles, members, API keys

blueprint:read        # View blueprints
blueprint:write       # Create/update blueprints
blueprint:delete      # Delete blueprints

entity:read           # View entities
entity:write          # Create/update entities
entity:delete         # Delete entities

integration:read      # View integrations (future)
integration:write     # Configure integrations (future)

scorecard:read        # View scorecards (future)
scorecard:write       # Configure scorecards (future)

action:read           # View actions (future)
action:write          # Configure actions (future)
action:execute        # Execute actions (future)
```

---

### Default Roles

#### Admin Role
Full permissions including team management and delete operations.

```json
{
  "name": "admin",
  "permissions": [
    "team:manage",
    "blueprint:read",
    "blueprint:write",
    "blueprint:delete",
    "entity:read",
    "entity:write",
    "entity:delete"
  ]
}
```

**Use Case**: Team owners, administrators

---

#### Editor Role
Read and write permissions without destructive operations.

```json
{
  "name": "editor",
  "permissions": [
    "blueprint:read",
    "blueprint:write",
    "entity:read",
    "entity:write"
  ]
}
```

**Use Case**: Developers, content managers

---

#### Viewer Role
Read-only access to all resources.

```json
{
  "name": "viewer",
  "permissions": [
    "blueprint:read",
    "entity:read"
  ]
}
```

**Use Case**: Auditors, stakeholders, read-only access

---

### Custom Roles

Teams can create custom roles with specific permission combinations:

```bash
curl -X POST http://localhost:8080/api/teams/$TEAM_ID/roles \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "developer",
    "permissions": [
      "blueprint:read",
      "entity:read",
      "entity:write"
    ]
  }'
```

**Best Practices**:
- Follow principle of least privilege
- Create role per job function
- Regularly audit role permissions
- Document custom role purposes

---

### Permission Enforcement

**Middleware Chain**:
```
Request → Authenticate → RequireTeam → RequirePermission → Handler
```

**Permission Check** (`internal/api/middleware/auth.go`):
```go
func RequirePermission(permission string) gin.HandlerFunc {
    return func(c *gin.Context) {
        permissions := c.GetStringSlice("permissions")

        if !contains(permissions, permission) {
            c.AbortWithStatusJSON(403, gin.H{
                "error": "insufficient permissions",
            })
            return
        }

        c.Next()
    }
}
```

**Endpoints**:
```go
// Example: Blueprint creation requires blueprint:write
blueprints.POST("",
    middleware.RequirePermission("blueprint:write"),
    handler.CreateBlueprint,
)
```

---

## Data Protection

### Multi-Tenancy Isolation

**Team-Based Isolation**:
- All resources scoped to `team_id`
- Database queries always filter by `team_id`
- Middleware validates team membership
- Users cannot access data from teams they don't belong to

**Implementation**:
```sql
-- All queries include team_id filter
SELECT * FROM entities
WHERE team_id = $1  -- Always required
  AND blueprint_id = $2;
```

**Isolation Guarantees**:
1. Data leak prevention between teams
2. User membership validation on every request
3. API keys team-scoped
4. Cascade delete maintains referential integrity

---

### Data Encryption

**At Rest**:
- Use PostgreSQL encryption features (e.g., `pgcrypto` for column encryption)
- Encrypt database backups
- Use encrypted volumes in cloud environments

**In Transit**:
- **Production**: Always use HTTPS/TLS
- Configure Gin with TLS certificates
- Redirect HTTP to HTTPS

```go
// Enable TLS in production
router.RunTLS(":443", "cert.pem", "key.pem")
```

---

### Sensitive Data Handling

**Never Store Plain Text**:
- ✅ Passwords: bcrypt hash
- ✅ API Keys: SHA-256 hash
- ❌ Never store: Credit cards, SSNs, plaintext passwords

**Never Log Sensitive Data**:
```go
// Bad
log.Printf("User login: %s with password %s", email, password)

// Good
log.Printf("User login attempt: %s", email)
```

**API Responses**:
```go
// Never return password_hash in user objects
type UserResponse struct {
    ID        string    `json:"id"`
    Email     string    `json:"email"`
    Name      string    `json:"name"`
    // password_hash omitted
}
```

---

## Input Validation

### JSON Schema Validation

**Entity Data Validation**:
- All entity data validated against blueprint's JSON Schema
- Validation runs before INSERT and UPDATE
- Detailed error messages with field names

**Implementation**: `internal/core/validation/validator.go`

**Example**:
```json
{
  "error": "Validation failed",
  "details": [
    {
      "field": "data.version",
      "message": "String length must be greater than or equal to 1"
    }
  ]
}
```

**Security Benefit**: Prevents malformed data, type confusion attacks

---

### SQL Injection Prevention

**Parameterized Queries**:
```go
// Safe: Uses parameterized query
query := "SELECT * FROM entities WHERE team_id = $1 AND id = $2"
row := db.QueryRow(query, teamID, entityID)
```

**Never**:
```go
// UNSAFE: SQL injection vulnerability
query := fmt.Sprintf("SELECT * FROM entities WHERE id = '%s'", entityID)
```

**Property Name Validation**:
```go
// Validate JSONB property names
propertyRegex := regexp.MustCompile(`^[a-zA-Z0-9_.]+$`)
if !propertyRegex.MatchString(property) {
    return errors.New("invalid property name")
}
```

**Location**: `internal/core/entity/repository.go:166-169`

**Protected**:
- All database queries use parameterized statements
- JSONB property names validated with regex
- ORDER BY columns use whitelist

---

### Request Validation

**Input Sanitization**:
```go
// Bind and validate JSON
var req CreateEntityRequest
if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(400, gin.H{"error": "Invalid request"})
    return
}
```

**Validation Rules**:
- Email format validation
- Password minimum length (8 characters)
- UUID format validation
- String length limits
- Required field checks

---

## Security Best Practices

### Deployment Security

#### Environment Variables

```bash
# Strong JWT secret (32+ characters)
JWT_SECRET=$(openssl rand -base64 32)

# Production mode
GIN_MODE=release

# Database SSL
DB_SSL_MODE=require
```

---

#### HTTPS/TLS

**Always use HTTPS in production**:

```nginx
# Nginx reverse proxy
server {
    listen 443 ssl;
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

---

#### Rate Limiting

**Implement rate limiting** (not currently built-in):

```go
// Example: Using go-rate
import "golang.org/x/time/rate"

limiter := rate.NewLimiter(rate.Limit(10), 100) // 10 req/sec, burst 100
```

**Recommended Limits**:
- `/api/auth/login`: 5 requests/minute per IP
- `/api/auth/register`: 3 requests/hour per IP
- API endpoints: 100 requests/minute per user/API key

---

#### CORS Configuration

```go
import "github.com/gin-contrib/cors"

router.Use(cors.New(cors.Config{
    AllowOrigins:     []string{"https://your-domain.com"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
    ExposeHeaders:    []string{"Content-Length"},
    AllowCredentials: true,
    MaxAge:           12 * time.Hour,
}))
```

---

### Operational Security

#### Logging

**Log Security Events**:
- Failed authentication attempts
- Permission denials
- API key creation/deletion
- Team member changes
- Bulk deletions

**Never Log**:
- Passwords (plain or hashed)
- API keys
- JWT tokens
- Sensitive JSONB data (PII)

---

#### Monitoring

**Monitor for Anomalies**:
- Unusual API key usage patterns
- Failed login spikes
- Permission denial spikes
- Large data exports
- Off-hours access

---

#### Secrets Management

**Use secrets managers**:
```bash
# AWS Secrets Manager
aws secretsmanager get-secret-value --secret-id baseplate/jwt-secret

# HashiCorp Vault
vault kv get secret/baseplate/jwt-secret

# Kubernetes Secrets
kubectl create secret generic baseplate-secrets --from-literal=jwt-secret=$JWT_SECRET
```

---

#### Database Security

**PostgreSQL Security**:
```postgresql.conf
# SSL required
ssl = on
ssl_cert_file = 'server.crt'
ssl_key_file = 'server.key'

# Connection limits
max_connections = 100

# Password encryption
password_encryption = scram-sha-256
```

**User Permissions**:
```sql
-- Create read-only user for reporting
CREATE USER readonly WITH PASSWORD 'secure-password';
GRANT CONNECT ON DATABASE baseplate TO readonly;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO readonly;
```

---

## Security Checklist

### Pre-Deployment

- [ ] Strong `JWT_SECRET` set (32+ characters)
- [ ] `GIN_MODE=release` in production
- [ ] HTTPS/TLS configured
- [ ] `DB_SSL_MODE=require` for production database
- [ ] Database backups configured and tested
- [ ] Secrets stored in secrets manager (not env files)
- [ ] CORS properly configured
- [ ] Rate limiting implemented

### Post-Deployment

- [ ] Monitor authentication logs
- [ ] Review API key usage patterns
- [ ] Audit team memberships and roles
- [ ] Test backup restoration process
- [ ] Review database connection limits
- [ ] Check for SQL injection vulnerabilities
- [ ] Validate HTTPS certificate expiration
- [ ] Review and rotate API keys

### Regular Maintenance

- [ ] Rotate API keys (quarterly)
- [ ] Update JWT_SECRET (annually or if compromised)
- [ ] Review and remove unused API keys
- [ ] Audit custom role permissions
- [ ] Check for outdated dependencies (`go list -m -u all`)
- [ ] Review security logs for anomalies
- [ ] Test disaster recovery procedures

---

## Vulnerability Reporting

If you discover a security vulnerability in Baseplate:

1. **Do not** open a public GitHub issue
2. **Do not** disclose publicly until patch is available
3. **Email** security contact (to be configured)
4. **Include**:
   - Description of vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

**Expected Response Time**:
- Acknowledgment: Within 48 hours
- Initial assessment: Within 1 week
- Patch and disclosure: Within 30 days (for critical issues)

---

## Security Resources

### External Resources

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [OWASP API Security Top 10](https://owasp.org/www-project-api-security/)
- [JWT Best Practices](https://tools.ietf.org/html/rfc8725)
- [Go Security](https://go.dev/doc/security/)

### Internal Documentation

- [API Documentation](./API.md) - Authentication usage
- [Architecture](./ARCHITECTURE.md) - Security architecture diagrams
- [Deployment](./DEPLOYMENT.md) - Production security setup
- [Database](./DATABASE.md) - Database security configuration

---

## Security Updates

Check for security updates regularly:

```bash
# Check for Go security advisories
go list -m -u all | grep security

# Update dependencies
go get -u ./...
go mod tidy
```

**Subscribe to**:
- Go security announcements
- PostgreSQL security mailing list
- Gin framework security advisories
- Dependency security alerts (GitHub Dependabot)
