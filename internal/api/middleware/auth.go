package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/baseplate/baseplate/internal/core/auth"
)

const (
	ContextUserID      = "user_id"
	ContextTeamID      = "team_id"
	ContextPermissions = "permissions"
)

type AuthMiddleware struct {
	authService *auth.Service
}

func NewAuthMiddleware(authService *auth.Service) *AuthMiddleware {
	return &AuthMiddleware{authService: authService}
}

func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			return
		}

		switch strings.ToLower(parts[0]) {
		case "bearer":
			m.handleJWT(c, parts[1])
		case "apikey":
			m.handleAPIKey(c, parts[1])
		default:
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unsupported authorization type"})
			return
		}
	}
}

func (m *AuthMiddleware) handleJWT(c *gin.Context, token string) {
	claims, err := m.authService.ValidateToken(token)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	c.Set(ContextUserID, claims.UserID)
	c.Next()
}

func (m *AuthMiddleware) handleAPIKey(c *gin.Context, key string) {
	apiKey, err := m.authService.ValidateAPIKey(c.Request.Context(), key)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
		return
	}

	c.Set(ContextTeamID, apiKey.TeamID)
	c.Set(ContextPermissions, apiKey.Permissions)
	if apiKey.UserID != nil {
		c.Set(ContextUserID, *apiKey.UserID)
	}
	c.Next()
}

func (m *AuthMiddleware) RequireTeam() gin.HandlerFunc {
	return func(c *gin.Context) {
		teamIDStr := c.Param("teamId")
		if teamIDStr == "" {
			teamIDStr = c.GetHeader("X-Team-ID")
		}

		if teamIDStr == "" {
			// Check if already set by API key
			if _, exists := c.Get(ContextTeamID); exists {
				c.Next()
				return
			}
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "team id required"})
			return
		}

		teamID, err := uuid.Parse(teamIDStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid team id"})
			return
		}

		// Verify user has access to this team
		userID, exists := c.Get(ContextUserID)
		if exists {

			userUUID, ok := userID.(uuid.UUID)
			if !ok {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "invalid user type"})
				return
			}

			permissions, err := m.authService.GetUserPermissions(c.Request.Context(), teamID, userUUID)

			if err != nil {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "access denied"})
				return
			}
			c.Set(ContextPermissions, permissions)
		}

		c.Set(ContextTeamID, teamID)
		c.Next()
	}
}

func (m *AuthMiddleware) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		perms, exists := c.Get(ContextPermissions)
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "no permissions found"})
			return
		}

		permissions, ok := perms.([]string)

		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "invalid permession type"})
			return
		}

		for _, p := range permissions {
			if p == permission {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "permission denied"})
	}
}

// Helper functions to get context values
func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get(ContextUserID)
	if !exists {
		return uuid.Nil, false
	}

	return val.(uuid.UUID), true
}

func GetTeamID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get(ContextTeamID)
	if !exists {
		return uuid.Nil, false
	}

	if id, ok := val.(uuid.UUID); ok {
		return id, true
	}

	return uuid.Nil, false
}

func GetPermissions(c *gin.Context) []string {
	val, exists := c.Get(ContextPermissions)
	if !exists {
		return nil
	}

	if perms, ok := val.([]string); ok {
		return perms
	}

	return nil
}
