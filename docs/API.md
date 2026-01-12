# Baseplate API Documentation

Complete REST API reference for Baseplate - a headless backend engine with dynamic schema management.

## Table of Contents

- [Overview](#overview)
- [Authentication](#authentication)
- [Team Context](#team-context)
- [Permissions](#permissions)
- [Error Handling](#error-handling)
- [Super Admin](#super-admin)
- [Endpoints](#endpoints)
  - [Health Check](#health-check)
  - [Authentication](#authentication-endpoints)
  - [Teams](#team-management)
  - [Roles](#role-management)
  - [Members](#member-management)
  - [API Keys](#api-key-management)
  - [Blueprints](#blueprint-management)
  - [Entities](#entity-management)
  - [Admin - Super Admin Only](#admin-super-admin-only)
- [Examples](#examples)

## Overview

**Base URL**: `http://localhost:8080/api`

All API endpoints return JSON responses and follow REST conventions. The API supports two authentication methods (JWT tokens and API keys) and implements comprehensive RBAC for access control.

## Authentication

Baseplate supports two authentication mechanisms:

### JWT Bearer Token

Used for user authentication. Obtained via `/api/auth/login` or `/api/auth/register`.

```http
Authorization: Bearer <jwt_token>
```

**Token Properties**:
- Algorithm: HS256 (HMAC-SHA256)
- Expiration: 24 hours (configurable)
- Contains: user_id, email, issued_at, expires_at

### API Key

Used for service-to-service authentication. Created via `/api/teams/:teamId/api-keys`.

```http
Authorization: ApiKey <api_key>
```

**API Key Properties**:
- Format: `bp_<64_hex_characters>`
- Storage: SHA-256 hash in database
- Team-scoped with optional permissions
- Optional expiration date
- Tracks last usage timestamp

## Team Context

Most endpoints require a team context. It can be provided via:

1. **URL Parameter**: `/teams/:teamId/...`
2. **Header**: `X-Team-ID: <team_uuid>`
3. **API Key**: Automatically set from key's team association

Team context is validated to ensure the authenticated user has access to the specified team.

## Permissions

### Available Permissions

| Permission | Description |
|------------|-------------|
| `team:manage` | Manage team settings, members, roles, and API keys |
| `blueprint:read` | View blueprints |
| `blueprint:write` | Create and update blueprints |
| `blueprint:delete` | Delete blueprints |
| `entity:read` | View entities |
| `entity:write` | Create and update entities |
| `entity:delete` | Delete entities |
| `integration:read` | View integrations (future feature) |
| `integration:write` | Configure integrations (future feature) |
| `scorecard:read` | View scorecards (future feature) |
| `scorecard:write` | Configure scorecards (future feature) |
| `action:read` | View actions (future feature) |
| `action:write` | Configure actions (future feature) |
| `action:execute` | Execute actions (future feature) |

### Default Roles

#### Admin
All permissions including team management and delete operations.

```json
{
  "permissions": [
    "team:manage",
    "blueprint:read", "blueprint:write", "blueprint:delete",
    "entity:read", "entity:write", "entity:delete"
  ]
}
```

#### Editor
Read and write permissions without delete operations.

```json
{
  "permissions": [
    "blueprint:read", "blueprint:write",
    "entity:read", "entity:write"
  ]
}
```

#### Viewer
Read-only access to all resources.

```json
{
  "permissions": [
    "blueprint:read",
    "entity:read"
  ]
}
```

## Super Admin

### Overview

The Super Admin role provides platform-level administrative access. Super admins can:
- Manage all teams in the system
- Manage all users across the platform
- Promote and demote other users to/from super admin
- Access and modify resources in any team without membership
- View comprehensive audit logs of all super admin actions

### Authorization

Super admin status is verified through:
1. **JWT Claim**: `is_super_admin: true` in the JWT token
2. **Middleware**: `RequireSuperAdmin()` enforces authorization on sensitive endpoints
3. **Permission Bypass**: Super admins bypass all team-level permission checks

### Super Admin Endpoints

See [Admin - Super Admin Only](#admin-super-admin-only) section for complete endpoint reference.

**Key Endpoints**:
- `GET /api/admin/teams` - List all teams
- `GET /api/admin/users` - List all users
- `POST /api/admin/users/:userId/promote` - Promote to super admin
- `POST /api/admin/users/:userId/demote` - Demote from super admin
- `GET /api/admin/audit-logs` - Query super admin actions

### Error Cases

| Status | Error | Meaning |
|--------|-------|---------|
| 403 | Forbidden | User is not a super admin |
| 409 | Conflict | Cannot demote the last super admin |
| 400 | Bad Request | User already super admin or not a super admin |

## Error Handling

### Standard Error Response

```json
{
  "error": "Error message description"
}
```

### Validation Error Response

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

### HTTP Status Codes

| Code | Description |
|------|-------------|
| 200 | Success |
| 201 | Created |
| 204 | No Content (successful deletion) |
| 400 | Bad Request (validation errors, malformed input) |
| 401 | Unauthorized (missing/invalid authentication) |
| 403 | Forbidden (insufficient permissions) |
| 404 | Not Found |
| 409 | Conflict (duplicate resources) |
| 500 | Internal Server Error |

---

## Endpoints

### Health Check

#### GET /api/health

Check API health status.

**Authentication**: None required

**Response** `200 OK`

```json
{
  "status": "ok"
}
```

---

## Authentication Endpoints

### POST /api/auth/register

Register a new user account.

**Authentication**: None required

**Request Body**

```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "securePassword123"
}
```

**Validation Rules**:
- `name`: Required, non-empty string
- `email`: Required, valid email format
- `password`: Required, minimum 8 characters

**Response** `201 Created`

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "John Doe",
    "email": "john@example.com",
    "status": "active",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

**Errors**:
- `400` - Validation error (invalid email, password too short)
- `409` - User already exists
- `500` - Server error

---

### POST /api/auth/login

Authenticate and receive a JWT token.

**Authentication**: None required

**Request Body**

```json
{
  "email": "john@example.com",
  "password": "securePassword123"
}
```

**Response** `200 OK`

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "John Doe",
    "email": "john@example.com",
    "status": "active",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

**Errors**:
- `400` - Validation error
- `401` - Invalid credentials
- `500` - Server error

---

### GET /api/auth/me

Get authenticated user information.

**Authentication**: JWT Bearer token required

**Request Headers**

```http
Authorization: Bearer <token>
```

**Response** `200 OK`

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "John Doe",
  "email": "john@example.com",
  "status": "active",
  "created_at": "2024-01-15T10:30:00Z"
}
```

**Errors**:
- `401` - Unauthorized (missing/invalid token)
- `404` - User not found
- `500` - Server error

---

## Team Management

### POST /api/teams

Create a new team. The creator is automatically added as an admin.

**Authentication**: JWT Bearer token required

**Request Headers**

```http
Authorization: Bearer <token>
```

**Request Body**

```json
{
  "name": "Acme Corp",
  "slug": "acme-corp"
}
```

**Validation Rules**:
- `name`: Required, non-empty string
- `slug`: Required, unique, lowercase alphanumeric with hyphens

**Response** `201 Created`

```json
{
  "id": "660e8400-e29b-41d4-a716-446655440001",
  "name": "Acme Corp",
  "slug": "acme-corp",
  "created_at": "2024-01-15T10:30:00Z"
}
```

**Errors**:
- `400` - Validation error
- `401` - Unauthorized
- `409` - Team slug already exists
- `500` - Server error

**Side Effects**:
- Creates three default roles (admin, editor, viewer)
- Adds creator as admin member

---

### GET /api/teams

List all teams the authenticated user belongs to.

**Authentication**: JWT Bearer token required

**Request Headers**

```http
Authorization: Bearer <token>
```

**Response** `200 OK`

```json
{
  "teams": [
    {
      "id": "660e8400-e29b-41d4-a716-446655440001",
      "name": "Acme Corp",
      "slug": "acme-corp",
      "created_at": "2024-01-15T10:30:00Z"
    }
  ]
}
```

**Errors**:
- `401` - Unauthorized
- `500` - Server error

---

### GET /api/teams/:teamId

Get team details.

**Authentication**: JWT Bearer token required
**Required Access**: User must be a member of the team

**Path Parameters**:
- `teamId` (UUID): Team identifier

**Request Headers**

```http
Authorization: Bearer <token>
```

**Response** `200 OK`

```json
{
  "id": "660e8400-e29b-41d4-a716-446655440001",
  "name": "Acme Corp",
  "slug": "acme-corp",
  "created_at": "2024-01-15T10:30:00Z"
}
```

**Errors**:
- `400` - Invalid team ID
- `401` - Unauthorized
- `403` - Access denied (not a team member)
- `404` - Team not found
- `500` - Server error

---

### PUT /api/teams/:teamId

Update team details.

**Authentication**: JWT Bearer token required
**Required Permission**: `team:manage`

**Path Parameters**:
- `teamId` (UUID): Team identifier

**Request Headers**

```http
Authorization: Bearer <token>
```

**Request Body**

```json
{
  "name": "Acme Corporation",
  "slug": "acme-corp"
}
```

**Response** `200 OK`

```json
{
  "id": "660e8400-e29b-41d4-a716-446655440001",
  "name": "Acme Corporation",
  "slug": "acme-corp",
  "created_at": "2024-01-15T10:30:00Z"
}
```

**Errors**:
- `400` - Validation error
- `401` - Unauthorized
- `403` - Permission denied
- `404` - Team not found
- `500` - Server error

---

### DELETE /api/teams/:teamId

Delete a team and all associated data (blueprints, entities, roles, memberships, API keys).

**Authentication**: JWT Bearer token required
**Required Permission**: `team:manage`

**Path Parameters**:
- `teamId` (UUID): Team identifier

**Request Headers**

```http
Authorization: Bearer <token>
```

**Response** `204 No Content`

**Errors**:
- `400` - Invalid team ID
- `401` - Unauthorized
- `403` - Permission denied
- `500` - Server error

**Warning**: This operation is irreversible and cascades to all team resources.

---

## Role Management

### GET /api/teams/:teamId/roles

List all roles for a team.

**Authentication**: JWT Bearer token required
**Required Access**: User must be a team member

**Path Parameters**:
- `teamId` (UUID): Team identifier

**Request Headers**

```http
Authorization: Bearer <token>
```

**Response** `200 OK`

```json
{
  "roles": [
    {
      "id": "770e8400-e29b-41d4-a716-446655440002",
      "team_id": "660e8400-e29b-41d4-a716-446655440001",
      "name": "admin",
      "permissions": [
        "team:manage",
        "blueprint:read",
        "blueprint:write",
        "blueprint:delete",
        "entity:read",
        "entity:write",
        "entity:delete"
      ],
      "created_at": "2024-01-15T10:30:00Z"
    }
  ]
}
```

**Errors**:
- `400` - Invalid team ID
- `401` - Unauthorized
- `403` - Access denied
- `500` - Server error

---

### POST /api/teams/:teamId/roles

Create a custom role for a team.

**Authentication**: JWT Bearer token required
**Required Permission**: `team:manage`

**Path Parameters**:
- `teamId` (UUID): Team identifier

**Request Headers**

```http
Authorization: Bearer <token>
```

**Request Body**

```json
{
  "name": "developer",
  "permissions": [
    "blueprint:read",
    "entity:read",
    "entity:write"
  ]
}
```

**Validation Rules**:
- `name`: Required, unique within team
- `permissions`: Required array of valid permission strings

**Response** `201 Created`

```json
{
  "id": "770e8400-e29b-41d4-a716-446655440004",
  "team_id": "660e8400-e29b-41d4-a716-446655440001",
  "name": "developer",
  "permissions": [
    "blueprint:read",
    "entity:read",
    "entity:write"
  ],
  "created_at": "2024-01-15T10:30:00Z"
}
```

**Errors**:
- `400` - Validation error
- `401` - Unauthorized
- `403` - Permission denied
- `500` - Server error

---

## Member Management

### GET /api/teams/:teamId/members

List all members of a team.

**Authentication**: JWT Bearer token required
**Required Access**: User must be a team member

**Path Parameters**:
- `teamId` (UUID): Team identifier

**Request Headers**

```http
Authorization: Bearer <token>
```

**Response** `200 OK`

```json
{
  "members": [
    {
      "id": "880e8400-e29b-41d4-a716-446655440005",
      "team_id": "660e8400-e29b-41d4-a716-446655440001",
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "role_id": "770e8400-e29b-41d4-a716-446655440002",
      "created_at": "2024-01-15T10:30:00Z"
    }
  ]
}
```

**Errors**:
- `400` - Invalid team ID
- `401` - Unauthorized
- `403` - Access denied
- `500` - Server error

---

### POST /api/teams/:teamId/members

Add a user to a team.

**Authentication**: JWT Bearer token required
**Required Permission**: `team:manage`

**Path Parameters**:
- `teamId` (UUID): Team identifier

**Request Headers**

```http
Authorization: Bearer <token>
```

**Request Body**

```json
{
  "email": "jane@example.com",
  "role_id": "770e8400-e29b-41d4-a716-446655440003"
}
```

**Validation Rules**:
- `email`: Required, must exist in users table
- `role_id`: Required, must exist in team's roles

**Response** `201 Created`

```json
{
  "id": "880e8400-e29b-41d4-a716-446655440006",
  "team_id": "660e8400-e29b-41d4-a716-446655440001",
  "user_id": "550e8400-e29b-41d4-a716-446655440001",
  "role_id": "770e8400-e29b-41d4-a716-446655440003",
  "created_at": "2024-01-15T10:30:00Z"
}
```

**Errors**:
- `400` - Validation error or invalid role ID
- `401` - Unauthorized
- `403` - Permission denied
- `404` - User not found
- `409` - User already a team member
- `500` - Server error

---

### DELETE /api/teams/:teamId/members/:userId

Remove a user from a team.

**Authentication**: JWT Bearer token required
**Required Permission**: `team:manage`

**Path Parameters**:
- `teamId` (UUID): Team identifier
- `userId` (UUID): User identifier

**Request Headers**

```http
Authorization: Bearer <token>
```

**Response** `204 No Content`

**Errors**:
- `400` - Invalid team/user ID
- `401` - Unauthorized
- `403` - Permission denied
- `404` - Membership not found
- `500` - Server error

---

## API Key Management

### GET /api/teams/:teamId/api-keys

List all API keys for a team.

**Authentication**: JWT Bearer token required
**Required Access**: User must be a team member

**Path Parameters**:
- `teamId` (UUID): Team identifier

**Request Headers**

```http
Authorization: Bearer <token>
```

**Response** `200 OK`

```json
{
  "api_keys": [
    {
      "id": "990e8400-e29b-41d4-a716-446655440006",
      "team_id": "660e8400-e29b-41d4-a716-446655440001",
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Production API Key",
      "permissions": ["blueprint:read", "entity:read", "entity:write"],
      "expires_at": "2025-01-15T10:30:00Z",
      "last_used_at": "2024-01-15T14:20:00Z",
      "created_at": "2024-01-15T10:30:00Z"
    }
  ]
}
```

**Note**: The actual key value is never returned after creation.

**Errors**:
- `400` - Invalid team ID
- `401` - Unauthorized
- `403` - Access denied
- `500` - Server error

---

### POST /api/teams/:teamId/api-keys

Create a new API key.

**Authentication**: JWT Bearer token required
**Required Permission**: `team:manage`

**Path Parameters**:
- `teamId` (UUID): Team identifier

**Request Headers**

```http
Authorization: Bearer <token>
```

**Request Body**

```json
{
  "name": "CI/CD Pipeline",
  "permissions": ["blueprint:read", "entity:read", "entity:write"],
  "expires_at": "2025-01-15T10:30:00Z"
}
```

**Validation Rules**:
- `name`: Required, descriptive name for the key
- `permissions`: Optional array (defaults to empty)
- `expires_at`: Optional ISO 8601 timestamp

**Response** `201 Created`

```json
{
  "api_key": {
    "id": "990e8400-e29b-41d4-a716-446655440007",
    "team_id": "660e8400-e29b-41d4-a716-446655440001",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "CI/CD Pipeline",
    "permissions": ["blueprint:read", "entity:read", "entity:write"],
    "expires_at": "2025-01-15T10:30:00Z",
    "created_at": "2024-01-15T10:30:00Z"
  },
  "key": "bp_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6a7b8c9d0e1f2"
}
```

**Important**: The `key` field is only returned once during creation. Store it securely - it cannot be retrieved later.

**Errors**:
- `400` - Validation error
- `401` - Unauthorized
- `403` - Permission denied
- `500` - Server error

---

### DELETE /api/api-keys/:keyId

Revoke an API key.

**Authentication**: JWT Bearer token or API Key
**Required Permission**: `team:manage`
**Required Header**: `X-Team-ID` (if using JWT)

**Path Parameters**:
- `keyId` (UUID): API key identifier

**Request Headers**

```http
Authorization: Bearer <token>
X-Team-ID: 660e8400-e29b-41d4-a716-446655440001
```

**Response** `204 No Content`

**Errors**:
- `400` - Invalid key ID or missing team ID
- `401` - Unauthorized
- `403` - Permission denied
- `404` - API key not found
- `500` - Server error

---

## Blueprint Management

Blueprints define schemas for entity types using JSON Schema.

### POST /api/blueprints

Create a new blueprint with JSON Schema definition.

**Authentication**: JWT Bearer token or API Key
**Required Permission**: `blueprint:write`
**Required Context**: Team ID

**Request Headers**

```http
Authorization: Bearer <token>
X-Team-ID: 660e8400-e29b-41d4-a716-446655440001
```

**Request Body**

```json
{
  "id": "service",
  "title": "Service",
  "description": "A microservice in our infrastructure",
  "icon": "🚀",
  "schema": {
    "type": "object",
    "properties": {
      "name": {
        "type": "string",
        "title": "Service Name"
      },
      "version": {
        "type": "string",
        "title": "Version",
        "pattern": "^\\d+\\.\\d+\\.\\d+$"
      },
      "language": {
        "type": "string",
        "title": "Programming Language"
      },
      "repository": {
        "type": "string",
        "format": "uri",
        "title": "Repository URL"
      },
      "team": {
        "type": "string",
        "title": "Owning Team"
      },
      "status": {
        "type": "string",
        "enum": ["active", "deprecated", "sunset"],
        "title": "Status"
      },
      "dependencies": {
        "type": "array",
        "items": {
          "type": "string"
        },
        "title": "Dependencies"
      }
    },
    "required": ["name", "version", "language"]
  }
}
```

**Validation Rules**:
- `id`: Required, unique within team, lowercase alphanumeric with hyphens
- `title`: Required, display name
- `schema`: Required, valid JSON Schema object

**Response** `201 Created`

```json
{
  "id": "service",
  "team_id": "660e8400-e29b-41d4-a716-446655440001",
  "title": "Service",
  "description": "A microservice in our infrastructure",
  "icon": "🚀",
  "schema": { /* full schema */ },
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

**Errors**:
- `400` - Validation error or missing team ID
- `401` - Unauthorized
- `403` - Permission denied
- `409` - Blueprint ID already exists
- `500` - Server error

---

### GET /api/blueprints

List all blueprints for a team.

**Authentication**: JWT Bearer token or API Key
**Required Permission**: `blueprint:read`
**Required Context**: Team ID

**Request Headers**

```http
Authorization: Bearer <token>
X-Team-ID: 660e8400-e29b-41d4-a716-446655440001
```

**Response** `200 OK`

```json
{
  "blueprints": [
    {
      "id": "service",
      "team_id": "660e8400-e29b-41d4-a716-446655440001",
      "title": "Service",
      "description": "A microservice in our infrastructure",
      "icon": "🚀",
      "schema": { /* full schema */ },
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 1
}
```

**Errors**:
- `400` - Missing team ID
- `401` - Unauthorized
- `403` - Permission denied
- `500` - Server error

---

### GET /api/blueprints/:id

Get a specific blueprint by ID.

**Authentication**: JWT Bearer token or API Key
**Required Permission**: `blueprint:read`
**Required Context**: Team ID

**Path Parameters**:
- `id` (string): Blueprint identifier

**Request Headers**

```http
Authorization: Bearer <token>
X-Team-ID: 660e8400-e29b-41d4-a716-446655440001
```

**Response** `200 OK`

```json
{
  "id": "service",
  "team_id": "660e8400-e29b-41d4-a716-446655440001",
  "title": "Service",
  "description": "A microservice in our infrastructure",
  "icon": "🚀",
  "schema": { /* full schema */ },
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

**Errors**:
- `400` - Missing team ID
- `401` - Unauthorized
- `403` - Permission denied
- `404` - Blueprint not found
- `500` - Server error

---

### PUT /api/blueprints/:id

Update an existing blueprint.

**Authentication**: JWT Bearer token or API Key
**Required Permission**: `blueprint:write`
**Required Context**: Team ID

**Path Parameters**:
- `id` (string): Blueprint identifier

**Request Headers**

```http
Authorization: Bearer <token>
X-Team-ID: 660e8400-e29b-41d4-a716-446655440001
```

**Request Body**

All fields are optional - only include fields to update:

```json
{
  "title": "Microservice",
  "description": "Updated description",
  "icon": "⚡",
  "schema": { /* updated schema */ }
}
```

**Response** `200 OK`

```json
{
  "id": "service",
  "team_id": "660e8400-e29b-41d4-a716-446655440001",
  "title": "Microservice",
  "description": "Updated description",
  "icon": "⚡",
  "schema": { /* updated schema */ },
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T11:00:00Z"
}
```

**Errors**:
- `400` - Validation error or missing team ID
- `401` - Unauthorized
- `403` - Permission denied
- `404` - Blueprint not found
- `500` - Server error

---

### DELETE /api/blueprints/:id

Delete a blueprint and all its entities.

**Authentication**: JWT Bearer token or API Key
**Required Permission**: `blueprint:delete`
**Required Context**: Team ID

**Path Parameters**:
- `id` (string): Blueprint identifier

**Request Headers**

```http
Authorization: Bearer <token>
X-Team-ID: 660e8400-e29b-41d4-a716-446655440001
```

**Response** `204 No Content`

**Errors**:
- `400` - Missing team ID
- `401` - Unauthorized
- `403` - Permission denied
- `404` - Blueprint not found
- `500` - Server error

**Warning**: This cascades to delete all entities of this blueprint type.

---

## Entity Management

Entities are instances of blueprints, validated against their blueprint's JSON Schema.

### POST /api/blueprints/:blueprintId/entities

Create a new entity instance.

**Authentication**: JWT Bearer token or API Key
**Required Permission**: `entity:write`
**Required Context**: Team ID

**Path Parameters**:
- `blueprintId` (string): Blueprint identifier

**Request Headers**

```http
Authorization: Bearer <token>
X-Team-ID: 660e8400-e29b-41d4-a716-446655440001
```

**Request Body**

```json
{
  "identifier": "auth-service",
  "title": "Authentication Service",
  "data": {
    "name": "auth-service",
    "version": "2.3.1",
    "language": "Go",
    "repository": "https://github.com/acme/auth-service",
    "team": "Platform",
    "status": "active",
    "dependencies": ["postgres", "redis"]
  }
}
```

**Validation Rules**:
- `identifier`: Required, unique within blueprint, lowercase alphanumeric with hyphens
- `title`: Optional, display name
- `data`: Required, must validate against blueprint's schema

**Response** `201 Created`

```json
{
  "id": "aa0e8400-e29b-41d4-a716-446655440008",
  "team_id": "660e8400-e29b-41d4-a716-446655440001",
  "blueprint_id": "service",
  "identifier": "auth-service",
  "title": "Authentication Service",
  "data": {
    "name": "auth-service",
    "version": "2.3.1",
    "language": "Go",
    "repository": "https://github.com/acme/auth-service",
    "team": "Platform",
    "status": "active",
    "dependencies": ["postgres", "redis"]
  },
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

**Errors**:
- `400` - Validation error (schema validation failure) or missing team ID
- `401` - Unauthorized
- `403` - Permission denied
- `404` - Blueprint not found
- `409` - Entity identifier already exists
- `500` - Server error

---

### GET /api/blueprints/:blueprintId/entities

List entities with pagination.

**Authentication**: JWT Bearer token or API Key
**Required Permission**: `entity:read`
**Required Context**: Team ID

**Path Parameters**:
- `blueprintId` (string): Blueprint identifier

**Query Parameters**:
- `limit` (integer): Items per page (default: 50, max: 100)
- `offset` (integer): Items to skip (default: 0)

**Request Headers**

```http
Authorization: Bearer <token>
X-Team-ID: 660e8400-e29b-41d4-a716-446655440001
```

**Response** `200 OK`

```json
{
  "entities": [
    {
      "id": "aa0e8400-e29b-41d4-a716-446655440008",
      "team_id": "660e8400-e29b-41d4-a716-446655440001",
      "blueprint_id": "service",
      "identifier": "auth-service",
      "title": "Authentication Service",
      "data": { /* full data */ },
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 1,
  "limit": 50,
  "offset": 0
}
```

**Errors**:
- `400` - Missing team ID
- `401` - Unauthorized
- `403` - Permission denied
- `500` - Server error

---

### POST /api/blueprints/:blueprintId/entities/search

Advanced search with filters on JSONB data.

**Authentication**: JWT Bearer token or API Key
**Required Permission**: `entity:read`
**Required Context**: Team ID

**Path Parameters**:
- `blueprintId` (string): Blueprint identifier

**Request Headers**

```http
Authorization: Bearer <token>
X-Team-ID: 660e8400-e29b-41d4-a716-446655440001
```

**Request Body**

```json
{
  "filters": [
    {
      "property": "language",
      "operator": "eq",
      "value": "Go"
    },
    {
      "property": "status",
      "operator": "eq",
      "value": "active"
    },
    {
      "property": "dependencies",
      "operator": "contains",
      "value": "postgres"
    }
  ],
  "order_by": "created_at",
  "order_dir": "desc",
  "limit": 20,
  "offset": 0
}
```

**Filter Operators**:

| Operator | Description | Example |
|----------|-------------|---------|
| `eq` | Equals | `{"property": "status", "operator": "eq", "value": "active"}` |
| `neq` | Not equals | `{"property": "status", "operator": "neq", "value": "deprecated"}` |
| `gt` | Greater than | `{"property": "version", "operator": "gt", "value": 5}` |
| `gte` | Greater than or equal | `{"property": "version", "operator": "gte", "value": 5}` |
| `lt` | Less than | `{"property": "version", "operator": "lt", "value": 10}` |
| `lte` | Less than or equal | `{"property": "version", "operator": "lte", "value": 10}` |
| `contains` | Contains (arrays/strings) | `{"property": "dependencies", "operator": "contains", "value": "redis"}` |
| `exists` | Property exists | `{"property": "metadata.tags", "operator": "exists", "value": true}` |
| `in` | Value in array | `{"property": "status", "operator": "in", "value": ["active", "beta"]}` |

**Nested Properties**: Use dot notation for nested JSONB properties:

```json
{
  "property": "metadata.tags.environment",
  "operator": "eq",
  "value": "production"
}
```

**Response** `200 OK`

```json
{
  "entities": [
    {
      "id": "aa0e8400-e29b-41d4-a716-446655440008",
      "team_id": "660e8400-e29b-41d4-a716-446655440001",
      "blueprint_id": "service",
      "identifier": "auth-service",
      "title": "Authentication Service",
      "data": { /* full data */ },
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 1,
  "limit": 20,
  "offset": 0
}
```

**Errors**:
- `400` - Validation error (invalid operator, property) or missing team ID
- `401` - Unauthorized
- `403` - Permission denied
- `500` - Server error

---

### GET /api/blueprints/:blueprintId/entities/by-identifier/:identifier

Get entity by its unique identifier within a blueprint.

**Authentication**: JWT Bearer token or API Key
**Required Permission**: `entity:read`
**Required Context**: Team ID

**Path Parameters**:
- `blueprintId` (string): Blueprint identifier
- `identifier` (string): Entity identifier

**Request Headers**

```http
Authorization: Bearer <token>
X-Team-ID: 660e8400-e29b-41d4-a716-446655440001
```

**Response** `200 OK`

```json
{
  "id": "aa0e8400-e29b-41d4-a716-446655440008",
  "team_id": "660e8400-e29b-41d4-a716-446655440001",
  "blueprint_id": "service",
  "identifier": "auth-service",
  "title": "Authentication Service",
  "data": { /* full data */ },
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

**Errors**:
- `400` - Missing team ID
- `401` - Unauthorized
- `403` - Permission denied
- `404` - Entity not found
- `500` - Server error

---

### GET /api/entities/:id

Get entity by its UUID.

**Authentication**: JWT Bearer token or API Key
**Required Permission**: `entity:read`
**Required Context**: Team ID

**Path Parameters**:
- `id` (UUID): Entity UUID

**Request Headers**

```http
Authorization: Bearer <token>
X-Team-ID: 660e8400-e29b-41d4-a716-446655440001
```

**Response** `200 OK`

```json
{
  "id": "aa0e8400-e29b-41d4-a716-446655440008",
  "team_id": "660e8400-e29b-41d4-a716-446655440001",
  "blueprint_id": "service",
  "identifier": "auth-service",
  "title": "Authentication Service",
  "data": { /* full data */ },
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

**Errors**:
- `400` - Invalid entity ID or missing team ID
- `401` - Unauthorized
- `403` - Permission denied
- `404` - Entity not found
- `500` - Server error

---

### PUT /api/entities/:id

Update an existing entity. Data is validated against the blueprint schema.

**Authentication**: JWT Bearer token or API Key
**Required Permission**: `entity:write`
**Required Context**: Team ID

**Path Parameters**:
- `id` (UUID): Entity UUID

**Request Headers**

```http
Authorization: Bearer <token>
X-Team-ID: 660e8400-e29b-41d4-a716-446655440001
```

**Request Body**

Both fields are optional - only include what you want to update:

```json
{
  "title": "Auth Service (Updated)",
  "data": {
    "name": "auth-service",
    "version": "2.4.0",
    "language": "Go",
    "repository": "https://github.com/acme/auth-service",
    "team": "Platform",
    "status": "active",
    "dependencies": ["postgres", "redis", "kafka"]
  }
}
```

**Response** `200 OK`

```json
{
  "id": "aa0e8400-e29b-41d4-a716-446655440008",
  "team_id": "660e8400-e29b-41d4-a716-446655440001",
  "blueprint_id": "service",
  "identifier": "auth-service",
  "title": "Auth Service (Updated)",
  "data": {
    "name": "auth-service",
    "version": "2.4.0",
    "language": "Go",
    "repository": "https://github.com/acme/auth-service",
    "team": "Platform",
    "status": "active",
    "dependencies": ["postgres", "redis", "kafka"]
  },
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T11:15:00Z"
}
```

**Errors**:
- `400` - Validation error (schema validation failure) or invalid entity ID
- `401` - Unauthorized
- `403` - Permission denied
- `404` - Entity not found
- `500` - Server error

---

### DELETE /api/entities/:id

Delete an entity.

**Authentication**: JWT Bearer token or API Key
**Required Permission**: `entity:delete`
**Required Context**: Team ID

**Path Parameters**:
- `id` (UUID): Entity UUID

**Request Headers**

```http
Authorization: Bearer <token>
X-Team-ID: 660e8400-e29b-41d4-a716-446655440001
```

**Response** `204 No Content`

**Errors**:
- `400` - Invalid entity ID or missing team ID
- `401` - Unauthorized
- `403` - Permission denied
- `404` - Entity not found
- `500` - Server error

---

## Admin - Super Admin Only

All admin endpoints require super admin privileges and are protected by the `RequireSuperAdmin()` middleware.

### Team Management

#### List All Teams

```
GET /api/admin/teams
```

Lists all teams in the system regardless of membership.

**Headers**:
```
Authorization: Bearer <jwt_token>
```

**Response** (200 OK):
```json
{
  "teams": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Team Name",
      "slug": "team-name",
      "created_at": "2026-01-12T10:00:00Z"
    }
  ]
}
```

#### Get Team Details

```
GET /api/admin/teams/:teamId
```

View details for a specific team.

**Parameters**:
- `teamId` (required) - UUID of the team

**Response** (200 OK):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Team Name",
  "slug": "team-name",
  "created_at": "2026-01-12T10:00:00Z"
}
```

### User Management

#### List All Users

```
GET /api/admin/users?limit=50&offset=0
```

Lists all platform users with pagination.

**Query Parameters**:
- `limit` (optional) - Items per page, max 500, default 50
- `offset` (optional) - Pagination offset, default 0

**Response** (200 OK):
```json
{
  "users": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "user@example.com",
      "name": "User Name",
      "status": "active",
      "is_super_admin": false,
      "created_at": "2026-01-12T10:00:00Z"
    }
  ],
  "limit": 50,
  "offset": 0
}
```

#### Get User Details

```
GET /api/admin/users/:userId
```

View user details including team memberships.

**Parameters**:
- `userId` (required) - UUID of the user

**Response** (200 OK):
```json
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "name": "User Name",
    "is_super_admin": false,
    "created_at": "2026-01-12T10:00:00Z"
  },
  "memberships": [
    {
      "id": "660e8400-e29b-41d4-a716-446655440000",
      "team_id": "770e8400-e29b-41d4-a716-446655440000",
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "role_id": "880e8400-e29b-41d4-a716-446655440000",
      "created_at": "2026-01-12T10:00:00Z"
    }
  ]
}
```

#### Update User

```
PUT /api/admin/users/:userId
```

Update user information (name, status).

**Parameters**:
- `userId` (required) - UUID of the user

**Request Body**:
```json
{
  "name": "Updated Name",
  "status": "active"
}
```

**Response** (200 OK):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "name": "Updated Name",
  "status": "active",
  "is_super_admin": false,
  "created_at": "2026-01-12T10:00:00Z"
}
```

### Super Admin Delegation

#### Promote User to Super Admin

```
POST /api/admin/users/:userId/promote
```

Grant super admin privileges to a user.

**Parameters**:
- `userId` (required) - UUID of the user to promote

**Response** (200 OK):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "name": "User Name",
  "is_super_admin": true,
  "super_admin_promoted_at": "2026-01-12T10:30:00Z",
  "super_admin_promoted_by": "admin-user-id",
  "created_at": "2026-01-12T10:00:00Z"
}
```

**Errors**:
- `403` - User is not a super admin (unauthorized)
- `404` - User not found
- `400` - User already super admin

#### Demote Super Admin

```
POST /api/admin/users/:userId/demote
```

Revoke super admin privileges from a user (except last super admin).

**Parameters**:
- `userId` (required) - UUID of the super admin to demote

**Response** (200 OK):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "name": "User Name",
  "is_super_admin": false,
  "created_at": "2026-01-12T10:00:00Z"
}
```

**Errors**:
- `403` - User is not a super admin (unauthorized)
- `404` - User not found
- `400` - User is not a super admin
- `409` - Cannot demote the last super admin

### Audit Logging

#### Query Audit Logs

```
GET /api/admin/audit-logs?limit=50&offset=0
```

View audit logs of all super admin actions.

**Query Parameters**:
- `limit` (optional) - Items per page, max 500, default 50
- `offset` (optional) - Pagination offset, default 0

**Response** (200 OK):
```json
{
  "logs": [
    {
      "id": "990e8400-e29b-41d4-a716-446655440000",
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "actor_type": "super_admin",
      "entity_type": "user",
      "entity_id": "aa0e8400-e29b-41d4-a716-446655440000",
      "action": "promote",
      "ip_address": "192.168.1.100",
      "user_agent": "curl/7.68.0",
      "result_status": "success",
      "created_at": "2026-01-12T10:30:00Z"
    }
  ],
  "limit": 50,
  "offset": 0
}
```

---

## Examples

### Complete Workflow Example

```bash
# 1. Register a user
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "securePassword123"
  }'

# Response: {"token": "eyJ...", "user": {...}}

# 2. Save token for subsequent requests
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# 3. Create a team
TEAM_RESPONSE=$(curl -X POST http://localhost:8080/api/teams \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Corp",
    "slug": "acme-corp"
  }')

TEAM_ID=$(echo $TEAM_RESPONSE | jq -r '.id')

# 4. Create a blueprint
curl -X POST http://localhost:8080/api/blueprints \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Team-ID: $TEAM_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "service",
    "title": "Service",
    "description": "A microservice",
    "icon": "🚀",
    "schema": {
      "type": "object",
      "properties": {
        "name": {"type": "string"},
        "version": {"type": "string"},
        "language": {"type": "string"}
      },
      "required": ["name", "version"]
    }
  }'

# 5. Create an entity
curl -X POST http://localhost:8080/api/blueprints/service/entities \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Team-ID: $TEAM_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "identifier": "auth-service",
    "title": "Auth Service",
    "data": {
      "name": "auth-service",
      "version": "1.0.0",
      "language": "Go"
    }
  }'

# 6. Search entities
curl -X POST http://localhost:8080/api/blueprints/service/entities/search \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Team-ID: $TEAM_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "filters": [
      {
        "property": "language",
        "operator": "eq",
        "value": "Go"
      }
    ]
  }'

# 7. Update entity
ENTITY_ID="aa0e8400-e29b-41d4-a716-446655440008"
curl -X PUT http://localhost:8080/api/entities/$ENTITY_ID \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Team-ID: $TEAM_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "data": {
      "name": "auth-service",
      "version": "1.1.0",
      "language": "Go"
    }
  }'
```

### Using API Keys

```bash
# 1. Create API key (using JWT token)
API_KEY_RESPONSE=$(curl -X POST http://localhost:8080/api/teams/$TEAM_ID/api-keys \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "CI/CD Pipeline",
    "permissions": ["entity:read", "entity:write"]
  }')

API_KEY=$(echo $API_KEY_RESPONSE | jq -r '.key')

# 2. Use API key for authentication (no X-Team-ID needed)
curl -X GET http://localhost:8080/api/blueprints \
  -H "Authorization: ApiKey $API_KEY"

# 3. Create entity with API key
curl -X POST http://localhost:8080/api/blueprints/service/entities \
  -H "Authorization: ApiKey $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "identifier": "api-gateway",
    "title": "API Gateway",
    "data": {
      "name": "api-gateway",
      "version": "2.0.0",
      "language": "Node.js"
    }
  }'
```

### Advanced Search Examples

```bash
# Search by multiple conditions
curl -X POST http://localhost:8080/api/blueprints/service/entities/search \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Team-ID: $TEAM_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "filters": [
      {"property": "language", "operator": "eq", "value": "Go"},
      {"property": "status", "operator": "eq", "value": "active"},
      {"property": "version", "operator": "gte", "value": "2.0.0"}
    ],
    "order_by": "created_at",
    "order_dir": "desc",
    "limit": 10
  }'

# Search with nested properties
curl -X POST http://localhost:8080/api/blueprints/service/entities/search \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Team-ID: $TEAM_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "filters": [
      {
        "property": "metadata.environment",
        "operator": "eq",
        "value": "production"
      },
      {
        "property": "metadata.tags",
        "operator": "contains",
        "value": "critical"
      }
    ]
  }'

# Search with 'in' operator
curl -X POST http://localhost:8080/api/blueprints/service/entities/search \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Team-ID: $TEAM_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "filters": [
      {
        "property": "status",
        "operator": "in",
        "value": ["active", "beta", "staging"]
      }
    ]
  }'
```

### Team Management Example

```bash
# List all teams
curl -X GET http://localhost:8080/api/teams \
  -H "Authorization: Bearer $TOKEN"

# Get team details
curl -X GET http://localhost:8080/api/teams/$TEAM_ID \
  -H "Authorization: Bearer $TOKEN"

# Create custom role
ROLE_RESPONSE=$(curl -X POST http://localhost:8080/api/teams/$TEAM_ID/roles \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "developer",
    "permissions": ["blueprint:read", "entity:read", "entity:write"]
  }')

ROLE_ID=$(echo $ROLE_RESPONSE | jq -r '.id')

# Add team member
curl -X POST http://localhost:8080/api/teams/$TEAM_ID/members \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "jane@example.com",
    "role_id": "'$ROLE_ID'"
  }'

# List team members
curl -X GET http://localhost:8080/api/teams/$TEAM_ID/members \
  -H "Authorization: Bearer $TOKEN"
```

---

## Rate Limiting

Currently, Baseplate does not implement rate limiting. This is planned for future releases.

## Pagination

List and search endpoints support pagination:

- **Query Parameters**: `limit` (max items), `offset` (skip count)
- **Defaults**: `limit=50`, `offset=0`
- **Maximum**: `limit` cannot exceed 100

**Response includes**:
- `entities`/`items`: Array of results
- `total`: Total count matching query
- `limit`: Items per page
- `offset`: Items skipped

## Webhooks

Webhooks are planned but not yet implemented. Future versions will support webhook notifications for entity changes, blueprint updates, and other events.

---

For architecture details, see [ARCHITECTURE.md](./ARCHITECTURE.md).
For deployment instructions, see [DEPLOYMENT.md](./DEPLOYMENT.md).
For development setup, see [DEVELOPMENT.md](./DEVELOPMENT.md).
