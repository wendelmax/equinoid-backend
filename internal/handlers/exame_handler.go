package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/equinoid/backend/internal/models"
	"github.com/gin-gonic/gin"
)

func (h *Handlers) ListExames(c *gin.Context) {
	var laboratorioID, solicitanteID *uint

	if id := c.Query("laboratorio_id"); id != "" {
		idInt, _ := strconv.ParseUint(id, 10, 32)
		idUint := uint(idInt)
		laboratorioID = &idUint
	}

	if id := c.Query("solicitante_id"); id != "" {
		idInt, _ := strconv.ParseUint(id, 10, 32)
		idUint := uint(idInt)
		solicitanteID = &idUint
	}

	exames, err := h.ExameLaboratorialService.List(laboratorioID, solicitanteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, exames)
}

func (h *Handlers) GetExame(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	exame, err := h.ExameLaboratorialService.Get(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Exame não encontrado"})
		return
	}

	c.JSON(http.StatusOK, exame)
}

func (h *Handlers) CreateExame(c *gin.Context) {
	var exame models.ExameLaboratorial
	if err := c.ShouldBindJSON(&exame); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}

		exame.VeterinarioSolicitanteID = userID.(uint)
	exame.DataSolicitacao = time.Now()
	exame.Status = "solicitado"

	if err := h.ExameLaboratorialService.Create(&exame); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, exame)
}

func (h *Handlers) UpdateExame(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.ExameLaboratorialService.Update(uint(id), updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Exame atualizado com sucesso"})
}

func (h *Handlers) AdicionarResultado(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var req struct {
		Resultado    string `json:"resultado" binding:"required"`
		Observacoes  string `json:"observacoes"`
		DocumentoURL string `json:"documento_url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.ExameLaboratorialService.AdicionarResultado(uint(id), req.Resultado, req.Observacoes, req.DocumentoURL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Resultado adicionado com sucesso"})
}

func (h *Handlers) AtribuirLaboratorio(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var req struct {
		LaboratorioID uint `json:"laboratorio_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.ExameLaboratorialService.AtribuirLaboratorio(uint(id), req.LaboratorioID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Laboratório atribuído com sucesso"})
}

func (h *Handlers) RegistrarColeta(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := h.ExameLaboratorialService.RegistrarColeta(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Coleta registrada com sucesso"})
}

func (h *Handlers) CancelarExame(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var req struct {
		Motivo string `json:"motivo" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.ExameLaboratorialService.Cancelar(uint(id), req.Motivo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Exame cancelado com sucesso"})
}
