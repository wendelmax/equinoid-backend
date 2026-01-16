package leiloes

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {
	leiloes := rg.Group("/leiloes")
	leiloes.Use(authMiddleware)
	{
		leiloes.GET("/:leilao_id/participacoes", handler.ListParticipacoes)
		leiloes.POST("/:leilao_id/participacoes", handler.CriarParticipacao)
		
		participacoes := leiloes.Group("/participacoes")
		{
			participacoes.POST("/:id/aprovar", handler.AprovarParticipacao)
			participacoes.POST("/:id/venda", handler.RegistrarVenda)
			participacoes.POST("/:id/ausencia", handler.MarcarAusencia)
			participacoes.POST("/:id/presenca", handler.MarcarPresenca)
		}
	}
}
