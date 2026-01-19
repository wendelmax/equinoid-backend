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

// ListAll godoc
// @Summary Listar tokenizações
// @Description Lista todas as tokenizações RWA com paginação e filtros
// @Tags Tokenização
// @Produce json
// @Param page query int false "Página" default(1)
// @Param limit query int false "Itens por página" default(20)
// @Param status query string false "Filtrar por status"
// @Param rating query string false "Filtrar por rating"
// @Success 200 {object} models.APIResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /tokenizacao [get]
// @Security BearerAuth
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

// GetByID godoc
// @Summary Buscar tokenização por ID
// @Description Retorna os dados completos de uma tokenização RWA
// @Tags Tokenização
// @Produce json
// @Param id path int true "ID da tokenização"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /tokenizacao/{id} [get]
// @Security BearerAuth
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

// GetByEquinoid godoc
// @Summary Buscar tokenização por Equinoid
// @Description Retorna a tokenização de um equino específico
// @Tags Tokenização
// @Produce json
// @Param equinoid path string true "Equinoid do equino"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /tokenizacao/equino/{equinoid} [get]
// @Security BearerAuth
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

// Create godoc
// @Summary Criar tokenização RWA
// @Description Tokeniza um equino criando um ativo digital (RWA)
// @Tags Tokenização
// @Accept json
// @Produce json
// @Param tokenizacao body models.CreateTokenizacaoRequest true "Dados da tokenização"
// @Success 201 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /tokenizacao [post]
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

// ListTransacoes godoc
// @Summary Listar transações de uma tokenização
// @Description Retorna o histórico de transações de uma tokenização RWA
// @Tags Tokenização
// @Produce json
// @Param id path int true "ID da tokenização"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /tokenizacao/{id}/transacoes [get]
// @Security BearerAuth
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

// ExecutarOrdem godoc
// @Summary Executar ordem de compra/venda
// @Description Executa uma ordem de compra ou venda de tokens RWA
// @Tags Tokenização
// @Accept json
// @Produce json
// @Param ordem body object true "Dados da ordem"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /tokenizacao/executar [post]
// @Security BearerAuth
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

// CriarOferta godoc
// @Summary Criar oferta de tokens
// @Description Cria uma oferta de compra ou venda de tokens RWA
// @Tags Tokenização
// @Accept json
// @Produce json
// @Param oferta body object true "Dados da oferta"
// @Success 201 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /tokenizacao/ofertas [post]
// @Security BearerAuth
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
