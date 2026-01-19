package nutricao

import (
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

// GetPlano godoc
// @Summary Obter plano nutricional
// @Description Retorna o plano nutricional de um equino
// @Tags Nutrição
// @Produce json
// @Param equinoid path string true "Equinoid do equino"
// @Success 200 {object} models.APIResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /nutricao/equino/{equinoid} [get]
// @Security BearerAuth
func (h *Handler) GetPlano(c *gin.Context) {
	equinoid := c.Param("equinoid")

	plano, err := h.service.GetPlanoByEquinoid(c.Request.Context(), equinoid)
	if err != nil {
		if apperrors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Success:   false,
				Error:     "Plano nutricional não encontrado",
				Timestamp: time.Now(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao buscar plano nutricional",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   "Plano nutricional do equino",
		Timestamp: time.Now(),
		Data:      plano,
	})
}

// CreatePlano godoc
// @Summary Criar plano nutricional
// @Description Cria um novo plano nutricional para um equino
// @Tags Nutrição
// @Accept json
// @Produce json
// @Param plano body models.CreatePlanoNutricionalRequest true "Dados do plano"
// @Success 201 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /nutricao/plano [post]
// @Security BearerAuth
func (h *Handler) CreatePlano(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success:   false,
			Error:     "Authentication required",
			Timestamp: time.Now(),
		})
		return
	}

	var req models.CreatePlanoNutricionalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Dados inválidos: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	plano, err := h.service.CreatePlano(c.Request.Context(), userID, &req)
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
			Error:     "Erro ao criar plano nutricional",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success:   true,
		Message:   "Plano nutricional criado com sucesso",
		Timestamp: time.Now(),
		Data:      plano,
	})
}

// GetSugestaoIA godoc
// @Summary Obter sugestão nutricional com IA
// @Description Gera uma sugestão de plano nutricional usando inteligência artificial
// @Tags Nutrição
// @Accept json
// @Produce json
// @Param sugestao body models.SugestaoIARequest true "Dados para sugestão"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /nutricao/equino/{equinoid}/ai-suggestion [post]
// @Security BearerAuth
func (h *Handler) GetSugestaoIA(c *gin.Context) {
	var req models.SugestaoIARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Dados inválidos: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	sugestao, err := h.service.GetSugestaoIA(c.Request.Context(), &req)
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
			Error:     "Erro ao gerar sugestão com IA",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   "Sugestão nutricional gerada com IA",
		Timestamp: time.Now(),
		Data:      sugestao,
	})
}

// CreateRefeicao godoc
// @Summary Criar registro de refeição
// @Description Registra uma refeição realizada pelo equino
// @Tags Nutrição
// @Accept json
// @Produce json
// @Param refeicao body models.CreateRefeicaoRequest true "Dados da refeição"
// @Success 201 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /nutricao/refeicoes [post]
// @Security BearerAuth
func (h *Handler) CreateRefeicao(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success:   false,
			Error:     "Authentication required",
			Timestamp: time.Now(),
		})
		return
	}

	var req models.CreateRefeicaoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Dados inválidos: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	refeicao, err := h.service.CreateRefeicao(c.Request.Context(), userID, &req)
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
			Error:     "Erro ao registrar refeição",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success:   true,
		Message:   "Refeição registrada com sucesso",
		Timestamp: time.Now(),
		Data:      refeicao,
	})
}
