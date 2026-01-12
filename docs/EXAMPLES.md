# Baseplate Usage Examples

Comprehensive examples demonstrating real-world usage of Baseplate API.

## Table of Contents

- [Getting Started](#getting-started)
- [User Management](#user-management)
- [Team Management](#team-management)
- [Blueprint and Entity Workflows](#blueprint-and-entity-workflows)
- [Advanced Search](#advanced-search)
- [API Key Usage](#api-key-usage)
- [Complete Application Example](#complete-application-example)

## Getting Started

### Prerequisites

```bash
# Baseplate server running
curl http://localhost:8080/api/health
# Response: {"status":"ok"}

# Tools needed
# - curl (or Postman, HTTPie)
# - jq (for JSON parsing, optional)
```

---

## User Management

### Register and Login

```bash
#!/bin/bash
# Example: User registration and authentication

BASE_URL="http://localhost:8080/api"

# 1. Register a new user
echo "Registering user..."
REGISTER_RESPONSE=$(curl -s -X POST $BASE_URL/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alice Johnson",
    "email": "alice@acme.com",
    "password": "SecurePass123!"
  }')

echo "Register response: $REGISTER_RESPONSE"

# Extract token
TOKEN=$(echo $REGISTER_RESPONSE | jq -r '.token')
USER_ID=$(echo $REGISTER_RESPONSE | jq -r '.user.id')

echo "Token: $TOKEN"
echo "User ID: $USER_ID"

# 2. Verify authentication with /me endpoint
echo -e "\nFetching user info..."
curl -s -X GET $BASE_URL/auth/me \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# 3. Login (alternative to registration)
echo -e "\nLogging in..."
LOGIN_RESPONSE=$(curl -s -X POST $BASE_URL/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "alice@acme.com",
    "password": "SecurePass123!"
  }')

echo "Login response: $LOGIN_RESPONSE"
```

**Output**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Alice Johnson",
    "email": "alice@acme.com",
    "status": "active",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

---

## Team Management

### Create Team and Manage Members

```bash
#!/bin/bash
# Example: Team setup with members and roles

BASE_URL="http://localhost:8080/api"
TOKEN="your-jwt-token-here"

# 1. Create a team
echo "Creating team..."
TEAM_RESPONSE=$(curl -s -X POST $BASE_URL/teams \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Engineering",
    "slug": "acme-engineering"
  }')

TEAM_ID=$(echo $TEAM_RESPONSE | jq -r '.id')
echo "Team ID: $TEAM_ID"
echo $TEAM_RESPONSE | jq '.'

# 2. List all teams
echo -e "\nListing teams..."
curl -s -X GET $BASE_URL/teams \
  -H "Authorization: Bearer $TOKEN" | jq '.teams'

# 3. List team roles (default: admin, editor, viewer)
echo -e "\nListing roles..."
ROLES_RESPONSE=$(curl -s -X GET $BASE_URL/teams/$TEAM_ID/roles \
  -H "Authorization: Bearer $TOKEN")

echo $ROLES_RESPONSE | jq '.roles'

# Extract editor role ID
EDITOR_ROLE_ID=$(echo $ROLES_RESPONSE | jq -r '.roles[] | select(.name=="editor") | .id')

# 4. Create custom role
echo -e "\nCreating custom role..."
CUSTOM_ROLE=$(curl -s -X POST $BASE_URL/teams/$TEAM_ID/roles \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "developer",
    "permissions": [
      "blueprint:read",
      "entity:read",
      "entity:write"
    ]
  }')

echo $CUSTOM_ROLE | jq '.'
DEVELOPER_ROLE_ID=$(echo $CUSTOM_ROLE | jq -r '.id')

# 5. Add team member (assuming bob@acme.com already registered)
echo -e "\nAdding team member..."
curl -s -X POST $BASE_URL/teams/$TEAM_ID/members \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"bob@acme.com\",
    \"role_id\": \"$DEVELOPER_ROLE_ID\"
  }" | jq '.'

# 6. List team members
echo -e "\nListing team members..."
curl -s -X GET $BASE_URL/teams/$TEAM_ID/members \
  -H "Authorization: Bearer $TOKEN" | jq '.members'
```

---

## Blueprint and Entity Workflows

### Microservices Catalog

```bash
#!/bin/bash
# Example: Building a microservices catalog

BASE_URL="http://localhost:8080/api"
TOKEN="your-jwt-token-here"
TEAM_ID="your-team-id-here"

# 1. Create "Service" blueprint
echo "Creating Service blueprint..."
SERVICE_BLUEPRINT=$(curl -s -X POST $BASE_URL/blueprints \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Team-ID: $TEAM_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "service",
    "title": "Microservice",
    "description": "A microservice in our platform",
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
          "title": "Programming Language",
          "enum": ["Go", "Python", "Node.js", "Java", "Rust"]
        },
        "repository": {
          "type": "string",
          "format": "uri",
          "title": "Git Repository"
        },
        "team": {
          "type": "string",
          "title": "Owning Team"
        },
        "status": {
          "type": "string",
          "enum": ["active", "beta", "deprecated", "sunset"],
          "title": "Status"
        },
        "dependencies": {
          "type": "array",
          "items": {
            "type": "string"
          },
          "title": "Dependencies"
        },
        "replicas": {
          "type": "integer",
          "minimum": 1,
          "title": "Replica Count"
        },
        "metadata": {
          "type": "object",
          "properties": {
            "environment": {
              "type": "string",
              "enum": ["development", "staging", "production"]
            },
            "region": {
              "type": "string"
            }
          }
        }
      },
      "required": ["name", "version", "language", "status"]
    }
  }')

echo $SERVICE_BLUEPRINT | jq '.'

# 2. Create service entities
echo -e "\nCreating auth-service..."
AUTH_SERVICE=$(curl -s -X POST $BASE_URL/blueprints/service/entities \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Team-ID: $TEAM_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "identifier": "auth-service",
    "title": "Authentication Service",
    "data": {
      "name": "auth-service",
      "version": "2.3.1",
      "language": "Go",
      "repository": "https://github.com/acme/auth-service",
      "team": "Platform",
      "status": "active",
      "dependencies": ["postgres", "redis"],
      "replicas": 3,
      "metadata": {
        "environment": "production",
        "region": "us-east-1"
      }
    }
  }')

echo $AUTH_SERVICE | jq '.'
AUTH_SERVICE_ID=$(echo $AUTH_SERVICE | jq -r '.id')

echo -e "\nCreating api-gateway..."
curl -s -X POST $BASE_URL/blueprints/service/entities \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Team-ID: $TEAM_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "identifier": "api-gateway",
    "title": "API Gateway",
    "data": {
      "name": "api-gateway",
      "version": "1.5.2",
      "language": "Node.js",
      "repository": "https://github.com/acme/api-gateway",
      "team": "Platform",
      "status": "active",
      "dependencies": ["auth-service", "user-service"],
      "replicas": 5,
      "metadata": {
        "environment": "production",
        "region": "us-east-1"
      }
    }
  }' | jq '.'

echo -e "\nCreating user-service..."
curl -s -X POST $BASE_URL/blueprints/service/entities \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Team-ID: $TEAM_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "identifier": "user-service",
    "title": "User Service",
    "data": {
      "name": "user-service",
      "version": "3.1.0",
      "language": "Python",
      "repository": "https://github.com/acme/user-service",
      "team": "Product",
      "status": "active",
      "dependencies": ["postgres", "elasticsearch"],
      "replicas": 2,
      "metadata": {
        "environment": "production",
        "region": "us-east-1"
      }
    }
  }' | jq '.'

# 3. List all services
echo -e "\nListing all services..."
curl -s -X GET "$BASE_URL/blueprints/service/entities?limit=50&offset=0" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Team-ID: $TEAM_ID" | jq '.entities[] | {identifier, title, language: .data.language, status: .data.status}'

# 4. Get service by identifier
echo -e "\nGetting auth-service..."
curl -s -X GET $BASE_URL/blueprints/service/entities/by-identifier/auth-service \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Team-ID: $TEAM_ID" | jq '.'

# 5. Update service
echo -e "\nUpdating auth-service version..."
curl -s -X PUT $BASE_URL/entities/$AUTH_SERVICE_ID \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Team-ID: $TEAM_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "data": {
      "name": "auth-service",
      "version": "2.4.0",
      "language": "Go",
      "repository": "https://github.com/acme/auth-service",
      "team": "Platform",
      "status": "active",
      "dependencies": ["postgres", "redis", "kafka"],
      "replicas": 3,
      "metadata": {
        "environment": "production",
        "region": "us-east-1"
      }
    }
  }' | jq '.'
```

---

### Database Catalog

```bash
#!/bin/bash
# Example: Creating a database catalog

BASE_URL="http://localhost:8080/api"
TOKEN="your-jwt-token-here"
TEAM_ID="your-team-id-here"

# 1. Create "Database" blueprint
echo "Creating Database blueprint..."
curl -s -X POST $BASE_URL/blueprints \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Team-ID: $TEAM_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "database",
    "title": "Database",
    "description": "Database instance",
    "icon": "🗄️",
    "schema": {
      "type": "object",
      "properties": {
        "name": {"type": "string"},
        "type": {
          "type": "string",
          "enum": ["PostgreSQL", "MySQL", "MongoDB", "Redis", "Elasticsearch"]
        },
        "version": {"type": "string"},
        "host": {"type": "string"},
        "port": {"type": "integer"},
        "size_gb": {"type": "integer"},
        "backup_enabled": {"type": "boolean"},
        "environment": {
          "type": "string",
          "enum": ["development", "staging", "production"]
        }
      },
      "required": ["name", "type", "version", "environment"]
    }
  }' | jq '.'

# 2. Create database entities
echo -e "\nCreating production PostgreSQL..."
curl -s -X POST $BASE_URL/blueprints/database/entities \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Team-ID: $TEAM_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "identifier": "prod-postgres-main",
    "title": "Production PostgreSQL",
    "data": {
      "name": "prod-postgres-main",
      "type": "PostgreSQL",
      "version": "15.2",
      "host": "db.prod.acme.com",
      "port": 5432,
      "size_gb": 500,
      "backup_enabled": true,
      "environment": "production"
    }
  }' | jq '.'

echo -e "\nCreating Redis cache..."
curl -s -X POST $BASE_URL/blueprints/database/entities \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Team-ID: $TEAM_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "identifier": "prod-redis-cache",
    "title": "Production Redis Cache",
    "data": {
      "name": "prod-redis-cache",
      "type": "Redis",
      "version": "7.0",
      "host": "redis.prod.acme.com",
      "port": 6379,
      "size_gb": 32,
      "backup_enabled": true,
      "environment": "production"
    }
  }' | jq '.'
```

---

## Advanced Search

### Search by Multiple Criteria

```bash
#!/bin/bash
# Example: Advanced entity search

BASE_URL="http://localhost:8080/api"
TOKEN="your-jwt-token-here"
TEAM_ID="your-team-id-here"

# 1. Find all active Go services
echo "Searching for active Go services..."
curl -s -X POST $BASE_URL/blueprints/service/entities/search \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Team-ID: $TEAM_ID" \
  -H "Content-Type: application/json" \
  -d '{
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
      }
    ],
    "order_by": "created_at",
    "order_dir": "desc"
  }' | jq '.entities[] | {identifier, title, version: .data.version}'

# 2. Find services with version >= 2.0.0
echo -e "\nSearching for services version >= 2.0.0..."
curl -s -X POST $BASE_URL/blueprints/service/entities/search \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Team-ID: $TEAM_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "filters": [
      {
        "property": "version",
        "operator": "gte",
        "value": "2.0.0"
      }
    ]
  }' | jq '.entities[] | {identifier, version: .data.version}'

# 3. Find services in production with postgres dependency
echo -e "\nSearching for production services using postgres..."
curl -s -X POST $BASE_URL/blueprints/service/entities/search \
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
        "property": "dependencies",
        "operator": "contains",
        "value": "postgres"
      }
    ]
  }' | jq '.entities[] | {identifier, dependencies: .data.dependencies}'

# 4. Find services with replica count > 3
echo -e "\nSearching for highly replicated services..."
curl -s -X POST $BASE_URL/blueprints/service/entities/search \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Team-ID: $TEAM_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "filters": [
      {
        "property": "replicas",
        "operator": "gt",
        "value": 3
      }
    ],
    "order_by": "data->replicas",
    "order_dir": "desc"
  }' | jq '.entities[] | {identifier, replicas: .data.replicas}'

# 5. Find services by multiple statuses (using 'in' operator)
echo -e "\nSearching for active or beta services..."
curl -s -X POST $BASE_URL/blueprints/service/entities/search \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Team-ID: $TEAM_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "filters": [
      {
        "property": "status",
        "operator": "in",
        "value": ["active", "beta"]
      }
    ]
  }' | jq '.entities[] | {identifier, status: .data.status}'

# 6. Complex search with pagination
echo -e "\nComplex search with pagination..."
curl -s -X POST $BASE_URL/blueprints/service/entities/search \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Team-ID: $TEAM_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "filters": [
      {
        "property": "status",
        "operator": "eq",
        "value": "active"
      },
      {
        "property": "metadata.environment",
        "operator": "eq",
        "value": "production"
      }
    ],
    "order_by": "created_at",
    "order_dir": "desc",
    "limit": 10,
    "offset": 0
  }' | jq '{total, entities: .entities | length, first: .entities[0].identifier}'
```

---

## API Key Usage

### Create and Use API Keys

```bash
#!/bin/bash
# Example: API key management and usage

BASE_URL="http://localhost:8080/api"
USER_TOKEN="your-jwt-token-here"
TEAM_ID="your-team-id-here"

# 1. Create API key for CI/CD
echo "Creating CI/CD API key..."
API_KEY_RESPONSE=$(curl -s -X POST $BASE_URL/teams/$TEAM_ID/api-keys \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "CI/CD Pipeline",
    "permissions": [
      "blueprint:read",
      "entity:read",
      "entity:write"
    ],
    "expires_at": "2025-12-31T23:59:59Z"
  }')

echo $API_KEY_RESPONSE | jq '.'

# IMPORTANT: Save this key! It won't be shown again
API_KEY=$(echo $API_KEY_RESPONSE | jq -r '.key')
echo -e "\n⚠️  SAVE THIS KEY: $API_KEY"

# 2. Use API key to create entity (no X-Team-ID needed!)
echo -e "\nUsing API key to create entity..."
curl -s -X POST $BASE_URL/blueprints/service/entities \
  -H "Authorization: ApiKey $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "identifier": "notification-service",
    "title": "Notification Service",
    "data": {
      "name": "notification-service",
      "version": "1.0.0",
      "language": "Python",
      "repository": "https://github.com/acme/notification-service",
      "team": "Product",
      "status": "beta",
      "dependencies": ["postgres", "sendgrid"],
      "replicas": 2,
      "metadata": {
        "environment": "staging",
        "region": "us-west-2"
      }
    }
  }' | jq '.'

# 3. List all API keys
echo -e "\nListing all API keys..."
curl -s -X GET $BASE_URL/teams/$TEAM_ID/api-keys \
  -H "Authorization: Bearer $USER_TOKEN" | jq '.api_keys[] | {name, permissions, expires_at, last_used_at}'

# 4. Delete API key
API_KEY_ID=$(echo $API_KEY_RESPONSE | jq -r '.api_key.id')
echo -e "\nDeleting API key..."
curl -s -X DELETE $BASE_URL/api-keys/$API_KEY_ID \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "X-Team-ID: $TEAM_ID"

echo "API key deleted successfully"
```

---

## Complete Application Example

### Building a Service Catalog Application

```bash
#!/bin/bash
# Complete example: Service catalog setup from scratch

set -e  # Exit on error

BASE_URL="http://localhost:8080/api"

echo "========================================="
echo "Baseplate Service Catalog Setup"
echo "========================================="

# Step 1: Register user
echo -e "\n[1/8] Registering user..."
REGISTER_RESPONSE=$(curl -s -X POST $BASE_URL/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "System Admin",
    "email": "admin@acme.com",
    "password": "AdminPass123!"
  }')

TOKEN=$(echo $REGISTER_RESPONSE | jq -r '.token')
USER_ID=$(echo $REGISTER_RESPONSE | jq -r '.user.id')
echo "✓ User registered: $USER_ID"

# Step 2: Create team
echo -e "\n[2/8] Creating team..."
TEAM_RESPONSE=$(curl -s -X POST $BASE_URL/teams \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Platform",
    "slug": "acme-platform"
  }')

TEAM_ID=$(echo $TEAM_RESPONSE | jq -r '.id')
echo "✓ Team created: $TEAM_ID"

# Step 3: Create Service blueprint
echo -e "\n[3/8] Creating Service blueprint..."
curl -s -X POST $BASE_URL/blueprints \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Team-ID: $TEAM_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "service",
    "title": "Microservice",
    "description": "A microservice in our platform",
    "icon": "🚀",
    "schema": {
      "type": "object",
      "properties": {
        "name": {"type": "string"},
        "version": {"type": "string"},
        "language": {"type": "string"},
        "status": {
          "type": "string",
          "enum": ["active", "beta", "deprecated"]
        },
        "replicas": {"type": "integer"}
      },
      "required": ["name", "version", "language", "status"]
    }
  }' > /dev/null

echo "✓ Service blueprint created"

# Step 4: Create Database blueprint
echo -e "\n[4/8] Creating Database blueprint..."
curl -s -X POST $BASE_URL/blueprints \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Team-ID: $TEAM_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "database",
    "title": "Database",
    "description": "Database instance",
    "icon": "🗄️",
    "schema": {
      "type": "object",
      "properties": {
        "name": {"type": "string"},
        "type": {"type": "string"},
        "version": {"type": "string"},
        "size_gb": {"type": "integer"}
      },
      "required": ["name", "type", "version"]
    }
  }' > /dev/null

echo "✓ Database blueprint created"

# Step 5: Create service entities
echo -e "\n[5/8] Creating service entities..."
SERVICES=("auth-service:Go:2.3.1:5" "api-gateway:Node.js:1.5.2:3" "user-service:Python:3.1.0:2")

for service in "${SERVICES[@]}"; do
  IFS=':' read -r name lang version replicas <<< "$service"
  curl -s -X POST $BASE_URL/blueprints/service/entities \
    -H "Authorization: Bearer $TOKEN" \
    -H "X-Team-ID: $TEAM_ID" \
    -H "Content-Type: application/json" \
    -d "{
      \"identifier\": \"$name\",
      \"title\": \"$(echo $name | sed 's/-/ /g' | awk '{for(i=1;i<=NF;i++) $i=toupper(substr($i,1,1)) tolower(substr($i,2))}1')\",
      \"data\": {
        \"name\": \"$name\",
        \"version\": \"$version\",
        \"language\": \"$lang\",
        \"status\": \"active\",
        \"replicas\": $replicas
      }
    }" > /dev/null
  echo "  ✓ Created: $name"
done

# Step 6: Create database entities
echo -e "\n[6/8] Creating database entities..."
curl -s -X POST $BASE_URL/blueprints/database/entities \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Team-ID: $TEAM_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "identifier": "prod-postgres",
    "title": "Production PostgreSQL",
    "data": {
      "name": "prod-postgres",
      "type": "PostgreSQL",
      "version": "15.2",
      "size_gb": 500
    }
  }' > /dev/null

echo "  ✓ Created: prod-postgres"

# Step 7: Create API key
echo -e "\n[7/8] Creating API key..."
API_KEY_RESPONSE=$(curl -s -X POST $BASE_URL/teams/$TEAM_ID/api-keys \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Read-only API Key",
    "permissions": ["blueprint:read", "entity:read"]
  }')

API_KEY=$(echo $API_KEY_RESPONSE | jq -r '.key')
echo "✓ API key created"
echo "  Key: $API_KEY"

# Step 8: Query catalog
echo -e "\n[8/8] Querying catalog..."
SERVICES_COUNT=$(curl -s -X GET "$BASE_URL/blueprints/service/entities" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Team-ID: $TEAM_ID" | jq '.total')

DATABASES_COUNT=$(curl -s -X GET "$BASE_URL/blueprints/database/entities" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Team-ID: $TEAM_ID" | jq '.total')

echo "✓ Catalog statistics:"
echo "  Services: $SERVICES_COUNT"
echo "  Databases: $DATABASES_COUNT"

echo -e "\n========================================="
echo "Setup Complete!"
echo "========================================="
echo "Team ID: $TEAM_ID"
echo "JWT Token: $TOKEN"
echo "API Key: $API_KEY"
echo ""
echo "Try these commands:"
echo "  # List services"
echo "  curl -H 'Authorization: ApiKey $API_KEY' $BASE_URL/blueprints/service/entities | jq '.'"
echo ""
echo "  # Search active services"
echo "  curl -X POST -H 'Authorization: ApiKey $API_KEY' -H 'Content-Type: application/json' \\"
echo "    $BASE_URL/blueprints/service/entities/search \\"
echo "    -d '{\"filters\": [{\"property\": \"status\", \"operator\": \"eq\", \"value\": \"active\"}]}' | jq '.'"
```

---

## Python Example

```python
#!/usr/bin/env python3
"""
Baseplate Python Client Example
"""

import requests
import json
from typing import Optional

class BaseplateClient:
    def __init__(self, base_url: str, token: Optional[str] = None):
        self.base_url = base_url
        self.token = token
        self.team_id = None

    def register(self, name: str, email: str, password: str) -> dict:
        """Register a new user"""
        response = requests.post(
            f"{self.base_url}/auth/register",
            json={"name": name, "email": email, "password": password}
        )
        response.raise_for_status()
        data = response.json()
        self.token = data["token"]
        return data

    def login(self, email: str, password: str) -> dict:
        """Login user"""
        response = requests.post(
            f"{self.base_url}/auth/login",
            json={"email": email, "password": password}
        )
        response.raise_for_status()
        data = response.json()
        self.token = data["token"]
        return data

    def create_team(self, name: str, slug: str) -> dict:
        """Create a team"""
        response = requests.post(
            f"{self.base_url}/teams",
            headers={"Authorization": f"Bearer {self.token}"},
            json={"name": name, "slug": slug}
        )
        response.raise_for_status()
        data = response.json()
        self.team_id = data["id"]
        return data

    def create_blueprint(self, blueprint_id: str, title: str, schema: dict) -> dict:
        """Create a blueprint"""
        response = requests.post(
            f"{self.base_url}/blueprints",
            headers={
                "Authorization": f"Bearer {self.token}",
                "X-Team-ID": self.team_id
            },
            json={
                "id": blueprint_id,
                "title": title,
                "schema": schema
            }
        )
        response.raise_for_status()
        return response.json()

    def create_entity(self, blueprint_id: str, identifier: str, data: dict) -> dict:
        """Create an entity"""
        response = requests.post(
            f"{self.base_url}/blueprints/{blueprint_id}/entities",
            headers={
                "Authorization": f"Bearer {self.token}",
                "X-Team-ID": self.team_id
            },
            json={
                "identifier": identifier,
                "data": data
            }
        )
        response.raise_for_status()
        return response.json()

    def search_entities(self, blueprint_id: str, filters: list) -> dict:
        """Search entities"""
        response = requests.post(
            f"{self.base_url}/blueprints/{blueprint_id}/entities/search",
            headers={
                "Authorization": f"Bearer {self.token}",
                "X-Team-ID": self.team_id
            },
            json={"filters": filters}
        )
        response.raise_for_status()
        return response.json()

# Usage example
if __name__ == "__main__":
    client = BaseplateClient("http://localhost:8080/api")

    # Register
    client.register("Alice", "alice@example.com", "password123")
    print(f"✓ Registered with token: {client.token[:20]}...")

    # Create team
    team = client.create_team("My Team", "my-team")
    print(f"✓ Created team: {team['id']}")

    # Create blueprint
    schema = {
        "type": "object",
        "properties": {
            "name": {"type": "string"},
            "version": {"type": "string"}
        },
        "required": ["name", "version"]
    }
    blueprint = client.create_blueprint("service", "Service", schema)
    print(f"✓ Created blueprint: {blueprint['id']}")

    # Create entity
    entity = client.create_entity(
        "service",
        "my-service",
        {"name": "my-service", "version": "1.0.0"}
    )
    print(f"✓ Created entity: {entity['identifier']}")

    # Search
    results = client.search_entities("service", [
        {"property": "version", "operator": "eq", "value": "1.0.0"}
    ])
    print(f"✓ Found {results['total']} entities")
```

---

For more details, see:
- [API Documentation](./API.md) - Complete API reference
- [Development Guide](./DEVELOPMENT.md) - Contributing code
- [Security Documentation](./SECURITY.md) - Authentication details
