# Baseplate

**Baseplate** is a "Headless" Backend Engine that allows defining data entities via **Blueprints** (schemas) rather than hard-coded structs.

Inspired by the internal architecture of tools like **Port.io** and **n8n**, this engine allows users to register new resources (e.g., "Service", "Cluster", "Employee") at runtime and instantly get validated CRUD APIs for them without database migrations or code changes.

Built with **Go (Gin)** and **PostgreSQL (JSONB)**.

## Features

- **Dynamic Blueprint Management**: Create and manage entity schemas at runtime
- **Auto-generated CRUD APIs**: Instantly get validated REST APIs for any blueprint
- **JSON Schema Validation**: Built-in validation using JSON Schema specification
- **Multi-tenancy**: Team-based isolation with JWT authentication
- **PostgreSQL JSONB**: Schema-less storage with powerful querying
- **RESTful API**: Clean HTTP endpoints with middleware support

## Tech Stack

- **Go 1.25+** with Gin web framework
- **PostgreSQL 15** with JSONB support
- **JWT** authentication
- **Docker & Docker Compose** for containerization
- **JSON Schema** validation

## Project Structure

```
Baseplate/
├── cmd/server/          # Application entry point
├── config/              # Configuration management
├── internal/
│   ├── api/            # HTTP layer (router, handlers, middleware)
│   ├── core/           # Business logic (auth, blueprint, entity, validation)
│   └── storage/        # Database clients
├── migrations/         # Database schema migrations
└── docs/              # Documentation
```

## Getting Started

### Prerequisites
- [Go 1.25+](https://go.dev/dl/)
- [Docker & Docker Compose](https://www.docker.com/)
- [Make](https://www.gnu.org/software/make/) (optional)

### Quick Start

1. **Clone the repository:**
   ```bash
   git clone <repo_url>
   cd Baseplate
   ```

2. **Start the database:**
   ```bash
   make db-up
   ```
   Or using Docker Compose directly:
   ```bash
   docker-compose up -d db
   ```

3. **Run the application:**
   ```bash
   make run
   ```
   Or:
   ```bash
   go run ./cmd/server
   ```

4. **Server runs on:** `http://localhost:8080`

## API Endpoints

### Authentication
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - Login user

### Teams
- `POST /api/v1/teams` - Create team
- `GET /api/v1/teams/:id` - Get team details

### Blueprints
- `POST /api/v1/blueprints` - Create blueprint
- `GET /api/v1/blueprints` - List blueprints
- `GET /api/v1/blueprints/:id` - Get blueprint
- `PUT /api/v1/blueprints/:id` - Update blueprint
- `DELETE /api/v1/blueprints/:id` - Delete blueprint

### Entities (Dynamic)
- `POST /api/v1/entities/:blueprint_id` - Create entity
- `GET /api/v1/entities/:blueprint_id` - List entities
- `GET /api/v1/entities/:blueprint_id/:id` - Get entity
- `PUT /api/v1/entities/:blueprint_id/:id` - Update entity
- `DELETE /api/v1/entities/:blueprint_id/:id` - Delete entity

## Development Commands

### Using Makefile
```bash
make build        # Build the application
make run          # Run the application
make test         # Run tests
make clean        # Clean build artifacts
make db-up        # Start database
make db-down      # Stop database
make db-reset     # Reset database (drop and recreate)
make migrate      # Run migrations manually
make fmt          # Format code
make tidy         # Tidy dependencies
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
```

## Configuration

The application uses environment variables for configuration:

```bash
# Server
SERVER_PORT=8080
SERVER_MODE=debug

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=user
DB_PASSWORD=password
DB_NAME=baseplate

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRY=24h
```

## Database

PostgreSQL with JSONB storage:
- **Container name**: `baseplate_db`
- **Port**: 5432
- **Database**: `baseplate`
- **User**: `user`
- **Password**: `password`

Migrations run automatically on first database startup via Docker init scripts.

## License

MIT

## Contributing

Contributions welcome! Please open an issue or submit a PR.
