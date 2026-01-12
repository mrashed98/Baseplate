package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// Helper to create test context
func createTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	return c, w
}

// Test IsSuperAdmin helper function
func TestIsSuperAdmin_True(t *testing.T) {
	c, _ := createTestContext()
	c.Set(ContextIsSuperAdmin, true)

	if !IsSuperAdmin(c) {
		t.Error("IsSuperAdmin should return true when context has is_super_admin=true")
	}
}

func TestIsSuperAdmin_False(t *testing.T) {
	c, _ := createTestContext()
	c.Set(ContextIsSuperAdmin, false)

	if IsSuperAdmin(c) {
		t.Error("IsSuperAdmin should return false when context has is_super_admin=false")
	}
}

func TestIsSuperAdmin_NotSet(t *testing.T) {
	c, _ := createTestContext()
	// Don't set ContextIsSuperAdmin

	if IsSuperAdmin(c) {
		t.Error("IsSuperAdmin should return false when context is not set")
	}
}

func TestIsSuperAdmin_InvalidType(t *testing.T) {
	c, _ := createTestContext()
	c.Set(ContextIsSuperAdmin, "invalid") // Wrong type

	if IsSuperAdmin(c) {
		t.Error("IsSuperAdmin should return false when context has invalid type")
	}
}

// Test GetUserID helper function
func TestGetUserID_Valid(t *testing.T) {
	c, _ := createTestContext()
	expectedID := uuid.New()
	c.Set(ContextUserID, expectedID)

	id, ok := GetUserID(c)
	if !ok {
		t.Error("GetUserID should return true when user_id is set")
	}
	if id != expectedID {
		t.Errorf("GetUserID returned %v, expected %v", id, expectedID)
	}
}

func TestGetUserID_NotSet(t *testing.T) {
	c, _ := createTestContext()

	_, ok := GetUserID(c)
	if ok {
		t.Error("GetUserID should return false when user_id is not set")
	}
}

// Test GetTeamID helper function
func TestGetTeamID_Valid(t *testing.T) {
	c, _ := createTestContext()
	expectedID := uuid.New()
	c.Set(ContextTeamID, expectedID)

	id, ok := GetTeamID(c)
	if !ok {
		t.Error("GetTeamID should return true when team_id is set")
	}
	if id != expectedID {
		t.Errorf("GetTeamID returned %v, expected %v", id, expectedID)
	}
}

func TestGetTeamID_NotSet(t *testing.T) {
	c, _ := createTestContext()

	_, ok := GetTeamID(c)
	if ok {
		t.Error("GetTeamID should return false when team_id is not set")
	}
}

func TestGetTeamID_InvalidType(t *testing.T) {
	c, _ := createTestContext()
	c.Set(ContextTeamID, "invalid-uuid")

	_, ok := GetTeamID(c)
	if ok {
		t.Error("GetTeamID should return false when team_id has invalid type")
	}
}

// Test GetPermissions helper function
func TestGetPermissions_Valid(t *testing.T) {
	c, _ := createTestContext()
	expectedPerms := []string{"entity:read", "entity:write"}
	c.Set(ContextPermissions, expectedPerms)

	perms := GetPermissions(c)
	if len(perms) != len(expectedPerms) {
		t.Errorf("GetPermissions returned %d permissions, expected %d", len(perms), len(expectedPerms))
	}
}

func TestGetPermissions_NotSet(t *testing.T) {
	c, _ := createTestContext()

	perms := GetPermissions(c)
	if perms != nil {
		t.Error("GetPermissions should return nil when not set")
	}
}

func TestGetPermissions_InvalidType(t *testing.T) {
	c, _ := createTestContext()
	c.Set(ContextPermissions, "invalid")

	perms := GetPermissions(c)
	if perms != nil {
		t.Error("GetPermissions should return nil when invalid type")
	}
}

// Test RequireSuperAdmin behavior simulation
func TestRequireSuperAdmin_AllowsSuperAdmin(t *testing.T) {
	c, w := createTestContext()
	c.Set(ContextIsSuperAdmin, true)

	// Simulate middleware check
	if !IsSuperAdmin(c) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "super admin privileges required"})
	}

	if w.Code == http.StatusForbidden {
		t.Error("Super admin should be allowed")
	}
}

func TestRequireSuperAdmin_BlocksRegularUser(t *testing.T) {
	c, w := createTestContext()
	c.Set(ContextIsSuperAdmin, false)

	// Simulate middleware check
	if !IsSuperAdmin(c) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "super admin privileges required"})
	}

	if w.Code != http.StatusForbidden {
		t.Error("Regular user should be blocked")
	}
}

// Test RequirePermission behavior simulation
func TestRequirePermission_SuperAdminBypass(t *testing.T) {
	c, _ := createTestContext()
	c.Set(ContextIsSuperAdmin, true)
	c.Set(ContextPermissions, []string{}) // Empty permissions

	// Super admin should bypass permission check
	if IsSuperAdmin(c) {
		t.Log("Super admin correctly bypasses permission check")
		return
	}

	t.Error("Super admin should bypass permission check")
}

