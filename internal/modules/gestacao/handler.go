package gestacao

import (
	"net/http"
	"strconv"
	"time"

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

// CriarUltrassonografia godoc
// @Summary Criar ultrassonografia
// @Description Registra uma nova ultrassonografia de gestação
// @Tags Gestação
// @Accept json
// @Produce json
// @Param gestacao_id path int true "ID da gestação"
// @Param ultrassom body models.CreateUltrassonografiaRequest true "Dados da ultrassonografia"
// @Success 201 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /gestacoes/{gestacao_id}/ultrassonografias [post]
// @Security BearerAuth
func (h *Handler) CriarUltrassonografia(c *gin.Context) {
	gestacaoID, err := strconv.ParseUint(c.Param("gestacao_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "ID de gestação inválido",
			Timestamp: time.Now(),
		})
		return
	}

	var req models.CreateUltrassonografiaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Dados inválidos: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	ultrassom, err := h.service.CriarUltrassonografia(c.Request.Context(), uint(gestacaoID), &req)
	if err != nil {
		if apperrors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Success:   false,
				Error:     "Gestação não encontrada",
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
			Error:     "Erro ao criar ultrassonografia",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success:   true,
		Message:   "Ultrassonografia criada com sucesso",
		Timestamp: time.Now(),
		Data:      ultrassom,
	})
}

// RegistrarParto godoc
// @Summary Registrar parto
// @Description Registra o parto de uma gestação
// @Tags Gestação
// @Accept json
// @Produce json
// @Param gestacao_id path int true "ID da gestação"
// @Param parto body models.RegistrarPartoRequest true "Dados do parto"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /gestacoes/{gestacao_id}/parto [post]
// @Security BearerAuth
func (h *Handler) RegistrarParto(c *gin.Context) {
	gestacaoID, err := strconv.ParseUint(c.Param("gestacao_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "ID de gestação inválido",
			Timestamp: time.Now(),
		})
		return
	}

	var req models.RegistrarPartoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Dados inválidos: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	if err := h.service.RegistrarParto(c.Request.Context(), uint(gestacaoID), &req); err != nil {
		if apperrors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Success:   false,
				Error:     "Gestação não encontrada",
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
			Error:     "Erro ao registrar parto",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   "Parto registrado com sucesso",
		Timestamp: time.Now(),
		Data:      nil,
	})
}

// RegistrarPerformanceMaterna godoc
// @Summary Registrar performance materna
// @Description Registra a performance materna de uma égua após o parto
// @Tags Gestação
// @Accept json
// @Produce json
// @Param equinoid path string true "Equinoid da égua"
// @Param performance body models.CreatePerformanceMaternaRequest true "Dados da performance"
// @Success 201 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /equinos/{equinoid}/performance-materna [post]
// @Security BearerAuth
func (h *Handler) RegistrarPerformanceMaterna(c *gin.Context) {
	equinoid := c.Param("equinoid")

	var req models.CreatePerformanceMaternaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Dados inválidos: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	if err := h.service.RegistrarPerformanceMaterna(c.Request.Context(), equinoid, &req); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao registrar performance materna",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   "Performance materna registrada com sucesso",
		Timestamp: time.Now(),
		Data:      nil,
	})
}
