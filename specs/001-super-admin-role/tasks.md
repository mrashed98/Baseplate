# Tasks: Super Admin Role

**Input**: Design documents from `/specs/001-super-admin-role/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/super-admin-api.yaml

**Tests**: Tests ARE REQUIRED per Constitution (70% service coverage, 60% handler coverage)

**Organization**: Tasks grouped by user story to enable independent implementation and testing

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Terminology

- **Bypass**: Super admins bypass team-level permission checks (do not use "skip", "ignore", or "override")
- **Promote/Demote**: Used for super admin status changes (not "grant/revoke" or "add/remove")
- **Platform-wide**: Refers to cross-team super admin access (not "global" or "system-wide")

## Path Conventions

Repository root structure:
- `internal/core/auth/` - Authentication/authorization logic
- `internal/api/handlers/` - HTTP handlers
- `internal/api/middleware/` - Middleware
- `migrations/` - Database migrations
- `cmd/` - Command-line tools
- `tests/integration/` - Integration tests
- `tests/contract/` - Contract tests

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Database migration and initial super admin setup tooling

- [ ] T001 Create database migration file at migrations/002_super_admin.sql
- [ ] T002 [P] Create initial super admin CLI tool directory at cmd/init-superadmin/
- [ ] T003 [P] Add Makefile target for init-superadmin command

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core data model and JWT infrastructure that MUST be complete before ANY user story

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T004 Extend User model in internal/core/auth/models.go with IsSuperAdmin, SuperAdminPromotedAt, SuperAdminPromotedBy fields
- [ ] T005 [P] Update JWTClaims struct in internal/core/auth/models.go with IsSuperAdmin pointer field
- [ ] T006 [P] Extend AuditLog model in internal/core/auth/models.go with ActorType, IPAddress, UserAgent, ResultStatus, RequestContext fields
- [ ] T007 Update generateToken() in internal/core/auth/service.go to include is_super_admin claim from User model
- [ ] T008 Add graceful degradation logic in ValidateToken() for missing is_super_admin claim (default to false)
- [ ] T009 Add ContextIsSuperAdmin constant in internal/api/middleware/auth.go
- [ ] T010 Update handleJWT() in internal/api/middleware/auth.go to extract and set is_super_admin in context
- [ ] T011 Implement init-superadmin CLI tool in cmd/init-superadmin/main.go (reads SUPER_ADMIN_EMAIL and SUPER_ADMIN_PASSWORD env vars)

**Checkpoint**: Foundation ready - JWT includes super admin claim, database schema supports super admin

---

## Phase 3: User Story 1 - Platform Owner Manages All Teams (Priority: P1) 🎯 MVP

**Goal**: Super admins can view and manage all teams in the system regardless of membership

**Independent Test**: Designate a user as super admin and verify they can list all teams and access any team's data

### Tests for User Story 1 (TDD Approach)

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T012 [P] [US1] Unit test for GetAllTeams() in internal/core/auth/service_test.go
- [ ] T013 [P] [US1] Unit test for RequireTeam() super admin bypass in internal/api/middleware/auth_test.go
- [ ] T014 [P] [US1] Integration test for GET /admin/teams endpoint in tests/integration/super_admin_test.go
- [ ] T015 [P] [US1] Contract test for super admin team endpoints in tests/contract/admin_api_test.go

### Implementation for User Story 1

- [ ] T016 [US1] Add GetAllTeams(ctx) method to auth.Repository in internal/core/auth/repository.go
- [ ] T017 [US1] Add GetAllTeams(ctx) method to auth.Service in internal/core/auth/service.go
- [ ] T018 [US1] Update RequireTeam() middleware in internal/api/middleware/auth.go to bypass membership check for super admins
- [ ] T019 [US1] Create AdminHandler struct in internal/api/handlers/admin.go
- [ ] T020 [US1] Implement ListTeams() handler in internal/api/handlers/admin.go
- [ ] T021 [US1] Implement GetTeamDetail() handler in internal/api/handlers/admin.go
- [ ] T022 [US1] Add RequireSuperAdmin() middleware in internal/api/middleware/auth.go
- [ ] T023 [US1] Add /admin/teams routes in internal/api/router.go with RequireSuperAdmin middleware

**Checkpoint**: Super admins can view all teams, access any team's details; regular admins only see their teams

---

## Phase 4: User Story 2 - Platform Owner Manages All Users (Priority: P1)

**Goal**: Super admins can view and manage all users across all teams

**Independent Test**: Verify super admin can list all users, view user details with team memberships, and modify user accounts

### Tests for User Story 2 (TDD Approach)

- [ ] T024 [P] [US2] Unit test for GetAllUsers() in internal/core/auth/service_test.go
- [ ] T025 [P] [US2] Unit test for GetUserDetail() in internal/core/auth/service_test.go
- [ ] T026 [P] [US2] Integration test for GET /admin/users endpoint in tests/integration/super_admin_test.go
- [ ] T027 [P] [US2] Contract test for super admin user endpoints in tests/contract/admin_api_test.go

### Implementation for User Story 2

- [ ] T028 [P] [US2] Add GetAllUsers(ctx, limit, offset, filters) to auth.Repository in internal/core/auth/repository.go
- [ ] T029 [P] [US2] Add GetUserWithMemberships(ctx, userID) to auth.Repository in internal/core/auth/repository.go
- [ ] T030 [US2] Add GetAllUsers() method to auth.Service in internal/core/auth/service.go
- [ ] T031 [US2] Add GetUserDetail() method to auth.Service in internal/core/auth/service.go
- [ ] T032 [US2] Implement ListUsers() handler in internal/api/handlers/admin.go with pagination
- [ ] T033 [US2] Implement GetUserDetail() handler in internal/api/handlers/admin.go
- [ ] T034 [US2] Implement UpdateUser() handler in internal/api/handlers/admin.go
- [ ] T035 [US2] Add /admin/users routes in internal/api/router.go

**Checkpoint**: Super admins can view/manage all users; regular admins only see their team users

---

## Phase 5: User Story 3 - Super Admin Promotion (Priority: P2)

**Goal**: Super admins can promote users to super admin and demote super admins (with last admin protection)

**Independent Test**: Super admin promotes a user, verify new super admin gains platform access; attempt to demote last super admin and verify it's blocked

### Tests for User Story 3 (TDD Approach)

- [ ] T036 [P] [US3] Unit test for PromoteToSuperAdmin() in internal/core/auth/service_test.go
- [ ] T037 [P] [US3] Unit test for DemoteFromSuperAdmin() with last admin protection in internal/core/auth/service_test.go
- [ ] T038 [P] [US3] Unit test for CountSuperAdminsForUpdate() in internal/core/auth/repository_test.go
- [ ] T039 [P] [US3] Integration test for POST /admin/users/:id/promote endpoint in tests/integration/super_admin_test.go
- [ ] T040 [P] [US3] Integration test for POST /admin/users/:id/demote with last admin edge case in tests/integration/super_admin_test.go

### Implementation for User Story 3

- [ ] T041 [P] [US3] Add CountSuperAdminsForUpdate(ctx, tx) to auth.Repository in internal/core/auth/repository.go (uses SELECT FOR UPDATE)
- [ ] T042 [P] [US3] Add UpdateUserSuperAdminStatus(ctx, tx, userID, isSuperAdmin, promotedBy) to auth.Repository
- [ ] T043 [US3] Implement PromoteToSuperAdmin(ctx, actorID, targetUserID) in internal/core/auth/service.go
- [ ] T044 [US3] Implement DemoteFromSuperAdmin(ctx, actorID, targetUserID) in internal/core/auth/service.go with transaction and last admin check
- [ ] T045 [US3] Add transaction helper BeginTx() to postgres.Client in internal/storage/postgres/client.go
- [ ] T046 [US3] Implement PromoteUser() handler in internal/api/handlers/admin.go
- [ ] T047 [US3] Implement DemoteUser() handler in internal/api/handlers/admin.go
- [ ] T048 [US3] Add /admin/users/:userId/promote and /demote routes in internal/api/router.go

**Checkpoint**: Super admin promotion/demotion works; last super admin protection prevents lockout

---

## Phase 6: User Story 4 - Bypass Team Restrictions (Priority: P2)

**Goal**: Super admins can access and modify resources in any team without being a member

**Independent Test**: Super admin (not a team member) accesses team and performs CRUD operations on blueprints/entities

### Tests for User Story 4 (TDD Approach)

- [ ] T049 [P] [US4] Unit test for RequirePermission() super admin bypass in internal/api/middleware/auth_test.go
- [ ] T050 [P] [US4] Integration test for super admin accessing non-member team resources in tests/integration/super_admin_test.go
- [ ] T051 [P] [US4] Integration test for super admin creating/modifying blueprints in non-member team in tests/integration/super_admin_test.go

### Implementation for User Story 4

- [ ] T052 [US4] Update RequireTeam() in internal/api/middleware/auth.go to set AllPermissions for super admins
- [ ] T053 [US4] Update RequirePermission() in internal/api/middleware/auth.go to bypass permission check for super admins
- [ ] T054 [US4] Add unit tests verifying existing blueprint/entity endpoints work for super admins in non-member teams

**Checkpoint**: Super admins can perform any action in any team without membership

---

## Phase 7: User Story 5 - Super Admin Audit Trail (Priority: P3)

**Goal**: All super admin actions are logged with detailed context (IP, user agent, before/after snapshots)

**Independent Test**: Super admin performs actions across teams; verify audit logs capture all details including actor_type, IP, request context

### Tests for User Story 5 (TDD Approach)

- [ ] T055 [P] [US5] Unit test for CreateAuditLog() with super admin actor type in internal/core/auth/repository_test.go
- [ ] T056 [P] [US5] Integration test for audit log writes on super admin actions in tests/integration/audit_log_test.go
- [ ] T057 [P] [US5] Integration test for GET /admin/audit-logs with filters in tests/integration/super_admin_test.go

### Implementation for User Story 5

- [ ] T058 [P] [US5] Add CreateAuditLog(ctx, log) to auth.Repository in internal/core/auth/repository.go
- [ ] T059 [P] [US5] Add GetAuditLogs(ctx, filters, limit, offset) to auth.Repository in internal/core/auth/repository.go
- [ ] T060 [US5] Add audit log helper function LogSuperAdminAction() in internal/core/auth/service.go
- [ ] T061 [US5] Update all super admin handlers (ListUsers, ListTeams, PromoteUser, DemoteUser) to call LogSuperAdminAction()
- [ ] T062 [US5] Add middleware to extract IP address and user agent in internal/api/middleware/audit.go
- [ ] T063 [US5] Implement QueryAuditLogs() handler in internal/api/handlers/admin.go
- [ ] T064 [US5] Add GET /admin/audit-logs route in internal/api/router.go

**Checkpoint**: All super admin actions logged with full context; audit trail queryable via API

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Edge cases, documentation, and final testing

- [ ] T065 [P] Add unit tests for last super admin account deletion prevention in internal/core/auth/service_test.go
- [ ] T066 [P] Add integration test for permission precedence (super admin also team member) in tests/integration/super_admin_test.go
- [ ] T067 [P] Add error handling tests for privilege escalation attempts in tests/integration/super_admin_test.go
- [ ] T068 [P] Update docs/super-admin.md with setup and usage instructions
- [ ] T069 Run quickstart.md validation scenarios locally
- [ ] T070 Run make test to verify 70% service coverage, 60% handler coverage
- [ ] T071 [P] Add database indexes validation (idx_users_super_admin, idx_audit_logs_actor_type)
- [ ] T072 Performance test: measure permission check latency (<100ms p95)
- [ ] T073 Performance test: measure last admin check latency with SELECT FOR UPDATE (<20ms)
- [ ] T074 Security review: verify JWT signature validation, privilege escalation prevention
- [ ] T075 [P] Update CLAUDE.md with super admin implementation patterns

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-7)**: All depend on Foundational phase completion
  - User stories CAN proceed in parallel (if staffed)
  - Or sequentially in priority order: US1 (P1) → US2 (P1) → US3 (P2) → US4 (P2) → US5 (P3)
- **Polish (Phase 8)**: Depends on all desired user stories being complete

### User Story Dependencies

- **US1 (P1) - Manage Teams**: Depends only on Foundational - No dependencies on other stories
- **US2 (P1) - Manage Users**: Depends only on Foundational - Independent from US1
- **US3 (P2) - Promotion**: Depends on Foundational + requires JWT with super admin claim (T007-T008)
- **US4 (P2) - Bypass Restrictions**: Depends on Foundational + US1 (RequireTeam bypass) + US2 (RequirePermission bypass)
- **US5 (P3) - Audit Trail**: Can integrate with any completed story; should add logging to US1-US4 handlers

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Models/Repository before Service
- Service before Handlers
- Handlers before Routes
- Core implementation before integration

### Parallel Opportunities

**Foundational Phase (Phase 2):**
- T005 (JWTClaims), T006 (AuditLog) can run in parallel with T004 (User model)
- T009 (middleware constant), T011 (CLI tool) can run in parallel with auth model updates

**User Story 1 (Phase 3):**
- All 4 test tasks (T012-T015) can run in parallel
- T016 (repository), T022 (middleware) can run in parallel
- T019 (AdminHandler struct) + T020/T021 (handlers) sequential

**User Story 2 (Phase 4):**
- All 4 test tasks (T024-T027) can run in parallel
- T028 (GetAllUsers repo), T029 (GetUserWithMemberships repo) can run in parallel
- Handler implementations (T032-T034) can run in parallel after service layer

**User Story 3 (Phase 5):**
- All 5 test tasks (T036-T040) can run in parallel
- T041 (count query), T042 (update query) can run in parallel
- T043 (promote), T044 (demote) sequential after repository layer

**User Story 4 (Phase 6):**
- All 3 test tasks (T049-T051) can run in parallel
- T052 (RequireTeam update), T053 (RequirePermission update) sequential (same file)

**User Story 5 (Phase 7):**
- All 3 test tasks (T055-T057) can run in parallel
- T058 (CreateAuditLog repo), T059 (GetAuditLogs repo) can run in parallel
- T062 (audit middleware) can run in parallel with repository layer

**Polish Phase (Phase 8):**
- All test tasks (T065-T067), docs (T068), indexes (T071), and CLAUDE.md (T075) can run in parallel
- Performance tests (T072-T073) sequential after implementation complete

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together (TDD approach):
Task T012: "Unit test for GetAllTeams() in internal/core/auth/service_test.go"
Task T013: "Unit test for RequireTeam() super admin bypass in internal/api/middleware/auth_test.go"
Task T014: "Integration test for GET /admin/teams endpoint in tests/integration/super_admin_test.go"
Task T015: "Contract test for super admin team endpoints in tests/contract/admin_api_test.go"

# After tests written and failing, launch repository implementations:
Task T016: "Add GetAllTeams(ctx) method to auth.Repository"
Task T022: "Add RequireSuperAdmin() middleware"

# Handlers can proceed in parallel after service layer:
Task T020: "Implement ListTeams() handler"
Task T021: "Implement GetTeamDetail() handler"
```

