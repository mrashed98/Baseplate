package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/baseplate/baseplate/internal/api/middleware"
	"github.com/baseplate/baseplate/internal/core/entity"
	"github.com/baseplate/baseplate/internal/core/validation"
)

type EntityHandler struct {
	entityService *entity.Service
}

func NewEntityHandler(entityService *entity.Service) *EntityHandler {
	return &EntityHandler{entityService: entityService}
}

func (h *EntityHandler) Create(c *gin.Context) {
	teamID, ok := middleware.GetTeamID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team id required"})
		return
	}

	blueprintID := c.Param("blueprintId")

	var req entity.CreateEntityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ent, err := h.entityService.Create(c.Request.Context(), teamID, blueprintID, &req)
	if err != nil {
		if errors.Is(err, entity.ErrAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, entity.ErrBlueprintNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if validation.IsValidationError(err) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "details": validation.GetValidationErrors(err)})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ent)
}

func (h *EntityHandler) List(c *gin.Context) {
	teamID, ok := middleware.GetTeamID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team id required"})
		return
	}

	blueprintID := c.Param("blueprintId")

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	resp, err := h.entityService.List(c.Request.Context(), teamID, blueprintID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *EntityHandler) Search(c *gin.Context) {
	teamID, ok := middleware.GetTeamID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team id required"})
		return
	}

	blueprintID := c.Param("blueprintId")

	var req entity.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.entityService.Search(c.Request.Context(), teamID, blueprintID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *EntityHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid entity id"})
		return
	}

	ent, err := h.entityService.Get(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ent)
}

func (h *EntityHandler) GetByIdentifier(c *gin.Context) {
	teamID, ok := middleware.GetTeamID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team id required"})
		return
	}

	blueprintID := c.Param("blueprintId")
	identifier := c.Param("identifier")

	ent, err := h.entityService.GetByIdentifier(c.Request.Context(), teamID, blueprintID, identifier)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ent)
}

func (h *EntityHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid entity id"})
		return
	}

	var req entity.UpdateEntityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ent, err := h.entityService.Update(c.Request.Context(), id, &req)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if validation.IsValidationError(err) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "details": validation.GetValidationErrors(err)})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ent)
}

func (h *EntityHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid entity id"})
		return
	}

	if err := h.entityService.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
