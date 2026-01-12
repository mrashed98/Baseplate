# Baseplate - Port.ai Clone Backend Architecture

## Overview
A headless backend engine for dynamic entity management with blueprints, relations, integrations, scorecards, and workflow actions.

---

## Tech Stack Recommendation

**Go + Gin + PostgreSQL** (keep current stack)

**Rationale:**
- Go's goroutines handle concurrent webhook receivers and integration workers efficiently
- Single binary deployment, easy containerization
- Strong typing for complex schema-driven system
- PostgreSQL JSONB provides schema flexibility with query power

---

## Core Architecture

### 1. Blueprint System (Enhanced)
Current schema is good. Need to add:
- **Property types**: string, number, boolean, array, object, date, url, email
- **Calculated properties**: Derived from other properties via expressions
- **Mirror properties**: Inherited from related entities
- **Relation definitions**: Links between blueprints

### 2. Relations System
Connect entities across blueprints:
- **Types**: one-to-one, one-to-many, many-to-many
- **Direction**: Bidirectional with configurable titles
- **Cascade**: Define delete/update behavior

### 3. Scorecards
Compliance and quality metrics:
- Define rules per blueprint
- Levels (Gold/Silver/Bronze) with conditions
- Automatic evaluation on entity changes

### 4. Integrations Framework
Plugin-based architecture:
- **Webhook receivers**: Generic endpoint + provider-specific handlers
- **Data mappers**: Transform external data to blueprint properties
- **OAuth manager**: Handle credentials securely
- **Initial providers**: GitHub, Kubernetes, Jira, AWS

### 5. Actions & Workflows
Self-service automation:
- **Triggers**: Manual, entity events, scheduled
- **Executors**: Webhook, script, integration-specific
- **Workflow engine**: Multi-step orchestration

---

## Database Schema Expansion

Note: Original `blueprints` and `entities` tables need `team_id` added for multi-tenancy:

```sql
-- Update existing tables (add team_id)
ALTER TABLE blueprints ADD COLUMN team_id UUID NOT NULL REFERENCES teams(id);
ALTER TABLE entities ADD COLUMN team_id UUID NOT NULL REFERENCES teams(id);
CREATE INDEX idx_blueprints_team ON blueprints(team_id);
CREATE INDEX idx_entities_team ON entities(team_id);
```

```sql
-- Blueprint Relations (schema level)
CREATE TABLE blueprint_relations (
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    source_blueprint_id VARCHAR(50) REFERENCES blueprints(id) ON DELETE CASCADE,
    target_blueprint_id VARCHAR(50) REFERENCES blueprints(id) ON DELETE CASCADE,
    relation_type VARCHAR(20) NOT NULL, -- 'one-to-one', 'one-to-many', 'many-to-many'
    source_title VARCHAR(100),
    target_title VARCHAR(100),
    required BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Entity Relations (instance level)
CREATE TABLE entity_relations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    relation_id UUID NOT NULL REFERENCES blueprint_relations(id) ON DELETE CASCADE,
    source_entity_id UUID NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    target_entity_id UUID NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(relation_id, source_entity_id, target_entity_id)
);

-- Scorecards
CREATE TABLE scorecards (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    blueprint_id VARCHAR(50) NOT NULL REFERENCES blueprints(id) ON DELETE CASCADE,
    title VARCHAR(100) NOT NULL,
    levels JSONB NOT NULL, -- [{name, color, order}]
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE scorecard_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    scorecard_id UUID NOT NULL REFERENCES scorecards(id) ON DELETE CASCADE,
    level_name VARCHAR(50) NOT NULL,
    property_path VARCHAR(200) NOT NULL,
    operator VARCHAR(20) NOT NULL, -- 'eq', 'neq', 'gt', 'lt', 'contains', 'exists'
    value JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Integrations
CREATE TABLE integrations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL, -- 'github', 'kubernetes', 'jira', 'aws'
    name VARCHAR(100) NOT NULL,
    config JSONB NOT NULL DEFAULT '{}', -- encrypted credentials, settings
    status VARCHAR(20) DEFAULT 'inactive',
    last_sync_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE integration_mappings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    integration_id UUID NOT NULL REFERENCES integrations(id) ON DELETE CASCADE,
    blueprint_id VARCHAR(50) NOT NULL REFERENCES blueprints(id) ON DELETE CASCADE,
    external_type VARCHAR(100) NOT NULL, -- 'repository', 'deployment', 'issue'
    mapping JSONB NOT NULL, -- {property: "source.path"}
    filter JSONB, -- conditions for which external entities to sync
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Actions
CREATE TABLE actions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    blueprint_id VARCHAR(50) REFERENCES blueprints(id) ON DELETE CASCADE,
    title VARCHAR(100) NOT NULL,
    description TEXT,
    trigger_type VARCHAR(20) NOT NULL, -- 'manual', 'entity_created', 'entity_updated', 'scheduled'
    trigger_config JSONB DEFAULT '{}',
    user_inputs JSONB DEFAULT '[]', -- form fields for manual actions
    steps JSONB NOT NULL, -- [{type, config}]
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- RBAC: Users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(100),
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- RBAC: Teams/Organizations
CREATE TABLE teams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(50) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- RBAC: Roles
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL, -- 'admin', 'editor', 'viewer'
    permissions JSONB NOT NULL DEFAULT '[]', -- ['blueprint:write', 'entity:read']
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(team_id, name)
);

-- RBAC: Team Memberships
CREATE TABLE team_memberships (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(team_id, user_id)
);

-- RBAC: API Keys
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    name VARCHAR(100) NOT NULL,
    key_hash VARCHAR(255) NOT NULL,
    permissions JSONB NOT NULL DEFAULT '[]',
    expires_at TIMESTAMP WITH TIME ZONE,
    last_used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Audit Log
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    team_id UUID REFERENCES teams(id) ON DELETE SET NULL,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    entity_type VARCHAR(50) NOT NULL, -- 'entity', 'blueprint', 'integration'
    entity_id UUID,
    action VARCHAR(20) NOT NULL, -- 'create', 'update', 'delete'
    old_data JSONB,
    new_data JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_entity_relations_source ON entity_relations(source_entity_id);
CREATE INDEX idx_entity_relations_target ON entity_relations(target_entity_id);
CREATE INDEX idx_scorecards_blueprint ON scorecards(blueprint_id);
CREATE INDEX idx_integrations_type ON integrations(type);
CREATE INDEX idx_audit_logs_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_logs_created ON audit_logs(created_at DESC);
```

