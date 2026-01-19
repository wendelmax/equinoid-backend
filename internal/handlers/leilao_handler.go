package handlers

import (
	"net/http"
	"strconv"

	"github.com/equinoid/backend/internal/models"
	"github.com/gin-gonic/gin"
)

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

func (h *Handlers) GetLeilao(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	leilao, err := h.LeilaoService.Get(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Leilão não encontrado"})
		return
	}

	c.JSON(http.StatusOK, leilao)
}

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

func (h *Handlers) FinalizarLeilao(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	leilao, err := h.LeilaoService.Finalizar(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, leilao)
}

func (h *Handlers) ListParticipacoes(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	participacoes, err := h.LeilaoService.ListParticipacoes(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, participacoes)
}

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

func (h *Handlers) AprovarParticipacao(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := h.LeilaoService.AprovarParticipacao(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Participação aprovada com sucesso"})
}

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

func (h *Handlers) MarcarAusencia(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := h.LeilaoService.MarcarAusencia(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ausência marcada com sucesso"})
}

func (h *Handlers) MarcarPresenca(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := h.LeilaoService.MarcarPresenca(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Presença confirmada com sucesso"})
}

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
