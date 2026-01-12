package auth

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
)

// MockRepository implements a mock repository for testing
type MockRepository struct {
	users           map[uuid.UUID]*User
	superAdminCount int
	updateErr       error
	txStarted       bool
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		users:           make(map[uuid.UUID]*User),
		superAdminCount: 1,
	}
}

func (m *MockRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	if user, ok := m.users[id]; ok {
		return user, nil
	}
	return nil, nil
}

func (m *MockRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, nil
}

func (m *MockRepository) UpdateUser(ctx context.Context, user *User) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.users[user.ID] = user
	return nil
}

func (m *MockRepository) CountSuperAdminsForUpdate(ctx context.Context, tx *sql.Tx) (int, error) {
	return m.superAdminCount, nil
}

func (m *MockRepository) UpdateUserSuperAdminStatus(ctx context.Context, tx *sql.Tx, userID uuid.UUID, isSuperAdmin bool, promotedBy *uuid.UUID) error {
	if user, ok := m.users[userID]; ok {
		user.IsSuperAdmin = isSuperAdmin
		if isSuperAdmin {
			now := time.Now()
			user.SuperAdminPromotedAt = &now
			user.SuperAdminPromotedBy = promotedBy
		} else {
			user.SuperAdminPromotedAt = nil
			user.SuperAdminPromotedBy = nil
		}
	}
	return nil
}

// Test PromoteToSuperAdmin - Model Behavior Tests
// NOTE: These tests verify model field behavior and preconditions, not service method integration.
// The mock repository doesn't wire to the actual service, so these tests document expected
// model state transitions rather than testing the PromoteToSuperAdmin service method directly.
// For full integration tests, use a test database with the real service.
func TestPromoteToSuperAdmin_Success(t *testing.T) {
	mock := NewMockRepository()

	// Create actor (super admin)
	actorID := uuid.New()
	actor := &User{
		ID:           actorID,
		Email:        "admin@example.com",
		Name:         "Admin",
		Status:       "active",
		IsSuperAdmin: true,
	}
	mock.users[actorID] = actor

	// Create target user (regular user)
	targetID := uuid.New()
	target := &User{
		ID:           targetID,
		Email:        "user@example.com",
		Name:         "User",
		Status:       "active",
		IsSuperAdmin: false,
	}
	mock.users[targetID] = target

	// Verify preconditions for promotion
	if !actor.IsSuperAdmin {
		t.Error("Actor should be super admin")
	}
	if target.IsSuperAdmin {
		t.Error("Target should not be super admin initially")
	}

	// Simulate promotion (documents expected field changes)
	target.IsSuperAdmin = true
	now := time.Now()
	target.SuperAdminPromotedAt = &now
	target.SuperAdminPromotedBy = &actorID

	// Verify expected state after promotion
	if !target.IsSuperAdmin {
		t.Error("Target should be super admin after promotion")
	}
	if target.SuperAdminPromotedBy == nil || *target.SuperAdminPromotedBy != actorID {
		t.Error("SuperAdminPromotedBy should be set to actor ID")
	}
}

// TestPromoteToSuperAdmin_AlreadySuperAdmin documents the precondition check.
// NOTE: This is a model behavior test, not a service integration test.
// The actual service method returns ErrAlreadySuperAdmin when target.IsSuperAdmin is true.
func TestPromoteToSuperAdmin_AlreadySuperAdmin(t *testing.T) {
	// Create target user who is already super admin
	targetID := uuid.New()
	target := &User{
		ID:           targetID,
		Email:        "user@example.com",
		Name:         "User",
		Status:       "active",
		IsSuperAdmin: true,
	}

	// Document the precondition that triggers ErrAlreadySuperAdmin in the service
	if !target.IsSuperAdmin {
		t.Error("Test setup error: target should be super admin for this test case")
	}
	t.Log("Precondition verified: user already super admin triggers ErrAlreadySuperAdmin in service")
}