---

## Project Structure

```
baseplate/
├── cmd/
│   └── server/
│       └── main.go              # Entry point
├── internal/
│   ├── api/
│   │   ├── handlers/
│   │   │   ├── auth.go          # Login, register, API keys
│   │   │   ├── team.go          # Team management
│   │   │   ├── blueprint.go     # Blueprint CRUD
│   │   │   ├── entity.go        # Entity CRUD + search
│   │   │   ├── relation.go      # Relation management
│   │   │   ├── scorecard.go     # Scorecard CRUD + evaluation
│   │   │   ├── integration.go   # Integration management
│   │   │   ├── action.go        # Action CRUD + execution
│   │   │   └── webhook.go       # Webhook receiver
│   │   ├── middleware/
│   │   │   ├── auth.go          # JWT + API key validation
│   │   │   ├── rbac.go          # Permission checking
│   │   │   ├── team.go          # Team context
│   │   │   ├── logging.go
│   │   │   └── error.go
│   │   └── router.go
│   ├── core/
│   │   ├── blueprint/
│   │   │   ├── service.go
│   │   │   └── repository.go
│   │   ├── entity/
│   │   │   ├── service.go
│   │   │   ├── repository.go
│   │   │   └── search.go        # Query builder
│   │   ├── relation/
│   │   │   ├── service.go
│   │   │   └── repository.go
│   │   ├── scorecard/
│   │   │   ├── service.go
│   │   │   ├── repository.go
│   │   │   └── evaluator.go     # Rule evaluation engine
│   │   └── validation/
│   │       └── jsonschema.go    # Schema validation
│   ├── integration/
│   │   ├── engine/
│   │   │   ├── manager.go       # Integration lifecycle
│   │   │   ├── mapper.go        # Data transformation
│   │   │   └── sync.go          # Sync orchestration
│   │   ├── providers/
│   │   │   ├── provider.go      # Interface
│   │   │   ├── github/
│   │   │   ├── kubernetes/
│   │   │   ├── jira/
│   │   │   └── aws/
│   │   └── webhook/
│   │       ├── handler.go       # Generic webhook handler
│   │       └── parser.go        # Payload parsing
│   ├── action/
│   │   ├── engine/
│   │   │   ├── executor.go      # Action execution
│   │   │   └── workflow.go      # Multi-step orchestration
│   │   └── executors/
│   │       ├── webhook.go       # HTTP call executor
│   │       └── script.go        # Script executor
│   └── storage/
│       ├── postgres/
│       │   └── client.go
│       └── cache/
│           └── redis.go         # Optional caching
├── pkg/
│   ├── jsonschema/
│   │   └── validator.go
│   └── query/
│       └── builder.go           # JSONB query builder
├── migrations/
│   └── 001_initial.sql
├── config/
│   └── config.go
├── docker-compose.yaml
└── Makefile
```

---

## Implementation Phases

### Phase 1: Core Foundation + Auth
1. Set up Go project structure with Gin
2. Database migrations setup
3. RBAC system: Users, Teams, Roles, Permissions
4. JWT authentication + API key auth
5. Permission middleware
6. Implement Blueprint service (CRUD + validation)
7. Implement Entity service (CRUD + search with JSONB queries)
8. JSON Schema validation engine
9. Basic API endpoints

### Phase 2: Relations
1. Blueprint relations schema and service
2. Entity relations CRUD
3. Graph traversal queries
4. Update entity search to include relations

### Phase 3: Scorecards
1. Scorecard CRUD
2. Rule evaluation engine
3. Auto-evaluation on entity changes
4. Scorecard results API

