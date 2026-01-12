package blueprint

import (
	"time"

	"github.com/google/uuid"
)

type Blueprint struct {
	ID          string                 `json:"id"`
	TeamID      uuid.UUID              `json:"team_id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description,omitempty"`
	Icon        string                 `json:"icon,omitempty"`
	Schema      map[string]interface{} `json:"schema"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type CreateBlueprintRequest struct {
	ID          string                 `json:"id" binding:"required"`
	Title       string                 `json:"title" binding:"required"`
	Description string                 `json:"description"`
	Icon        string                 `json:"icon"`
	Schema      map[string]interface{} `json:"schema" binding:"required"`
}

type UpdateBlueprintRequest struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Icon        string                 `json:"icon"`
	Schema      map[string]interface{} `json:"schema"`
}

type ListBlueprintsResponse struct {
	Blueprints []*Blueprint `json:"blueprints"`
	Total      int          `json:"total"`
}

// JSON Schema property types
type PropertyType string

const (
	PropertyTypeString  PropertyType = "string"
	PropertyTypeNumber  PropertyType = "number"
	PropertyTypeInteger PropertyType = "integer"
	PropertyTypeBoolean PropertyType = "boolean"
	PropertyTypeArray   PropertyType = "array"
	PropertyTypeObject  PropertyType = "object"
)

// Schema builder helpers
type SchemaProperty struct {
	Type        PropertyType           `json:"type"`
	Title       string                 `json:"title,omitempty"`
	Description string                 `json:"description,omitempty"`
	Format      string                 `json:"format,omitempty"`
	Enum        []interface{}          `json:"enum,omitempty"`
	Default     interface{}            `json:"default,omitempty"`
	Items       *SchemaProperty        `json:"items,omitempty"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
	Required    []string               `json:"required,omitempty"`
}

func NewSchema(title string, properties map[string]*SchemaProperty, required []string) map[string]interface{} {
	props := make(map[string]interface{})
	for k, v := range properties {
		props[k] = v
	}

	return map[string]interface{}{
		"type":       "object",
		"title":      title,
		"properties": props,
		"required":   required,
	}
}
