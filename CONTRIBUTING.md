# Contributing to Baseplate

Thank you for your interest in contributing to Baseplate! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [How to Contribute](#how-to-contribute)
- [Development Guidelines](#development-guidelines)
- [Pull Request Process](#pull-request-process)
- [Coding Standards](#coding-standards)
- [Testing Requirements](#testing-requirements)
- [Documentation](#documentation)
- [Community](#community)

## Code of Conduct

### Our Pledge

We pledge to make participation in Baseplate a harassment-free experience for everyone, regardless of age, body size, disability, ethnicity, gender identity and expression, level of experience, nationality, personal appearance, race, religion, or sexual identity and orientation.

### Our Standards

**Positive behavior includes**:
- Using welcoming and inclusive language
- Being respectful of differing viewpoints
- Gracefully accepting constructive criticism
- Focusing on what is best for the community
- Showing empathy towards other community members

**Unacceptable behavior includes**:
- Trolling, insulting/derogatory comments, and personal or political attacks
- Public or private harassment
- Publishing others' private information without permission
- Other conduct which could reasonably be considered inappropriate

### Enforcement

Project maintainers have the right and responsibility to remove, edit, or reject comments, commits, code, issues, and other contributions that are not aligned with this Code of Conduct.

## Getting Started

### Prerequisites

Before contributing, ensure you have:
- **Go 1.25.1+** installed
- **Docker and Docker Compose** for local development
- **Git** for version control
- **Make** for build automation
- Familiarity with Go, PostgreSQL, and REST APIs

### Setup Development Environment

1. **Fork the repository** on GitHub

2. **Clone your fork**:
```bash
git clone https://github.com/YOUR_USERNAME/baseplate.git
cd baseplate
```

3. **Add upstream remote**:
```bash
git remote add upstream https://github.com/original-org/baseplate.git
```

4. **Install dependencies**:
```bash
go mod download
```

5. **Start database**:
```bash
make db-up
```

6. **Set environment variables**:
```bash
export JWT_SECRET=$(openssl rand -base64 32)
```

7. **Run the application**:
```bash
make run
```

8. **Verify setup**:
```bash
curl http://localhost:8080/api/health
# Response: {"status":"ok"}
```

See [DEVELOPMENT.md](./docs/DEVELOPMENT.md) for detailed setup instructions.

## How to Contribute

### Reporting Bugs

**Before submitting a bug report**:
- Check existing issues to avoid duplicates
- Collect information about the bug
- Try to reproduce the issue in a clean environment

**When submitting a bug report, include**:
- **Title**: Clear and descriptive summary
- **Description**: Detailed explanation of the issue
- **Steps to reproduce**: Numbered list of steps
- **Expected behavior**: What you expected to happen
- **Actual behavior**: What actually happened
- **Environment**: OS, Go version, Docker version
- **Logs**: Relevant error messages or logs
- **Screenshots**: If applicable

**Bug report template**:
```markdown
### Description
Brief description of the bug

### Steps to Reproduce
1. Start the application
2. Call POST /api/auth/register
3. Observe error

### Expected Behavior
User should be registered successfully

### Actual Behavior
Receives 500 internal server error

### Environment
- OS: macOS 14.1
- Go: 1.25.1
- Docker: 24.0.6

### Logs
```
[error] database connection failed: ...
```
```

### Suggesting Enhancements

**Before suggesting an enhancement**:
- Check if the feature already exists
- Review existing feature requests
- Consider if it aligns with project goals

**When suggesting enhancements, include**:
- **Title**: Clear feature name
- **Problem statement**: What problem does this solve?
- **Proposed solution**: How should it work?
- **Alternatives considered**: Other approaches you've thought about
- **Use cases**: Real-world scenarios
- **Priority**: Low, medium, or high

### Contributing Code

**Types of contributions we welcome**:
- Bug fixes
- New features (discuss in an issue first)
- Performance improvements
- Documentation improvements
- Test coverage improvements
- Code refactoring

**Before starting work**:
1. Check if an issue exists for your contribution
2. If not, create an issue to discuss your idea
3. Wait for maintainer feedback before starting work
4. Assign yourself to the issue

## Development Guidelines

### Branch Strategy

**Branch naming**:
```
feature/add-webhooks
fix/entity-validation-bug
docs/update-api-reference
refactor/simplify-auth-flow
test/add-integration-tests
```

**Branch workflow**:
```bash
# Create feature branch from main
git checkout main
git pull upstream main
git checkout -b feature/your-feature

# Make changes and commit
git add .
git commit -m "feat: add your feature"

# Keep your branch updated
git fetch upstream
git rebase upstream/main

# Push to your fork
git push origin feature/your-feature
```

### Commit Message Guidelines

Follow **Conventional Commits** specification:

**Format**:
```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types**:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code formatting (no functional changes)
- `refactor`: Code restructuring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**Examples**:
```bash
# Good commits
git commit -m "feat(auth): add API key expiration support"
git commit -m "fix(entity): correct JSON schema validation error messages"
git commit -m "docs: update deployment guide with SSL configuration"
git commit -m "test(blueprint): add integration tests for CRUD operations"

# Bad commits
git commit -m "fixed stuff"
git commit -m "wip"
git commit -m "updates"
```

**Detailed commit**:
```
feat(search): add support for nested property filters

Add ability to filter entities using dot notation for nested
JSONB properties (e.g., "metadata.environment").

- Add property path parsing in search filter
- Update SQL query generation for nested paths
- Add validation for property path format
- Add tests for nested property search

Closes #123
```

### Code Organization

Follow the project's layered architecture:

```
Handler → Service → Repository → Database
```

**Example**:
```go
// Handler (internal/api/handlers/entity.go)
func (h *EntityHandler) Create(c *gin.Context) {
    var req CreateEntityRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    entity, err := h.service.Create(c.Request.Context(), &req)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(201, entity)
}

// Service (internal/core/entity/service.go)
func (s *EntityService) Create(ctx context.Context, req *CreateEntityRequest) (*Entity, error) {
    // 1. Validate
    if err := s.validator.Validate(req.Data, blueprint.Schema); err != nil {
        return nil, err
    }

    // 2. Business logic
    entity := &Entity{
        ID: uuid.New(),
        // ...
    }

    // 3. Persist
    return s.repo.Create(ctx, entity)
}

// Repository (internal/core/entity/repository.go)
func (r *EntityRepository) Create(ctx context.Context, entity *Entity) error {
    query := `
        INSERT INTO entities (id, team_id, blueprint_id, identifier, title, data)
        VALUES ($1, $2, $3, $4, $5, $6)
    `
    _, err := r.db.ExecContext(ctx, query, entity.ID, entity.TeamID, ...)
    return err
}
```

## Pull Request Process

### Before Submitting

1. **Update your branch**:
```bash
git fetch upstream
git rebase upstream/main
```

2. **Run tests**:
```bash
make test
```

3. **Format code**:
```bash
make fmt
```

4. **Run linter** (if available):
```bash
make lint
```

5. **Update documentation** if needed

6. **Test manually** with curl or Postman

### Submitting Pull Request

1. **Push to your fork**:
```bash
git push origin feature/your-feature
```

2. **Create pull request** on GitHub

3. **Fill out PR template**:
```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Changes Made
- Added X functionality
- Fixed Y bug
- Updated Z documentation

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing completed
- [ ] All tests passing

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Comments added for complex logic
- [ ] Documentation updated
- [ ] No new warnings generated
- [ ] Tests added for new features
- [ ] All tests pass

## Related Issues
Closes #123
Relates to #456

## Screenshots (if applicable)
```

4. **Wait for review**

### Code Review Process

**What reviewers check**:
- Code quality and style
- Test coverage
- Documentation completeness
- Breaking changes
- Performance implications
- Security concerns

**Responding to feedback**:
- Address all comments
- Explain your reasoning if you disagree
- Make requested changes
- Push updates to your branch
- Re-request review when ready

**After approval**:
- Maintainers will merge your PR
- Your contribution will be credited
- Celebrate! 🎉

## Coding Standards

### Go Style

Follow [Effective Go](https://go.dev/doc/effective_go) and [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).

**Key principles**:

**Naming**:
```go
// Good
type UserService struct {}
var errNotFound = errors.New("not found")
const maxRetries = 3

// Bad
type UserServiceImpl struct {}
var ErrNotFound = errors.New("not found") // Don't export if not needed
const MAX_RETRIES = 3 // Use camelCase, not SCREAMING_SNAKE
```

**Error handling**:
```go
// Good
result, err := doSomething()
if err != nil {
    return nil, fmt.Errorf("failed to do something: %w", err)
}

// Bad
result, _ := doSomething() // Never ignore errors
```

**Comments**:
```go
// Good
// GetByID retrieves a user by their unique identifier.
// Returns ErrNotFound if the user doesn't exist.
func GetByID(id uuid.UUID) (*User, error) {
    // ...
}

// Bad
// get user
func GetByID(id uuid.UUID) (*User, error) {
    // ...
}
```

**Package structure**:
- Keep packages small and focused
- Avoid circular dependencies
- Use internal/ for private packages

### SQL

**Use parameterized queries**:
```go
// Good
query := "SELECT * FROM users WHERE email = $1"
row := db.QueryRow(query, email)

// Bad - SQL injection vulnerability
query := fmt.Sprintf("SELECT * FROM users WHERE email = '%s'", email)
```

**Format queries**:
```sql
-- Good
query := `
    SELECT id, name, email, created_at
    FROM users
    WHERE team_id = $1
      AND status = $2
    ORDER BY created_at DESC
    LIMIT $3
`

-- Bad
query := "SELECT id,name,email,created_at FROM users WHERE team_id=$1 AND status=$2 ORDER BY created_at DESC LIMIT $3"
```

## Testing Requirements

### Test Coverage

**Minimum requirements**:
- Unit tests for business logic: 80%+
- Integration tests for critical paths
- API endpoint tests

### Writing Tests

**Test file naming**: `*_test.go`

**Test naming**: `TestFunctionName_Scenario_ExpectedResult`

**Example**:
```go
func TestCreateEntity_ValidData_Success(t *testing.T) {
    // Arrange
    db := setupTestDB(t)
    defer db.Close()

    service := entity.NewEntityService(repo, validator)
    req := &entity.CreateEntityRequest{
        BlueprintID: "service",
        Identifier:  "test-service",
        Data: map[string]interface{}{
            "name": "test-service",
            "version": "1.0.0",
        },
    }

    // Act
    result, err := service.Create(context.Background(), req)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, "test-service", result.Identifier)
}

func TestCreateEntity_InvalidSchema_ReturnsError(t *testing.T) {
    // Test validation failure
}
```

**Run tests**:
```bash
# All tests
make test

# Specific package
go test -v ./internal/core/entity/...

# With coverage
go test -cover ./...
```

## Documentation

### Code Documentation

**Comment exported functions**:
```go
// CreateEntity creates a new entity instance and validates it against
// the blueprint's JSON Schema. Returns ErrInvalidData if validation fails.
func CreateEntity(ctx context.Context, req *CreateEntityRequest) (*Entity, error) {
    // ...
}
```

### Documentation Files

**Update relevant docs**:
- `docs/API.md` - API changes
- `docs/ARCHITECTURE.md` - Architecture changes
- `docs/DEVELOPMENT.md` - Development process changes
- `docs/DEPLOYMENT.md` - Deployment changes
- `README.md` - Major feature additions

**Documentation style**:
- Clear and concise
- Include code examples
- Use proper markdown formatting
- Add table of contents for long documents

## Community

### Communication Channels

- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: General questions and ideas
- **Pull Requests**: Code contributions

### Getting Help

**Before asking for help**:
1. Check documentation
2. Search existing issues
3. Review closed issues

**When asking for help**:
- Be specific and clear
- Provide context
- Include relevant code/logs
- Show what you've tried

### Recognition

Contributors are recognized in:
- Git commit history
- Release notes
- Contributors list (if maintained)

---

## Thank You!

Your contributions make Baseplate better for everyone. We appreciate your time and effort!

**Questions?**
- Open an issue with the `question` label
- Start a discussion on GitHub Discussions

**Ready to contribute?**
- Check out [good first issues](https://github.com/your-org/baseplate/labels/good%20first%20issue)
- Read the [Development Guide](./docs/DEVELOPMENT.md)
- Join the community!
