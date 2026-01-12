package entity

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/baseplate/baseplate/internal/storage/postgres"
)

type Repository struct {
	db *postgres.Client
}

func NewRepository(db *postgres.Client) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, entity *Entity) error {
	data, err := json.Marshal(entity.Data)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO entities (id, team_id, blueprint_id, identifier, title, data)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at`

	return r.db.DB.QueryRowContext(ctx, query,
		entity.ID, entity.TeamID, entity.BlueprintID, entity.Identifier, entity.Title, data,
	).Scan(&entity.CreatedAt, &entity.UpdatedAt)
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Entity, error) {
	query := `
		SELECT id, team_id, blueprint_id, identifier, title, data, created_at, updated_at
		FROM entities
		WHERE id = $1`

	return r.scanEntity(r.db.DB.QueryRowContext(ctx, query, id))
}

func (r *Repository) GetByIdentifier(ctx context.Context, teamID uuid.UUID, blueprintID, identifier string) (*Entity, error) {
	query := `
		SELECT id, team_id, blueprint_id, identifier, title, data, created_at, updated_at
		FROM entities
		WHERE team_id = $1 AND blueprint_id = $2 AND identifier = $3`

	return r.scanEntity(r.db.DB.QueryRowContext(ctx, query, teamID, blueprintID, identifier))
}

func (r *Repository) List(ctx context.Context, teamID uuid.UUID, blueprintID string, limit, offset int) ([]*Entity, int, error) {
	countQuery := `SELECT COUNT(*) FROM entities WHERE team_id = $1 AND blueprint_id = $2`
	var total int
	if err := r.db.DB.QueryRowContext(ctx, countQuery, teamID, blueprintID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, team_id, blueprint_id, identifier, title, data, created_at, updated_at
		FROM entities
		WHERE team_id = $1 AND blueprint_id = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`

	rows, err := r.db.DB.QueryContext(ctx, query, teamID, blueprintID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	entities, err := r.scanEntities(rows)
	return entities, total, err
}

func (r *Repository) Search(ctx context.Context, teamID uuid.UUID, blueprintID string, req *SearchRequest) ([]*Entity, int, error) {
	whereClause := []string{"team_id = $1", "blueprint_id = $2"}
	args := []interface{}{teamID, blueprintID}
	argIndex := 3

	for _, filter := range req.Filters {
		clause, newArgs, idx := r.buildFilterClause(filter, argIndex)
		if clause != "" {
			whereClause = append(whereClause, clause)
			args = append(args, newArgs...)
			argIndex = idx
		}
	}

	where := strings.Join(whereClause, " AND ")

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM entities WHERE %s", where)
	var total int
	if err := r.db.DB.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Build order clause
	orderClause := "created_at DESC"
	if req.OrderBy != "" {
		dir := "ASC"
		if strings.ToUpper(req.OrderDir) == "DESC" {
			dir = "DESC"
		}
		if req.OrderBy == "created_at" || req.OrderBy == "updated_at" || req.OrderBy == "identifier" || req.OrderBy == "title" {
			orderClause = fmt.Sprintf("%s %s", req.OrderBy, dir)
		} else {
			// Order by JSONB property
			orderClause = fmt.Sprintf("data->>'%s' %s", req.OrderBy, dir)
		}
	}

	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	query := fmt.Sprintf(`
		SELECT id, team_id, blueprint_id, identifier, title, data, created_at, updated_at
		FROM entities
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d`, where, orderClause, argIndex, argIndex+1)

	args = append(args, limit, req.Offset)

	rows, err := r.db.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	entities, err := r.scanEntities(rows)
	return entities, total, err
}

func (r *Repository) buildFilterClause(filter SearchFilter, argIndex int) (string, []interface{}, int) {
	var clause string
	var args []interface{}

	// Property path - supports nested properties like "metadata.version"
	propPath := fmt.Sprintf("data->'%s'", strings.Replace(filter.Property, ".", "'->'", -1))

	switch filter.Operator {
	case "eq":
		valueJSON, _ := json.Marshal(filter.Value)
		clause = fmt.Sprintf("%s = $%d", propPath, argIndex)
		args = append(args, string(valueJSON))
		argIndex++
	case "neq":
		valueJSON, _ := json.Marshal(filter.Value)
		clause = fmt.Sprintf("%s != $%d", propPath, argIndex)
		args = append(args, string(valueJSON))
		argIndex++
	case "gt":
		clause = fmt.Sprintf("(%s)::numeric > $%d", propPath, argIndex)
		args = append(args, filter.Value)
		argIndex++
	case "gte":
		clause = fmt.Sprintf("(%s)::numeric >= $%d", propPath, argIndex)
		args = append(args, filter.Value)
		argIndex++
	case "lt":
		clause = fmt.Sprintf("(%s)::numeric < $%d", propPath, argIndex)
		args = append(args, filter.Value)
		argIndex++
	case "lte":
		clause = fmt.Sprintf("(%s)::numeric <= $%d", propPath, argIndex)
		args = append(args, filter.Value)
		argIndex++
	case "contains":
		// Text contains
		clause = fmt.Sprintf("%s::text ILIKE $%d", propPath, argIndex)
		args = append(args, "%"+fmt.Sprint(filter.Value)+"%")
		argIndex++
	case "exists":
		if filter.Value == true {
			clause = fmt.Sprintf("data ? '%s'", filter.Property)
		} else {
			clause = fmt.Sprintf("NOT (data ? '%s')", filter.Property)
		}
	case "in":
		if arr, ok := filter.Value.([]interface{}); ok {
			placeholders := make([]string, len(arr))
			for i, v := range arr {
				valueJSON, _ := json.Marshal(v)
				placeholders[i] = fmt.Sprintf("$%d", argIndex)
				args = append(args, string(valueJSON))
				argIndex++
			}
			clause = fmt.Sprintf("%s IN (%s)", propPath, strings.Join(placeholders, ","))
		}
	}

	return clause, args, argIndex
}

func (r *Repository) Update(ctx context.Context, entity *Entity) error {
	data, err := json.Marshal(entity.Data)
	if err != nil {
		return err
	}

	query := `
		UPDATE entities
		SET title = $2, data = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at`

	return r.db.DB.QueryRowContext(ctx, query, entity.ID, entity.Title, data).Scan(&entity.UpdatedAt)
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM entities WHERE id = $1`
	_, err := r.db.DB.ExecContext(ctx, query, id)
	return err
}

func (r *Repository) DeleteByBlueprint(ctx context.Context, teamID uuid.UUID, blueprintID string) error {
	query := `DELETE FROM entities WHERE team_id = $1 AND blueprint_id = $2`
	_, err := r.db.DB.ExecContext(ctx, query, teamID, blueprintID)
	return err
}

func (r *Repository) scanEntity(row *sql.Row) (*Entity, error) {
	entity := &Entity{}
	var data []byte
	var title sql.NullString

	err := row.Scan(
		&entity.ID, &entity.TeamID, &entity.BlueprintID,
		&entity.Identifier, &title, &data,
		&entity.CreatedAt, &entity.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	entity.Title = title.String
	json.Unmarshal(data, &entity.Data)
	return entity, nil
}

func (r *Repository) scanEntities(rows *sql.Rows) ([]*Entity, error) {
	var entities []*Entity
	for rows.Next() {
		entity := &Entity{}
		var data []byte
		var title sql.NullString

		if err := rows.Scan(
			&entity.ID, &entity.TeamID, &entity.BlueprintID,
			&entity.Identifier, &title, &data,
			&entity.CreatedAt, &entity.UpdatedAt,
		); err != nil {
			return nil, err
		}

		entity.Title = title.String
		json.Unmarshal(data, &entity.Data)
		entities = append(entities, entity)
	}
	return entities, rows.Err()
}
