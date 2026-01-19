package exames

import (
	"fmt"
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

// List godoc
// @Summary Listar exames laboratoriais
// @Description Lista todos os exames laboratoriais com filtros opcionais
// @Tags Exames
// @Produce json
// @Param equinoid query string false "Filtrar por Equinoid"
// @Param status query string false "Filtrar por status"
// @Param tipo_exame query string false "Filtrar por tipo de exame"
// @Param veterinario_id query int false "Filtrar por veterinário"
// @Param laboratorio_id query int false "Filtrar por laboratório"
// @Success 200 {object} models.APIResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /exames-laboratoriais [get]
// @Security BearerAuth
func (h *Handler) List(c *gin.Context) {
	filters := make(map[string]interface{})
	
	if equinoid := c.Query("equinoid"); equinoid != "" {
		filters["equinoid"] = equinoid
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if tipoExame := c.Query("tipo_exame"); tipoExame != "" {
		filters["tipo_exame"] = tipoExame
	}
	if vetID := c.Query("veterinario_id"); vetID != "" {
		if id, err := strconv.ParseUint(vetID, 10, 32); err == nil {
			filters["veterinario_id"] = uint(id)
		}
	}
	if labID := c.Query("laboratorio_id"); labID != "" {
		if id, err := strconv.ParseUint(labID, 10, 32); err == nil {
			filters["laboratorio_id"] = uint(id)
		}
	}

	exames, err := h.service.ListAll(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao listar exames",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   fmt.Sprintf("Lista de exames (total: %d)", len(exames)),
		Timestamp: time.Now(),
		Data:      exames,
	})
}

// GetByID godoc
// @Summary Buscar exame por ID
// @Description Retorna os dados completos de um exame laboratorial
// @Tags Exames
// @Produce json
// @Param id path int true "ID do exame"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /exames-laboratoriais/{id} [get]
// @Security BearerAuth
func (h *Handler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "ID inválido",
			Timestamp: time.Now(),
		})
		return
	}

	exame, err := h.service.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		if apperrors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Success:   false,
				Error:     "Exame não encontrado",
				Timestamp: time.Now(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao buscar exame",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   "Exame encontrado",
		Timestamp: time.Now(),
		Data:      exame,
	})
}

// Create godoc
// @Summary Criar solicitação de exame
// @Description Cria uma nova solicitação de exame laboratorial
// @Tags Exames
// @Accept json
// @Produce json
// @Param exame body models.CreateExameRequest true "Dados do exame"
// @Success 201 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /exames-laboratoriais [post]
// @Security BearerAuth
func (h *Handler) Create(c *gin.Context) {
	var req models.CreateExameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Dados inválidos: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	exame, err := h.service.Create(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao criar exame",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success:   true,
		Message:   "Exame solicitado com sucesso",
		Timestamp: time.Now(),
		Data:      exame,
	})
}

// Update godoc
// @Summary Atualizar exame
// @Description Atualiza informações de um exame laboratorial
// @Tags Exames
// @Accept json
// @Produce json
// @Param id path int true "ID do exame"
// @Param exame body models.UpdateExameRequest true "Dados para atualização"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /exames-laboratoriais/{id} [put]
// @Security BearerAuth
func (h *Handler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "ID inválido",
			Timestamp: time.Now(),
		})
		return
	}

	var req models.UpdateExameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Dados inválidos: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	exame, err := h.service.Update(c.Request.Context(), uint(id), &req)
	if err != nil {
		if apperrors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Success:   false,
				Error:     "Exame não encontrado",
				Timestamp: time.Now(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao atualizar exame",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   "Exame atualizado com sucesso",
		Timestamp: time.Now(),
		Data:      exame,
	})
}

// Delete godoc
// @Summary Deletar exame
// @Description Remove um exame laboratorial do sistema
// @Tags Exames
// @Produce json
// @Param id path int true "ID do exame"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /exames-laboratoriais/{id} [delete]
// @Security BearerAuth
func (h *Handler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "ID inválido",
			Timestamp: time.Now(),
		})
		return
	}

	if err := h.service.Delete(c.Request.Context(), uint(id)); err != nil {
		if apperrors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Success:   false,
				Error:     "Exame não encontrado",
				Timestamp: time.Now(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao deletar exame",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   "Exame deletado com sucesso",
		Timestamp: time.Now(),
		Data:      nil,
	})
}

// ReceberAmostra godoc
// @Summary Receber amostra no laboratório
// @Description Registra o recebimento da amostra no laboratório
// @Tags Exames
// @Accept json
// @Produce json
// @Param id path int true "ID do exame"
// @Param recebimento body object{data_recebimento=string} false "Data de recebimento"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /exames-laboratoriais/{id}/receber-amostra [put]
// @Security BearerAuth
func (h *Handler) ReceberAmostra(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "ID inválido",
			Timestamp: time.Now(),
		})
		return
	}

	var req struct {
		DataRecebimento *string `json:"data_recebimento"`
	}
	c.ShouldBindJSON(&req)

	exame, err := h.service.ReceberAmostra(c.Request.Context(), uint(id), req.DataRecebimento)
	if err != nil {
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
			Error:     "Erro ao receber amostra",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   "Amostra recebida",
		Timestamp: time.Now(),
		Data:      exame,
	})
}

// IniciarAnalise godoc
// @Summary Iniciar análise do exame
// @Description Marca o início da análise laboratorial
// @Tags Exames
// @Produce json
// @Param id path int true "ID do exame"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /exames-laboratoriais/{id}/iniciar-analise [put]
// @Security BearerAuth
func (h *Handler) IniciarAnalise(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "ID inválido",
			Timestamp: time.Now(),
		})
		return
	}

	exame, err := h.service.IniciarAnalise(c.Request.Context(), uint(id))
	if err != nil {
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
			Error:     "Erro ao iniciar análise",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   "Análise iniciada",
		Timestamp: time.Now(),
		Data:      exame,
	})
}

// ConcluirExame godoc
// @Summary Concluir exame com resultado
// @Description Finaliza o exame registrando resultado, valores e laudo
// @Tags Exames
// @Accept json
// @Produce json
// @Param id path int true "ID do exame"
// @Param resultado body object{resultado=string,valores=object,laudo=string} true "Resultado do exame"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /exames-laboratoriais/{id}/concluir [put]
// @Security BearerAuth
func (h *Handler) ConcluirExame(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "ID inválido",
			Timestamp: time.Now(),
		})
		return
	}

	var req struct {
		Resultado models.ResultadoExame     `json:"resultado" binding:"required"`
		Valores   map[string]interface{}    `json:"valores"`
		Laudo     *string                   `json:"laudo"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Dados inválidos: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	exame, err := h.service.ConcluirExame(c.Request.Context(), uint(id), req.Resultado, req.Valores, req.Laudo)
	if err != nil {
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
			Error:     "Erro ao concluir exame",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   "Exame concluído com sucesso",
		Timestamp: time.Now(),
		Data:      exame,
	})
}
