# Baseplate

**Baseplate** is a production-ready headless backend engine that enables dynamic schema management through "Blueprints". Define entity types at runtime without database migrations or code changes - inspired by Port.io's architecture.

Built with **Go (Gin)** and **PostgreSQL (JSONB)**.

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://go.dev)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-336791?style=flat&logo=postgresql)](https://www.postgresql.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## Features

### Core Features
- **Dynamic Schema Management** - Define entity types via JSON Schema at runtime
- **Zero-Migration Architecture** - Add new entity types without database migrations
- **Auto-Validated CRUD APIs** - Instant REST APIs with JSON Schema validation
- **Advanced Search** - Query entities with 9 filter operators on JSONB data
- **Multi-Tenancy** - Team-based isolation with complete data segregation
- **Dual Authentication** - JWT tokens for users, API keys for services
- **Comprehensive RBAC** - 13 permissions across 6 resource types with custom roles

### Technical Features
- **PostgreSQL JSONB** - Flexible schema storage with GIN indexing
- **Clean Architecture** - Handler → Service → Repository pattern
- **Production-Ready** - Security best practices, connection pooling, error handling
- **Well-Documented** - Comprehensive API docs, architecture diagrams, examples
- **RESTful API** - 28 endpoints following REST conventions

## Quick Start

### Prerequisites