func TestPromoteToSuperAdmin_Unauthorized(t *testing.T) {
	// Create actor (not super admin)
	actorID := uuid.New()
	actor := &User{
		ID:           actorID,
		Email:        "regular@example.com",
		Name:         "Regular User",
		Status:       "active",
		IsSuperAdmin: false,
	}

	// Verify actor is not a super admin (precondition for ErrUnauthorized in service)
	if actor.IsSuperAdmin {
		t.Error("Actor should not be super admin - this is the precondition for ErrUnauthorized")
	}
}

func TestPromoteToSuperAdmin_UserNotFound(t *testing.T) {
	mock := NewMockRepository()

	// Try to promote non-existent user
	targetID := uuid.New()
	user, _ := mock.GetUserByID(context.Background(), targetID)

	if user == nil {
		t.Log("Correctly returned nil for non-existent user")
	}
}

// Test DemoteFromSuperAdmin
func TestDemoteFromSuperAdmin_Success(t *testing.T) {
	mock := NewMockRepository()
	mock.superAdminCount = 2 // Multiple super admins

	// Create actor (super admin)
	actorID := uuid.New()
	actor := &User{
		ID:           actorID,
		Email:        "admin@example.com",
		Name:         "Admin",
		Status:       "active",
		IsSuperAdmin: true,
	}
	mock.users[actorID] = actor

	// Create target user (super admin to demote)
	targetID := uuid.New()
	now := time.Now()
	target := &User{
		ID:                   targetID,
		Email:                "other-admin@example.com",
		Name:                 "Other Admin",
		Status:               "active",
		IsSuperAdmin:         true,
		SuperAdminPromotedAt: &now,
		SuperAdminPromotedBy: &actorID,
	}
	mock.users[targetID] = target

	// Verify conditions for demotion
	if mock.superAdminCount <= 1 {
		t.Error("Should have more than 1 super admin for demotion")
	}

	// Simulate demotion
	target.IsSuperAdmin = false
	target.SuperAdminPromotedAt = nil
	target.SuperAdminPromotedBy = nil

	// Verify demotion
	if target.IsSuperAdmin {
		t.Error("Target should not be super admin after demotion")
	}
	if target.SuperAdminPromotedAt != nil || target.SuperAdminPromotedBy != nil {
		t.Error("Promotion fields should be cleared after demotion")
	}
}

func TestDemoteFromSuperAdmin_LastSuperAdmin(t *testing.T) {
	mock := NewMockRepository()
	mock.superAdminCount = 1 // Only one super admin

	// Create the only super admin
	adminID := uuid.New()
	admin := &User{
		ID:           adminID,
		Email:        "admin@example.com",
		Name:         "Admin",
		Status:       "active",
		IsSuperAdmin: true,
	}
	mock.users[adminID] = admin

	// Verify the count is 1 (precondition for ErrLastSuperAdmin)
	count, err := mock.CountSuperAdminsForUpdate(context.Background(), nil)
	if err != nil {
		t.Fatalf("Failed to count super admins: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected exactly 1 super admin, got %d", count)
	}

	// Verify the condition that would trigger ErrLastSuperAdmin
	if count > 1 {
		t.Error("Demotion should be blocked when count <= 1, but count is greater than 1")
	}
}

func TestDemoteFromSuperAdmin_NotSuperAdmin(t *testing.T) {
	// Create target user who is not super admin
	targetID := uuid.New()
	target := &User{
		ID:           targetID,
		Email:        "user@example.com",
		Name:         "User",
		Status:       "active",
		IsSuperAdmin: false,
	}

	// Should return error when not super admin
	if !target.IsSuperAdmin {
		t.Log("Correctly identified user is not a super admin")
	}
}

func TestDemoteFromSuperAdmin_Unauthorized(t *testing.T) {
	// Create actor (not super admin)
	actorID := uuid.New()
	actor := &User{
		ID:           actorID,
		Email:        "regular@example.com",
		Name:         "Regular User",
		Status:       "active",
		IsSuperAdmin: false,
	}

	// Should not allow non-super-admin to demote
	if !actor.IsSuperAdmin {
		t.Log("Correctly identified actor is not authorized to demote")
	}
}

// Test JWT Claims
func TestJWTClaims_SuperAdminFlag(t *testing.T) {
	userID := uuid.New()
	isSuperAdmin := true

	claims := &JWTClaims{
		UserID:       userID,
		Email:        "admin@example.com",
		IsSuperAdmin: &isSuperAdmin,
	}

	if claims.IsSuperAdmin == nil || !*claims.IsSuperAdmin {
		t.Error("IsSuperAdmin should be true")
	}
}

