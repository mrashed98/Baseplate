package handlers

import (
	"net/http"

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
