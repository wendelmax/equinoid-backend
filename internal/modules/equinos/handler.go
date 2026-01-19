package equinos

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/equinoid/backend/internal/middleware"
	"github.com/equinoid/backend/internal/models"
	apperrors "github.com/equinoid/backend/pkg/errors"
	"github.com/equinoid/backend/pkg/logging"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
	logger  *logging.Logger
}

func NewHandler(service Service, logger *logging.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// ListEquinos godoc
// @Summary Listar equinos
// @Description Lista todos os equinos com filtros e paginação
// @Tags Equinos
// @Produce json
// @Param page query int false "Página" default(1)
// @Param limit query int false "Itens por página" default(20)
// @Param search query string false "Buscar por nome ou Equinoid"
// @Param status query string false "Filtrar por status"
// @Param raca query string false "Filtrar por raça"
// @Param owner_id query int false "Filtrar por proprietário"
// @Param veterinario_id query int false "Filtrar por veterinário"
// @Success 200 {object} models.APIResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /equinos [get]
// @Security BearerAuth
func (h *Handler) ListEquinos(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")
	status := c.Query("status")
	raca := c.Query("raca")

	if limit > 100 {
		limit = 100
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success:   false,
			Error:     "Authentication required",
			Timestamp: time.Now(),
		})
		return
	}

	userType, _ := middleware.GetUserTypeFromContext(c)

	filters := make(map[string]interface{})
	if search != "" {
		filters["search"] = search
	}
	if status != "" {
		filters["status"] = status
	}
	if raca != "" {
		filters["raca"] = raca
	}
	if ownerID := c.Query("owner_id"); ownerID != "" {
		if id, err := strconv.ParseUint(ownerID, 10, 32); err == nil {
			filters["proprietario_id"] = uint(id)
		}
	}
	if vetID := c.Query("veterinario_id"); vetID != "" {
		if id, err := strconv.ParseUint(vetID, 10, 32); err == nil {
			filters["veterinario_id"] = uint(id)
		}
	}

	if userType == "criador" {
		filters["proprietario_id"] = userID
	}

	equinos, total, err := h.service.List(c.Request.Context(), page, limit, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao listar equinos",
			Timestamp: time.Now(),
		})
		return
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   fmt.Sprintf("Lista de equinos (total: %d)", total),
		Timestamp: time.Now(),
		Data: models.PaginatedResponse{
			Data: equinos,
			Pagination: &models.Pagination{
				Page:  page,
				Limit: limit,
				Total: total,
				Pages: totalPages,
			},
		},
	})
}

// GetEquino godoc
// @Summary Buscar equino por Equinoid
// @Description Retorna os dados completos de um equino
// @Tags Equinos
// @Produce json
// @Param equinoid path string true "Equinoid do equino"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /equinos/{equinoid} [get]
// @Security BearerAuth
func (h *Handler) GetEquino(c *gin.Context) {
	equinoid := c.Param("equinoid")

	equino, err := h.service.GetByEquinoid(c.Request.Context(), equinoid)
	if err != nil {
		if apperrors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Success:   false,
				Error:     "Equino não encontrado",
				Timestamp: time.Now(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao buscar equino",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   "Equino encontrado",
		Timestamp: time.Now(),
		Data:      equino,
	})
}

// CreateEquino godoc
// @Summary Criar novo equino
// @Description Registra um novo equino no sistema
// @Tags Equinos
// @Accept json
// @Produce json
// @Param equino body models.CreateEquinoRequest true "Dados do equino"
// @Success 201 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /equinos [post]
// @Security BearerAuth
func (h *Handler) CreateEquino(c *gin.Context) {
	var req models.CreateEquinoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Dados inválidos: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success:   false,
			Error:     "Authentication required",
			Timestamp: time.Now(),
		})
		return
	}

	equino, err := h.service.Create(c.Request.Context(), &req, userID)
	if err != nil {
		if apperrors.IsConflict(err) {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Success:   false,
				Error:     err.Error(),
				Timestamp: time.Now(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao criar equino",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success:   true,
		Message:   "Equino criado com sucesso",
		Timestamp: time.Now(),
		Data:      equino,
	})
}

// UpdateEquino godoc
// @Summary Atualizar equino
// @Description Atualiza os dados de um equino existente
// @Tags Equinos
// @Accept json
// @Produce json
// @Param equinoid path string true "Equinoid do equino"
// @Param equino body models.UpdateEquinoRequest true "Dados para atualização"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /equinos/{equinoid} [put]
// @Security BearerAuth
func (h *Handler) UpdateEquino(c *gin.Context) {
	equinoid := c.Param("equinoid")

	var req models.UpdateEquinoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Dados inválidos: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	equino, err := h.service.Update(c.Request.Context(), equinoid, &req)
	if err != nil {
		if apperrors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Success:   false,
				Error:     "Equino não encontrado",
				Timestamp: time.Now(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao atualizar equino",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   "Equino atualizado com sucesso",
		Timestamp: time.Now(),
		Data:      equino,
	})
}

// DeleteEquino godoc
// @Summary Deletar equino
// @Description Remove um equino do sistema
// @Tags Equinos
// @Produce json
// @Param equinoid path string true "Equinoid do equino"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /equinos/{equinoid} [delete]
// @Security BearerAuth
func (h *Handler) DeleteEquino(c *gin.Context) {
	equinoid := c.Param("equinoid")

	if err := h.service.Delete(c.Request.Context(), equinoid); err != nil {
		if apperrors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Success:   false,
				Error:     "Equino não encontrado",
				Timestamp: time.Now(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao deletar equino",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   "Equino deletado com sucesso",
		Timestamp: time.Now(),
		Data:      nil,
	})
}

func (h *Handler) TransferOwnership(c *gin.Context) {
	equinoid := c.Param("equinoid")

	var req struct {
		NovoProprietarioID uint `json:"novo_proprietario_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Dados inválidos: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	if err := h.service.TransferOwnership(c.Request.Context(), equinoid, req.NovoProprietarioID); err != nil {
		if apperrors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Success:   false,
				Error:     "Equino não encontrado",
				Timestamp: time.Now(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao transferir propriedade",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   "Propriedade transferida com sucesso",
		Timestamp: time.Now(),
		Data:      nil,
	})
}
