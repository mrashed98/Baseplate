# Baseplate Development Guide

Complete guide for setting up a development environment and contributing to Baseplate.

## Table of Contents

- [Getting Started](#getting-started)
- [Project Structure](#project-structure)
- [Development Workflow](#development-workflow)
- [Code Organization](#code-organization)
- [Adding Features](#adding-features)
- [Testing](#testing)
- [Debugging](#debugging)
- [Git Workflow](#git-workflow)
- [Code Style](#code-style)

## Getting Started

### Prerequisites

**Required**:
- **Go**: 1.25.1 or higher ([install](https://go.dev/doc/install))
- **Docker**: 20.10+ ([install](https://docs.docker.com/get-docker/))
- **Docker Compose**: 2.0+ (included with Docker Desktop)
- **Make**: Build automation tool
- **Git**: Version control

**Recommended**:
- **IDE**: VS Code with Go extension, GoLand, or similar
- **Database Client**: psql, DBeaver, or TablePlus
- **API Client**: curl, Postman, or HTTPie

### Initial Setup

#### 1. Clone Repository

```bash
git clone https://github.com/your-org/baseplate.git
cd baseplate
```

#### 2. Install Dependencies

```bash
# Download Go dependencies
go mod download

# Verify installation
go mod verify
```

#### 3. Set Environment Variables

```bash
# Generate JWT secret
export JWT_SECRET=$(openssl rand -base64 32)

# Optional: Set custom values
export SERVER_PORT=8080
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=user
export DB_PASSWORD=password
export DB_NAME=baseplate
```

**Or create `.env` file**:
```bash
JWT_SECRET=your-secure-secret-here
SERVER_PORT=8080
GIN_MODE=debug
DB_HOST=localhost
DB_PORT=5432
DB_USER=user
DB_PASSWORD=password
DB_NAME=baseplate
DB_SSL_MODE=disable
```

#### 4. Start Database

```bash
make db-up
```

Wait 3 seconds for migrations to complete.

#### 5. Run Application

```bash
make run
```

#### 6. Verify Setup

```bash
# Health check
curl http://localhost:8080/api/health
# Response: {"status":"ok"}

# Register a user
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test User",
    "email": "test@example.com",
    "password": "password123"
  }'
```

---

## Project Structure

```
baseplate/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── config/
│   └── config.go                # Configuration loading
├── internal/
│   ├── api/
│   │   ├── router.go           # Route definitions
│   │   ├── handlers/           # HTTP handlers
│   │   │   ├── auth.go         # Auth endpoints (3)
│   │   │   ├── team.go         # Team/role/member/API key (11)
│   │   │   ├── blueprint.go    # Blueprint CRUD (5)
│   │   │   └── entity.go       # Entity CRUD + search (8)
│   │   └── middleware/         # HTTP middleware
│   │       ├── auth.go         # Authentication & RBAC
│   │       └── error.go        # Error handling
│   ├── core/                   # Business logic
│   │   ├── auth/
│   │   │   ├── models.go       # Domain models
│   │   │   ├── service.go      # Business logic
│   │   │   └── repository.go   # Data access
│   │   ├── blueprint/
│   │   │   ├── models.go
│   │   │   ├── service.go
│   │   │   └── repository.go
│   │   ├── entity/
│   │   │   ├── models.go
│   │   │   ├── service.go
│   │   │   └── repository.go
│   │   └── validation/
│   │       └── validator.go    # JSON Schema validator
│   └── storage/
│       └── postgres/
│           └── client.go       # Database connection
├── migrations/
│   └── 001_initial.sql         # Database schema
├── docs/                       # Documentation
├── .gitignore
├── go.mod                      # Go dependencies
├── go.sum                      # Dependency checksums
├── Makefile                    # Build commands
├── docker-compose.yaml         # Local PostgreSQL
└── README.md
```

### Directory Descriptions

**`cmd/server/`**: Application entry point, initializes components and starts server

**`config/`**: Configuration loading from environment variables

**`internal/api/`**: HTTP layer - routing, handlers, middleware

**`internal/core/`**: Business logic layer - domain models, services, repositories

**`internal/storage/`**: Infrastructure layer - database connection

**`migrations/`**: SQL migration files for database schema

**`docs/`**: Project documentation (API, architecture, etc.)

---

## Development Workflow

### Makefile Commands

```bash
# Database
make db-up          # Start PostgreSQL container
make db-down        # Stop PostgreSQL container
make db-reset       # Drop and recreate database (deletes all data!)
make migrate        # Run migrations manually

# Development
make run            # Run server (hot reload via go run)
make build          # Build binary to bin/server
make clean          # Remove bin/ directory

# Super Admin Setup
make init-superadmin # Initialize first super admin (requires SUPER_ADMIN_EMAIL, SUPER_ADMIN_PASSWORD env vars)

# Code Quality
make fmt            # Format code (go fmt)
make lint           # Run linter (requires golangci-lint)
make test           # Run tests
make tidy           # Run go mod tidy

# Documentation
make swagger        # Generate Swagger docs (requires swag)
```

### Daily Development Cycle

```bash
# 1. Start database (once per session)
make db-up

# 2. Run application (hot reload)
make run

# 3. Make code changes

# 4. Test changes with curl/Postman

# 5. Format code
make fmt

# 6. Run tests
make test

# 7. Commit changes
git add .
git commit -m "Add feature X"
```

---

## Code Organization

### Layered Architecture

Baseplate follows a clean layered architecture:

```
Handler → Service → Repository → Database
```

**Responsibilities**:

1. **Handler** (`internal/api/handlers/`):
   - HTTP request/response binding
   - Input validation
   - Call service methods
   - Format responses
   - **No business logic**

2. **Service** (`internal/core/*/service.go`):
   - Business logic orchestration
   - Cross-domain operations
   - Validation coordination
   - **No HTTP concerns**

3. **Repository** (`internal/core/*/repository.go`):
   - SQL queries
   - Data mapping
   - **No business logic**

### Code Conventions

**Package Naming**:
```go
package auth     // lowercase, singular
```

**File Naming**:
```go
models.go        // Domain models
service.go       // Business logic
repository.go    // Data access
```

**Struct Naming**:
```go
type User struct { }          // Domain model
type UserService struct { }   // Service
type UserRepository struct { } // Repository
```

**Interface Naming**:
```go
type EntityRepository interface { } // Noun, not "EntityRepositoryInterface"
```

**Error Handling**:
```go
// Return errors, don't panic
func Create() (*Entity, error) {
    if err := validate(); err != nil {
        return nil, err
    }
    return entity, nil
}
```

---

## Adding Features

### Adding a New Endpoint

**Example**: Add `GET /api/teams/:teamId/stats` endpoint

#### 1. Define Model (if needed)

**`internal/core/auth/models.go`**:
```go
type TeamStats struct {
    TotalMembers    int `json:"total_members"`
    TotalBlueprints int `json:"total_blueprints"`
    TotalEntities   int `json:"total_entities"`
}
```

#### 2. Add Repository Method

**`internal/core/auth/repository.go`**:
```go
func (r *AuthRepository) GetTeamStats(ctx context.Context, teamID uuid.UUID) (*TeamStats, error) {
    stats := &TeamStats{}

    // Count members
    err := r.db.QueryRow(`
        SELECT COUNT(*) FROM team_memberships WHERE team_id = $1
    `, teamID).Scan(&stats.TotalMembers)
    if err != nil {
        return nil, err
    }

    // Count blueprints
    err = r.db.QueryRow(`
        SELECT COUNT(*) FROM blueprints WHERE team_id = $1
    `, teamID).Scan(&stats.TotalBlueprints)
    if err != nil {
        return nil, err
    }

    // Count entities
    err = r.db.QueryRow(`
        SELECT COUNT(*) FROM entities WHERE team_id = $1
    `, teamID).Scan(&stats.TotalEntities)
    if err != nil {
        return nil, err
    }

    return stats, nil
}
```

#### 3. Add Service Method

**`internal/core/auth/service.go`**:
```go
func (s *AuthService) GetTeamStats(ctx context.Context, teamID uuid.UUID) (*TeamStats, error) {
    // Verify team exists
    team, err := s.repo.GetTeamByID(ctx, teamID)
    if err != nil {
        return nil, err
    }
    if team == nil {
        return nil, ErrTeamNotFound
    }

    // Get stats
    return s.repo.GetTeamStats(ctx, teamID)
}
```

#### 4. Add Handler

**`internal/api/handlers/team.go`**:
```go
func (h *TeamHandler) GetTeamStats(c *gin.Context) {
    // Parse team ID
    teamIDStr := c.Param("teamId")
    teamID, err := uuid.Parse(teamIDStr)
    if err != nil {
        c.JSON(400, gin.H{"error": "invalid team ID"})
        return
    }

    // Get team ID from context (verify access)
    contextTeamID := c.GetString("team_id")
    if teamID.String() != contextTeamID {
        c.JSON(403, gin.H{"error": "access denied"})
        return
    }

    // Get stats
    stats, err := h.authService.GetTeamStats(c.Request.Context(), teamID)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, stats)
}
```

#### 5. Register Route

**`internal/api/router.go`**:
```go
teams := api.Group("/teams")
teams.Use(middleware.Authenticate())
{
    teams.GET("/:teamId", middleware.RequireTeam(), teamHandler.GetTeam)
    teams.GET("/:teamId/stats", middleware.RequireTeam(), teamHandler.GetTeamStats) // Add this
}
```

#### 6. Test Endpoint

```bash
curl -X GET http://localhost:8080/api/teams/$TEAM_ID/stats \
  -H "Authorization: Bearer $TOKEN"
```

---

### Adding a New Permission

**Example**: Add `report:generate` permission

#### 1. Document Permission

**`docs/SECURITY.md`** and **`docs/API.md`**:
```markdown
| `report:generate` | Generate and download reports |
```

#### 2. Update Default Roles (if needed)

**`internal/core/auth/service.go`** in `createDefaultRoles()`:
```go
adminPermissions := []string{
    "team:manage",
    "blueprint:read", "blueprint:write", "blueprint:delete",
    "entity:read", "entity:write", "entity:delete",
    "report:generate", // Add to admin
}
```

#### 3. Use in Middleware

**`internal/api/router.go`**:
```go
reports := api.Group("/reports")
reports.Use(middleware.Authenticate())
reports.Use(middleware.RequireTeam())
{
    reports.POST("", middleware.RequirePermission("report:generate"), reportHandler.Generate)
}
```

---

### Adding a Database Migration

**Example**: Add `archived` column to entities

#### 1. Create Migration File

**`migrations/002_add_entity_archived.sql`**:
```sql
-- Migration: Add archived status to entities
-- Date: 2024-01-15
-- Description: Allow entities to be archived without deletion

BEGIN;

-- Add column
ALTER TABLE entities
ADD COLUMN archived BOOLEAN DEFAULT FALSE;

-- Add index
CREATE INDEX idx_entities_archived ON entities(archived);

-- Add comment
COMMENT ON COLUMN entities.archived IS 'Soft delete flag';

COMMIT;
```

#### 2. Update Model

**`internal/core/entity/models.go`**:
```go
type Entity struct {
    ID          uuid.UUID              `json:"id"`
    TeamID      uuid.UUID              `json:"team_id"`
    BlueprintID string                 `json:"blueprint_id"`
    Identifier  string                 `json:"identifier"`
    Title       string                 `json:"title"`
    Data        map[string]interface{} `json:"data"`
    Archived    bool                   `json:"archived"` // Add this
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
}
```

#### 3. Update Queries

**`internal/core/entity/repository.go`**:
```go
// Update GetAll to exclude archived
func (r *EntityRepository) GetAll(ctx context.Context, teamID uuid.UUID, blueprintID string, limit, offset int) ([]*Entity, int, error) {
    query := `
        SELECT id, team_id, blueprint_id, identifier, title, data, archived, created_at, updated_at
        FROM entities
        WHERE team_id = $1 AND blueprint_id = $2 AND archived = false
        ORDER BY created_at DESC
        LIMIT $3 OFFSET $4
    `
    // ...
}
```

#### 4. Apply Migration

```bash
# Automated (on container start)
make db-reset

# Manual
psql -U user -d baseplate -f migrations/002_add_entity_archived.sql
```

---

## Testing

### Running Tests

```bash
# Run all tests
make test

# Run specific package
go test -v ./internal/core/auth/...

# Run specific test
go test -v -run TestCreateUser ./internal/core/auth/...

# Run with coverage
go test -v -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Writing Tests

**Example Test** (`internal/core/auth/service_test.go`):

```go
package auth_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/your-org/baseplate/internal/core/auth"
)

func TestRegister(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    defer db.Close()

    repo := auth.NewAuthRepository(db)
    service := auth.NewAuthService(repo, "test-jwt-secret")

    // Test
    req := &auth.RegisterRequest{
        Name:     "Test User",
        Email:    "test@example.com",
        Password: "password123",
    }

    token, user, err := service.Register(context.Background(), req)

    // Assert
    assert.NoError(t, err)
    assert.NotEmpty(t, token)
    assert.Equal(t, "test@example.com", user.Email)
    assert.Equal(t, "Test User", user.Name)
}

func setupTestDB(t *testing.T) *sql.DB {
    // Connect to test database
    db, err := sql.Open("postgres", "postgres://user:password@localhost:5432/baseplate_test")
    if err != nil {
        t.Fatalf("Failed to connect to test DB: %v", err)
    }

    // Run migrations
    // ...

    return db
}
```

### Test Best Practices

- **Isolation**: Each test should be independent
- **Setup/Teardown**: Use test database, clean up after tests
- **Naming**: `TestFunctionName_Scenario_ExpectedResult`
- **Coverage**: Aim for 80%+ coverage on business logic
- **Mocking**: Mock external dependencies (database, APIs)

---

## Super Admin Testing

### Manual Testing Flow

```bash
# 1. Start database and app
make db-up
make run

# 2. Create super admin
export SUPER_ADMIN_EMAIL="admin@test.com"
export SUPER_ADMIN_PASSWORD="AdminPass123"
make init-superadmin

# 3. Get JWT token
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@test.com",
    "password": "AdminPass123"
  }' | jq -r '.token')

# 4. Test super admin endpoints
curl -X GET http://localhost:8080/api/admin/users \
  -H "Authorization: Bearer $TOKEN"

# 5. Test promotion (promote a different user first)
curl -s -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test User",
    "email": "test@example.com",
    "password": "TestPass123"
  }' | jq .

# Get user ID from registration response, then promote
USER_ID="<user-id-from-registration>"
curl -X POST http://localhost:8080/api/admin/users/$USER_ID/promote \
  -H "Authorization: Bearer $TOKEN"

# 6. Test demotion
curl -X POST http://localhost:8080/api/admin/users/$USER_ID/demote \
  -H "Authorization: Bearer $TOKEN"

# 7. Test audit logs
curl -X GET "http://localhost:8080/api/admin/audit-logs?limit=10&offset=0" \
  -H "Authorization: Bearer $TOKEN" | jq '.logs'
```

### Transaction Safety Testing

```bash
# Test that last super admin cannot be demoted:

# 1. Get the super admin's user ID
ADMIN_ID=$(curl -s -X GET http://localhost:8080/api/admin/users \
  -H "Authorization: Bearer $TOKEN" | jq -r '.users[] | select(.is_super_admin==true) | .id' | head -1)

# 2. Attempt to demote (should fail with 409 Conflict)
curl -X POST http://localhost:8080/api/admin/users/$ADMIN_ID/demote \
  -H "Authorization: Bearer $TOKEN" \
  -w "\nStatus: %{http_code}\n"

# Expected: 409 Conflict with error "cannot demote the last super admin"
```

### Permission Bypass Testing

```bash
# Super admins can access resources without team membership:

# 1. Create a team (as regular user)
TOKEN_USER=$(curl -s -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Regular User",
    "email": "user@example.com",
    "password": "UserPass123"
  }' | jq -r '.token')

TEAM_ID=$(curl -s -X POST http://localhost:8080/api/teams \
  -H "Authorization: Bearer $TOKEN_USER" \
  -H "Content-Type: application/json" \
  -d '{"name": "Test Team", "slug": "test-team"}' | jq -r '.id')

# 2. Super admin can access team without membership
curl -X GET http://localhost:8080/api/admin/teams/$TEAM_ID \
  -H "Authorization: Bearer $TOKEN"

# Expected: 200 OK with team details (super admin bypasses membership check)
```

---

## Debugging

### Using VS Code Debugger

**`.vscode/launch.json`**:
```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch Server",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "program": "${workspaceFolder}/cmd/server",
            "env": {
                "JWT_SECRET": "test-secret-for-debugging",
                "GIN_MODE": "debug",
                "DB_HOST": "localhost",
                "DB_PORT": "5432",
                "DB_USER": "user",
                "DB_PASSWORD": "password",
                "DB_NAME": "baseplate"
            },
            "args": []
        }
    ]
}
```

**Set breakpoints** in code, press F5 to start debugging.

---

### Logging

**Add Debug Logs**:
```go
import "log"

func (s *EntityService) Create(ctx context.Context, req *CreateEntityRequest) (*Entity, error) {
    log.Printf("Creating entity: blueprint=%s, identifier=%s", req.BlueprintID, req.Identifier)

    // ... business logic

    log.Printf("Entity created: id=%s", entity.ID)
    return entity, nil
}
```

**View Logs**:
```bash
# With make run
make run

# Logs appear in terminal
```

---

### Database Queries

**psql Access**:
```bash
# Connect to database
docker exec -it baseplate_db psql -U user -d baseplate

# List tables
\dt

# Describe table
\d entities

# Query data
SELECT * FROM users;
SELECT * FROM entities WHERE team_id = 'uuid-here';

# Quit
\q
```

**Query Debugging**:
```go
// Print query before execution
log.Printf("Query: %s", query)
log.Printf("Params: %v", params)

// Measure query time
start := time.Now()
rows, err := db.Query(query, params...)
log.Printf("Query took: %v", time.Since(start))
```

---

## Git Workflow

### Branch Strategy

**Main Branches**:
- `main`: Production-ready code
- `develop`: Integration branch for features

**Feature Branches**:
```bash
# Create feature branch
git checkout -b feature/add-reports

# Make changes, commit
git add .
git commit -m "Add reports endpoint"

# Push to remote
git push origin feature/add-reports

# Create pull request on GitHub
```

**Branch Naming**:
- `feature/description`: New features
- `fix/description`: Bug fixes
- `docs/description`: Documentation
- `refactor/description`: Code refactoring

---

### Commit Messages

**Format**:
```
<type>: <subject>

<body>

<footer>
```

**Types**:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `style`: Formatting
- `refactor`: Code restructuring
- `test`: Adding tests
- `chore`: Maintenance

**Examples**:
```bash
# Good
git commit -m "feat: add team statistics endpoint"
git commit -m "fix: correct entity search filter validation"
git commit -m "docs: update API documentation for new endpoint"

# Bad
git commit -m "changes"
git commit -m "fix stuff"
git commit -m "wip"
```

---

### Pull Request Process

1. **Create Feature Branch**
2. **Make Changes and Commit**
3. **Push to Remote**
4. **Create Pull Request**
5. **Code Review**
6. **Address Feedback**
7. **Merge to Main**

**Pull Request Template**:
```markdown
## Description
Brief description of changes

## Changes
- Added X feature
- Fixed Y bug
- Updated Z documentation

## Testing
- [ ] Unit tests added/updated
- [ ] Manual testing completed
- [ ] Documentation updated

## Related Issues
Closes #123
```

---

## Code Style

### Go Style Guide

Follow official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).

**Key Points**:

**Formatting**:
```bash
# Format all code
make fmt

# Or
go fmt ./...
```

**Naming**:
```go
// Good
type UserService struct {}
func (s *UserService) GetByID() {}

// Bad
type UserServiceImpl struct {}
func (s *UserService) getUserById() {}
```

**Error Handling**:
```go
// Check all errors
result, err := someFunction()
if err != nil {
    return nil, err
}

// Don't ignore errors
result, _ := someFunction() // Bad
```

**Comments**:
```go
// Exported functions must have comments
// GetByID retrieves a user by their unique identifier.
func GetByID(id uuid.UUID) (*User, error) {
    // ...
}
```

---

### Linting

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
make lint

# Or
golangci-lint run
```

**Configure** (`.golangci.yml`):
```yaml
linters:
  enable:
    - gofmt
    - govet
    - errcheck
    - staticcheck
    - unused
    - gosimple
    - structcheck
    - varcheck
    - ineffassign
    - deadcode
```

---

## Useful Tools

**Development**:
- [Air](https://github.com/cosmtrek/air): Live reload for Go apps
- [Delve](https://github.com/go-delve/delve): Go debugger
- [Postman](https://www.postman.com/): API testing

**Database**:
- [DBeaver](https://dbeaver.io/): Universal database client
- [pgcli](https://www.pgcli.com/): PostgreSQL CLI with auto-completion

**Code Quality**:
- [golangci-lint](https://golangci-lint.run/): Fast Go linters runner
- [gofumpt](https://github.com/mvdan/gofumpt): Stricter gofmt

---

## Additional Resources

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Standard Library](https://pkg.go.dev/std)
- [Gin Framework Docs](https://gin-gonic.com/docs/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)

---

For deployment instructions, see [DEPLOYMENT.md](./DEPLOYMENT.md).
For API reference, see [API.md](./API.md).
For architecture details, see [ARCHITECTURE.md](./ARCHITECTURE.md).