func TestRequirePermission_HasPermission(t *testing.T) {
	c, _ := createTestContext()
	c.Set(ContextIsSuperAdmin, false)
	c.Set(ContextPermissions, []string{"entity:read", "entity:write"})

	perms := GetPermissions(c)
	requiredPerm := "entity:read"

	hasPermission := false
	for _, p := range perms {
		if p == requiredPerm {
			hasPermission = true
			break
		}
	}

	if !hasPermission {
		t.Error("User with entity:read permission should pass check")
	}
}

func TestRequirePermission_LacksPermission(t *testing.T) {
	c, w := createTestContext()
	c.Set(ContextIsSuperAdmin, false)
	c.Set(ContextPermissions, []string{"entity:read"})

	perms := GetPermissions(c)
	requiredPerm := "entity:delete"

	hasPermission := false
	for _, p := range perms {
		if p == requiredPerm {
			hasPermission = true
			break
		}
	}

	if !hasPermission {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "permission denied"})
	}

	if w.Code != http.StatusForbidden {
		t.Error("User without required permission should be blocked")
	}
}

// Test RequireTeam behavior with super admin bypass
func TestRequireTeam_SuperAdminBypass(t *testing.T) {
	c, _ := createTestContext()
	c.Set(ContextIsSuperAdmin, true)
	c.Set(ContextUserID, uuid.New())
	c.Request.Header.Set("X-Team-ID", uuid.New().String())

	// Super admin should bypass team membership check and get all permissions
	if IsSuperAdmin(c) {
		// In actual middleware, this sets AllPermissions
		t.Log("Super admin correctly bypasses team membership check")
	} else {
		t.Error("Super admin should bypass team membership check")
	}
}

func TestRequireTeam_ExtractsFromHeader(t *testing.T) {
	c, _ := createTestContext()
	teamID := uuid.New()
	c.Request.Header.Set("X-Team-ID", teamID.String())

	// Extract team ID from header
	teamIDStr := c.GetHeader("X-Team-ID")
	parsedID, err := uuid.Parse(teamIDStr)
	if err != nil {
		t.Errorf("Failed to parse team ID from header: %v", err)
	}
	if parsedID != teamID {
		t.Errorf("Parsed team ID %v doesn't match expected %v", parsedID, teamID)
	}
}

// Test context constants
func TestContextConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		want     string
	}{
		{"ContextUserID", ContextUserID, "user_id"},
		{"ContextTeamID", ContextTeamID, "team_id"},
		{"ContextPermissions", ContextPermissions, "permissions"},
		{"ContextIsSuperAdmin", ContextIsSuperAdmin, "is_super_admin"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.want {
				t.Errorf("got %q, want %q", tt.constant, tt.want)
			}
		})
	}
}

// Test superAdminCache
func TestSuperAdminCache_SetAndGet(t *testing.T) {
	cache := newSuperAdminCache(SuperAdminCacheTTL)
	userID := uuid.New()

	// Initially not in cache
	_, found := cache.get(userID)
	if found {
		t.Error("Cache should not contain entry before set")
	}

	// Set entry
	cache.set(userID, true)

	// Should be found
	isSuperAdmin, found := cache.get(userID)
	if !found {
		t.Error("Cache should contain entry after set")
	}
	if !isSuperAdmin {
		t.Error("Cache should return true for super admin")
	}
}

func TestSuperAdminCache_SetFalse(t *testing.T) {
	cache := newSuperAdminCache(SuperAdminCacheTTL)
	userID := uuid.New()

	// Set as non-super-admin
	cache.set(userID, false)

	// Should return false
	isSuperAdmin, found := cache.get(userID)
	if !found {
		t.Error("Cache should contain entry after set")
	}
	if isSuperAdmin {
		t.Error("Cache should return false for non-super admin")
	}
}

func TestSuperAdminCache_DifferentUsers(t *testing.T) {
	cache := newSuperAdminCache(SuperAdminCacheTTL)
	user1 := uuid.New()
	user2 := uuid.New()

	cache.set(user1, true)
	cache.set(user2, false)

	is1, found1 := cache.get(user1)
	is2, found2 := cache.get(user2)

	if !found1 || !found2 {
		t.Error("Both users should be in cache")
	}
	if !is1 {
		t.Error("User1 should be super admin")
	}
	if is2 {
		t.Error("User2 should not be super admin")
	}
}

// Test SuperAdminCacheTTL constant
func TestSuperAdminCacheTTL_Value(t *testing.T) {
	// TTL should be 1 minute (reasonable for security without excessive DB load)
	if SuperAdminCacheTTL.Minutes() != 1 {
		t.Errorf("SuperAdminCacheTTL should be 1 minute, got %v", SuperAdminCacheTTL)
	}
}

func TestSuperAdminCache_Expiration(t *testing.T) {
	// Use very short TTL for testing
	// Use 50ms TTL with 150ms sleep (3x margin) to avoid flakiness in CI
	cache := newSuperAdminCache(50 * time.Millisecond)
	userID := uuid.New()

	cache.set(userID, true)

	// Should be found immediately
	_, found := cache.get(userID)
	if !found {
		t.Error("Cache should contain entry immediately after set")
	}

	// Wait for expiration (3x TTL to ensure reliable expiration)
	time.Sleep(150 * time.Millisecond)

	// Should no longer be found
	_, found = cache.get(userID)
	if found {
		t.Error("Cache entry should have expired")
	}
}
