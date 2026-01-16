package handlers

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/equinoid/backend/internal/models"
	"github.com/gin-gonic/gin"
)

type SimulacaoRequest struct {
	PaiID string `json:"pai_equinoid"`
	MaeID string `json:"mae_equinoid"`
}

type SimulacaoResponse struct {
	Inbreeding  float64 `json:"inbreeding"`
	Aptidao     int     `json:"aptidao_esportiva"`
	Valorizacao string  `json:"valorizacao_estimada"`
	Rating      string  `json:"rating"`
	Mensagem    string  `json:"mensagem"`
}

// CruzarSimulacao simula um cruzamento genético
func (h *Handlers) CruzarSimulacao(c *gin.Context) {
	var req SimulacaoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Selecione o pai e a mãe",
			Timestamp: time.Now(),
		})
		return
	}

	// Lógica de Simulação (Dummy funcional para Demo)
	// Em um cenário real, buscaríamos os ancestrais no banco e calcularíamos o COI
	rand.Seed(time.Now().UnixNano())

	response := SimulacaoResponse{
		Inbreeding:  0.5 + rand.Float64()*(5.0-0.5), // Entre 0.5% e 5%
		Aptidao:     70 + rand.Intn(30),             // Entre 70% e 100%
		Valorizacao: "Alta",
		Rating:      "AAA",
		Mensagem:    "Cruzamento com baixo risco de inbreeding e alto potencial atlético.",
	}

	if response.Inbreeding > 4.0 {
		response.Mensagem = "Atenção: Nível de inbreeding acima da média recomendada."
		response.Rating = "AA"
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      response,
		Timestamp: time.Now(),
	})
}
