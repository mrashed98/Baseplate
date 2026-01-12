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
		INSERT INTO users (id, email, password_hash, name, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at`
	return r.db.DB.QueryRowContext(ctx, query,
		user.ID, user.Email, user.PasswordHash, user.Name, user.Status,
	).Scan(&user.CreatedAt)
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `SELECT id, email, password_hash, name, status, created_at FROM users WHERE email = $1`
	user := &User{}
	err := r.db.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.Status, &user.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func (r *Repository) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	query := `SELECT id, email, password_hash, name, status, created_at FROM users WHERE id = $1`
	user := &User{}
	err := r.db.DB.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.Status, &user.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
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

func (r *Repository) DeleteAPIKey(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM api_keys WHERE id = $1`
	_, err := r.db.DB.ExecContext(ctx, query, id)
	return err
}
