package participacoes

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {
	eventos := rg.Group("/eventos")
	eventos.Use(authMiddleware)
	{
		eventos.GET("/:evento_id/participacoes", handler.ListByEvento)

		participacoes := eventos.Group("/participacoes")
		{
			participacoes.POST("", handler.Create)
			participacoes.PUT("/:id", handler.Update)
			participacoes.DELETE("/:id", handler.Delete)
			participacoes.POST("/:id/ausencia", handler.MarcarAusencia)
			participacoes.POST("/:id/presenca", handler.MarcarPresenca)
		}
	}

	equinos := rg.Group("/equinos")
	equinos.Use(authMiddleware)
	{
		equinos.GET("/:equinoid/participacoes-eventos", handler.ListByEquino)
	}
}
