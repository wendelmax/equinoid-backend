package leiloes

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

// ListParticipacoes godoc
// @Summary Listar participações de um leilão
// @Description Retorna todas as participações de um leilão específico
// @Tags Leilões
// @Produce json
// @Param leilao_id path int true "ID do leilão"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /leiloes/{leilao_id}/participacoes [get]
// @Security BearerAuth
func (h *Handler) ListParticipacoes(c *gin.Context) {
	idStr := c.Param("leilao_id")
	leilaoID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "ID de leilão inválido",
			Timestamp: time.Now(),
		})
		return
	}

	participacoes, err := h.service.ListParticipacoes(c.Request.Context(), uint(leilaoID))
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
		Message:   fmt.Sprintf("Participações do leilão (total: %d)", len(participacoes)),
		Timestamp: time.Now(),
		Data:      participacoes,
	})
}

// CriarParticipacao godoc
// @Summary Criar participação em leilão
// @Description Cria uma nova participação em um leilão
// @Tags Leilões
// @Accept json
// @Produce json
// @Param leilao_id path int true "ID do leilão"
// @Param participacao body models.CreateParticipacaoLeilaoRequest true "Dados da participação"
// @Success 201 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /leiloes/{leilao_id}/participacoes [post]
// @Security BearerAuth
func (h *Handler) CriarParticipacao(c *gin.Context) {
	criadorID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success:   false,
			Error:     "Authentication required",
			Timestamp: time.Now(),
		})
		return
	}

	idStr := c.Param("leilao_id")
	leilaoID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "ID de leilão inválido",
			Timestamp: time.Now(),
		})
		return
	}

	var req models.CreateParticipacaoLeilaoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Dados inválidos: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	participacao, err := h.service.CriarParticipacao(c.Request.Context(), uint(leilaoID), criadorID, &req)
	if err != nil {
		if apperrors.IsValidation(err) {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Success:   false,
				Error:     err.Error(),
				Timestamp: time.Now(),
			})
			return
		}
		if apperrors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Success:   false,
				Error:     err.Error(),
				Timestamp: time.Now(),
			})
			return
		}
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

// AprovarParticipacao godoc
// @Summary Aprovar participação em leilão
// @Description Aprova uma participação pendente em um leilão
// @Tags Leilões
// @Produce json
// @Param id path int true "ID da participação"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /leiloes/participacoes/{id}/aprovar [post]
// @Security BearerAuth
func (h *Handler) AprovarParticipacao(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "ID de participação inválido",
			Timestamp: time.Now(),
		})
		return
	}

	participacao, err := h.service.AprovarParticipacao(c.Request.Context(), uint(id))
	if err != nil {
		if apperrors.IsValidation(err) {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Success:   false,
				Error:     err.Error(),
				Timestamp: time.Now(),
			})
			return
		}
		if apperrors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Success:   false,
				Error:     err.Error(),
				Timestamp: time.Now(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao aprovar participação",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   "Participação aprovada com sucesso",
		Timestamp: time.Now(),
		Data:      participacao,
	})
}

// RegistrarVenda godoc
// @Summary Registrar venda em participação
// @Description Registra a venda de um equino em uma participação de leilão
// @Tags Leilões
// @Accept json
// @Produce json
// @Param id path int true "ID da participação"
// @Param venda body models.RegistrarVendaRequest true "Dados da venda"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /leiloes/participacoes/{id}/venda [post]
// @Security BearerAuth
func (h *Handler) RegistrarVenda(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "ID de participação inválido",
			Timestamp: time.Now(),
		})
		return
	}

	var req models.RegistrarVendaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Dados inválidos: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	participacao, err := h.service.RegistrarVenda(c.Request.Context(), uint(id), &req)
	if err != nil {
		if apperrors.IsValidation(err) {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Success:   false,
				Error:     err.Error(),
				Timestamp: time.Now(),
			})
			return
		}
		if apperrors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Success:   false,
				Error:     err.Error(),
				Timestamp: time.Now(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao registrar venda",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   "Venda registrada com sucesso",
		Timestamp: time.Now(),
		Data:      participacao,
	})
}

// MarcarAusencia godoc
// @Summary Marcar ausência em leilão
// @Description Marca um participante como ausente no leilão
// @Tags Leilões
// @Produce json
// @Param id path int true "ID da participação"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /leiloes/participacoes/{id}/ausencia [post]
// @Security BearerAuth
func (h *Handler) MarcarAusencia(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "ID de participação inválido",
			Timestamp: time.Now(),
		})
		return
	}

	participacao, err := h.service.MarcarAusencia(c.Request.Context(), uint(id))
	if err != nil {
		if apperrors.IsValidation(err) {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Success:   false,
				Error:     err.Error(),
				Timestamp: time.Now(),
			})
			return
		}
		if apperrors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Success:   false,
				Error:     err.Error(),
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
// @Summary Marcar presença em leilão
// @Description Marca um participante como presente no leilão
// @Tags Leilões
// @Produce json
// @Param id path int true "ID da participação"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /leiloes/participacoes/{id}/presenca [post]
// @Security BearerAuth
func (h *Handler) MarcarPresenca(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "ID de participação inválido",
			Timestamp: time.Now(),
		})
		return
	}

	participacao, err := h.service.MarcarPresenca(c.Request.Context(), uint(id))
	if err != nil {
		if apperrors.IsValidation(err) {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Success:   false,
				Error:     err.Error(),
				Timestamp: time.Now(),
			})
			return
		}
		if apperrors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Success:   false,
				Error:     err.Error(),
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
