package tokenizacao

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {
	tokenizacao := rg.Group("/tokenizacao")
	tokenizacao.Use(authMiddleware)
	{
		tokenizacao.GET("", handler.ListAll)
		tokenizacao.POST("", handler.Create)
		tokenizacao.GET("/:id", handler.GetByID)
		tokenizacao.GET("/:id/transacoes", handler.ListTransacoes)
		
		tokenizacao.POST("/executar", handler.ExecutarOrdem)
		tokenizacao.POST("/ofertas", handler.CriarOferta)
	}

	equinosToken := rg.Group("/tokenizacao/equino")
	equinosToken.Use(authMiddleware)
	{
		equinosToken.GET("/:equinoid", handler.GetByEquinoid)
	}
}
