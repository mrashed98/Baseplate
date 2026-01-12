# Baseplate Database Documentation

Complete database schema documentation, optimization strategies, and operational guide for Baseplate's PostgreSQL database.

## Table of Contents

- [Overview](#overview)
- [Database Schema](#database-schema)
- [Table Descriptions](#table-descriptions)
- [Indexes and Performance](#indexes-and-performance)
- [JSONB Usage](#jsonb-usage)
- [Relationships and Constraints](#relationships-and-constraints)
- [Migrations](#migrations)
- [Query Patterns](#query-patterns)
- [Backup and Recovery](#backup-and-recovery)
- [Performance Tuning](#performance-tuning)

## Overview

**Database**: PostgreSQL 15+
**Key Features**: UUID support, JSONB for flexible schemas, GIN indexes for JSONB queries
**Character Set**: UTF-8
**Timezone**: All timestamps with time zone

### Extensions Required

```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
```

Provides `uuid_generate_v4()` function for UUID primary keys.

## Database Schema

### Core Tables Summary

| Table | Purpose | Size Estimate | Growth Rate |
|-------|---------|---------------|-------------|
| `users` | User accounts | Low | Slow |
| `teams` | Organizations/tenants | Low | Slow |
| `roles` | RBAC role definitions | Low | Slow |
| `team_memberships` | User-team associations | Medium | Medium |
| `api_keys` | API authentication | Low | Slow |
| `blueprints` | Schema definitions | Low | Medium |
| `entities` | Entity instances | **High** | **Fast** |
| `blueprint_relations` | Schema-level relations | Low | Slow |
| `entity_relations` | Instance-level relations | High | Fast |
| `scorecards` | Quality metrics | Low | Slow |
| `scorecard_rules` | Scorecard rules | Low | Slow |
| `integrations` | External connectors | Low | Slow |
| `integration_mappings` | Integration configs | Low | Slow |
| `actions` | Workflow definitions | Low | Slow |
| `audit_logs` | Change history | **High** | **Fast** |

## Table Descriptions

### RBAC Tables

#### `users`

User accounts with authentication credentials.

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(100),
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

**Columns**:
- `id`: Unique identifier
- `email`: Unique email address (login username)
- `password_hash`: bcrypt hash of password
- `name`: Display name
- `status`: `active` | `inactive` | `suspended`
- `created_at`: Registration timestamp

**Constraints**:
- `email` must be unique
- `password_hash` never returned in API responses

**Growth**: Slow (per user registration)

---

#### `teams`

Organizations/tenants for multi-tenancy.

```sql
CREATE TABLE teams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(50) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

**Columns**:
- `id`: Unique identifier
- `name`: Display name
- `slug`: URL-friendly identifier (unique, lowercase)
- `created_at`: Creation timestamp

**Constraints**:
- `slug` must be unique across all teams
- On creation, three default roles are created

**Cascade**: Deleting a team cascades to all team resources

**Growth**: Slow (per team creation)

---

#### `roles`

Team-scoped roles with permission arrays.

```sql
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    permissions JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(team_id, name)
);
```

**Columns**:
- `id`: Unique identifier
- `team_id`: Team reference
- `name`: Role name (unique within team)
- `permissions`: JSONB array of permission strings
- `created_at`: Creation timestamp

**Default Roles**:
- `admin`: All permissions
- `editor`: Read/write without team management
- `viewer`: Read-only access

**Permissions Format**:
```json
["team:manage", "blueprint:read", "blueprint:write", "entity:read", "entity:write"]
```

**Growth**: Slow (few custom roles per team)

---

#### `team_memberships`

Associates users with teams and roles.

```sql
CREATE TABLE team_memberships (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(team_id, user_id)
);
```

**Columns**:
- `id`: Unique identifier
- `team_id`: Team reference
- `user_id`: User reference
- `role_id`: Role reference
- `created_at`: Join timestamp

**Constraints**:
- User can only have one role per team
- Unique constraint on `(team_id, user_id)`

**Indexes**:
- `idx_team_memberships_user` on `user_id`
- `idx_team_memberships_team` on `team_id`

**Growth**: Medium (proportional to team size)

---

#### `api_keys`

API keys for service authentication.

```sql
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
```

**Columns**:
- `id`: Unique identifier
- `team_id`: Team owner
- `user_id`: Creator (nullable, SET NULL on user delete)
- `name`: Descriptive name
- `key_hash`: SHA-256 hash of API key
- `permissions`: JSONB array of permissions
- `expires_at`: Optional expiration timestamp
- `last_used_at`: Last usage timestamp (async updated)
- `created_at`: Creation timestamp

**Security**:
- Raw key never stored (only SHA-256 hash)
- Key format: `bp_<64_hex_characters>`
- Last used tracking via async goroutine

**Indexes**:
- `idx_api_keys_team` on `team_id`
- `idx_api_keys_hash` on `key_hash` (critical for auth performance)

**Growth**: Slow (few keys per team)

---

### Core Domain Tables

#### `blueprints`

Schema definitions for entity types.

```sql
CREATE TABLE blueprints (
    id VARCHAR(50) PRIMARY KEY,
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    title VARCHAR(100) NOT NULL,
    description TEXT,
    icon VARCHAR(50),
    schema JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

**Columns**:
- `id`: String identifier (e.g., "service", "database", "cluster")
- `team_id`: Team owner
- `title`: Display name
- `description`: Optional description
- `icon`: Emoji or icon identifier
- `schema`: JSON Schema definition
- `created_at`, `updated_at`: Timestamps

**Schema Format**:
Standard JSON Schema object:
```json
{
  "type": "object",
  "properties": {
    "name": {"type": "string"},
    "version": {"type": "string"}
  },
  "required": ["name"]
}
```

**Indexes**:
- `idx_blueprints_team` on `team_id`

**Growth**: Low to medium (typically 10-50 blueprints per team)

---

#### `entities`

Instances of blueprints with validated data.

```sql
CREATE TABLE entities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    blueprint_id VARCHAR(50) NOT NULL REFERENCES blueprints(id) ON DELETE CASCADE,
    identifier VARCHAR(255) NOT NULL,
    title VARCHAR(255),
    data JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(team_id, blueprint_id, identifier)
);
```

**Columns**:
- `id`: UUID primary key
- `team_id`: Team owner (for isolation)
- `blueprint_id`: Blueprint reference
- `identifier`: Human-readable ID (unique within blueprint)
- `title`: Display title
- `data`: JSONB validated against blueprint schema
- `created_at`, `updated_at`: Timestamps

**Constraints**:
- Unique `(team_id, blueprint_id, identifier)`
- Data validated against blueprint schema before INSERT/UPDATE

**Indexes**:
- `idx_entities_team` on `team_id`
- `idx_entities_blueprint` on `blueprint_id`
- **`idx_entities_data` GIN index on `data`** (critical for search performance)

**Growth**: **High** - primary data storage table

**Example Data**:
```json
{
  "name": "auth-service",
  "version": "2.3.1",
  "language": "Go",
  "status": "active",
  "dependencies": ["postgres", "redis"]
}
```

---

### Relationship Tables

#### `blueprint_relations`

Defines relations between blueprint types (schema level).

```sql
CREATE TABLE blueprint_relations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    source_blueprint_id VARCHAR(50) NOT NULL REFERENCES blueprints(id) ON DELETE CASCADE,
    target_blueprint_id VARCHAR(50) NOT NULL REFERENCES blueprints(id) ON DELETE CASCADE,
    identifier VARCHAR(100) NOT NULL,
    title VARCHAR(100),
    relation_type VARCHAR(20) NOT NULL DEFAULT 'many-to-many',
    required BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(team_id, source_blueprint_id, identifier)
);
```

**Relation Types**:
- `one-to-one`
- `one-to-many`
- `many-to-one`
- `many-to-many`

**Status**: Table defined but not implemented in API

---

#### `entity_relations`

Actual relations between entity instances.

```sql
CREATE TABLE entity_relations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    relation_id UUID NOT NULL REFERENCES blueprint_relations(id) ON DELETE CASCADE,
    source_entity_id UUID NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    target_entity_id UUID NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(relation_id, source_entity_id, target_entity_id)
);
```

**Indexes**:
- `idx_entity_relations_source` on `source_entity_id`
- `idx_entity_relations_target` on `target_entity_id`

**Status**: Table defined but not implemented in API

---

### Future Feature Tables

#### `scorecards`, `scorecard_rules`

Quality/compliance metrics (planned feature).

#### `integrations`, `integration_mappings`

External system connectors (planned feature).

#### `actions`

Workflow automation (planned feature).

#### `audit_logs`

Change history tracking (planned feature).

---

## Indexes and Performance

### Primary Indexes

| Index | Table | Type | Purpose |
|-------|-------|------|---------|
| Primary Key | All tables | B-tree | Unique row identification |
| `email` | users | B-tree | Unique constraint + fast lookup |
| `slug` | teams | B-tree | Unique constraint + URL routing |
| `(team_id, name)` | roles | B-tree | Unique role names per team |
| `(team_id, user_id)` | team_memberships | B-tree | One role per user per team |
| `(team_id, blueprint_id, identifier)` | entities | B-tree | Unique entity identifiers |

### Performance Indexes

#### Critical for Performance

**`idx_entities_data` (GIN index on JSONB)**:
```sql
CREATE INDEX idx_entities_data ON entities USING GIN (data);
```

**Why**: Enables fast JSONB queries with operators:
- `@>` (contains)
- `?` (key exists)
- `->` (path navigation)
- `->` (text extraction)

**Query Example**:
```sql
SELECT * FROM entities
WHERE data @> '{"status": "active"}'
  AND data->'version' > '"2.0.0"';
```

**Impact**: 100-1000x faster than sequential scan for JSONB queries

---

**`idx_api_keys_hash` (B-tree on key_hash)**:
```sql
CREATE INDEX idx_api_keys_hash ON api_keys(key_hash);
```

**Why**: Fast API key authentication lookup

**Query**: `SELECT * FROM api_keys WHERE key_hash = $1`

**Impact**: Sub-millisecond authentication

---

### Relationship Indexes

For fast foreign key lookups and JOIN operations:

```sql
CREATE INDEX idx_team_memberships_user ON team_memberships(user_id);
CREATE INDEX idx_team_memberships_team ON team_memberships(team_id);
CREATE INDEX idx_entities_team ON entities(team_id);
CREATE INDEX idx_entities_blueprint ON entities(blueprint_id);
CREATE INDEX idx_entity_relations_source ON entity_relations(source_entity_id);
CREATE INDEX idx_entity_relations_target ON entity_relations(target_entity_id);
```

---

## JSONB Usage

### Why JSONB?

Baseplate uses JSONB extensively for flexible schema storage:

**Advantages**:
- Schema changes without migrations
- Rich querying with GIN indexes
- Validation via JSON Schema
- Compressed binary storage
- Fast key lookups

**Tables Using JSONB**:
- `blueprints.schema` - JSON Schema definitions
- `entities.data` - Entity data validated against schema
- `roles.permissions` - Permission arrays
- `api_keys.permissions` - Permission arrays
- `scorecards.levels` - Scorecard level definitions
- `scorecard_rules.value` - Rule values
- `integrations.config` - Integration configurations
- `integration_mappings.mapping`, `filter` - Field mappings
- `actions.trigger_config`, `user_inputs`, `steps` - Action definitions
- `audit_logs.old_data`, `new_data` - Change tracking

### JSONB Query Patterns

#### 1. Exact Match

```sql
SELECT * FROM entities
WHERE data @> '{"status": "active"}';
```

Uses GIN index for fast lookup.

---

#### 2. Key Existence

```sql
SELECT * FROM entities
WHERE data ? 'version';
```

Checks if `version` key exists.

---

#### 3. Nested Path

```sql
SELECT * FROM entities
WHERE data->'metadata'->>'environment' = 'production';
```

Navigate nested JSONB structure.

---

#### 4. Numeric Comparison

```sql
SELECT * FROM entities
WHERE (data->>'replica_count')::int > 3;
```

Cast to numeric for comparison.

---

#### 5. Array Contains

```sql
SELECT * FROM entities
WHERE data->'dependencies' ? 'postgres';
```

Check if array contains value.

---

#### 6. Text Search

```sql
SELECT * FROM entities
WHERE data->>'name' ILIKE '%auth%';
```

Case-insensitive text search.

---

### JSONB Best Practices

1. **Always use GIN index** on JSONB columns for search
2. **Validate before INSERT** using JSON Schema
3. **Use `@>` operator** for containment checks (uses index)
4. **Avoid deep nesting** (max 3-4 levels for performance)
5. **Cast appropriately** for numeric/boolean comparisons
6. **Use `->` for navigation**, `->>` for text extraction

---

## Relationships and Constraints

### Foreign Key Relationships

```
users
  ├─→ team_memberships.user_id (CASCADE)
  └─→ api_keys.user_id (SET NULL)

teams
  ├─→ roles.team_id (CASCADE)
  ├─→ team_memberships.team_id (CASCADE)
  ├─→ api_keys.team_id (CASCADE)
  ├─→ blueprints.team_id (CASCADE)
  ├─→ entities.team_id (CASCADE)
  ├─→ blueprint_relations.team_id (CASCADE)
  ├─→ scorecards.team_id (CASCADE)
  ├─→ integrations.team_id (CASCADE)
  └─→ actions.team_id (CASCADE)

blueprints
  ├─→ entities.blueprint_id (CASCADE)
  ├─→ blueprint_relations.source/target_blueprint_id (CASCADE)
  ├─→ scorecards.blueprint_id (CASCADE)
  └─→ actions.blueprint_id (CASCADE)

roles
  └─→ team_memberships.role_id (CASCADE)

entities
  └─→ entity_relations.source/target_entity_id (CASCADE)

blueprint_relations
  └─→ entity_relations.relation_id (CASCADE)

scorecards
  └─→ scorecard_rules.scorecard_id (CASCADE)

integrations
  └─→ integration_mappings.integration_id (CASCADE)
```

### Cascade Behavior

**DELETE team**:
- Cascades to ALL team resources (blueprints, entities, roles, memberships, API keys)
- Use with extreme caution

**DELETE user**:
- Cascades to team_memberships
- Sets api_keys.user_id to NULL (preserves keys)

**DELETE blueprint**:
- Cascades to ALL entities of that blueprint
- Cascades to relations, scorecards, actions

**DELETE entity**:
- Cascades to entity_relations

---

## Migrations

### Migration System

**Location**: `/migrations/001_initial.sql`

**Execution**: Auto-runs via Docker init scripts on first container startup

**Manual Execution**:
```bash
docker exec -i baseplate_db psql -U user -d baseplate < migrations/001_initial.sql
```

### Migration Best Practices

1. **Always backup** before running migrations
2. **Test on staging** environment first
3. **Use transactions** for rollback capability
4. **Document changes** in migration comments
5. **Version migrations** with timestamps or sequence numbers

### Adding New Migrations

Create new file: `migrations/002_description.sql`

```sql
-- Migration: Add new column
-- Date: 2024-01-15
-- Description: Add 'archived' status for entities

BEGIN;

ALTER TABLE entities
ADD COLUMN archived BOOLEAN DEFAULT FALSE;

CREATE INDEX idx_entities_archived ON entities(archived);

COMMIT;
```

---

## Query Patterns

### Common Queries

#### Get User Teams with Roles

```sql
SELECT t.*, r.name as role_name
FROM teams t
JOIN team_memberships tm ON t.id = tm.team_id
JOIN roles r ON tm.role_id = r.id
WHERE tm.user_id = $1;
```

---

#### Search Entities by Blueprint with Filters

```sql
SELECT * FROM entities
WHERE team_id = $1
  AND blueprint_id = $2
  AND data @> '{"status": "active"}'
  AND (data->>'version') > '2.0.0'
ORDER BY created_at DESC
LIMIT 50 OFFSET 0;
```

---

#### Count Entities by Status

```sql
SELECT
  data->>'status' as status,
  COUNT(*) as count
FROM entities
WHERE team_id = $1
  AND blueprint_id = $2
GROUP BY data->>'status';
```

---

#### Find Related Entities

```sql
SELECT e.*
FROM entities e
JOIN entity_relations er ON e.id = er.target_entity_id
WHERE er.source_entity_id = $1;
```

---

#### Audit Trail Query

```sql
SELECT
  al.*,
  u.name as user_name,
  u.email as user_email
FROM audit_logs al
LEFT JOIN users u ON al.user_id = u.id
WHERE al.team_id = $1
  AND al.entity_type = 'entity'
ORDER BY al.created_at DESC
LIMIT 100;
```

---

## Backup and Recovery

### Backup Strategies

#### 1. Docker Volume Backup

```bash
# Stop container
docker-compose down

# Backup volume
docker run --rm -v baseplate_postgres_data:/data -v $(pwd):/backup \
  alpine tar czf /backup/postgres_backup_$(date +%Y%m%d).tar.gz /data

# Restart container
docker-compose up -d
```

---

#### 2. pg_dump Backup

```bash
# Full database backup
docker exec baseplate_db pg_dump -U user baseplate > backup_$(date +%Y%m%d).sql

# Compressed backup
docker exec baseplate_db pg_dump -U user baseplate | gzip > backup_$(date +%Y%m%d).sql.gz

# Schema only
docker exec baseplate_db pg_dump -U user --schema-only baseplate > schema_$(date +%Y%m%d).sql

# Data only
docker exec baseplate_db pg_dump -U user --data-only baseplate > data_$(date +%Y%m%d).sql
```

---

#### 3. Continuous Backup (WAL Archiving)

For production, configure PostgreSQL WAL archiving:

```postgresql.conf
wal_level = replica
archive_mode = on
archive_command = 'cp %p /archive/%f'
```

---

### Recovery

#### Restore from pg_dump

```bash
# Drop and recreate database
docker exec -i baseplate_db psql -U user -c "DROP DATABASE baseplate;"
docker exec -i baseplate_db psql -U user -c "CREATE DATABASE baseplate;"

# Restore from backup
docker exec -i baseplate_db psql -U user baseplate < backup_20240115.sql
```

---

#### Restore from Volume

```bash
# Stop container
docker-compose down

# Remove old volume
docker volume rm baseplate_postgres_data

# Restore volume
docker run --rm -v baseplate_postgres_data:/data -v $(pwd):/backup \
  alpine tar xzf /backup/postgres_backup_20240115.tar.gz -C /

# Restart container
docker-compose up -d
```

---

## Performance Tuning

### PostgreSQL Configuration

**For Development** (docker-compose.yaml):
```yaml
POSTGRES_SHARED_BUFFERS: 256MB
POSTGRES_WORK_MEM: 4MB
POSTGRES_MAINTENANCE_WORK_MEM: 64MB
```

**For Production** (adjust based on available RAM):
```postgresql.conf
shared_buffers = 25% of RAM
work_mem = RAM / max_connections / 4
maintenance_work_mem = RAM / 16
effective_cache_size = 75% of RAM

# For JSONB heavy workload
random_page_cost = 1.1  # SSD optimization
```

---

### Connection Pooling

**Application Settings** (internal/storage/postgres/client.go):
```go
db.SetMaxOpenConns(25)        // Max connections
db.SetMaxIdleConns(5)         // Idle connections
db.SetConnMaxLifetime(5m)     // Connection lifetime
db.SetConnMaxIdleTime(1m)     // Idle connection timeout
```

**Tuning**:
- Increase `MaxOpenConns` for high concurrency
- Monitor connection usage with `SELECT * FROM pg_stat_activity;`
- Adjust based on `max_connections` in PostgreSQL

---

### Query Optimization

#### Use EXPLAIN ANALYZE

```sql
EXPLAIN ANALYZE
SELECT * FROM entities
WHERE team_id = 'uuid-here'
  AND data @> '{"status": "active"}';
```

Look for:
- Index usage (should show "Bitmap Index Scan on idx_entities_data")
- Sequential scans (bad for large tables)
- High execution time

---

#### Vacuum and Analyze

```bash
# Analyze tables for query planner
docker exec baseplate_db psql -U user -d baseplate -c "ANALYZE;"

# Vacuum to reclaim storage
docker exec baseplate_db psql -U user -d baseplate -c "VACUUM;"

# Full vacuum (requires exclusive lock)
docker exec baseplate_db psql -U user -d baseplate -c "VACUUM FULL;"
```

**Schedule**: Run ANALYZE daily, VACUUM weekly

---

#### Monitor Table Bloat

```sql
SELECT
  schemaname,
  tablename,
  pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

---

### Monitoring Queries

#### Active Queries

```sql
SELECT
  pid,
  now() - pg_stat_activity.query_start AS duration,
  query,
  state
FROM pg_stat_activity
WHERE state != 'idle'
ORDER BY duration DESC;
```

---

#### Slow Queries

Enable slow query logging:

```postgresql.conf
log_min_duration_statement = 1000  # Log queries > 1 second
```

---

#### Index Usage

```sql
SELECT
  schemaname,
  tablename,
  indexname,
  idx_scan as index_scans
FROM pg_stat_user_indexes
ORDER BY idx_scan DESC;
```

Low `idx_scan` means index is unused (consider dropping).

---

#### Cache Hit Ratio

```sql
SELECT
  sum(heap_blks_read) as heap_read,
  sum(heap_blks_hit)  as heap_hit,
  sum(heap_blks_hit) / (sum(heap_blks_hit) + sum(heap_blks_read)) as ratio
FROM pg_statio_user_tables;
```

Target: > 0.99 (99% cache hit rate)

---

## Maintenance Checklist

### Daily
- [ ] Monitor connection count
- [ ] Check slow query log
- [ ] Review error logs

### Weekly
- [ ] Run VACUUM
- [ ] Run ANALYZE
- [ ] Check table sizes
- [ ] Review index usage

### Monthly
- [ ] Full database backup
- [ ] Test backup restoration
- [ ] Review and archive audit logs
- [ ] Check for unused indexes
- [ ] Analyze query performance trends

---

For API documentation, see [API.md](./API.md).
For architecture details, see [ARCHITECTURE.md](./ARCHITECTURE.md).
For security information, see [SECURITY.md](./SECURITY.md).
