package simulador

import (
	"net/http"
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

type SimularCruzamentoRequest struct {
	PaiEquinoid string `json:"pai_equinoid" binding:"required"`
	MaeEquinoid string `json:"mae_equinoid" binding:"required"`
}

// SimularCruzamento godoc
// @Summary Simular cruzamento
// @Description Simula um cruzamento entre dois equinos e retorna projeções genéticas
// @Tags Simulador
// @Accept json
// @Produce json
// @Param simulacao body object{pai_equinoid=string,mae_equinoid=string} true "Dados do cruzamento"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /reproducao/simular [post]
// @Security BearerAuth
func (h *Handler) SimularCruzamento(c *gin.Context) {
	var req SimularCruzamentoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Dados inválidos: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	if req.PaiEquinoid == req.MaeEquinoid {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Pai e mãe não podem ser o mesmo animal",
			Timestamp: time.Now(),
		})
		return
	}

	result, err := h.service.SimularCruzamento(c.Request.Context(), req.PaiEquinoid, req.MaeEquinoid)
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
			Error:     "Erro ao simular cruzamento",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   "Simulação realizada com sucesso",
		Timestamp: time.Now(),
		Data:      result,
	})
}
