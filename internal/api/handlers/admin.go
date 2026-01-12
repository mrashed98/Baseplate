package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/baseplate/baseplate/internal/api/middleware"
	"github.com/baseplate/baseplate/internal/core/auth"
)

type AdminHandler struct {
	authService *auth.Service
}

func NewAdminHandler(authService *auth.Service) *AdminHandler {
	return &AdminHandler{authService: authService}
}

// ListTeams returns all teams in the system (super admin only)
func (h *AdminHandler) ListTeams(c *gin.Context) {
	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 1000 {
			limit = parsed
		}
	}

	offset := 0
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	teams, err := h.authService.GetAllTeams(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"teams":  teams,
		"limit":  limit,
		"offset": offset,
	})
}

// GetTeamDetail returns details for a specific team (super admin only)
func (h *AdminHandler) GetTeamDetail(c *gin.Context) {
	teamIDStr := c.Param("teamId")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team id"})
		return
	}

	team, err := h.authService.GetTeam(c.Request.Context(), teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if team == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "team not found"})
		return
	}

	c.JSON(http.StatusOK, team)
}

// ListUsers returns all users in the system with pagination (super admin only)
func (h *AdminHandler) ListUsers(c *gin.Context) {
	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 500 {
			limit = parsed
		}
	}

	offset := 0
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	users, err := h.authService.GetAllUsers(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users":  users,
		"limit":  limit,
		"offset": offset,
	})
}

// GetUserDetail returns details for a specific user with their team memberships (super admin only)
func (h *AdminHandler) GetUserDetail(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	user, memberships, err := h.authService.GetUserDetail(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":        user,
		"memberships": memberships,
	})
}

// UpdateUser updates a user's information (super admin only)
func (h *AdminHandler) UpdateUser(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Update allowed fields
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Status != "" {
		// Validate status value
		if req.Status != "active" && req.Status != "deleted" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status value, must be 'active' or 'deleted'"})
			return
		}
		user.Status = req.Status
	}

	if err := h.authService.UpdateUser(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// PromoteUser promotes a user to super admin (super admin only)
func (h *AdminHandler) PromoteUser(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	// Get actor from context
	actorID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user id"})
		return
	}

	// Get audit context
	ipAddress := middleware.GetIPAddress(c)
	userAgent := middleware.GetUserAgent(c)
	var ipPtr, uaPtr *string
	if ipAddress != "" {
		ipPtr = &ipAddress
	}
	if userAgent != "" {
		uaPtr = &userAgent
	}

	user, err := h.authService.PromoteToSuperAdmin(c.Request.Context(), actorID, userID, ipPtr, uaPtr)
	if err != nil {
		if err == auth.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"error": "only super admins can promote users"})
			return
		}
		if err == auth.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		if err == auth.ErrAlreadySuperAdmin {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user is already a super admin"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// DemoteUser demotes a user from super admin (super admin only)
func (h *AdminHandler) DemoteUser(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	// Get actor from context
	actorID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user id"})
		return
	}

	// Get audit context
	ipAddress := middleware.GetIPAddress(c)
	userAgent := middleware.GetUserAgent(c)
	var ipPtr, uaPtr *string
	if ipAddress != "" {
		ipPtr = &ipAddress
	}
	if userAgent != "" {
		uaPtr = &userAgent
	}

	user, err := h.authService.DemoteFromSuperAdmin(c.Request.Context(), actorID, userID, ipPtr, uaPtr)
	if err != nil {
		if err == auth.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"error": "only super admins can demote users"})
			return
		}
		if err == auth.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		if err == auth.ErrNotSuperAdmin {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user is not a super admin"})
			return
		}
		if err == auth.ErrLastSuperAdmin {
			c.JSON(http.StatusConflict, gin.H{"error": "cannot demote the last super admin"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// QueryAuditLogs returns audit logs for super admin actions (super admin only)
func (h *AdminHandler) QueryAuditLogs(c *gin.Context) {
	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 500 {
			limit = parsed
		}
	}

	offset := 0
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	logs, err := h.authService.GetSuperAdminAuditLogs(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":   logs,
		"limit":  limit,
		"offset": offset,
	})
}

type UpdateUserRequest struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}