---

## Parallel Example: User Story 3

```bash
# Launch all tests for User Story 3 together:
Task T036: "Unit test for PromoteToSuperAdmin()"
Task T037: "Unit test for DemoteFromSuperAdmin() with last admin protection"
Task T038: "Unit test for CountSuperAdminsForUpdate()"
Task T039: "Integration test for POST /admin/users/:id/promote"
Task T040: "Integration test for POST /admin/users/:id/demote with last admin edge case"

# Repository methods in parallel:
Task T041: "Add CountSuperAdminsForUpdate(ctx, tx) with SELECT FOR UPDATE"
Task T042: "Add UpdateUserSuperAdminStatus(ctx, tx, userID, ...)"
Task T045: "Add transaction helper BeginTx()"

# Service methods sequential (depend on repository):
Task T043: "Implement PromoteToSuperAdmin() service method"
Task T044: "Implement DemoteFromSuperAdmin() with transaction and last admin check"

# Handlers in parallel:
Task T046: "Implement PromoteUser() handler"
Task T047: "Implement DemoteUser() handler"
```

---

## Implementation Strategy

### MVP First (User Stories 1 & 2 Only - Both P1)

1. Complete Phase 1: Setup (T001-T003)
2. Complete Phase 2: Foundational (T004-T011) - CRITICAL checkpoint
3. Complete Phase 3: User Story 1 (T012-T023)
4. **VALIDATE US1**: Test super admin can view/manage all teams
5. Complete Phase 4: User Story 2 (T024-T035)
6. **VALIDATE US2**: Test super admin can view/manage all users
7. **STOP and VALIDATE MVP**: Both P1 stories functional, platform oversight enabled

