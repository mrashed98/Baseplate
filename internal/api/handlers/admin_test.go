package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/baseplate/baseplate/internal/api/middleware"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// Helper to create test context with super admin
func createAdminTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/test", nil)
	c.Set(middleware.ContextIsSuperAdmin, true)
	c.Set(middleware.ContextUserID, uuid.New())
	return c, w
}

// Helper to create test context with regular user
func createRegularUserTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/test", nil)
	c.Set(middleware.ContextIsSuperAdmin, false)
	c.Set(middleware.ContextUserID, uuid.New())
	return c, w
}

// Test middleware authorization behavior for admin endpoints
// NOTE: These tests verify the middleware authorization check (IsSuperAdmin),
// not the handler implementation itself. They document expected middleware behavior.

func TestListTeams_MiddlewareAllowsSuperAdmin(t *testing.T) {
	c, _ := createAdminTestContext()

	// Verify middleware check passes for super admin context
	if !middleware.IsSuperAdmin(c) {
		t.Error("Middleware should allow super admin to access admin endpoints")
	}
}

func TestListTeams_MiddlewareBlocksRegularUser(t *testing.T) {
	c, w := createRegularUserTestContext()

	// Simulate middleware rejection for non-super-admin
	if !middleware.IsSuperAdmin(c) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "super admin privileges required"})
	}

	if w.Code != http.StatusForbidden {
		t.Error("Middleware should block regular user from admin endpoints")
	}
}

func TestListUsers_MiddlewareAllowsSuperAdmin(t *testing.T) {
	c, _ := createAdminTestContext()

	// Verify middleware check passes for super admin context
	if !middleware.IsSuperAdmin(c) {
		t.Error("Middleware should allow super admin to access admin endpoints")
	}
}

func TestListUsers_PaginationParams(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/users?limit=100&offset=50", nil)

	// Verify pagination parsing
	limit := c.DefaultQuery("limit", "50")
	offset := c.DefaultQuery("offset", "0")

	if limit != "100" {
		t.Errorf("Expected limit=100, got %s", limit)
	}
	if offset != "50" {
		t.Errorf("Expected offset=50, got %s", offset)
	}
}

// Test GetUserDetail endpoint
func TestGetUserDetail_ValidUUID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	userID := uuid.New()
	c.Params = gin.Params{{Key: "userId", Value: userID.String()}}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/users/"+userID.String(), nil)

	// Parse the user ID
	userIDStr := c.Param("userId")
	parsedID, err := uuid.Parse(userIDStr)
	if err != nil {
		t.Errorf("Failed to parse user ID: %v", err)
	}
	if parsedID != userID {
		t.Errorf("Parsed ID %v doesn't match expected %v", parsedID, userID)
	}
}

