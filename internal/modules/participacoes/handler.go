package participacoes

import (
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

// ListByEvento godoc
// @Summary Listar participações de um evento
// @Description Retorna todas as participações de um evento
// @Tags Participações
// @Produce json
// @Param evento_id path int true "ID do evento"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /eventos/{evento_id}/participacoes [get]
// @Security BearerAuth
func (h *Handler) ListByEvento(c *gin.Context) {
	eventoID, err := strconv.ParseUint(c.Param("evento_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "ID de evento inválido",
			Timestamp: time.Now(),
		})
		return
	}

	participacoes, err := h.service.List(c.Request.Context(), uint(eventoID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao listar participações",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   "Participações do evento",
		Timestamp: time.Now(),
		Data:      participacoes,
	})
}

// ListByEquino godoc
// @Summary Listar participações de um equino
// @Description Retorna todas as participações em eventos de um equino
// @Tags Participações
// @Produce json
// @Param equinoid path string true "Equinoid do equino"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /equinos/{equinoid}/participacoes-eventos [get]
// @Security BearerAuth
func (h *Handler) ListByEquino(c *gin.Context) {
	equinoID, err := strconv.ParseUint(c.Param("equinoid"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "ID de equino inválido",
			Timestamp: time.Now(),
		})
		return
	}

	participacoes, err := h.service.ListByEquino(c.Request.Context(), uint(equinoID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao listar participações",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   "Participações do equino",
		Timestamp: time.Now(),
		Data:      participacoes,
	})
}

// Create godoc
// @Summary Criar participação em evento
// @Description Cria uma nova participação em um evento
// @Tags Participações
// @Accept json
// @Produce json
// @Param participacao body models.CreateParticipacaoEventoRequest true "Dados da participação"
// @Success 201 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /eventos/participacoes [post]
// @Security BearerAuth
func (h *Handler) Create(c *gin.Context) {
	var req models.CreateParticipacaoEventoRequest
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

	participacao, err := h.service.Create(c.Request.Context(), &req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao criar participação",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success:   true,
		Message:   "Participação criada com sucesso",
		Timestamp: time.Now(),
		Data:      participacao,
	})
}

// Update godoc
// @Summary Atualizar participação
// @Description Atualiza os dados de uma participação em evento
// @Tags Participações
// @Accept json
// @Produce json
// @Param id path int true "ID da participação"
// @Param participacao body models.UpdateParticipacaoEventoRequest true "Dados para atualização"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /eventos/participacoes/{id} [put]
// @Security BearerAuth
func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "ID inválido",
			Timestamp: time.Now(),
		})
		return
	}

	var req models.UpdateParticipacaoEventoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Dados inválidos: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	participacao, err := h.service.Update(c.Request.Context(), uint(id), &req)
	if err != nil {
		if apperrors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Success:   false,
				Error:     "Participação não encontrada",
				Timestamp: time.Now(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao atualizar participação",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   "Participação atualizada com sucesso",
		Timestamp: time.Now(),
		Data:      participacao,
	})
}

// Delete godoc
// @Summary Deletar participação
// @Description Remove uma participação em evento
// @Tags Participações
// @Produce json
// @Param id path int true "ID da participação"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /eventos/participacoes/{id} [delete]
// @Security BearerAuth
func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "ID inválido",
			Timestamp: time.Now(),
		})
		return
	}

	if err := h.service.Delete(c.Request.Context(), uint(id)); err != nil {
		if apperrors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Success:   false,
				Error:     "Participação não encontrada",
				Timestamp: time.Now(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao deletar participação",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   "Participação deletada com sucesso",
		Timestamp: time.Now(),
		Data:      nil,
	})
}

// MarcarAusencia godoc
// @Summary Marcar ausência em evento
// @Description Marca um equino como ausente em um evento
// @Tags Participações
// @Produce json
// @Param id path int true "ID da participação"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /eventos/participacoes/{id}/ausencia [post]
// @Security BearerAuth
func (h *Handler) MarcarAusencia(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "ID inválido",
			Timestamp: time.Now(),
		})
		return
	}

	participacao, err := h.service.MarcarAusencia(c.Request.Context(), uint(id))
	if err != nil {
		if apperrors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Success:   false,
				Error:     "Participação não encontrada",
				Timestamp: time.Now(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao marcar ausência",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   "Ausência marcada com penalização de -50 pontos",
		Timestamp: time.Now(),
		Data:      participacao,
	})
}

// MarcarPresenca godoc
// @Summary Marcar presença em evento
// @Description Marca um equino como presente em um evento
// @Tags Participações
// @Produce json
// @Param id path int true "ID da participação"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /eventos/participacoes/{id}/presenca [post]
// @Security BearerAuth
func (h *Handler) MarcarPresenca(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "ID inválido",
			Timestamp: time.Now(),
		})
		return
	}

	participacao, err := h.service.MarcarPresenca(c.Request.Context(), uint(id))
	if err != nil {
		if apperrors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Success:   false,
				Error:     "Participação não encontrada",
				Timestamp: time.Now(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao marcar presença",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   "Presença confirmada",
		Timestamp: time.Now(),
		Data:      participacao,
	})
}