**MVP Delivers**: Super admin can view and manage all teams and users - core platform oversight capability

### Incremental Delivery

1. **Foundation** (Phase 1-2) → JWT with super admin claim, database schema ready
2. **MVP** (Phase 3-4, US1+US2) → Platform-wide team & user management → Deploy/Demo
3. **Add Promotion** (Phase 5, US3) → Super admin delegation capability → Deploy/Demo
4. **Add Bypass** (Phase 6, US4) → Cross-team resource access → Deploy/Demo
5. **Add Audit** (Phase 7, US5) → Compliance and security logging → Deploy/Demo
6. **Polish** (Phase 8) → Production-ready with documentation

Each phase adds value without breaking previous functionality.

### Parallel Team Strategy

With 3 developers after Foundational phase:

1. **Team completes Phase 1-2 together** (foundation)
2. **Once Foundational done:**
   - Developer A: User Story 1 (T012-T023)
   - Developer B: User Story 2 (T024-T035)
   - Developer C: User Story 3 (T036-T048) - starts tests/infrastructure
3. **After US1 + US2 validate:**
   - Developer A: User Story 4 (T049-T054)
   - Developer B: User Story 5 (T055-T064)
   - Developer C: Polish tasks (T065-T075)

Stories complete and integrate independently.

---

