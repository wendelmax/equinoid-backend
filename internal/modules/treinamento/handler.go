package treinamento

import (
	"fmt"
	"net/http"
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

// GetSessoes godoc
// @Summary Listar sessões de treinamento
// @Description Retorna todas as sessões de treinamento de um equino
// @Tags Treinamento
// @Produce json
// @Param equinoid query string true "Equinoid do equino"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /treinamento/sessoes [get]
// @Security BearerAuth
func (h *Handler) GetSessoes(c *gin.Context) {
	equinoid := c.Query("equinoid")
	if equinoid == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Parâmetro equinoid é obrigatório",
			Timestamp: time.Now(),
		})
		return
	}

	sessoes, err := h.service.GetSessoesByEquinoid(c.Request.Context(), equinoid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao buscar sessões de treinamento",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   fmt.Sprintf("Sessões de treinamento (total: %d)", len(sessoes)),
		Timestamp: time.Now(),
		Data:      sessoes,
	})
}

// CreateSessao godoc
// @Summary Criar sessão de treinamento
// @Description Registra uma nova sessão de treinamento
// @Tags Treinamento
// @Accept json
// @Produce json
// @Param sessao body models.CreateSessaoTreinamentoRequest true "Dados da sessão"
// @Success 201 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /treinamento/sessoes [post]
// @Security BearerAuth
func (h *Handler) CreateSessao(c *gin.Context) {
	treinadorID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success:   false,
			Error:     "Authentication required",
			Timestamp: time.Now(),
		})
		return
	}

	var req models.CreateSessaoTreinamentoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Dados inválidos: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	sessao, err := h.service.CreateSessao(c.Request.Context(), treinadorID, &req)
	if err != nil {
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
			Error:     "Erro ao criar sessão de treinamento",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success:   true,
		Message:   "Sessão de treinamento registrada com sucesso",
		Timestamp: time.Now(),
		Data:      sessao,
	})
}

// GetProgramas godoc
// @Summary Listar programas de treinamento
// @Description Retorna todos os programas de treinamento de um equino
// @Tags Treinamento
// @Produce json
// @Param equinoid query string true "Equinoid do equino"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /treinamento/programas [get]
// @Security BearerAuth
func (h *Handler) GetProgramas(c *gin.Context) {
	equinoid := c.Query("equinoid")
	if equinoid == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Parâmetro equinoid é obrigatório",
			Timestamp: time.Now(),
		})
		return
	}

	programas, err := h.service.GetProgramasByEquinoid(c.Request.Context(), equinoid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao buscar programas de treinamento",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   fmt.Sprintf("Programas de treinamento (total: %d)", len(programas)),
		Timestamp: time.Now(),
		Data:      programas,
	})
}

// CreatePrograma godoc
// @Summary Criar programa de treinamento
// @Description Cria um novo programa de treinamento para um equino
// @Tags Treinamento
// @Accept json
// @Produce json
// @Param programa body models.CreateProgramaTreinamentoRequest true "Dados do programa"
// @Success 201 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /treinamento/programas [post]
// @Security BearerAuth
func (h *Handler) CreatePrograma(c *gin.Context) {
	treinadorID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success:   false,
			Error:     "Authentication required",
			Timestamp: time.Now(),
		})
		return
	}

	var req models.CreateProgramaTreinamentoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Dados inválidos: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	programa, err := h.service.CreatePrograma(c.Request.Context(), treinadorID, &req)
	if err != nil {
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
			Error:     "Erro ao criar programa de treinamento",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success:   true,
		Message:   "Programa de treinamento criado com sucesso",
		Timestamp: time.Now(),
		Data:      programa,
	})
}
