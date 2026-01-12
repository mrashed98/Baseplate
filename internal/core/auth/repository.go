package auth

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"

	"github.com/baseplate/baseplate/internal/storage/postgres"
)

type Repository struct {
	db *postgres.Client
}

func NewRepository(db *postgres.Client) *Repository {
	return &Repository{db: db}
}

// User methods
func (r *Repository) CreateUser(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (id, email, password_hash, name, status, is_super_admin, super_admin_promoted_at, super_admin_promoted_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at`
	return r.db.DB.QueryRowContext(ctx, query,
		user.ID, user.Email, user.PasswordHash, user.Name, user.Status,
		user.IsSuperAdmin, user.SuperAdminPromotedAt, user.SuperAdminPromotedBy,
	).Scan(&user.CreatedAt)
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `SELECT id, email, password_hash, name, status, is_super_admin, super_admin_promoted_at, super_admin_promoted_by, created_at FROM users WHERE email = $1`
	user := &User{}
	err := r.db.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.Status,
		&user.IsSuperAdmin, &user.SuperAdminPromotedAt, &user.SuperAdminPromotedBy, &user.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func (r *Repository) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	query := `SELECT id, email, password_hash, name, status, is_super_admin, super_admin_promoted_at, super_admin_promoted_by, created_at FROM users WHERE id = $1`
	user := &User{}
	err := r.db.DB.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.Status,
		&user.IsSuperAdmin, &user.SuperAdminPromotedAt, &user.SuperAdminPromotedBy, &user.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func (r *Repository) UpdateUser(ctx context.Context, user *User) error {
	query := `
		UPDATE users
		SET email = $2, password_hash = $3, name = $4, status = $5,
		    is_super_admin = $6, super_admin_promoted_at = $7, super_admin_promoted_by = $8
		WHERE id = $1`
	_, err := r.db.DB.ExecContext(ctx, query,
		user.ID, user.Email, user.PasswordHash, user.Name, user.Status,
		user.IsSuperAdmin, user.SuperAdminPromotedAt, user.SuperAdminPromotedBy,
	)
	return err
}

func (r *Repository) GetAllUsers(ctx context.Context, limit int, offset int) ([]*User, error) {
	query := `
		SELECT id, email, password_hash, name, status, is_super_admin, super_admin_promoted_at, super_admin_promoted_by, created_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`
	rows, err := r.db.DB.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		if err := rows.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.Status,
			&user.IsSuperAdmin, &user.SuperAdminPromotedAt, &user.SuperAdminPromotedBy, &user.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (r *Repository) GetUserWithMemberships(ctx context.Context, userID uuid.UUID) (*User, []*TeamMembership, error) {
	user, err := r.GetUserByID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	if user == nil {
		return nil, nil, nil
	}

	query := `SELECT id, team_id, user_id, role_id, created_at FROM team_memberships WHERE user_id = $1`
	rows, err := r.db.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var memberships []*TeamMembership
	for rows.Next() {
		tm := &TeamMembership{}
		if err := rows.Scan(&tm.ID, &tm.TeamID, &tm.UserID, &tm.RoleID, &tm.CreatedAt); err != nil {
			return nil, nil, err
		}
		memberships = append(memberships, tm)
	}

	return user, memberships, rows.Err()
}

func (r *Repository) CountSuperAdminsForUpdate(ctx context.Context, tx *sql.Tx) (int, error) {
	query := `SELECT COUNT(*) FROM users WHERE is_super_admin = true FOR UPDATE`
	var count int
	err := tx.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}

func (r *Repository) UpdateUserSuperAdminStatus(ctx context.Context, tx *sql.Tx, userID uuid.UUID, isSuperAdmin bool, promotedBy *uuid.UUID) error {
	query := `
		UPDATE users
		SET is_super_admin = $2, super_admin_promoted_at = CASE WHEN $2 THEN NOW() ELSE NULL END, super_admin_promoted_by = $3
		WHERE id = $1`

	_, err := tx.ExecContext(ctx, query, userID, isSuperAdmin, promotedBy)
	return err
}

func (r *Repository) CreateAuditLog(ctx context.Context, log *AuditLog) error {
	query := `
		INSERT INTO audit_logs (id, team_id, user_id, actor_type, entity_type, entity_id, action, old_data, new_data, ip_address, user_agent, result_status, request_context)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING created_at`

	var oldDataJSON, newDataJSON, requestContextJSON []byte
	if log.OldData != nil {
		var err error
		oldDataJSON, err = json.Marshal(log.OldData)
		if err != nil {
			return err
		}
	}
	if log.NewData != nil {
		var err error
		newDataJSON, err = json.Marshal(log.NewData)
		if err != nil {
			return err
		}
	}
	if log.RequestContext != nil {
		var err error
		requestContextJSON, err = json.Marshal(log.RequestContext)
		if err != nil {
			return err
		}
	}

	return r.db.DB.QueryRowContext(ctx, query,
		log.ID, log.TeamID, log.UserID, log.ActorType, log.EntityType, log.EntityID, log.Action,
		oldDataJSON, newDataJSON, log.IPAddress, log.UserAgent, log.ResultStatus, requestContextJSON,
	).Scan(&log.CreatedAt)
}

func (r *Repository) GetAuditLogs(ctx context.Context, limit int, offset int) ([]*AuditLog, error) {
	query := `
		SELECT id, team_id, user_id, actor_type, entity_type, entity_id, action, old_data, new_data, ip_address, user_agent, result_status, request_context, created_at
		FROM audit_logs
		WHERE actor_type = 'super_admin'
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.DB.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*AuditLog
	for rows.Next() {
		log := &AuditLog{}
		var oldDataJSON, newDataJSON, requestContextJSON sql.NullString

		if err := rows.Scan(&log.ID, &log.TeamID, &log.UserID, &log.ActorType, &log.EntityType, &log.EntityID, &log.Action,
			&oldDataJSON, &newDataJSON, &log.IPAddress, &log.UserAgent, &log.ResultStatus, &requestContextJSON, &log.CreatedAt); err != nil {
			return nil, err
		}

		if oldDataJSON.Valid {
			var data map[string]any
			if err := json.Unmarshal([]byte(oldDataJSON.String), &data); err != nil {
				return nil, err
			}
			log.OldData = data
		}

		if newDataJSON.Valid {
			var data map[string]any
			if err := json.Unmarshal([]byte(newDataJSON.String), &data); err != nil {
				return nil, err
			}
			log.NewData = data
		}

		if requestContextJSON.Valid {
			var data map[string]any
			if err := json.Unmarshal([]byte(requestContextJSON.String), &data); err != nil {
				return nil, err
			}
			log.RequestContext = data
		}

		logs = append(logs, log)
	}
	return logs, rows.Err()
}

// Team methods
func (r *Repository) CreateTeam(ctx context.Context, team *Team) error {
	query := `
		INSERT INTO teams (id, name, slug)
		VALUES ($1, $2, $3)
		RETURNING created_at`
	return r.db.DB.QueryRowContext(ctx, query,
		team.ID, team.Name, team.Slug,
	).Scan(&team.CreatedAt)
}

func (r *Repository) GetTeamByID(ctx context.Context, id uuid.UUID) (*Team, error) {
	query := `SELECT id, name, slug, created_at FROM teams WHERE id = $1`
	team := &Team{}
	err := r.db.DB.QueryRowContext(ctx, query, id).Scan(
		&team.ID, &team.Name, &team.Slug, &team.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return team, err
}

func (r *Repository) GetTeamBySlug(ctx context.Context, slug string) (*Team, error) {
	query := `SELECT id, name, slug, created_at FROM teams WHERE slug = $1`
	team := &Team{}
	err := r.db.DB.QueryRowContext(ctx, query, slug).Scan(
		&team.ID, &team.Name, &team.Slug, &team.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return team, err
}

func (r *Repository) GetTeamsByUserID(ctx context.Context, userID uuid.UUID) ([]*Team, error) {
	query := `
		SELECT t.id, t.name, t.slug, t.created_at
		FROM teams t
		INNER JOIN team_memberships tm ON t.id = tm.team_id
		WHERE tm.user_id = $1
		ORDER BY t.created_at DESC`
	rows, err := r.db.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []*Team
	for rows.Next() {
		team := &Team{}
		if err := rows.Scan(&team.ID, &team.Name, &team.Slug, &team.CreatedAt); err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}
	return teams, rows.Err()
}

func (r *Repository) GetAllTeams(ctx context.Context) ([]*Team, error) {
	query := `SELECT id, name, slug, created_at FROM teams ORDER BY created_at DESC`
	rows, err := r.db.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []*Team
	for rows.Next() {
		team := &Team{}
		if err := rows.Scan(&team.ID, &team.Name, &team.Slug, &team.CreatedAt); err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}
	return teams, rows.Err()
}

func (r *Repository) UpdateTeam(ctx context.Context, team *Team) error {
	query := `UPDATE teams SET name = $2, slug = $3 WHERE id = $1`
	_, err := r.db.DB.ExecContext(ctx, query, team.ID, team.Name, team.Slug)
	return err
}

func (r *Repository) DeleteTeam(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM teams WHERE id = $1`
	_, err := r.db.DB.ExecContext(ctx, query, id)
	return err
}

// Role methods
func (r *Repository) CreateRole(ctx context.Context, role *Role) error {
	permissions, _ := json.Marshal(role.Permissions)
	query := `
		INSERT INTO roles (id, team_id, name, permissions)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at`
	return r.db.DB.QueryRowContext(ctx, query,
		role.ID, role.TeamID, role.Name, permissions,
	).Scan(&role.CreatedAt)
}

func (r *Repository) GetRoleByID(ctx context.Context, id uuid.UUID) (*Role, error) {
	query := `SELECT id, team_id, name, permissions, created_at FROM roles WHERE id = $1`
	role := &Role{}
	var permissions []byte
	err := r.db.DB.QueryRowContext(ctx, query, id).Scan(
		&role.ID, &role.TeamID, &role.Name, &permissions, &role.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(permissions, &role.Permissions); err != nil {
		return nil, err
	}
	return role, nil
}

func (r *Repository) GetRolesByTeamID(ctx context.Context, teamID uuid.UUID) ([]*Role, error) {
	query := `SELECT id, team_id, name, permissions, created_at FROM roles WHERE team_id = $1 ORDER BY name`
	rows, err := r.db.DB.QueryContext(ctx, query, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*Role
	for rows.Next() {
		role := &Role{}
		var permissions []byte
		if err := rows.Scan(&role.ID, &role.TeamID, &role.Name, &permissions, &role.CreatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(permissions, &role.Permissions)
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func (r *Repository) GetRoleByTeamAndName(ctx context.Context, teamID uuid.UUID, name string) (*Role, error) {
	query := `SELECT id, team_id, name, permissions, created_at FROM roles WHERE team_id = $1 AND name = $2`
	role := &Role{}
	var permissions []byte
	err := r.db.DB.QueryRowContext(ctx, query, teamID, name).Scan(
		&role.ID, &role.TeamID, &role.Name, &permissions, &role.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	json.Unmarshal(permissions, &role.Permissions)
	return role, nil
}

func (r *Repository) UpdateRole(ctx context.Context, role *Role) error {
	permissions, _ := json.Marshal(role.Permissions)
	query := `UPDATE roles SET name = $2, permissions = $3 WHERE id = $1`
	_, err := r.db.DB.ExecContext(ctx, query, role.ID, role.Name, permissions)
	return err
}

// Team Membership methods
func (r *Repository) CreateMembership(ctx context.Context, membership *TeamMembership) error {
	query := `
		INSERT INTO team_memberships (id, team_id, user_id, role_id)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at`
	return r.db.DB.QueryRowContext(ctx, query,
		membership.ID, membership.TeamID, membership.UserID, membership.RoleID,
	).Scan(&membership.CreatedAt)
}

func (r *Repository) GetMembership(ctx context.Context, teamID, userID uuid.UUID) (*TeamMembership, error) {
	query := `SELECT id, team_id, user_id, role_id, created_at FROM team_memberships WHERE team_id = $1 AND user_id = $2`
	m := &TeamMembership{}
	err := r.db.DB.QueryRowContext(ctx, query, teamID, userID).Scan(
		&m.ID, &m.TeamID, &m.UserID, &m.RoleID, &m.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return m, err
}

func (r *Repository) GetMembershipsByTeamID(ctx context.Context, teamID uuid.UUID) ([]*TeamMembership, error) {
	query := `SELECT id, team_id, user_id, role_id, created_at FROM team_memberships WHERE team_id = $1`
	rows, err := r.db.DB.QueryContext(ctx, query, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memberships []*TeamMembership
	for rows.Next() {
		m := &TeamMembership{}
		if err := rows.Scan(&m.ID, &m.TeamID, &m.UserID, &m.RoleID, &m.CreatedAt); err != nil {
			return nil, err
		}
		memberships = append(memberships, m)
	}
	return memberships, rows.Err()
}

func (r *Repository) DeleteMembership(ctx context.Context, teamID, userID uuid.UUID) error {
	query := `DELETE FROM team_memberships WHERE team_id = $1 AND user_id = $2`
	_, err := r.db.DB.ExecContext(ctx, query, teamID, userID)
	return err
}

// API Key methods
func (r *Repository) CreateAPIKey(ctx context.Context, key *APIKey) error {
	permissions, _ := json.Marshal(key.Permissions)
	query := `
		INSERT INTO api_keys (id, team_id, user_id, name, key_hash, permissions, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at`
	return r.db.DB.QueryRowContext(ctx, query,
		key.ID, key.TeamID, key.UserID, key.Name, key.KeyHash, permissions, key.ExpiresAt,
	).Scan(&key.CreatedAt)
}

func (r *Repository) GetAPIKeyByHash(ctx context.Context, keyHash string) (*APIKey, error) {
	query := `SELECT id, team_id, user_id, name, key_hash, permissions, expires_at, last_used_at, created_at
		FROM api_keys WHERE key_hash = $1`
	key := &APIKey{}
	var permissions []byte
	err := r.db.DB.QueryRowContext(ctx, query, keyHash).Scan(
		&key.ID, &key.TeamID, &key.UserID, &key.Name, &key.KeyHash,
		&permissions, &key.ExpiresAt, &key.LastUsedAt, &key.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	json.Unmarshal(permissions, &key.Permissions)
	return key, nil
}

func (r *Repository) GetAPIKeysByTeamID(ctx context.Context, teamID uuid.UUID) ([]*APIKey, error) {
	query := `SELECT id, team_id, user_id, name, permissions, expires_at, last_used_at, created_at
		FROM api_keys WHERE team_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.DB.QueryContext(ctx, query, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*APIKey
	for rows.Next() {
		key := &APIKey{}
		var permissions []byte
		if err := rows.Scan(&key.ID, &key.TeamID, &key.UserID, &key.Name,
			&permissions, &key.ExpiresAt, &key.LastUsedAt, &key.CreatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(permissions, &key.Permissions)
		keys = append(keys, key)
	}
	return keys, rows.Err()
}

func (r *Repository) UpdateAPIKeyLastUsed(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE api_keys SET last_used_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.DB.ExecContext(ctx, query, id)
	return err
}

func (r *Repository) DeleteAPIKey(ctx context.Context, teamID, id uuid.UUID) error {
	query := `DELETE FROM api_keys WHERE id = $1 AND team_id = $2`
	_, err := r.db.DB.ExecContext(ctx, query, id, teamID)
	return err
}
