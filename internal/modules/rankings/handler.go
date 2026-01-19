package rankings

import (
	"fmt"
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

// GetRankingGeral godoc
// @Summary Obter ranking geral
// @Description Retorna ranking geral por tipo (geral, raca, categoria, etc)
// @Tags Rankings
// @Produce json
// @Param tipo path string true "Tipo de ranking"
// @Success 200 {object} models.APIResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /rankings/{tipo} [get]
// @Security BearerAuth
func (h *Handler) GetRankingGeral(c *gin.Context) {
	tipo := c.Param("tipo")
	if tipo == "" {
		tipo = "geral"
	}

	items, err := h.service.GetRankingGeral(c.Request.Context(), tipo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao buscar ranking",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   fmt.Sprintf("Ranking %s (total: %d)", tipo, len(items)),
		Timestamp: time.Now(),
		Data:      items,
	})
}

// GetRankingsEquino godoc
// @Summary Obter rankings de um equino
// @Description Retorna todos os rankings de um equino específico
// @Tags Rankings
// @Produce json
// @Param equinoid path string true "Equinoid do equino"
// @Success 200 {object} models.APIResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /equinos/{equinoid}/rankings [get]
// @Security BearerAuth
func (h *Handler) GetRankingsEquino(c *gin.Context) {
	equinoid := c.Param("equinoid")

	items, err := h.service.GetRankingsEquino(c.Request.Context(), equinoid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao buscar rankings do equino",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   fmt.Sprintf("Rankings do equino %s", equinoid),
		Timestamp: time.Now(),
		Data:      items,
	})
}