- **Go 1.25+** ([install](https://go.dev/dl/))
- **Docker & Docker Compose** ([install](https://www.docker.com/))
- **Make** (optional, included on most systems)

### 5-Minute Setup

```bash
# 1. Clone repository
git clone https://github.com/your-org/baseplate.git
cd baseplate

# 2. Generate JWT secret
export JWT_SECRET=$(openssl rand -base64 32)

# 3. Start database
make db-up

# 4. Run application
make run

# 5. Verify
curl http://localhost:8080/api/health
# Response: {"status":"ok"}
```

Server runs on `http://localhost:8080`

### First Steps

```bash
# Register a user
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alice",
    "email": "alice@example.com",
    "password": "SecurePass123!"
  }'

# Create a team, blueprint, and entities
# See docs/EXAMPLES.md for complete workflows
```

## Documentation

### Core Documentation

| Document | Description |
|----------|-------------|
| **[API Reference](docs/API.md)** | Complete API documentation with all 28 endpoints, examples, and cURL commands |
| **[Architecture Guide](docs/ARCHITECTURE.md)** | System architecture with Mermaid diagrams (ERD, request flow, auth flows) |
| **[Database Guide](docs/DATABASE.md)** | Schema documentation, JSONB patterns, optimization, backup procedures |
| **[Security Guide](docs/SECURITY.md)** | Authentication mechanisms, RBAC, security best practices |
| **[Deployment Guide](docs/DEPLOYMENT.md)** | Production deployment with Docker, systemd, Nginx, SSL setup |
| **[Development Guide](docs/DEVELOPMENT.md)** | Setup, code organization, adding features, testing |
| **[Usage Examples](docs/EXAMPLES.md)** | Real-world examples: service catalog, database inventory, searches |
| **[Contributing Guide](CONTRIBUTING.md)** | Contribution guidelines, code standards, PR process |

## Architecture Overview

### System Design

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ HTTP
       ▼
┌─────────────────────────────────────────┐
│         Gin Router + Middleware         │
│  (Auth, Team Context, Permissions)      │
└──────┬──────────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────────┐
│           Handler Layer                  │
│  (Request/Response Binding)              │
└──────┬──────────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────────┐
│          Service Layer                   │
│  (Business Logic, Validation)            │
└──────┬──────────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────────┐
│        Repository Layer                  │
│  (SQL Queries, Data Mapping)             │
└──────┬──────────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────────┐
│      PostgreSQL 15 (JSONB)               │
└─────────────────────────────────────────┘
```

### Key Components

- **Blueprints**: Define entity schemas using JSON Schema
- **Entities**: Instances of blueprints with validated JSONB data
- **Teams**: Multi-tenant organizations with isolated data
- **Roles**: RBAC with 13 permissions (default: admin, editor, viewer)
- **API Keys**: Service authentication with team-scoped permissions

See [ARCHITECTURE.md](docs/ARCHITECTURE.md) for detailed diagrams and component descriptions.

## API Endpoints

### Authentication
```
POST   /api/auth/register          Register new user
POST   /api/auth/login             Login and get JWT token
GET    /api/auth/me                Get current user info
```

### Teams & RBAC
```
POST   /api/teams                  Create team
GET    /api/teams                  List user's teams
GET    /api/teams/:teamId          Get team details
PUT    /api/teams/:teamId          Update team
DELETE /api/teams/:teamId          Delete team

GET    /api/teams/:teamId/roles    List roles
POST   /api/teams/:teamId/roles    Create custom role

GET    /api/teams/:teamId/members  List members
POST   /api/teams/:teamId/members  Add member
DELETE /api/teams/:teamId/members/:userId  Remove member

GET    /api/teams/:teamId/api-keys Create API key
POST   /api/teams/:teamId/api-keys List API keys
DELETE /api/api-keys/:keyId        Delete API key
```

### Blueprints (Schema Definitions)
```
POST   /api/blueprints             Create blueprint
GET    /api/blueprints             List blueprints
GET    /api/blueprints/:id         Get blueprint
PUT    /api/blueprints/:id         Update blueprint
DELETE /api/blueprints/:id         Delete blueprint
```

### Entities (Dynamic Data)
```
POST   /api/blueprints/:blueprintId/entities                Create entity
GET    /api/blueprints/:blueprintId/entities                List entities
POST   /api/blueprints/:blueprintId/entities/search         Search entities
GET    /api/blueprints/:blueprintId/entities/by-identifier/:identifier  Get by identifier
GET    /api/entities/:id                                    Get entity by ID
PUT    /api/entities/:id                                    Update entity
DELETE /api/entities/:id                                    Delete entity
```

### Health Check
```
GET    /api/health                 Check API health
```

**Total**: 28 endpoints

See [API.md](docs/API.md) for complete documentation with request/response examples.

## Technology Stack

| Category | Technology |
|----------|-----------|
| **Language** | Go 1.25.1 |
| **Web Framework** | Gin v1.11.0 |
| **Database** | PostgreSQL 15 with JSONB |
| **Authentication** | JWT (HS256) + API Keys (SHA-256) |
| **Password Hashing** | bcrypt |
| **JSON Schema** | gojsonschema v1.2.0 |
| **Containerization** | Docker & Docker Compose |

## Project Structure

```
baseplate/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── config/
│   └── config.go                # Configuration management
├── internal/
│   ├── api/
│   │   ├── router.go           # Route setup
│   │   ├── handlers/           # HTTP handlers (auth, team, blueprint, entity)
│   │   └── middleware/         # Authentication, RBAC, error handling
│   ├── core/
│   │   ├── auth/               # Auth domain (models, service, repository)
│   │   ├── blueprint/          # Blueprint domain
│   │   ├── entity/             # Entity domain with search
│   │   └── validation/         # JSON Schema validator
│   └── storage/
│       └── postgres/           # Database connection
├── migrations/
│   └── 001_initial.sql         # Database schema
├── docs/                       # Documentation
│   ├── API.md
│   ├── ARCHITECTURE.md
│   ├── DATABASE.md
│   ├── SECURITY.md
│   ├── DEPLOYMENT.md
│   ├── DEVELOPMENT.md
│   └── EXAMPLES.md
├── .gitignore
├── go.mod
├── go.sum
├── Makefile
├── docker-compose.yaml
├── README.md
└── CONTRIBUTING.md
```

## Development Commands

### Using Makefile

```bash
# Database
make db-up          # Start PostgreSQL container
make db-down        # Stop PostgreSQL container
make db-reset       # Drop and recreate database (⚠️ deletes all data)
make migrate        # Run migrations manually

# Development
make run            # Run server with hot reload
make build          # Build binary to bin/server
make clean          # Remove bin/ directory

# Code Quality
make fmt            # Format code (go fmt)
make lint           # Run linter (requires golangci-lint)
make test           # Run tests
make tidy           # Run go mod tidy

# Documentation
make swagger        # Generate Swagger docs (requires swag)
```

### Manual Commands

```bash
# Build
go build -o bin/server ./cmd/server

# Run
go run ./cmd/server

# Test
go test -v ./...

# Format
go fmt ./...

# Dependencies
go mod download
go mod tidy
```

## Configuration

### Environment Variables

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `JWT_SECRET` | - | **Yes** | JWT signing secret (32+ characters) |
| `SERVER_PORT` | `8080` | No | HTTP server port |
| `GIN_MODE` | `debug` | No | Gin mode (`debug` or `release`) |
| `DB_HOST` | `localhost` | No | PostgreSQL host |
| `DB_PORT` | `5432` | No | PostgreSQL port |
| `DB_USER` | `user` | No | PostgreSQL username |
| `DB_PASSWORD` | `password` | No | PostgreSQL password |
| `DB_NAME` | `baseplate` | No | PostgreSQL database name |
| `DB_SSL_MODE` | `disable` | No | PostgreSQL SSL mode |
| `JWT_EXPIRATION_HOURS` | `24` | No | JWT token lifetime (hours) |

### Configuration File (.env)

Create `.env` file in project root:

```bash
JWT_SECRET=your-secure-secret-minimum-32-characters
SERVER_PORT=8080
GIN_MODE=debug
DB_HOST=localhost
DB_PORT=5432
DB_USER=user
DB_PASSWORD=password
DB_NAME=baseplate
DB_SSL_MODE=disable
JWT_EXPIRATION_HOURS=24
```

**Load environment**:
```bash
export $(cat .env | xargs)
```

**⚠️ Never commit `.env` to version control!**

## Database

**PostgreSQL 15** with JSONB for flexible schema storage.

### Docker Setup

```yaml
# Included docker-compose.yaml
services:
  db:
    image: postgres:15-alpine
    container_name: baseplate_db
    ports:
      - "5432:5432"
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
      POSTGRES_DB: baseplate
```

**Migrations**: Auto-run via Docker init scripts on first startup.

### Key Tables

- `users` - User accounts with bcrypt passwords
- `teams` - Organizations/tenants
- `roles` - RBAC role definitions with JSONB permissions
- `team_memberships` - User-team-role relationships
- `api_keys` - SHA-256 hashed API keys
- **`blueprints`** - JSON Schema definitions (JSONB)
- **`entities`** - Entity instances with validated JSONB data

### JSONB Usage

Entities store flexible data validated against blueprint schemas:

```json
{
  "name": "auth-service",
  "version": "2.3.1",
  "language": "Go",
  "dependencies": ["postgres", "redis"],
  "metadata": {
    "environment": "production"
  }
}
```

**Search with GIN indexes**: Fast queries on nested JSONB properties.

See [DATABASE.md](docs/DATABASE.md) for schema details and optimization.

## Security

### Authentication

- **JWT Tokens**: HS256, 24-hour expiration, for user sessions
- **API Keys**: SHA-256 hashing, team-scoped, optional expiration

### Authorization

- **RBAC**: 13 permissions across 6 resource types
- **Default Roles**: admin, editor, viewer
- **Custom Roles**: Create roles with specific permission combinations
- **Team Isolation**: All queries filtered by team_id

### Best Practices

- Strong JWT secrets (32+ characters)
- HTTPS/TLS in production
- Rate limiting (implement external)
- Regular API key rotation
- Parameterized SQL queries (SQL injection prevention)
- JSON Schema validation (input validation)

See [SECURITY.md](docs/SECURITY.md) for comprehensive security documentation.

## Examples

### Complete Workflow

```bash
# 1. Register user
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com","password":"pass123"}'

# 2. Create team
curl -X POST http://localhost:8080/api/teams \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Acme Corp","slug":"acme"}'

# 3. Create blueprint (service catalog)
curl -X POST http://localhost:8080/api/blueprints \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Team-ID: $TEAM_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "service",
    "title": "Microservice",
    "schema": {
      "type": "object",
      "properties": {
        "name": {"type": "string"},
        "version": {"type": "string"}
      },
      "required": ["name", "version"]
    }
  }'

# 4. Create entity
curl -X POST http://localhost:8080/api/blueprints/service/entities \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Team-ID: $TEAM_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "identifier": "auth-service",
    "data": {"name": "auth-service", "version": "2.3.1"}
  }'

# 5. Search entities
curl -X POST http://localhost:8080/api/blueprints/service/entities/search \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Team-ID: $TEAM_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "filters": [
      {"property": "version", "operator": "gte", "value": "2.0.0"}
    ]
  }'
