package handlers

import (
	"net/http"
	"strconv"

	"github.com/equinoid/backend/internal/models"
	"github.com/gin-gonic/gin"
)

// ListLeiloes godoc
// @Summary Lista leilões
// @Description Retorna lista de leilões, opcionalmente filtrada por leiloeiro
// @Tags leiloes
// @Produce json
// @Param leiloeiro_id query int false "ID do leiloeiro"
// @Success 200 {array} models.Leilao
// @Failure 500 {object} map[string]string
// @Router /leiloes [get]
// @Security BearerAuth
func (h *Handlers) ListLeiloes(c *gin.Context) {
	var leiloeiroID *uint
	if id := c.Query("leiloeiro_id"); id != "" {
		idInt, _ := strconv.ParseUint(id, 10, 32)
		idUint := uint(idInt)
		leiloeiroID = &idUint
	}

	leiloes, err := h.LeilaoService.List(leiloeiroID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, leiloes)
}

// GetLeilao godoc
// @Summary Busca leilão por ID
// @Description Retorna detalhes de um leilão específico
// @Tags leiloes
// @Produce json
// @Param id path int true "ID do leilão"
// @Success 200 {object} models.Leilao
// @Failure 404 {object} map[string]string
// @Router /leiloes/{id} [get]
// @Security BearerAuth
func (h *Handlers) GetLeilao(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	leilao, err := h.LeilaoService.Get(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Leilão não encontrado"})
		return
	}

	c.JSON(http.StatusOK, leilao)
}

// CreateLeilao godoc
// @Summary Cria novo leilão
// @Description Cria um novo leilão (apenas leiloeiros)
// @Tags leiloes
// @Accept json
// @Produce json
// @Param leilao body models.Leilao true "Dados do leilão"
// @Success 201 {object} models.Leilao
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /leiloes [post]
// @Security BearerAuth
func (h *Handlers) CreateLeilao(c *gin.Context) {
	var leilao models.Leilao
	if err := c.ShouldBindJSON(&leilao); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Pegar ID do usuário do contexto
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}

	leilao.LeiloeiroID = userID.(uint)
	leilao.Status = "agendado"

	if err := h.LeilaoService.Create(&leilao); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, leilao)
}

// FinalizarLeilao godoc
// @Summary Finaliza leilão
// @Description Encerra um leilão e calcula totais
// @Tags leiloes
// @Produce json
// @Param id path int true "ID do leilão"
// @Success 200 {object} models.Leilao
// @Failure 500 {object} map[string]string
// @Router /leiloes/{id}/finalizar [post]
// @Security BearerAuth
func (h *Handlers) FinalizarLeilao(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	leilao, err := h.LeilaoService.Finalizar(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, leilao)
}

// ListParticipacoes godoc
// @Summary Lista participações de um leilão
// @Description Retorna todas as participações de um leilão específico
// @Tags leiloes
// @Produce json
// @Param id path int true "ID do leilão"
// @Success 200 {array} models.ParticipacaoLeilao
// @Failure 500 {object} map[string]string
// @Router /leiloes/{id}/participacoes [get]
// @Security BearerAuth
func (h *Handlers) ListParticipacoes(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	participacoes, err := h.LeilaoService.ListParticipacoes(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, participacoes)
}

// CreateParticipacao godoc
// @Summary Cria participação em leilão
// @Description Inscreve um equino em um leilão
// @Tags leiloes
// @Accept json
// @Produce json
// @Param id path int true "ID do leilão"
// @Param participacao body models.ParticipacaoLeilao true "Dados da participação"
// @Success 201 {object} models.ParticipacaoLeilao
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /leiloes/{id}/participacoes [post]
// @Security BearerAuth
func (h *Handlers) CreateParticipacao(c *gin.Context) {
	leilaoID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var participacao models.ParticipacaoLeilao
	if err := c.ShouldBindJSON(&participacao); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	participacao.LeilaoID = uint(leilaoID)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}
	participacao.CriadorID = userID.(uint)
	participacao.Status = "inscrito"

	if err := h.LeilaoService.CreateParticipacao(&participacao); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, participacao)
}

// AprovarParticipacao godoc
// @Summary Aprova participação
// @Description Aprova a participação de um equino no leilão
// @Tags leiloes
// @Produce json
// @Param id path int true "ID da participação"
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /leiloes/participacoes/{id}/aprovar [post]
// @Security BearerAuth
func (h *Handlers) AprovarParticipacao(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := h.LeilaoService.AprovarParticipacao(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Participação aprovada com sucesso"})
}

// RegistrarVenda godoc
// @Summary Registra venda em leilão
// @Description Registra a venda de um equino no leilão
// @Tags leiloes
// @Accept json
// @Produce json
// @Param id path int true "ID da participação"
// @Param venda body object{valor_vendido=number,comprador_id=number} true "Dados da venda"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /leiloes/participacoes/{id}/venda [post]
// @Security BearerAuth
func (h *Handlers) RegistrarVenda(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var req struct {
		ValorVendido float64 `json:"valor_vendido" binding:"required"`
		CompradorID  uint    `json:"comprador_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Buscar participação para obter o leilão
	var participacao models.ParticipacaoLeilao
	if err := h.DB.First(&participacao, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Participação não encontrada"})
		return
	}

	// Buscar leilão para calcular comissão
	leilao, err := h.LeilaoService.Get(participacao.LeilaoID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Leilão não encontrado"})
		return
	}

	if err := h.LeilaoService.RegistrarVenda(uint(id), req.ValorVendido, req.CompradorID, leilao); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Venda registrada com sucesso"})
}

// MarcarAusencia godoc
// @Summary Marca ausência em leilão presencial
// @Description Marca a ausência de um equino em leilão presencial e aplica penalização
// @Tags leiloes
// @Produce json
// @Param id path int true "ID da participação"
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /leiloes/participacoes/{id}/ausencia [post]
// @Security BearerAuth
func (h *Handlers) MarcarAusencia(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := h.LeilaoService.MarcarAusencia(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ausência marcada com sucesso"})
}

// MarcarPresenca godoc
// @Summary Marca presença em leilão presencial
// @Description Marca a presença de um equino em leilão presencial
// @Tags leiloes
// @Produce json
// @Param id path int true "ID da participação"
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /leiloes/participacoes/{id}/presenca [post]
// @Security BearerAuth
func (h *Handlers) MarcarPresenca(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := h.LeilaoService.MarcarPresenca(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Presença confirmada com sucesso"})
}

// GetRelatorioGanhos godoc
// @Summary Relatório de ganhos de leilões
// @Description Retorna relatório de ganhos de leilões finalizados
// @Tags leiloes
// @Produce json
// @Param leiloeiro_id query int false "ID do leiloeiro"
// @Success 200 {array} object
// @Failure 500 {object} map[string]string
// @Router /leiloes/relatorio-ganhos [get]
// @Security BearerAuth
func (h *Handlers) GetRelatorioGanhos(c *gin.Context) {
	var leiloeiroID *uint
	if id := c.Query("leiloeiro_id"); id != "" {
		idInt, _ := strconv.ParseUint(id, 10, 32)
		idUint := uint(idInt)
		leiloeiroID = &idUint
	}

	relatorios, err := h.LeilaoService.GetRelatorioGanhos(leiloeiroID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, relatorios)
}
