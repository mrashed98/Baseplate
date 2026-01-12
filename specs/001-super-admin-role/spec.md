# Feature Specification: Super Admin Role

**Feature Branch**: `001-super-admin-role`
**Created**: 2026-01-12
**Status**: Draft
**Input**: User description: "Create SUPER ADMIN role that the user have this role can manage Everything in the system"

## Clarifications

### Session 2026-01-12

- Q: What should happen when the last super admin attempts to demote themselves? → A: Block it - system prevents the last super admin from removing their own status (per FR-012)
- Q: How does the system prevent privilege escalation where a regular admin attempts to mark themselves as super admin? → A: Authorization check - only users with existing super admin status can promote others (enforced in middleware/service layer)
- Q: When a super admin is also a regular member of a team, which permissions take precedence? → A: Super admin always takes precedence - super admin permissions apply universally regardless of team membership
- Q: What level of detail should be captured in audit logs for super admin actions? → A: Detailed logging - capture timestamp, actor, action type, target entity, IP address, user agent, result status, before/after snapshots for modifications, and full request context; apply to all super admin actions
- Q: How should the system handle a super admin attempting to delete their own account? → A: Conditional block - prevent deletion if last super admin, otherwise allow with confirmation

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Platform Owner Manages All Teams (Priority: P1)

As a platform owner, I need to view and manage all teams in the system so that I can oversee the entire platform and provide support when needed.

**Why this priority**: This is the core capability that distinguishes super admins from regular team admins. Without this, the super admin role has no practical value.

**Independent Test**: Can be fully tested by designating a user as super admin and verifying they can list all teams in the system and access any team's data, delivering immediate platform-wide oversight capability.

**Acceptance Scenarios**:

1. **Given** I am logged in as a super admin, **When** I request to view all teams, **Then** I see a complete list of all teams in the system regardless of my team memberships
2. **Given** I am logged in as a super admin, **When** I select any team from the list, **Then** I can access that team's details, members, and resources
3. **Given** I am logged in as a regular admin, **When** I attempt to view all teams, **Then** I only see teams where I have membership

---

### User Story 2 - Platform Owner Manages All Users (Priority: P1)

As a platform owner, I need to view and manage all users across all teams so that I can handle user issues, security concerns, and account management at the platform level.

**Why this priority**: Essential for platform administration, security management, and user support. Equally critical as team management.

**Independent Test**: Can be fully tested by verifying a super admin can list all users, view any user's details and team memberships, and modify user accounts across the entire platform.

**Acceptance Scenarios**:

1. **Given** I am logged in as a super admin, **When** I request to view all users, **Then** I see a complete list of all registered users regardless of their team memberships
2. **Given** I am logged in as a super admin, **When** I view a user's profile, **Then** I see all teams they belong to and their roles in each team
3. **Given** I am logged in as a super admin, **When** I need to suspend a user account, **Then** I can change their status across all teams
4. **Given** I am logged in as a regular admin, **When** I attempt to view all users, **Then** I only see users within my team

---

### User Story 3 - Super Admin Promotion (Priority: P2)

As a platform owner, I need to designate other users as super admins so that I can delegate platform-wide administrative responsibilities.

**Why this priority**: Important for operational scalability and team growth, but the first super admin can function alone initially.

**Independent Test**: Can be tested by having an existing super admin promote another user to super admin status and verifying the new super admin gains all platform-wide permissions.

**Acceptance Scenarios**:

1. **Given** I am logged in as a super admin, **When** I select a user and promote them to super admin, **Then** that user gains platform-wide access to all teams and users
2. **Given** I am logged in as a super admin, **When** I demote a super admin to regular user, **Then** they lose platform-wide access and only retain their team-specific permissions
3. **Given** I am logged in as a regular admin, **When** I attempt to promote someone to super admin, **Then** the action is denied

---

### User Story 4 - Bypass Team Restrictions (Priority: P2)

As a super admin, I need to perform any action across all teams without being explicitly added as a member so that I can efficiently manage the platform and respond to issues.

**Why this priority**: Operational efficiency feature that prevents the need to manually add super admins to every team, but the core functionality works without it.

**Independent Test**: Can be tested by verifying a super admin (who is not a member of a specific team) can still create, read, update, and delete blueprints, entities, and other resources in that team.

**Acceptance Scenarios**:

1. **Given** I am logged in as a super admin, **When** I access a team I am not a member of, **Then** I can view and modify all team resources (blueprints, entities, integrations, etc.)
2. **Given** I am logged in as a super admin, **When** I create or modify resources in any team, **Then** the action is logged with my super admin identity
3. **Given** team-level permission checks are enforced, **When** a super admin performs an action, **Then** permission checks are bypassed

---

### User Story 5 - Super Admin Audit Trail (Priority: P3)

As a platform owner, I need to see a log of all actions taken by super admins so that I can maintain accountability and investigate security incidents.

**Why this priority**: Important for security and compliance but not required for basic super admin functionality. Can be added after core features are stable.

**Independent Test**: Can be tested by having a super admin perform various actions across different teams and verifying all actions are logged with clear identification of the super admin user.