```

See [EXAMPLES.md](docs/EXAMPLES.md) for complete examples including Python client.

## Deployment

### Production Checklist

- [ ] Set strong `JWT_SECRET` (32+ characters)
- [ ] Set `GIN_MODE=release`
- [ ] Configure `DB_SSL_MODE=require`
- [ ] Enable HTTPS/TLS with valid certificates
- [ ] Set up reverse proxy (Nginx)
- [ ] Configure database backups
- [ ] Set up monitoring and logging
- [ ] Implement rate limiting
- [ ] Use secrets manager for sensitive data

### Docker Deployment

```bash
# Build and run with Docker Compose
docker-compose -f docker-compose.prod.yaml up -d
```

### Systemd Service

```bash
# Install as systemd service
sudo systemctl enable baseplate
sudo systemctl start baseplate
sudo systemctl status baseplate
```

See [DEPLOYMENT.md](docs/DEPLOYMENT.md) for complete deployment guide with Nginx, SSL, and production configuration.

## Contributing

We welcome contributions! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for:

- Code of Conduct
- Development setup
- Coding standards
- Pull request process
- Testing requirements

**Quick Links**:
- [Report a Bug](https://github.com/your-org/baseplate/issues/new?template=bug_report.md)
- [Request a Feature](https://github.com/your-org/baseplate/issues/new?template=feature_request.md)
- [Development Guide](docs/DEVELOPMENT.md)

## Roadmap

### Current (v1.0)
- ✅ Dynamic blueprint and entity management
- ✅ Multi-tenancy with RBAC
- ✅ JWT and API key authentication
- ✅ Advanced search with 9 filter operators
- ✅ JSON Schema validation

### Planned Features
- [ ] Entity relations (blueprint-level and instance-level)
- [ ] Scorecards (quality/compliance metrics)
- [ ] Integrations (external system connectors)
- [ ] Actions (workflow automation)
- [ ] Audit logging (change history)
- [ ] Webhooks (event notifications)
- [ ] GraphQL API
- [ ] OpenAPI/Swagger documentation

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

Inspired by:
- **Port.io** - Internal developer portal architecture
- **n8n** - Workflow automation flexibility

## Support

- **Documentation**: Check [docs/](docs/) directory
- **Issues**: [GitHub Issues](https://github.com/your-org/baseplate/issues)
- **Discussions**: [GitHub Discussions](https://github.com/your-org/baseplate/discussions)

---

**Built with ❤️ using Go and PostgreSQL**

[⭐ Star us on GitHub](https://github.com/your-org/baseplate) if you find this useful!