### Phase 4: Integrations Framework
1. Integration manager and provider interface
2. Generic webhook receiver
3. Data mapper engine
4. First provider: GitHub (webhooks for repos, PRs)

### Phase 5: Additional Integrations
1. Kubernetes provider (via API)
2. Jira provider (webhooks)
3. AWS provider (CloudWatch events)

### Phase 6: Actions & Workflows
1. Action definitions CRUD
2. Webhook executor
3. Event triggers (entity created/updated)
4. Workflow orchestration

### Phase 7: Polish
1. Audit logging
2. Caching layer (Redis)
3. Rate limiting
4. API documentation (OpenAPI)

---

## API Endpoints

### Auth
- `POST   /api/auth/register` - Register user
- `POST   /api/auth/login` - Login (returns JWT)
- `POST   /api/auth/refresh` - Refresh token
- `GET    /api/auth/me` - Get current user

### Teams
- `POST   /api/teams` - Create team
- `GET    /api/teams` - List user's teams
- `GET    /api/teams/:id` - Get team
- `PUT    /api/teams/:id` - Update team
- `POST   /api/teams/:id/members` - Invite member
- `DELETE /api/teams/:id/members/:userId` - Remove member

### Roles
- `POST   /api/teams/:id/roles` - Create role
- `GET    /api/teams/:id/roles` - List roles
- `PUT    /api/roles/:id` - Update role

### API Keys
- `POST   /api/teams/:id/api-keys` - Create API key
- `GET    /api/teams/:id/api-keys` - List API keys
- `DELETE /api/api-keys/:id` - Revoke API key

### Blueprints
- `POST   /api/blueprints` - Create blueprint
- `GET    /api/blueprints` - List blueprints
- `GET    /api/blueprints/:id` - Get blueprint
- `PUT    /api/blueprints/:id` - Update blueprint
- `DELETE /api/blueprints/:id` - Delete blueprint

### Entities
- `POST   /api/blueprints/:blueprintId/entities` - Create entity
- `GET    /api/blueprints/:blueprintId/entities` - List/search entities
- `GET    /api/entities/:id` - Get entity
- `PUT    /api/entities/:id` - Update entity
- `DELETE /api/entities/:id` - Delete entity

### Relations
- `POST   /api/blueprint-relations` - Define relation between blueprints
- `GET    /api/blueprint-relations` - List relations
- `POST   /api/entity-relations` - Create entity relation
- `DELETE /api/entity-relations/:id` - Remove entity relation
- `GET    /api/entities/:id/relations` - Get entity's relations

### Scorecards
- `POST   /api/scorecards` - Create scorecard
- `GET    /api/blueprints/:id/scorecards` - List scorecards for blueprint
- `GET    /api/entities/:id/scores` - Get entity scores

### Integrations
- `POST   /api/integrations` - Create integration
- `GET    /api/integrations` - List integrations
- `POST   /api/integrations/:id/mappings` - Create mapping
- `POST   /api/webhooks/:integrationId` - Receive webhook

### Actions
- `POST   /api/actions` - Create action
- `GET    /api/blueprints/:id/actions` - List actions
- `POST   /api/actions/:id/execute` - Execute action manually

---

## Verification Plan

1. **Unit tests**: Each service layer
2. **Integration tests**: API endpoints with test database
3. **Manual testing**:
   - Create blueprints via API
   - Create entities with validation
   - Set up relations between entities
   - Configure scorecards and verify evaluation
   - Connect GitHub integration and receive webhooks
   - Execute actions

---

## RBAC Permission Model

```
Permissions (stored as JSON array in roles):
- team:manage           # Manage team settings, members
- blueprint:read        # View blueprints
- blueprint:write       # Create/update blueprints
- blueprint:delete      # Delete blueprints
- entity:read           # View entities
- entity:write          # Create/update entities
- entity:delete         # Delete entities
- integration:read      # View integrations
- integration:write     # Configure integrations
- scorecard:read        # View scorecards
- scorecard:write       # Configure scorecards
- action:read           # View actions
- action:write          # Configure actions
- action:execute        # Execute actions

Default Roles:
- Admin: All permissions
- Editor: All except team:manage and *:delete
- Viewer: All *:read permissions
```

---

## Key Design Decisions

1. **JSONB everywhere**: Flexible schema storage with PostgreSQL query power
2. **Plugin-based integrations**: Each provider implements a common interface
3. **Event-driven**: Webhooks as primary sync, polling as fallback
4. **Separation of concerns**: Blueprint-level configs vs Entity-level instances
5. **Audit trail**: All changes logged for compliance

---

## Files to Create/Modify

- `schema.sql` - Expand with new tables
- `cmd/server/main.go` - New
- `internal/api/router.go` - New
- `internal/api/handlers/*.go` - New (7 files)
- `internal/core/**/service.go` - New (5 services)
- `internal/core/**/repository.go` - New (5 repos)
- `internal/integration/**` - New (provider framework)
- `internal/action/**` - New (action engine)
- `config/config.go` - New
- `docker-compose.yaml` - Update (add Redis optional)
