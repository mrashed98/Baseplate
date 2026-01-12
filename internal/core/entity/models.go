package entity

import (
	"time"

	"github.com/google/uuid"
)

type Entity struct {
	ID          uuid.UUID              `json:"id"`
	TeamID      uuid.UUID              `json:"team_id"`
	BlueprintID string                 `json:"blueprint_id"`
	Identifier  string                 `json:"identifier"`
	Title       string                 `json:"title,omitempty"`
	Data        map[string]interface{} `json:"data"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type CreateEntityRequest struct {
	Identifier string                 `json:"identifier" binding:"required"`
	Title      string                 `json:"title"`
	Data       map[string]interface{} `json:"data" binding:"required"`
}

type UpdateEntityRequest struct {
	Title string                 `json:"title"`
	Data  map[string]interface{} `json:"data"`
}

type SearchFilter struct {
	Property string      `json:"property"`
	Operator string      `json:"operator"` // eq, neq, gt, lt, gte, lte, contains, exists
	Value    interface{} `json:"value"`
}

type SearchRequest struct {
	Filters  []SearchFilter `json:"filters"`
	OrderBy  string         `json:"order_by"`
	OrderDir string         `json:"order_dir"` // asc, desc
	Limit    int            `json:"limit"`
	Offset   int            `json:"offset"`
}

type ListEntitiesResponse struct {
	Entities []*Entity `json:"entities"`
	Total    int       `json:"total"`
	Limit    int       `json:"limit"`
	Offset   int       `json:"offset"`
}
