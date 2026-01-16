package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/equinoid/backend/internal/middleware"
	"github.com/equinoid/backend/internal/models"
	apperrors "github.com/equinoid/backend/pkg/errors"
	"github.com/gin-gonic/gin"
)

// ListPropriedades lista as propriedades do usuário ou filtradas
func (h *Handlers) ListPropriedades(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	filters := make(map[string]interface{})
	if tipo := c.Query("tipo"); tipo != "" {
		filters["tipo"] = tipo
	}

	// Obter ID do usuário do contexto
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success:   false,
			Error:     "Authentication required",
			Timestamp: time.Now(),
		})
		return
	}

	// Se não for admin, filtra pelo responsavel_id
	userType, _ := middleware.GetUserTypeFromContext(c)
	if userType != "admin" {
		filters["responsavel_id"] = userID
	}

	propriedades, total, err := h.PropriedadeService.List(c.Request.Context(), page, limit, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	responses := make([]*models.PropriedadeResponse, len(propriedades))
	for i, p := range propriedades {
		responses[i] = p.ToResponse()
	}

	pages := int(total) / limit
	if int(total)%limit != 0 {
		pages++
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: gin.H{
			"items": responses,
			"total": total,
			"page":  page,
			"limit": limit,
			"pages": pages,
		},
		Timestamp: time.Now(),
	})
}

// GetPropriedade busca uma propriedade pelo ID
func (h *Handlers) GetPropriedade(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "ID inválido",
			Timestamp: time.Now(),
		})
		return
	}

	propriedade, err := h.PropriedadeService.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		if apperrors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Success:   false,
				Error:     "Propriedade não encontrada",
				Timestamp: time.Now(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      propriedade.ToResponse(),
		Timestamp: time.Now(),
	})
}

// CreatePropriedade cria uma nova propriedade
func (h *Handlers) CreatePropriedade(c *gin.Context) {
	var req models.CreatePropriedadeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Dados inválidos",
			Timestamp: time.Now(),
		})
		return
	}

	userID, _ := middleware.GetUserIDFromContext(c)

	propriedade, err := h.PropriedadeService.Create(c.Request.Context(), &req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success:   true,
		Data:      propriedade.ToResponse(),
		Timestamp: time.Now(),
	})
}

// UpdatePropriedade atualiza uma propriedade
func (h *Handlers) UpdatePropriedade(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "ID inválido",
			Timestamp: time.Now(),
		})
		return
	}

	var req models.UpdatePropriedadeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Dados inválidos",
			Timestamp: time.Now(),
		})
		return
	}

	propriedade, err := h.PropriedadeService.Update(c.Request.Context(), uint(id), &req)
	if err != nil {
		if apperrors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Success:   false,
				Error:     "Propriedade não encontrada",
				Timestamp: time.Now(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      propriedade.ToResponse(),
		Timestamp: time.Now(),
	})
}

// DeletePropriedade exclui uma propriedade
func (h *Handlers) DeletePropriedade(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "ID inválido",
			Timestamp: time.Now(),
		})
		return
	}

	if err := h.PropriedadeService.Delete(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   "Propriedade excluída com sucesso",
		Timestamp: time.Now(),
	})
}
