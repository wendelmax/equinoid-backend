package nutricao

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {
	nutricao := rg.Group("/nutricao")
	nutricao.Use(authMiddleware)
	{
		nutricao.GET("/equino/:equinoid", handler.GetPlano)
		nutricao.POST("/plano", handler.CreatePlano)
		nutricao.POST("/equino/:equinoid/ai-suggestion", handler.GetSugestaoIA)
		nutricao.POST("/refeicoes", handler.CreateRefeicao)
	}
}
