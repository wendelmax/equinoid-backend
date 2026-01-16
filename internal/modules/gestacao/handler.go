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
