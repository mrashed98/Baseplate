# Baseplate Architecture

Comprehensive technical architecture documentation for Baseplate - a headless backend engine with dynamic schema management.

## Table of Contents

- [System Overview](#system-overview)
- [Technology Stack](#technology-stack)
- [System Architecture](#system-architecture)
- [Request Flow](#request-flow)
- [Database Schema](#database-schema)
- [Authentication System](#authentication-system)
- [Multi-Tenancy](#multi-tenancy)
- [Component Details](#component-details)
- [Design Patterns](#design-patterns)

## System Overview

Baseplate is a production-ready headless backend engine inspired by Port.ai that enables dynamic schema management through "Blueprints". It eliminates the need for database migrations or code changes when defining new entity types, making it highly flexible for evolving data models.

**Key Capabilities**:
- Dynamic schema definition using JSON Schema
- Multi-tenant architecture with team-based isolation
- Comprehensive RBAC (Role-Based Access Control)
- Dual authentication (JWT tokens and API keys)
- Advanced entity search with JSONB queries
- RESTful API with 28 endpoints

## Technology Stack

**Core Technologies**:
- **Language**: Go 1.25.1
- **Web Framework**: Gin v1.11.0
- **Database**: PostgreSQL 15 with JSONB support
- **Authentication**: JWT (golang-jwt/jwt/v5) + API Keys
- **Password Hashing**: bcrypt (golang.org/x/crypto)
- **JSON Schema Validation**: gojsonschema v1.2.0
- **UUID Generation**: google/uuid v1.6.0

**Infrastructure**:
- **Containerization**: Docker with docker-compose
- **Database Driver**: lib/pq v1.10.9

## System Architecture

```mermaid
graph TB
    subgraph "Client Layer"
        CLIENT[HTTP Client]
        BROWSER[Web Browser]
        SERVICE[External Service]
    end

    subgraph "API Layer"
        ROUTER[Gin Router]
        MW_AUTH[Auth Middleware]
        MW_TEAM[Team Context Middleware]
        MW_SUPER[Super Admin Check]
        MW_PERM[Permission Middleware]
        MW_AUDIT[Audit Middleware]
        MW_ERR[Error Handler]

        HANDLER_AUTH[Auth Handler]
        HANDLER_TEAM[Team Handler]
        HANDLER_BP[Blueprint Handler]
        HANDLER_ENT[Entity Handler]
        HANDLER_ADMIN[Admin Handler]
    end

    subgraph "Service Layer"
        SVC_AUTH[Auth Service]
        SVC_BP[Blueprint Service]
        SVC_ENT[Entity Service]
        VALIDATOR[JSON Schema Validator]
    end

    subgraph "Repository Layer"
        REPO_AUTH[Auth Repository]
        REPO_BP[Blueprint Repository]
        REPO_ENT[Entity Repository]
    end

    subgraph "Data Layer"
        POSTGRES[(PostgreSQL 15)]
        JSONB[JSONB Storage]
    end

    CLIENT --> ROUTER
    BROWSER --> ROUTER
    SERVICE --> ROUTER

    ROUTER --> MW_ERR
    MW_ERR --> MW_AUDIT
    MW_AUDIT --> MW_AUTH
    MW_AUTH --> MW_TEAM
    MW_TEAM --> MW_SUPER
    MW_SUPER --> MW_PERM

    MW_PERM --> HANDLER_AUTH
    MW_PERM --> HANDLER_TEAM
    MW_PERM --> HANDLER_BP
    MW_PERM --> HANDLER_ENT
    MW_PERM --> HANDLER_ADMIN

    HANDLER_AUTH --> SVC_AUTH
    HANDLER_TEAM --> SVC_AUTH
    HANDLER_ADMIN --> SVC_AUTH
    HANDLER_BP --> SVC_BP
    HANDLER_ENT --> SVC_ENT

    SVC_ENT --> VALIDATOR
    SVC_ENT --> SVC_BP

    SVC_AUTH --> REPO_AUTH
    SVC_BP --> REPO_BP
    SVC_ENT --> REPO_ENT

    REPO_AUTH --> POSTGRES
    REPO_BP --> POSTGRES
    REPO_ENT --> POSTGRES

    POSTGRES --> JSONB
```

## Request Flow

### Complete Request-Response Cycle

```mermaid
sequenceDiagram
    participant Client
    participant Router
    participant Recovery
    participant Logger
    participant ErrorHandler
    participant AuthMW
    participant TeamMW
    participant PermMW
    participant Handler
    participant Service
    participant Validator
    participant Repository
    participant Database

    Client->>Router: HTTP Request
    Router->>Recovery: Pass through
    Recovery->>Logger: Log request
    Logger->>ErrorHandler: Error wrapper
    ErrorHandler->>AuthMW: Authenticate

    alt JWT Token
        AuthMW->>AuthMW: Validate JWT
        AuthMW->>AuthMW: Set user_id in context
    else API Key
        AuthMW->>Database: Lookup key hash
        Database-->>AuthMW: API key + team_id + permissions
        AuthMW->>AuthMW: Set user_id, team_id, permissions
    end

    AuthMW->>TeamMW: Team context required?

    alt Team Required
        TeamMW->>TeamMW: Extract team_id from URL/Header/Context
        TeamMW->>Database: Verify user has team access
        Database-->>TeamMW: Membership confirmed
        TeamMW->>TeamMW: Set team_id, permissions in context
    end

    TeamMW->>PermMW: Permission check
    PermMW->>PermMW: Check required permission in context

    alt Permission Denied
        PermMW-->>Client: 403 Forbidden
    else Permission Granted
        PermMW->>Handler: Route to handler
        Handler->>Handler: Bind JSON & validate
        Handler->>Service: Business logic

        alt Entity Operation
            Service->>Validator: Validate against schema
            Validator-->>Service: Validation result
        end

        Service->>Repository: Data operation
        Repository->>Database: SQL query
        Database-->>Repository: Result set
        Repository-->>Service: Mapped data
        Service-->>Handler: Response data
        Handler-->>Client: JSON Response
    end
```

### Middleware Chain

```mermaid
graph LR
    A[Request] --> B[gin.Recovery]
    B --> C[gin.Logger]
    C --> D[ErrorHandler]
    D --> E[AuditMiddleware]
    E --> F[Authenticate]
    F --> G{Auth Type}

    G -->|JWT| H[Extract user_id + is_super_admin]
    G -->|API Key| I[Extract user_id + team_id + permissions]

    H --> J[RequireTeam]
    I --> J

    J --> K{Team Needed?}
    K -->|Yes| L[Extract team_id from URL/Header]
    K -->|No| M_SUPER[RequireSuperAdmin]

    L --> M_CHECK{Is Super Admin?}
    M_CHECK -->|Yes| N[Grant AllPermissions]
    M_CHECK -->|No| O[RequirePermission]

    M_SUPER --> M_SUPER_CHECK{Is Super Admin?}
    M_SUPER_CHECK -->|Yes| P[Handler]
    M_SUPER_CHECK -->|No| R[403 Forbidden]

    N --> P
    O --> Q{Has Permission?}
    Q -->|Yes| P
    Q -->|No| R

    P --> S[Response]
```

## Database Schema

### Entity Relationship Diagram

```mermaid
erDiagram
    USERS ||--o{ TEAM_MEMBERSHIPS : "has"
    USERS ||--o{ API_KEYS : "creates"

    TEAMS ||--o{ TEAM_MEMBERSHIPS : "contains"
    TEAMS ||--o{ ROLES : "defines"
    TEAMS ||--o{ API_KEYS : "owns"
    TEAMS ||--o{ BLUEPRINTS : "owns"
    TEAMS ||--o{ ENTITIES : "owns"
    TEAMS ||--o{ SCORECARDS : "owns"
    TEAMS ||--o{ INTEGRATIONS : "configures"
    TEAMS ||--o{ ACTIONS : "defines"

    ROLES ||--o{ TEAM_MEMBERSHIPS : "assigned"

    TEAM_MEMBERSHIPS {
        uuid id PK
        uuid team_id FK
        uuid user_id FK
        uuid role_id FK
        timestamp created_at
    }

    USERS {
        uuid id PK
        varchar email UK
        varchar password_hash
        varchar name
        varchar status
        timestamp created_at
    }

    TEAMS {
        uuid id PK
        varchar name
        varchar slug UK
        timestamp created_at
    }

    ROLES {
        uuid id PK
        uuid team_id FK
        varchar name
        jsonb permissions
        timestamp created_at
    }

    API_KEYS {
        uuid id PK
        uuid team_id FK
        uuid user_id FK
        varchar name
        varchar key_hash
        jsonb permissions
        timestamp expires_at
        timestamp last_used_at
        timestamp created_at
    }

    BLUEPRINTS ||--o{ ENTITIES : "defines"
    BLUEPRINTS ||--o{ BLUEPRINT_RELATIONS : "source"
    BLUEPRINTS ||--o{ BLUEPRINT_RELATIONS : "target"
    BLUEPRINTS ||--o{ SCORECARDS : "measured"
    BLUEPRINTS ||--o{ INTEGRATION_MAPPINGS : "mapped"
    BLUEPRINTS ||--o{ ACTIONS : "triggers"

    BLUEPRINTS {
        varchar id PK
        uuid team_id FK
        varchar title
        text description
        varchar icon
        jsonb schema
        timestamp created_at
        timestamp updated_at
    }

    ENTITIES {
        uuid id PK
        uuid team_id FK
        varchar blueprint_id FK
        varchar identifier UK
        varchar title
        jsonb data
        timestamp created_at
        timestamp updated_at
    }

    ENTITIES ||--o{ ENTITY_RELATIONS : "source"
    ENTITIES ||--o{ ENTITY_RELATIONS : "target"

    BLUEPRINT_RELATIONS ||--o{ ENTITY_RELATIONS : "instantiates"

    BLUEPRINT_RELATIONS {
        uuid id PK
        uuid team_id FK
        varchar source_blueprint_id FK
        varchar target_blueprint_id FK
        varchar identifier
        varchar title
        varchar relation_type
        boolean required
        timestamp created_at
    }

    ENTITY_RELATIONS {
        uuid id PK
        uuid relation_id FK
        uuid source_entity_id FK
        uuid target_entity_id FK
        timestamp created_at
    }

    SCORECARDS {
        uuid id PK
        uuid team_id FK
        varchar blueprint_id FK
        varchar identifier
        varchar title
        jsonb levels
        timestamp created_at
    }

    SCORECARDS ||--o{ SCORECARD_RULES : "contains"

    SCORECARD_RULES {
        uuid id PK
        uuid scorecard_id FK
        varchar level_name
        varchar property_path
        varchar operator
        jsonb value
        timestamp created_at
    }

    INTEGRATIONS {
        uuid id PK
        uuid team_id FK
        varchar type
        varchar name
        jsonb config
        varchar status
        timestamp last_sync_at
        timestamp created_at
    }

    INTEGRATIONS ||--o{ INTEGRATION_MAPPINGS : "uses"

    INTEGRATION_MAPPINGS {
        uuid id PK
        uuid integration_id FK
        varchar blueprint_id FK
        varchar external_type
        jsonb mapping
        jsonb filter
        timestamp created_at
    }

    ACTIONS {
        uuid id PK
        uuid team_id FK
        varchar blueprint_id FK
        varchar identifier
        varchar title
        text description
        varchar trigger_type
        jsonb trigger_config
        jsonb user_inputs
        jsonb steps
        timestamp created_at
    }

    AUDIT_LOGS {
        uuid id PK
        uuid team_id FK
        uuid user_id FK
        varchar entity_type
        varchar entity_id
        varchar action
        jsonb old_data
        jsonb new_data
        timestamp created_at
    }
```

### Key Indexes

**Performance Indexes**:
- `idx_entities_data` - GIN index on JSONB data for fast queries
- `idx_api_keys_hash` - Hash lookup for API key authentication
- `idx_team_memberships_user` - Fast user membership lookups
- `idx_team_memberships_team` - Fast team member listings

**Relationship Indexes**:
- `idx_entities_blueprint` - Entity filtering by blueprint
- `idx_entity_relations_source/target` - Relationship traversal
- `idx_scorecards_blueprint` - Scorecard lookups

**Cascade Behavior**:
- Deleting a team cascades to all team resources (blueprints, entities, roles, etc.)
- Deleting a user cascades to memberships, sets API keys' user_id to NULL
- Deleting a blueprint cascades to entities and relations

## Authentication System

### JWT Token Flow

```mermaid
sequenceDiagram
    participant User
    participant API
    participant AuthService
    participant Database
    participant JWTLib

    User->>API: POST /api/auth/login
    API->>AuthService: Login(email, password)
    AuthService->>Database: SELECT user WHERE email = ?
    Database-->>AuthService: User record
    AuthService->>AuthService: bcrypt.CompareHashAndPassword

    alt Password Valid
        AuthService->>JWTLib: Generate token (HS256)
        JWTLib-->>AuthService: JWT token
        AuthService-->>API: {token, user}
        API-->>User: 200 OK + JWT

        User->>API: Request with Authorization: Bearer <token>
        API->>JWTLib: Parse and validate token
        JWTLib-->>API: Claims {user_id, email, exp}
        API->>API: Set user_id in context
        API->>API: Continue to handler
    else Password Invalid
        AuthService-->>API: Error
        API-->>User: 401 Unauthorized
    end
```

**JWT Properties**:
- Algorithm: HS256 (HMAC-SHA256)
- Secret: From `JWT_SECRET` environment variable
- Expiration: 24 hours (configurable)
- Claims: `user_id`, `email`, `issued_at`, `expires_at`
- Location: `internal/core/auth/service.go`

### API Key Flow

```mermaid
sequenceDiagram
    participant Admin
    participant API
    participant AuthService
    participant Database
    participant CryptoLib

    Admin->>API: POST /api/teams/:teamId/api-keys
    API->>AuthService: CreateAPIKey(name, permissions)
    AuthService->>CryptoLib: Generate 32 random bytes
    CryptoLib-->>AuthService: Random bytes
    AuthService->>AuthService: Hex encode + "bp_" prefix
    AuthService->>CryptoLib: SHA256 hash
    CryptoLib-->>AuthService: Hash
    AuthService->>Database: INSERT api_key (hash, team_id, permissions)
    Database-->>AuthService: Success
    AuthService-->>API: {api_key object + raw key}
    API-->>Admin: 201 Created + key (shown once!)

    Note over Admin: Store key securely

    Admin->>API: Request with Authorization: ApiKey <key>
    API->>CryptoLib: SHA256 hash incoming key
    CryptoLib-->>API: Hash
    API->>Database: SELECT api_key WHERE hash = ?
    Database-->>API: API key record
    API->>API: Check expiration
    API->>API: Set team_id, permissions, user_id in context

    par Async Update
        API->>Database: UPDATE last_used_at (async)
    end

    API->>API: Continue to handler
```

**API Key Properties**:
- Format: `bp_<64_hex_characters>`
- Storage: SHA-256 hash in database
- Team-scoped with optional permissions
- Optional expiration timestamp
- Async last-used tracking
- Location: `internal/core/auth/service.go:325-384`

### Password Security

- **Algorithm**: bcrypt with default cost (10)
- **Salt**: Automatically generated per password
- **Storage**: In `users.password_hash` column
- **Validation**: `bcrypt.CompareHashAndPassword`
- **Location**: `internal/core/auth/service.go`

## Multi-Tenancy

### Team Isolation Model

```mermaid
graph TB
    subgraph "Team A"
        USER_A1[User 1]
        USER_A2[User 2]
        ROLE_A[Admin Role]
        BP_A[Blueprints]
        ENT_A[Entities]
        API_A[API Keys]
    end

    subgraph "Team B"
        USER_B1[User 3]
        ROLE_B[Editor Role]
        BP_B[Blueprints]
        ENT_B[Entities]
        API_B[API Keys]
    end

    subgraph "Shared User"
        USER_SHARED[User 2]
    end

    USER_A1 --> ROLE_A
    USER_A2 --> ROLE_A
    ROLE_A --> BP_A
    ROLE_A --> ENT_A
    ROLE_A --> API_A

    USER_B1 --> ROLE_B
    USER_SHARED --> ROLE_B
    ROLE_B --> BP_B
    ROLE_B --> ENT_B
    ROLE_B --> API_B

    style USER_A2 fill:#ff9
    style USER_SHARED fill:#ff9
    style BP_A fill:#e1f5ff
    style BP_B fill:#f3e5f5
    style ENT_A fill:#e1f5ff
    style ENT_B fill:#f3e5f5
```

**Isolation Guarantees**:
1. All resources (blueprints, entities, roles) scoped to `team_id`
2. Users can belong to multiple teams with different roles
3. API keys are team-specific
4. Database queries always filter by `team_id`
5. Middleware enforces team membership before access

**Team Context Resolution Order**:
1. API Key → Automatic from key's team association
2. URL Parameter → `/teams/:teamId/...`
3. Header → `X-Team-ID: <uuid>`

## Component Details

### Layer Responsibilities

| Layer | Location | Responsibilities | Rules |
|-------|----------|-----------------|-------|
| **Handlers** | `internal/api/handlers/` | - HTTP request/response binding<br>- Input validation<br>- Response formatting<br>- Error handling | - No business logic<br>- No database access<br>- Thin layer |
| **Services** | `internal/core/*/service.go` | - Business logic orchestration<br>- Cross-domain operations<br>- Validation coordination<br>- Transaction management | - No HTTP concerns<br>- Testable without HTTP<br>- Core domain logic |
| **Repositories** | `internal/core/*/repository.go` | - SQL query execution<br>- Data mapping (SQL ↔ Go)<br>- JSONB operations<br>- Query optimization | - No business logic<br>- Pure data access<br>- SQL expertise |
| **Validation** | `internal/core/validation/` | - JSON Schema validation<br>- Full and partial validation<br>- Error reporting | - Schema-driven<br>- Framework-agnostic |

### Package Structure

```
internal/
├── api/
│   ├── router.go                 # Route setup, middleware chain
│   ├── handlers/
│   │   ├── auth.go              # Auth endpoints (3)
│   │   ├── team.go              # Team/role/member/API key (11)
│   │   ├── blueprint.go         # Blueprint CRUD (5)
│   │   └── entity.go            # Entity CRUD + search (8)
│   └── middleware/
│       ├── auth.go              # JWT/API key auth + RBAC
│       └── error.go             # Global error handling
├── core/
│   ├── auth/
│   │   ├── models.go            # User, Team, Role, APIKey
│   │   ├── service.go           # Auth business logic
│   │   └── repository.go        # Auth data access
│   ├── blueprint/
│   │   ├── models.go            # Blueprint structs
│   │   ├── service.go           # Blueprint business logic
│   │   └── repository.go        # Blueprint data access
│   ├── entity/
│   │   ├── models.go            # Entity, SearchRequest
│   │   ├── service.go           # Entity business logic
│   │   └── repository.go        # Entity data access + search
│   └── validation/
│       └── validator.go         # JSON Schema validator
└── storage/
    └── postgres/
        └── client.go            # Database connection
```

### Dependency Flow

```mermaid
graph LR
    A[main.go] --> B[config]
    A --> C[router]
    C --> D[handlers]
    D --> E[services]
    E --> F[repositories]
    F --> G[postgres client]
    E --> H[validator]

    style A fill:#e1f5ff
    style E fill:#f3e5f5
    style F fill:#fff3e0
```

**Key Dependencies**:
- Services depend on repositories
- Entity service depends on blueprint service
- Entity service depends on validator
- No circular dependencies

## Design Patterns

### 1. Repository Pattern

Separates data access logic from business logic.

```go
// Service uses repository interface
type EntityService struct {
    repo EntityRepository
    blueprintService *blueprint.Service
    validator *validation.Validator
}

// Repository implements data access
type EntityRepository interface {
    Create(ctx context.Context, entity *Entity) error
    GetByID(ctx context.Context, id uuid.UUID) (*Entity, error)
    // ...
}
```

### 2. Service Layer

Encapsulates business logic, coordinates between repositories.

```go
func (s *EntityService) Create(ctx context.Context, req *CreateEntityRequest) (*Entity, error) {
    // 1. Fetch blueprint schema
    blueprint := s.blueprintService.GetByID(req.BlueprintID)

    // 2. Validate against schema
    if err := s.validator.Validate(req.Data, blueprint.Schema); err != nil {
        return nil, err
    }

    // 3. Save to database
    return s.repo.Create(ctx, entity)
}
```

### 3. Dependency Injection

Constructor-based injection for testability.

```go
func NewEntityService(repo EntityRepository,
                      blueprintService *blueprint.Service,
                      validator *validation.Validator) *EntityService {
    return &EntityService{
        repo: repo,
        blueprintService: blueprintService,
        validator: validator,
    }
}
```

### 4. Middleware Chain

Composable request processing pipeline.

```go
router.Use(
    gin.Recovery(),
    gin.Logger(),
    middleware.ErrorHandler(),
    middleware.Authenticate(),
    middleware.RequireTeam(),
    middleware.RequirePermission("entity:write"),
)
```

### 5. Context Passing

Request-scoped data via Gin context.

```go
// Middleware sets values
c.Set("user_id", userID)
c.Set("team_id", teamID)
c.Set("permissions", permissions)

// Handler retrieves values
userID := c.GetString("user_id")
teamID := c.GetString("team_id")
```

### 6. JSONB for Flexibility

Dynamic schema storage without migrations.

```go
// Blueprint defines schema
type Blueprint struct {
    Schema map[string]interface{} `json:"schema"`
}

// Entity stores validated data
type Entity struct {
    Data map[string]interface{} `json:"data"`
}

// PostgreSQL stores as JSONB
data JSONB NOT NULL DEFAULT '{}'
```

### 7. Strategy Pattern (Search Filters)

Dynamic SQL generation based on operator.

```go
switch filter.Operator {
case "eq":
    query += fmt.Sprintf("data->>'%s' = $%d", property, paramCount)
case "gt":
    query += fmt.Sprintf("(data->>'%s')::numeric > $%d", property, paramCount)
case "contains":
    query += fmt.Sprintf("data->>'%s' ILIKE $%d", property, paramCount)
}
```

## Security Architecture

### Security Layers

```mermaid
graph TB
    A[Request] --> B{HTTPS?}
    B -->|No| C[Reject]
    B -->|Yes| D[Authentication]
    D --> E{Valid JWT/API Key?}
    E -->|No| F[401 Unauthorized]
    E -->|Yes| G[Team Isolation]
    G --> H{Team Member?}
    H -->|No| I[403 Forbidden]
    H -->|Yes| J[Permission Check]
    J --> K{Has Permission?}
    K -->|No| L[403 Forbidden]
    K -->|Yes| M[Input Validation]
    M --> N{Valid Schema?}
    N -->|No| O[400 Bad Request]
    N -->|Yes| P[SQL Injection Prevention]
    P --> Q[Parameterized Query]
    Q --> R[Execute]
```

### Security Features

1. **Authentication**:
   - bcrypt password hashing (cost 10)
   - JWT with HS256 signing
   - API keys with SHA-256 hashing
   - Token expiration enforcement

2. **Authorization**:
   - RBAC with 13 permissions
   - Team-based isolation
   - Permission middleware
   - Default roles (admin, editor, viewer)

3. **Input Validation**:
   - JSON Schema validation
   - Property name regex: `^[a-zA-Z0-9_.]+$`
   - SQL injection prevention via parameterized queries
   - Order by column whitelist

4. **Data Protection**:
   - Team-scoped queries (multi-tenancy)
   - Cascade deletes for data consistency
   - API key hashing (never stored plain text)
   - Password never returned in responses

## Performance Considerations

### Database Optimizations

1. **Connection Pooling**:
   - MaxOpenConns: 25
   - MaxIdleConns: 5
   - ConnMaxLifetime: 5 minutes
   - ConnMaxIdleTime: 1 minute

2. **Indexes**:
   - GIN index on `entities.data` for JSONB queries
   - Standard B-tree indexes on foreign keys
   - Hash index on `api_keys.key_hash`
   - Composite unique indexes for constraints

3. **Query Optimization**:
   - Pagination support (limit/offset)
   - Count queries before fetch
   - Efficient JSONB path navigation
   - Prepared statement reuse

### Async Operations

**API Key Last Used Tracking**:
```go
go s.repo.UpdateAPIKeyLastUsed(context.Background(), apiKey.ID)
```

Runs in separate goroutine to avoid blocking request.

## Future Architecture

### Planned Features (Tables Defined)

1. **Relations System**:
   - Blueprint-level relation definitions
   - Entity-level relation instances
   - Support for many-to-many, one-to-many

2. **Scorecards**:
   - Quality/compliance metrics
   - Rule-based evaluation
   - Level-based scoring

3. **Integrations**:
   - External system connectors
   - Data synchronization
   - Field mapping

4. **Actions**:
   - Workflow automation
   - Manual and automatic triggers
   - Multi-step execution

5. **Audit Logging**:
   - Change tracking
   - User attribution
   - Historical data

---

For API documentation, see [API.md](./API.md).
For database details, see [DATABASE.md](./DATABASE.md).
For deployment instructions, see [DEPLOYMENT.md](./DEPLOYMENT.md).