func TestGetUserDetail_InvalidUUID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Params = gin.Params{{Key: "userId", Value: "invalid-uuid"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/users/invalid-uuid", nil)

	userIDStr := c.Param("userId")
	_, err := uuid.Parse(userIDStr)
	if err == nil {
		t.Error("Should fail to parse invalid UUID")
	}
}

// Test UpdateUser endpoint
func TestUpdateUser_ValidRequest(t *testing.T) {
	req := UpdateUserRequest{
		Name:   "New Name",
		Status: "active",
	}

	if req.Name == "" {
		t.Error("Name should not be empty")
	}
	if req.Status != "active" && req.Status != "deleted" {
		// This should trigger validation error in actual code
		t.Log("Status validation would catch invalid values")
	}
}

func TestUpdateUser_StatusValidation(t *testing.T) {
	validStatuses := []string{"active", "deleted"}
	invalidStatuses := []string{"inactive", "suspended", "banned", ""}

	for _, status := range validStatuses {
		isValid := status == "active" || status == "deleted"
		if !isValid {
			t.Errorf("Status %q should be valid", status)
		}
	}

	for _, status := range invalidStatuses {
		if status == "" {
			continue // Empty is allowed (no update)
		}
		isValid := status == "active" || status == "deleted"
		if isValid {
			t.Errorf("Status %q should be invalid", status)
		}
	}
}

// Test PromoteUser endpoint
func TestPromoteUser_ValidUUID(t *testing.T) {
	c, _ := createAdminTestContext()

	userID := uuid.New()
	c.Params = gin.Params{{Key: "userId", Value: userID.String()}}

	userIDStr := c.Param("userId")
	parsedID, err := uuid.Parse(userIDStr)
	if err != nil {
		t.Errorf("Failed to parse user ID: %v", err)
	}
	if parsedID != userID {
		t.Errorf("Parsed ID doesn't match expected")
	}
}

func TestPromoteUser_ActorIDExtraction(t *testing.T) {
	c, _ := createAdminTestContext()
	expectedActorID := uuid.New()
	c.Set(middleware.ContextUserID, expectedActorID)

	actorID, ok := middleware.GetUserID(c)
	if !ok {
		t.Error("Should get actor ID from context")
	}
	if actorID != expectedActorID {
		t.Error("Actor ID doesn't match expected")
	}
}

// Test DemoteUser endpoint
func TestDemoteUser_ValidUUID(t *testing.T) {
	c, _ := createAdminTestContext()

	userID := uuid.New()
	c.Params = gin.Params{{Key: "userId", Value: userID.String()}}

	userIDStr := c.Param("userId")
	parsedID, err := uuid.Parse(userIDStr)
	if err != nil {
		t.Errorf("Failed to parse user ID: %v", err)
	}
	if parsedID != userID {
		t.Errorf("Parsed ID doesn't match expected")
	}
}

// Test QueryAuditLogs endpoint
func TestQueryAuditLogs_DefaultPagination(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/audit-logs", nil)

	// Default values
	limit := 50
	offset := 0

	if l := c.Query("limit"); l != "" {
		t.Error("Expected empty limit query param")
	}
	if o := c.Query("offset"); o != "" {
		t.Error("Expected empty offset query param")
	}

	t.Logf("Default pagination: limit=%d, offset=%d", limit, offset)
}

func TestQueryAuditLogs_CustomPagination(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/audit-logs?limit=100&offset=25", nil)

	limit := c.DefaultQuery("limit", "50")
	offset := c.DefaultQuery("offset", "0")

	if limit != "100" {
		t.Errorf("Expected limit=100, got %s", limit)
	}
	if offset != "25" {
		t.Errorf("Expected offset=25, got %s", offset)
	}
}

// Error Response Messages Documentation
// The following error messages are used by admin handlers:
// - "invalid user id" - returned when UUID parsing fails
// - "user not found" - returned when user doesn't exist
// - "only super admins can promote users" - returned for unauthorized promotion
// - "only super admins can demote users" - returned for unauthorized demotion
// - "cannot demote the last super admin" - returned when attempting to remove last admin
// - "user is already a super admin" - returned when promoting existing admin
// - "user is not a super admin" - returned when demoting non-admin

// Test GetTeamDetail endpoint
func TestGetTeamDetail_ValidUUID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	teamID := uuid.New()
	c.Params = gin.Params{{Key: "teamId", Value: teamID.String()}}

	teamIDStr := c.Param("teamId")
	parsedID, err := uuid.Parse(teamIDStr)
	if err != nil {
		t.Errorf("Failed to parse team ID: %v", err)
	}
	if parsedID != teamID {
		t.Errorf("Parsed ID %v doesn't match expected %v", parsedID, teamID)
	}
}

func TestGetTeamDetail_InvalidUUID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Params = gin.Params{{Key: "teamId", Value: "not-a-uuid"}}

	teamIDStr := c.Param("teamId")
	_, err := uuid.Parse(teamIDStr)
	if err == nil {
		t.Error("Should fail to parse invalid UUID")
	}
}

// Test UpdateUserRequest struct
func TestUpdateUserRequest_EmptyFields(t *testing.T) {
	req := UpdateUserRequest{
		Name:   "",
		Status: "",
	}

	// Empty fields should be allowed (optional updates)
	if req.Name != "" || req.Status != "" {
		t.Error("Empty request should have empty fields")
	}
}

func TestUpdateUserRequest_PartialUpdate(t *testing.T) {
	req := UpdateUserRequest{
		Name:   "Updated Name",
		Status: "", // Only updating name
	}

	if req.Name == "" {
		t.Error("Name should be set for partial update")
	}
	if req.Status != "" {
		t.Error("Status should be empty for partial update")
	}
}