**Acceptance Scenarios**:

1. **Given** a super admin performs any action in any team, **When** I review the audit log, **Then** I see the action clearly marked as performed by a super admin
2. **Given** I am reviewing system activity, **When** I filter audit logs for super admin actions, **Then** I see all cross-team and administrative actions performed by super admins
3. **Given** a security incident occurs, **When** I investigate the audit trail, **Then** I can distinguish between actions taken by team admins vs super admins

---

### Edge Cases

- **Last super admin demotion**: System blocks the last super admin from removing their own status, preventing platform lockout (requires at least one super admin to exist at all times)
- **Privilege escalation prevention**: Authorization middleware enforces that only existing super admins can promote users to super admin status; regular admins attempting to self-promote are denied at the service layer
- **Permission precedence**: When a super admin is also a regular team member, super admin permissions always take precedence universally (no context switching between roles)
- **Super admin account deletion**: System prevents the last super admin from deleting their own account; non-last super admins can delete their accounts with explicit confirmation
- **Super admin cross-team attribution**: When a super admin creates or modifies resources in a team they're not a member of, the resource's `created_by` and `updated_by` fields reflect the super admin's user ID. Team owners see these resources as created by the super admin user (with their name and email visible), not as "external" or "system" users. Audit logs capture the action with `actor_type = 'super_admin'` to distinguish platform-level actions from team member actions.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a mechanism to designate specific users as super admins with platform-wide privileges
- **FR-002**: System MUST allow super admins to view all teams in the system regardless of their team memberships
- **FR-003**: System MUST allow super admins to view all users in the system regardless of team boundaries
- **FR-004**: System MUST allow super admins to access and modify resources in any team without being a member
- **FR-005**: System MUST bypass team-level permission checks for users with super admin status
- **FR-006**: System MUST allow super admins to promote other users to super admin status
- **FR-007**: System MUST allow super admins to demote other super admins to regular user status
- **FR-008**: System MUST prevent users without super admin status from granting super admin privileges
- **FR-009**: System MUST apply super admin permissions universally for users with super admin status, even when they are also regular team members (super admin permissions always take precedence)
- **FR-010**: System MUST log all actions performed by super admins with detailed audit information including timestamp, actor, action type, target entity, IP address, user agent, result status, before/after snapshots for modifications, and full request context
- **FR-011**: System MUST include super admin status in user authentication tokens or session data for efficient permission checks
- **FR-012**: System MUST prevent the last super admin from removing their own super admin status
- **FR-013**: System MUST create the initial super admin through a database migration during system setup, requiring the admin's email to be specified during deployment
- **FR-014**: System MUST prevent the last super admin from deleting their own account while allowing non-last super admins to delete their accounts with explicit confirmation

### Key Entities

- **Super Admin**: A user designation that grants platform-wide access to all teams, users, and resources. Key attributes include:
  - User identifier
  - Super admin status (boolean flag)
  - Date designated as super admin
  - Designated by (which super admin granted this status)

- **User** (modified): Extended to include super admin designation attribute

- **Audit Log**: Records of all super admin actions with detailed information including:
  - Actor (user performing action)
  - Actor type (super admin vs team admin)
  - Target entity (user, team, resource)
  - Action type (create, read, update, delete, promote, demote)
  - Timestamp
  - Result (success/failure)
  - IP address
  - User agent
  - Before snapshot (state prior to modification)
  - After snapshot (state after modification)
  - Full request context (headers, parameters)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Super admins can access and manage any team's resources without being added as team members (with permission checks completing in <100ms p95 per SC-004)
- **SC-002**: Super admins can view a complete list of all platform users and teams in a single query
- **SC-003**: 100% of super admin actions across teams are logged and identifiable in audit trails
- **SC-004**: Permission checks for super admins complete in under 100ms regardless of team size
- **SC-005**: Platform owners can promote new super admins in under 30 seconds
- **SC-006**: System prevents privilege escalation with 100% success rate (no regular users can grant themselves super admin status)

## Assumptions *(mandatory)*

- Super admin is a platform-level role, not a team-level role
- Super admins can perform any action that team-level admins can perform, across all teams
- Super admins are trusted users (employees, platform owners) not external customers
- The number of super admins will be relatively small (< 50) compared to total users
- Super admin actions should be audited for security and compliance purposes
- Existing authentication mechanism (JWT tokens) can be extended to include super admin status
- Super admin status persists across sessions (not temporary)

## Dependencies *(optional)*

- Existing RBAC (Role-Based Access Control) system with teams, roles, and permissions
- Existing authentication and authorization middleware
- Existing audit logging infrastructure (or needs to be built)
- User management system with ability to update user attributes

## Out of Scope *(optional)*

- Granular super admin permissions (e.g., "super admin for users only" or "super admin for teams only") - this feature provides all-or-nothing platform-wide access
- Time-limited super admin access (temporary elevation) - super admin status is permanent until explicitly revoked
- Super admin role inheritance or delegation to sub-administrators
- Multi-factor authentication specifically for super admin actions (assumes existing MFA applies)
- Super admin-specific UI/dashboard - super admins use existing admin interfaces with expanded scope
