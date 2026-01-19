package eventos

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

// ListAll godoc
// @Summary Listar eventos
// @Description Lista todos os eventos com paginação e filtros
// @Tags Eventos
// @Produce json
// @Param page query int false "Página" default(1)
// @Param limit query int false "Itens por página" default(20)
// @Param categoria query string false "Filtrar por categoria"
// @Param tipo_evento query string false "Filtrar por tipo"
// @Param data_inicio query string false "Data início"
// @Param data_fim query string false "Data fim"
// @Success 200 {object} models.APIResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /eventos [get]
// @Security BearerAuth
func (h *Handler) ListAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if limit > 100 {
		limit = 100
	}

	filters := make(map[string]interface{})
	if categoria := c.Query("categoria"); categoria != "" {
		filters["categoria"] = categoria
	}
	if tipoEvento := c.Query("tipo_evento"); tipoEvento != "" {
		filters["tipo_evento"] = tipoEvento
	}
	if dataInicio := c.Query("data_inicio"); dataInicio != "" {
		filters["data_inicio"] = dataInicio
	}
	if dataFim := c.Query("data_fim"); dataFim != "" {
		filters["data_fim"] = dataFim
	}

	eventos, total, err := h.service.ListAll(c.Request.Context(), page, limit, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao listar eventos",
			Timestamp: time.Now(),
		})
		return
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   fmt.Sprintf("Lista de eventos (total: %d)", total),
		Timestamp: time.Now(),
		Data: models.PaginatedResponse{
			Data: eventos,
			Pagination: &models.Pagination{
				Page:  page,
				Limit: limit,
				Total: total,
				Pages: totalPages,
			},
		},
	})
}

// Create godoc
// @Summary Criar evento
// @Description Cria um novo evento
// @Tags Eventos
// @Accept json
// @Produce json
// @Param evento body models.CreateEventoRequest true "Dados do evento"
// @Success 201 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /eventos [post]
// @Security BearerAuth
func (h *Handler) Create(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success:   false,
			Error:     "Authentication required",
			Timestamp: time.Now(),
		})
		return
	}

	var req models.CreateEventoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Dados inválidos: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	evento, err := h.service.Create(c.Request.Context(), userID, &req)
	if err != nil {
		if apperrors.IsValidation(err) {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Success:   false,
				Error:     err.Error(),
				Timestamp: time.Now(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao criar evento",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success:   true,
		Message:   "Evento criado com sucesso",
		Timestamp: time.Now(),
		Data:      evento,
	})
}

// GetByID godoc
// @Summary Buscar evento por ID
// @Description Retorna os dados de um evento específico
// @Tags Eventos
// @Produce json
// @Param evento_id path int true "ID do evento"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /eventos/{evento_id} [get]
// @Security BearerAuth
func (h *Handler) GetByID(c *gin.Context) {
	idStr := c.Param("evento_id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "ID de evento inválido",
			Timestamp: time.Now(),
		})
		return
	}

	evento, err := h.service.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		if apperrors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Success:   false,
				Error:     "Evento não encontrado",
				Timestamp: time.Now(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao buscar evento",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   "Evento encontrado",
		Timestamp: time.Now(),
		Data:      evento,
	})
}

// Update godoc
// @Summary Atualizar evento
// @Description Atualiza os dados de um evento
// @Tags Eventos
// @Accept json
// @Produce json
// @Param evento_id path int true "ID do evento"
// @Param evento body object true "Dados para atualização"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /eventos/{evento_id} [put]
// @Security BearerAuth
func (h *Handler) Update(c *gin.Context) {
	idStr := c.Param("evento_id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "ID de evento inválido",
			Timestamp: time.Now(),
		})
		return
	}

	var req models.CreateEventoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Dados inválidos: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	evento, err := h.service.Update(c.Request.Context(), uint(id), &req)
	if err != nil {
		if apperrors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Success:   false,
				Error:     "Evento não encontrado",
				Timestamp: time.Now(),
			})
			return
		}
		if apperrors.IsValidation(err) {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Success:   false,
				Error:     err.Error(),
				Timestamp: time.Now(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao atualizar evento",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   "Evento atualizado com sucesso",
		Timestamp: time.Now(),
		Data:      evento,
	})
}

// Delete godoc
// @Summary Deletar evento
// @Description Remove um evento do sistema
// @Tags Eventos
// @Produce json
// @Param evento_id path int true "ID do evento"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /eventos/{evento_id} [delete]
// @Security BearerAuth
func (h *Handler) Delete(c *gin.Context) {
	idStr := c.Param("evento_id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "ID de evento inválido",
			Timestamp: time.Now(),
		})
		return
	}

	if err := h.service.Delete(c.Request.Context(), uint(id)); err != nil {
		if apperrors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Success:   false,
				Error:     "Evento não encontrado",
				Timestamp: time.Now(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao deletar evento",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   "Evento deletado com sucesso",
		Timestamp: time.Now(),
		Data:      nil,
	})
}

// ListByEquino godoc
// @Summary Listar eventos de um equino
// @Description Retorna todos os eventos de um equino específico
// @Tags Eventos
// @Produce json
// @Param equinoid path string true "Equinoid do equino"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /equinos/{equinoid}/eventos [get]
// @Security BearerAuth
func (h *Handler) ListByEquino(c *gin.Context) {
	equinoid := c.Param("equinoid")

	eventos, err := h.service.ListByEquino(c.Request.Context(), equinoid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao listar eventos do equino",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   fmt.Sprintf("Eventos do equino %s (total: %d)", equinoid, len(eventos)),
		Timestamp: time.Now(),
		Data:      eventos,
	})
}
