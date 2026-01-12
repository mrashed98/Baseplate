package blueprint

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

func (r *Repository) Create(ctx context.Context, bp *Blueprint) error {
	schema, err := json.Marshal(bp.Schema)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO blueprints (id, team_id, title, description, icon, schema)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at`

	return r.db.DB.QueryRowContext(ctx, query,
		bp.ID, bp.TeamID, bp.Title, bp.Description, bp.Icon, schema,
	).Scan(&bp.CreatedAt, &bp.UpdatedAt)
}

func (r *Repository) GetByID(ctx context.Context, teamID uuid.UUID, id string) (*Blueprint, error) {
	query := `
		SELECT id, team_id, title, description, icon, schema, created_at, updated_at
		FROM blueprints
		WHERE team_id = $1 AND id = $2`

	bp := &Blueprint{}
	var schema []byte
	var description, icon sql.NullString

	err := r.db.DB.QueryRowContext(ctx, query, teamID, id).Scan(
		&bp.ID, &bp.TeamID, &bp.Title, &description, &icon, &schema, &bp.CreatedAt, &bp.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	bp.Description = description.String
	bp.Icon = icon.String
	if err := json.Unmarshal(schema, &bp.Schema); err != nil {
		return nil, err
	}

	return bp, nil
}

func (r *Repository) List(ctx context.Context, teamID uuid.UUID) ([]*Blueprint, error) {
	query := `
		SELECT id, team_id, title, description, icon, schema, created_at, updated_at
		FROM blueprints
		WHERE team_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.DB.QueryContext(ctx, query, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blueprints []*Blueprint
	for rows.Next() {
		bp := &Blueprint{}
		var schema []byte
		var description, icon sql.NullString

		if err := rows.Scan(&bp.ID, &bp.TeamID, &bp.Title, &description, &icon, &schema, &bp.CreatedAt, &bp.UpdatedAt); err != nil {
			return nil, err
		}

		bp.Description = description.String
		bp.Icon = icon.String
		json.Unmarshal(schema, &bp.Schema)
		blueprints = append(blueprints, bp)
	}

	return blueprints, rows.Err()
}

func (r *Repository) Update(ctx context.Context, bp *Blueprint) error {
	schema, err := json.Marshal(bp.Schema)
	if err != nil {
		return err
	}

	query := `
		UPDATE blueprints
		SET title = $3, description = $4, icon = $5, schema = $6, updated_at = CURRENT_TIMESTAMP
		WHERE team_id = $1 AND id = $2
		RETURNING updated_at`

	return r.db.DB.QueryRowContext(ctx, query,
		bp.TeamID, bp.ID, bp.Title, bp.Description, bp.Icon, schema,
	).Scan(&bp.UpdatedAt)
}

func (r *Repository) Delete(ctx context.Context, teamID uuid.UUID, id string) error {
	query := `DELETE FROM blueprints WHERE team_id = $1 AND id = $2`
	_, err := r.db.DB.ExecContext(ctx, query, teamID, id)
	return err
}

func (r *Repository) Exists(ctx context.Context, teamID uuid.UUID, id string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM blueprints WHERE team_id = $1 AND id = $2)`
	var exists bool
	err := r.db.DB.QueryRowContext(ctx, query, teamID, id).Scan(&exists)
	return exists, err
}
