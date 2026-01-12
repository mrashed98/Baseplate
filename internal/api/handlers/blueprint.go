package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/baseplate/baseplate/internal/api/middleware"
	"github.com/baseplate/baseplate/internal/core/blueprint"
)

type BlueprintHandler struct {
	blueprintService *blueprint.Service
}

func NewBlueprintHandler(blueprintService *blueprint.Service) *BlueprintHandler {
	return &BlueprintHandler{blueprintService: blueprintService}
}

func (h *BlueprintHandler) Create(c *gin.Context) {
	teamID, ok := middleware.GetTeamID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team id required"})
		return
	}

	var req blueprint.CreateBlueprintRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bp, err := h.blueprintService.Create(c.Request.Context(), teamID, &req)
	if err != nil {
		if errors.Is(err, blueprint.ErrAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, bp)
}

func (h *BlueprintHandler) List(c *gin.Context) {
	teamID, ok := middleware.GetTeamID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team id required"})
		return
	}

	resp, err := h.blueprintService.List(c.Request.Context(), teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *BlueprintHandler) Get(c *gin.Context) {
	teamID, ok := middleware.GetTeamID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team id required"})
		return
	}

	id := c.Param("id")
	bp, err := h.blueprintService.Get(c.Request.Context(), teamID, id)
	if err != nil {
		if errors.Is(err, blueprint.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, bp)
}

func (h *BlueprintHandler) Update(c *gin.Context) {
	teamID, ok := middleware.GetTeamID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team id required"})
		return
	}

	id := c.Param("id")

	var req blueprint.UpdateBlueprintRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bp, err := h.blueprintService.Update(c.Request.Context(), teamID, id, &req)
	if err != nil {
		if errors.Is(err, blueprint.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, bp)
}

func (h *BlueprintHandler) Delete(c *gin.Context) {
	teamID, ok := middleware.GetTeamID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team id required"})
		return
	}

	id := c.Param("id")

	if err := h.blueprintService.Delete(c.Request.Context(), teamID, id); err != nil {
		if errors.Is(err, blueprint.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