## Notes

- **[P] tasks** = different files or independent components, can run in parallel
- **[Story] label** = maps task to specific user story for traceability
- **TDD required**: Write tests first (per Constitution testing standards)
- **Target coverage**: 70% services, 60% handlers
- **Verify tests fail** before implementing
- **Commit** after each task or logical group
- **Stop at checkpoints** to validate story independently
- **Performance validation**: <100ms p95 for permission checks (T072-T073)
- **Security validation**: Verify JWT signature, privilege escalation prevention (T074)

---

## Task Count Summary

- **Total Tasks**: 75
- **Phase 1 (Setup)**: 3 tasks
- **Phase 2 (Foundational)**: 8 tasks - BLOCKS all stories
- **Phase 3 (US1 - Manage Teams)**: 12 tasks (4 tests + 8 implementation)
- **Phase 4 (US2 - Manage Users)**: 12 tasks (4 tests + 8 implementation)
- **Phase 5 (US3 - Promotion)**: 13 tasks (5 tests + 8 implementation)
- **Phase 6 (US4 - Bypass)**: 6 tasks (3 tests + 3 implementation)
- **Phase 7 (US5 - Audit)**: 10 tasks (3 tests + 7 implementation)
- **Phase 8 (Polish)**: 11 tasks (testing, documentation, validation)

**Parallel Opportunities**: 42 tasks marked [P] (56% can run in parallel within constraints)

**MVP Scope**: Phase 1-2 (foundation) + Phase 3-4 (US1+US2) = 35 tasks → Platform-wide oversight capability
