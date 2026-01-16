package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/equinoid/backend/internal/models"
	"github.com/gin-gonic/gin"
)

// ListExames godoc
// @Summary Lista exames laboratoriais
// @Description Retorna lista de exames, opcionalmente filtrada por laboratório ou solicitante
// @Tags exames
// @Produce json
// @Param laboratorio_id query int false "ID do laboratório"
// @Param solicitante_id query int false "ID do solicitante"
// @Success 200 {array} models.ExameLaboratorial
// @Failure 500 {object} map[string]string
// @Router /exames [get]
// @Security BearerAuth
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

// GetExame godoc
// @Summary Busca exame por ID
// @Description Retorna detalhes de um exame específico
// @Tags exames
// @Produce json
// @Param id path int true "ID do exame"
// @Success 200 {object} models.ExameLaboratorial
// @Failure 404 {object} map[string]string
// @Router /exames/{id} [get]
// @Security BearerAuth
func (h *Handlers) GetExame(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	exame, err := h.ExameLaboratorialService.Get(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Exame não encontrado"})
		return
	}

	c.JSON(http.StatusOK, exame)
}

// CreateExame godoc
// @Summary Solicita novo exame
// @Description Cria uma solicitação de exame laboratorial
// @Tags exames
// @Accept json
// @Produce json
// @Param exame body models.ExameLaboratorial true "Dados do exame"
// @Success 201 {object} models.ExameLaboratorial
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /exames [post]
// @Security BearerAuth
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

// UpdateExame godoc
// @Summary Atualiza exame
// @Description Atualiza informações de um exame
// @Tags exames
// @Accept json
// @Produce json
// @Param id path int true "ID do exame"
// @Param exame body object true "Dados para atualizar"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /exames/{id} [put]
// @Security BearerAuth
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

// AdicionarResultado godoc
// @Summary Adiciona resultado ao exame
// @Description Adiciona resultado e finaliza um exame
// @Tags exames
// @Accept json
// @Produce json
// @Param id path int true "ID do exame"
// @Param resultado body object{resultado=string,observacoes=string,documento_url=string} true "Resultado do exame"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /exames/{id}/resultado [post]
// @Security BearerAuth
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

// AtribuirLaboratorio godoc
// @Summary Atribui laboratório ao exame
// @Description Atribui um laboratório para realizar o exame
// @Tags exames
// @Accept json
// @Produce json
// @Param id path int true "ID do exame"
// @Param laboratorio body object{laboratorio_id=number} true "ID do laboratório"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /exames/{id}/atribuir-laboratorio [post]
// @Security BearerAuth
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

// RegistrarColeta godoc
// @Summary Registra coleta de material
// @Description Registra que a coleta de material foi realizada
// @Tags exames
// @Produce json
// @Param id path int true "ID do exame"
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /exames/{id}/registrar-coleta [post]
// @Security BearerAuth
func (h *Handlers) RegistrarColeta(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := h.ExameLaboratorialService.RegistrarColeta(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Coleta registrada com sucesso"})
}

// CancelarExame godoc
// @Summary Cancela exame
// @Description Cancela um exame laboratorial
// @Tags exames
// @Accept json
// @Produce json
// @Param id path int true "ID do exame"
// @Param cancelamento body object{motivo=string} true "Motivo do cancelamento"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /exames/{id}/cancelar [post]
// @Security BearerAuth
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
