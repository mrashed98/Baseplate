package auth

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                   uuid.UUID  `json:"id"`
	Email                string     `json:"email"`
	PasswordHash         string     `json:"-"`
	Name                 string     `json:"name"`
	Status               string     `json:"status"`
	IsSuperAdmin         bool       `json:"is_super_admin"`
	SuperAdminPromotedAt *time.Time `json:"super_admin_promoted_at,omitempty"`
	SuperAdminPromotedBy *uuid.UUID `json:"super_admin_promoted_by,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
}

type Team struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
}

type Role struct {
	ID          uuid.UUID `json:"id"`
	TeamID      uuid.UUID `json:"team_id"`
	Name        string    `json:"name"`
	Permissions []string  `json:"permissions"`
	CreatedAt   time.Time `json:"created_at"`
}

type TeamMembership struct {
	ID        uuid.UUID `json:"id"`
	TeamID    uuid.UUID `json:"team_id"`
	UserID    uuid.UUID `json:"user_id"`
	RoleID    uuid.UUID `json:"role_id"`
	CreatedAt time.Time `json:"created_at"`
}

type APIKey struct {
	ID          uuid.UUID  `json:"id"`
	TeamID      uuid.UUID  `json:"team_id"`
	UserID      *uuid.UUID `json:"user_id,omitempty"`
	Name        string     `json:"name"`
	KeyHash     string     `json:"-"`
	Permissions []string   `json:"permissions"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// Request/Response types
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

type CreateTeamRequest struct {
	Name string `json:"name" binding:"required"`
	Slug string `json:"slug" binding:"required"`
}

type InviteMemberRequest struct {
	Email  string `json:"email" binding:"required,email"`
	RoleID string `json:"role_id" binding:"required"`
}

type CreateAPIKeyRequest struct {
	Name        string   `json:"name" binding:"required"`
	Permissions []string `json:"permissions"`
	ExpiresAt   *string  `json:"expires_at"`
}

type CreateAPIKeyResponse struct {
	APIKey *APIKey `json:"api_key"`
	Key    string  `json:"key"`
}

type AuditLog struct {
	ID             uuid.UUID         `json:"id"`
	TeamID         *uuid.UUID        `json:"team_id,omitempty"`
	UserID         *uuid.UUID        `json:"user_id,omitempty"`
	ActorType      string            `json:"actor_type"`
	EntityType     string            `json:"entity_type"`
	EntityID       string            `json:"entity_id"`
	Action         string            `json:"action"`
	OldData        map[string]any    `json:"old_data,omitempty"`
	NewData        map[string]any    `json:"new_data,omitempty"`
	IPAddress      *string           `json:"ip_address,omitempty"`
	UserAgent      *string           `json:"user_agent,omitempty"`
	ResultStatus   *string           `json:"result_status,omitempty"`
	RequestContext map[string]any    `json:"request_context,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
}

// Permission constants
const (
	PermTeamManage       = "team:manage"
	PermBlueprintRead    = "blueprint:read"
	PermBlueprintWrite   = "blueprint:write"
	PermBlueprintDelete  = "blueprint:delete"
	PermEntityRead       = "entity:read"
	PermEntityWrite      = "entity:write"
	PermEntityDelete     = "entity:delete"
	PermIntegrationRead  = "integration:read"
	PermIntegrationWrite = "integration:write"
	PermScorecardRead    = "scorecard:read"
	PermScorecardWrite   = "scorecard:write"
	PermActionRead       = "action:read"
	PermActionWrite      = "action:write"
	PermActionExecute    = "action:execute"
)

var AllPermissions = []string{
	PermTeamManage,
	PermBlueprintRead, PermBlueprintWrite, PermBlueprintDelete,
	PermEntityRead, PermEntityWrite, PermEntityDelete,
	PermIntegrationRead, PermIntegrationWrite,
	PermScorecardRead, PermScorecardWrite,
	PermActionRead, PermActionWrite, PermActionExecute,
}

var AdminPermissions = append([]string{}, AllPermissions...)

var EditorPermissions = []string{
	PermBlueprintRead, PermBlueprintWrite,
	PermEntityRead, PermEntityWrite,
	PermIntegrationRead, PermIntegrationWrite,
	PermScorecardRead, PermScorecardWrite,
	PermActionRead, PermActionWrite, PermActionExecute,
}

var ViewerPermissions = []string{
	PermBlueprintRead,
	PermEntityRead,
	PermIntegrationRead,
	PermScorecardRead,
	PermActionRead,
}
