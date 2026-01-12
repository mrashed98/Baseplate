<!--
Sync Impact Report:
Version: 0.0.0 → 1.0.0
Change Type: MAJOR (Initial ratification)
Modified Principles: All new
Added Sections:
  - Core Principles (5 principles)
  - Performance Standards
  - Development Workflow
  - Governance
Templates Requiring Updates:
  ✅ plan-template.md - Constitution Check section ready
  ✅ spec-template.md - Requirements and success criteria align
  ✅ tasks-template.md - Task organization supports test-first workflow
Follow-up TODOs: None
-->

# Baseplate Constitution

## Core Principles

### I. Code Quality & Maintainability

All code MUST be clean, readable, and maintainable. This includes:
- Clear, descriptive naming for variables, functions, and types
- Single Responsibility Principle: each function/struct does one thing well
- Proper error handling at all boundaries (HTTP, database, external APIs)
- No commented-out code in production branches
- Go idioms followed consistently (e.g., errors returned not thrown, interfaces small and focused)

**Rationale**: Baseplate is a dynamic schema engine where complexity is inherent. Code clarity is non-negotiable to prevent technical debt from becoming unmanageable as blueprints and entities proliferate.

### II. Testing Standards

Testing MUST follow this hierarchy:
- **Unit tests**: Required for all business logic in `internal/core/*/service.go`
- **Integration tests**: Required for database interactions in repositories and critical API flows
- **Contract tests**: Required for JSON Schema validation and API endpoint contracts
- Test coverage minimum: 70% for services, 60% for handlers
- Tests MUST be runnable via `make test` and pass before any merge to main
- Tests MUST run in isolation without side effects

**Rationale**: Dynamic schema validation and JSONB storage make runtime behavior unpredictable. Comprehensive testing catches schema violations and data corruption before they reach production.

### III. User Experience Consistency

All API responses MUST follow consistent patterns:
- Success responses: `200 OK` with data payload
- Created resources: `201 Created` with resource location
- Validation errors: `400 Bad Request` with structured error details
- Auth failures: `401 Unauthorized` or `403 Forbidden` appropriately
- Not found: `404 Not Found`
- Server errors: `500 Internal Server Error` with safe error messages (no stack traces exposed)
- Error response format: `{"error": {"code": "string", "message": "string", "details": object}}`
- All timestamps in ISO 8601 format
- Pagination via `limit`/`offset` query params with `total` in response

**Rationale**: Baseplate is a headless engine consumed by other systems. Predictable API contracts reduce integration friction and prevent cascading failures in dependent services.

### IV. Performance Requirements

The system MUST meet these performance benchmarks:
- Blueprint CRUD operations: <50ms p95 latency
- Entity CRUD operations: <100ms p95 latency
- Entity search with filters: <200ms p95 for result sets up to 1000 items
- JSON Schema validation: <10ms per entity
- Database connection pool: minimum 10 connections, maximum 100
- Memory usage: <512MB under normal load (10k entities)
- API endpoints MUST handle 1000 concurrent requests without degradation

Performance violations MUST be justified or resolved before merge.

**Rationale**: Dynamic schema evaluation adds computational overhead. Without strict performance gates, JSONB queries and repeated validation can degrade to unusable response times.

### V. Security & Data Integrity

Security MUST be enforced at every layer:
- JWT validation on all authenticated endpoints
- API key validation on all team-scoped endpoints
- Team isolation: queries MUST filter by `team_id` at the repository layer
- SQL injection prevention: use parameterized queries exclusively
- JSON Schema validation MUST occur before any database write
- Secrets (JWT_SECRET, DB_PASSWORD) MUST be environment variables, never hardcoded
- Role-based access control (RBAC) enforced via middleware before handler execution
- Blueprint deletion MUST check for dependent entities

**Rationale**: Multi-tenancy and dynamic schemas create attack vectors for data leakage and injection. Defense in depth is mandatory.

## Performance Standards

**Benchmarking**: All performance claims MUST be validated via:
- `go test -bench` for critical paths
- Load testing with `wrk` or `k6` for API endpoints before release

**Database Optimization**:
- JSONB columns MUST have GIN indexes for queried fields
- Queries MUST use `EXPLAIN ANALYZE` to verify index usage
- N+1 query patterns MUST be avoided (use joins or batch fetching)

**Profiling**: If a feature degrades performance by >10%, profile with `pprof` and optimize before merge.

## Development Workflow

**Branch Strategy**: Feature branches off `main`, named `feat/<feature-name>`.

**Commit Standards**:
- Atomic commits: one logical change per commit
- Commit messages: `<type>: <description>` (e.g., `feat: add blueprint versioning`, `fix: prevent duplicate entities`)
- Types: `feat`, `fix`, `refactor`, `test`, `docs`, `chore`

**Code Review Requirements**:
- All PRs MUST pass CI (tests + linting)
- At least one approving review required
- Reviewers MUST verify Constitution compliance (especially testing and performance gates)
- Breaking changes MUST include migration plan

**Deployment Gates**:
- `make test` passes
- `make build` succeeds
- No new linting errors (`make fmt` clean)
- Performance benchmarks within acceptable range

## Governance

**Authority**: This Constitution supersedes all other development practices. In conflicts, Constitution rules apply.

**Amendments**:
- Amendments require documentation of rationale in Sync Impact Report
- Breaking changes require MAJOR version increment
- New principles or sections require MINOR version increment
- Clarifications or wording fixes require PATCH version increment

**Compliance Review**:
- All PRs MUST verify compliance with relevant principles
- Violations MUST be justified in PR description or rejected
- Quarterly review of Constitution effectiveness by team

**Versioning Policy**: Constitution follows semantic versioning (MAJOR.MINOR.PATCH). Current version tracked at bottom of this document.

**Enforcement**: CI pipeline MUST enforce testable rules (test coverage, linting, build success). Human review enforces subjective rules (code clarity, security best practices).

**Guidance File**: See `CLAUDE.md` for runtime development guidance and common patterns.

---

**Version**: 1.0.0 | **Ratified**: 2026-01-12 | **Last Amended**: 2026-01-12
