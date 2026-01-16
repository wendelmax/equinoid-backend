package simulador

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {
	reproducao := rg.Group("/reproducao")
	reproducao.Use(authMiddleware)
	{
		reproducao.POST("/simular", handler.SimularCruzamento)
	}
}
