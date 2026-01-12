package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/baseplate/baseplate/internal/api/middleware"
	"github.com/baseplate/baseplate/internal/core/auth"
)

type TeamHandler struct {
	authService *auth.Service
}

func NewTeamHandler(authService *auth.Service) *TeamHandler {
	return &TeamHandler{authService: authService}
}

func (h *TeamHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req auth.CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	team, err := h.authService.CreateTeam(c.Request.Context(), userID, &req)
	if err != nil {
		if errors.Is(err, auth.ErrTeamExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, team)
}

func (h *TeamHandler) List(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	teams, err := h.authService.GetTeamsByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"teams": teams})
}

func (h *TeamHandler) Get(c *gin.Context) {
	teamID, ok := middleware.GetTeamID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team id required"})
		return
	}

	team, err := h.authService.GetTeam(c.Request.Context(), teamID)
	if err != nil {
		if errors.Is(err, auth.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, team)
}

func (h *TeamHandler) Update(c *gin.Context) {
	teamID, ok := middleware.GetTeamID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team id required"})
		return
	}

	team, err := h.authService.GetTeam(c.Request.Context(), teamID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "team not found"})
		return
	}

	var req auth.CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	team.Name = req.Name
	team.Slug = req.Slug

	if err := h.authService.UpdateTeam(c.Request.Context(), team); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, team)
}

func (h *TeamHandler) Delete(c *gin.Context) {
	teamID, ok := middleware.GetTeamID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team id required"})
		return
	}

	if err := h.authService.DeleteTeam(c.Request.Context(), teamID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// Role endpoints
func (h *TeamHandler) ListRoles(c *gin.Context) {
	teamID, ok := middleware.GetTeamID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team id required"})
		return
	}

	roles, err := h.authService.GetRoles(c.Request.Context(), teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"roles": roles})
}

func (h *TeamHandler) CreateRole(c *gin.Context) {
	teamID, ok := middleware.GetTeamID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team id required"})
		return
	}

	var req struct {
		Name        string   `json:"name" binding:"required"`
		Permissions []string `json:"permissions" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role, err := h.authService.CreateRole(c.Request.Context(), teamID, req.Name, req.Permissions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, role)
}

// Member endpoints
func (h *TeamHandler) ListMembers(c *gin.Context) {
	teamID, ok := middleware.GetTeamID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team id required"})
		return
	}

	memberships, err := h.authService.GetMemberships(c.Request.Context(), teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"members": memberships})
}

func (h *TeamHandler) AddMember(c *gin.Context) {
	teamID, ok := middleware.GetTeamID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team id required"})
		return
	}

	var req auth.InviteMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role id"})
		return
	}

	membership, err := h.authService.AddMember(c.Request.Context(), teamID, req.Email, roleID)
	if err != nil {
		if errors.Is(err, auth.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, membership)
}

func (h *TeamHandler) RemoveMember(c *gin.Context) {
	teamID, ok := middleware.GetTeamID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team id required"})
		return
	}

	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	if err := h.authService.RemoveMember(c.Request.Context(), teamID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// API Key endpoints
func (h *TeamHandler) ListAPIKeys(c *gin.Context) {
	teamID, ok := middleware.GetTeamID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team id required"})
		return
	}

	keys, err := h.authService.GetAPIKeys(c.Request.Context(), teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"api_keys": keys})
}

func (h *TeamHandler) CreateAPIKey(c *gin.Context) {
	teamID, ok := middleware.GetTeamID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team id required"})
		return
	}

	userID, _ := middleware.GetUserID(c)

	var req auth.CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authService.CreateAPIKey(c.Request.Context(), teamID, &userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *TeamHandler) DeleteAPIKey(c *gin.Context) {
	keyIDStr := c.Param("keyId")
	keyID, err := uuid.Parse(keyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid key id"})
		return
	}

	if err := h.authService.DeleteAPIKey(c.Request.Context(), keyID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
