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
	teams, err := h.authService.GetAllTeams(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"teams": teams})
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
		user.Status = req.Status
	}

	if err := h.authService.UpdateUser(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

type UpdateUserRequest struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}