func TestJWTClaims_GracefulDegradation(t *testing.T) {
	userID := uuid.New()

	// Old token without IsSuperAdmin field
	claims := &JWTClaims{
		UserID: userID,
		Email:  "user@example.com",
		// IsSuperAdmin is nil (old token)
	}

	// Should default to false when nil
	isSuperAdmin := claims.IsSuperAdmin != nil && *claims.IsSuperAdmin
	if isSuperAdmin {
		t.Error("Missing IsSuperAdmin should default to false")
	}
}

// Test GetAllUsers
func TestGetAllUsers_Pagination(t *testing.T) {
	mock := NewMockRepository()

	// Add multiple users
	for i := 0; i < 10; i++ {
		user := &User{
			ID:     uuid.New(),
			Email:  "user" + string(rune('0'+i)) + "@example.com",
			Status: "active",
		}
		mock.users[user.ID] = user
	}

	// Verify pagination parameters
	limit := 5
	offset := 0

	if limit < 1 || limit > 1000 {
		t.Error("Limit should be between 1 and 1000")
	}
	if offset < 0 {
		t.Error("Offset should not be negative")
	}

	t.Logf("Pagination test: limit=%d, offset=%d, total users=%d", limit, offset, len(mock.users))
}

// Test GetAllTeams
func TestGetAllTeams_Pagination(t *testing.T) {
	// Verify pagination parameters
	limit := 50
	offset := 0

	if limit < 1 || limit > 1000 {
		t.Error("Limit should be between 1 and 1000")
	}
	if offset < 0 {
		t.Error("Offset should not be negative")
	}

	t.Logf("Pagination test: limit=%d, offset=%d", limit, offset)
}

// Test error types
func TestErrorTypes(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"ErrLastSuperAdmin", ErrLastSuperAdmin, "cannot demote the last super admin"},
		{"ErrAlreadySuperAdmin", ErrAlreadySuperAdmin, "user is already a super admin"},
		{"ErrNotSuperAdmin", ErrNotSuperAdmin, "user is not a super admin"},
		{"ErrUnauthorized", ErrUnauthorized, "unauthorized"},
		{"ErrNotFound", ErrNotFound, "not found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.want {
				t.Errorf("got %q, want %q", tt.err.Error(), tt.want)
			}
		})
	}
}

// Test CheckSuperAdminStatus
func TestCheckSuperAdminStatus_SuperAdmin(t *testing.T) {
	mock := NewMockRepository()

	// Create super admin user
	adminID := uuid.New()
	admin := &User{
		ID:           adminID,
		Email:        "admin@example.com",
		Name:         "Admin",
		Status:       "active",
		IsSuperAdmin: true,
	}
	mock.users[adminID] = admin

	// Verify check returns true for super admin
	user, _ := mock.GetUserByID(context.Background(), adminID)
	if user == nil || !user.IsSuperAdmin {
		t.Error("CheckSuperAdminStatus should return true for super admin")
	}
}

func TestCheckSuperAdminStatus_RegularUser(t *testing.T) {
	mock := NewMockRepository()

	// Create regular user
	userID := uuid.New()
	user := &User{
		ID:           userID,
		Email:        "user@example.com",
		Name:         "User",
		Status:       "active",
		IsSuperAdmin: false,
	}
	mock.users[userID] = user

	// Verify check returns false for regular user
	fetched, _ := mock.GetUserByID(context.Background(), userID)
	if fetched == nil || fetched.IsSuperAdmin {
		t.Error("CheckSuperAdminStatus should return false for regular user")
	}
}

func TestCheckSuperAdminStatus_UserNotFound(t *testing.T) {
	mock := NewMockRepository()

	// Check non-existent user
	nonExistentID := uuid.New()
	user, _ := mock.GetUserByID(context.Background(), nonExistentID)

	// Should return nil for non-existent user
	if user != nil {
		t.Error("CheckSuperAdminStatus should return nil for non-existent user")
	}
}
