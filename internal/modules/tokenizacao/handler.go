package tokenizacao

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

func (h *Handler) ListAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if limit > 100 {
		limit = 100
	}

	filters := make(map[string]interface{})
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if rating := c.Query("rating"); rating != "" {
		filters["rating"] = rating
	}

	tokenizacoes, total, err := h.service.ListAll(c.Request.Context(), page, limit, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao listar tokenizações",
			Timestamp: time.Now(),
		})
		return
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   fmt.Sprintf("Lista de tokenizações (total: %d)", total),
		Timestamp: time.Now(),
		Data: models.PaginatedResponse{
			Data: tokenizacoes,
			Pagination: &models.Pagination{
				Page:  page,
				Limit: limit,
				Total: total,
				Pages: totalPages,
			},
		},
	})
}

func (h *Handler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "ID de tokenização inválido",
			Timestamp: time.Now(),
		})
		return
	}

	tokenizacao, err := h.service.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		if apperrors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Success:   false,
				Error:     "Tokenização não encontrada",
				Timestamp: time.Now(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao buscar tokenização",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   "Tokenização encontrada",
		Timestamp: time.Now(),
		Data:      tokenizacao,
	})
}

func (h *Handler) GetByEquinoid(c *gin.Context) {
	equinoid := c.Param("equinoid")

	tokenizacao, err := h.service.GetByEquinoid(c.Request.Context(), equinoid)
	if err != nil {
		if apperrors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Success:   false,
				Error:     "Tokenização não encontrada para este equino",
				Timestamp: time.Now(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao buscar tokenização",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   "Tokenização do equino",
		Timestamp: time.Now(),
		Data:      tokenizacao,
	})
}

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

	var req models.CreateTokenizacaoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Dados inválidos: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	tokenizacao, err := h.service.Create(c.Request.Context(), userID, &req)
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
			Error:     "Erro ao criar tokenização",
			Timestamp: time.Now(),
		})
		return
	}

	h.logger.WithFields(logging.Fields{
		"tokenizacao_id": tokenizacao.ID,
		"equinoid":       tokenizacao.Equinoid,
		"user_id":        userID,
	}).Info("Tokenização criada via API")

	c.JSON(http.StatusCreated, models.APIResponse{
		Success:   true,
		Message:   "Tokenização criada com sucesso! Equino agora é um ativo tokenizado (RWA).",
		Timestamp: time.Now(),
		Data:      tokenizacao,
	})
}

func (h *Handler) ListTransacoes(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "ID de tokenização inválido",
			Timestamp: time.Now(),
		})
		return
	}

	transacoes, err := h.service.ListTransacoes(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao listar transações",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   fmt.Sprintf("Histórico de transações (total: %d)", len(transacoes)),
		Timestamp: time.Now(),
		Data:      transacoes,
	})
}

func (h *Handler) ExecutarOrdem(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success:   false,
			Error:     "Authentication required",
			Timestamp: time.Now(),
		})
		return
	}

	var req models.OrdemCompraTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Dados inválidos: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	transacao, err := h.service.ExecutarOrdem(c.Request.Context(), userID, &req)
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
			Error:     "Erro ao executar ordem de compra",
			Timestamp: time.Now(),
		})
		return
	}

	h.logger.WithFields(logging.Fields{
		"transacao_id": transacao.ID,
		"comprador_id": userID,
		"quantidade":   transacao.Quantidade,
	}).Info("Ordem de compra executada via API")

	c.JSON(http.StatusCreated, models.APIResponse{
		Success:   true,
		Message:   "Ordem executada com sucesso! Tokens adquiridos.",
		Timestamp: time.Now(),
		Data:      transacao,
	})
}

func (h *Handler) CriarOferta(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success:   false,
			Error:     "Authentication required",
			Timestamp: time.Now(),
		})
		return
	}

	var req models.OfertaTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Dados inválidos: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	if err := h.service.CriarOferta(c.Request.Context(), userID, &req); err != nil {
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
			Error:     "Erro ao criar oferta",
			Timestamp: time.Now(),
		})
		return
	}

	h.logger.WithFields(logging.Fields{
		"tokenizacao_id": req.TokenizacaoID,
		"vendedor_id":    userID,
	}).Info("Oferta de venda criada via API")

	c.JSON(http.StatusCreated, models.APIResponse{
		Success:   true,
		Message:   "Oferta criada com sucesso!",
		Timestamp: time.Now(),
		Data:      nil,
	})
}
