package financeiro

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/equinoid/backend/internal/models"
	"github.com/equinoid/backend/pkg/logging"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	repoService RepositoryService
	logger      *logging.Logger
}

type RepositoryService interface {
	FindAll(ctx context.Context, filters map[string]interface{}) ([]*models.TransacaoFinanceira, error)
	Create(ctx context.Context, transacao *models.TransacaoFinanceira) error
	GetStats(ctx context.Context) (*models.FinanceiroStats, error)
	GetMonthlyData(ctx context.Context) ([]*models.MonthlyData, error)
	GetExpenseBreakdown(ctx context.Context) ([]*models.ExpenseBreakdown, error)
}

func NewHandler(repoService RepositoryService, logger *logging.Logger) *Handler {
	return &Handler{
		repoService: repoService,
		logger:      logger,
	}
}

// GetStats godoc
// @Summary Obter estatísticas financeiras
// @Description Retorna estatísticas financeiras gerais
// @Tags Financeiro
// @Produce json
// @Success 200 {object} models.APIResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /financeiro/stats [get]
// @Security BearerAuth
func (h *Handler) GetStats(c *gin.Context) {
	stats, err := h.repoService.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao buscar estatísticas",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   "Estatísticas financeiras",
		Timestamp: time.Now(),
		Data:      stats,
	})
}

// GetMonthlyData godoc
// @Summary Obter dados mensais
// @Description Retorna dados financeiros mensais
// @Tags Financeiro
// @Produce json
// @Success 200 {object} models.APIResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /financeiro/monthly [get]
// @Security BearerAuth
func (h *Handler) GetMonthlyData(c *gin.Context) {
	data, err := h.repoService.GetMonthlyData(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao buscar dados mensais",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   "Dados mensais",
		Timestamp: time.Now(),
		Data:      data,
	})
}

// GetExpenseBreakdown godoc
// @Summary Obter breakdown de despesas
// @Description Retorna breakdown detalhado de despesas por categoria
// @Tags Financeiro
// @Produce json
// @Success 200 {object} models.APIResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /financeiro/breakdown [get]
// @Security BearerAuth
func (h *Handler) GetExpenseBreakdown(c *gin.Context) {
	breakdown, err := h.repoService.GetExpenseBreakdown(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao buscar breakdown de despesas",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   "Breakdown de despesas",
		Timestamp: time.Now(),
		Data:      breakdown,
	})
}

// ListTransactions godoc
// @Summary Listar transações financeiras
// @Description Lista todas as transações com filtros opcionais
// @Tags Financeiro
// @Produce json
// @Param tipo query string false "Filtrar por tipo"
// @Param categoria query string false "Filtrar por categoria"
// @Success 200 {object} models.APIResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /financeiro/transactions [get]
// @Security BearerAuth
func (h *Handler) ListTransactions(c *gin.Context) {
	filters := make(map[string]interface{})
	
	if tipo := c.Query("tipo"); tipo != "" {
		filters["tipo"] = tipo
	}
	if categoria := c.Query("categoria"); categoria != "" {
		filters["categoria"] = categoria
	}

	transacoes, err := h.repoService.FindAll(c.Request.Context(), filters)
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
		Message:   fmt.Sprintf("Transações (total: %d)", len(transacoes)),
		Timestamp: time.Now(),
		Data:      transacoes,
	})
}

// CreateTransaction godoc
// @Summary Criar transação financeira
// @Description Registra uma nova transação financeira
// @Tags Financeiro
// @Accept json
// @Produce json
// @Param transacao body models.CreateTransacaoRequest true "Dados da transação"
// @Success 201 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /financeiro/transactions [post]
// @Security BearerAuth
func (h *Handler) CreateTransaction(c *gin.Context) {
	var req models.CreateTransacaoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Dados inválidos: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	dataTime, err := time.Parse("2006-01-02", req.Data)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Data inválida",
			Timestamp: time.Now(),
		})
		return
	}

	transacao := &models.TransacaoFinanceira{
		Tipo:      req.Tipo,
		Categoria: req.Categoria,
		Descricao: req.Descricao,
		Valor:     req.Valor,
		Data:      dataTime,
		EquinoID:  req.EquinoID,
		Status:    models.StatusPagamentoPendente,
	}

	if err := h.repoService.Create(c.Request.Context(), transacao); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao criar transação",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success:   true,
		Message:   "Transação criada com sucesso",
		Timestamp: time.Now(),
		Data:      transacao,
	})
}
