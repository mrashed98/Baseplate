package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/baseplate/baseplate/internal/core/auth"
)

// superAdminCache provides a simple TTL cache for super admin status checks.
// This reduces DB load while ensuring demoted users lose access within the cache TTL.
type superAdminCache struct {
	mu      sync.RWMutex
	entries map[uuid.UUID]cacheEntry
	ttl     time.Duration
}

type cacheEntry struct {
	isSuperAdmin bool
	expiresAt    time.Time
}

func newSuperAdminCache(ttl time.Duration) *superAdminCache {
	return &superAdminCache{
		entries: make(map[uuid.UUID]cacheEntry),
		ttl:     ttl,
	}
}

func (c *superAdminCache) get(userID uuid.UUID) (bool, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[userID]
	if !exists || time.Now().After(entry.expiresAt) {
		return false, false
	}
	return entry.isSuperAdmin, true
}

func (c *superAdminCache) set(userID uuid.UUID, isSuperAdmin bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Lazy cleanup: remove expired entries to prevent memory leak
	now := time.Now()
	for id, entry := range c.entries {
		if now.After(entry.expiresAt) {
			delete(c.entries, id)
		}
	}

	c.entries[userID] = cacheEntry{
		isSuperAdmin: isSuperAdmin,
		expiresAt:    now.Add(c.ttl),
	}
}

const (
	ContextUserID       = "user_id"
	ContextTeamID       = "team_id"
	ContextPermissions  = "permissions"
	ContextIsSuperAdmin = "is_super_admin"
)

type AuthMiddleware struct {
	authService     *auth.Service
	superAdminCache *superAdminCache
}

// SuperAdminCacheTTL is the duration super admin status is cached before re-checking the database.
// After demotion, a user will lose super admin access within this time window.
const SuperAdminCacheTTL = 1 * time.Minute

func NewAuthMiddleware(authService *auth.Service) *AuthMiddleware {
	return &AuthMiddleware{
		authService:     authService,
		superAdminCache: newSuperAdminCache(SuperAdminCacheTTL),
	}
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

	// Set is_super_admin flag in context
	isSuperAdmin := false
	if claims.IsSuperAdmin != nil && *claims.IsSuperAdmin {
		isSuperAdmin = true
	}
	c.Set(ContextIsSuperAdmin, isSuperAdmin)

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

			// Super admins bypass team membership checks and have all permissions
			if IsSuperAdmin(c) {
				c.Set(ContextPermissions, auth.AllPermissions)
			} else {
				permissions, err := m.authService.GetUserPermissions(c.Request.Context(), teamID, userUUID)

				if err != nil {
					c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "access denied"})
					return
				}
				c.Set(ContextPermissions, permissions)
			}
		}

		c.Set(ContextTeamID, teamID)
		c.Next()
	}
}

func (m *AuthMiddleware) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Super admins bypass all permission checks
		if IsSuperAdmin(c) {
			c.Next()
			return
		}

		perms, exists := c.Get(ContextPermissions)
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "no permissions found"})
			return
		}

		permissions, ok := perms.([]string)

		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "invalid permission type"})
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

func IsSuperAdmin(c *gin.Context) bool {
	val, exists := c.Get(ContextIsSuperAdmin)
	if !exists {
		return false
	}

	if isSuperAdmin, ok := val.(bool); ok {
		return isSuperAdmin
	}

	return false
}

// RequireSuperAdmin middleware ensures user is a super admin.
// It verifies the claim against the database (with caching) to handle demoted users
// whose JWT tokens still contain is_super_admin=true.
func (m *AuthMiddleware) RequireSuperAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First check JWT claim - if not super admin in JWT, reject immediately
		if !IsSuperAdmin(c) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "super admin privileges required"})
			return
		}

		// Get user ID for DB verification
		userID, exists := GetUserID(c)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user id not found"})
			return
		}

		// Check cache first
		if isSuperAdmin, found := m.superAdminCache.get(userID); found {
			if !isSuperAdmin {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "super admin privileges required"})
				return
			}
			c.Next()
			return
		}

		// Cache miss - verify against database
		isSuperAdmin, err := m.authService.CheckSuperAdminStatus(c.Request.Context(), userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to verify super admin status"})
			return
		}

		// Update cache
		m.superAdminCache.set(userID, isSuperAdmin)

		if !isSuperAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "super admin privileges required"})
			return
		}

		c.Next()
	}
}
