package relatorios

import (
	"net/http"
	"time"

	"github.com/equinoid/backend/internal/models"
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

// GetDashboardStats godoc
// @Summary Obter estatísticas do dashboard
// @Description Retorna estatísticas gerais para o dashboard
// @Tags Relatórios
// @Produce json
// @Success 200 {object} models.APIResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /reports/dashboard [get]
// @Security BearerAuth
func (h *Handler) GetDashboardStats(c *gin.Context) {
	stats, err := h.service.GetDashboardStats(c.Request.Context())
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
		Message:   "Estatísticas do dashboard",
		Timestamp: time.Now(),
		Data:      stats,
	})
}
