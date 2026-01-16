package treinamento

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {
	treinamento := rg.Group("/treinamento")
	treinamento.Use(authMiddleware)
	{
		treinamento.GET("/sessoes", handler.GetSessoes)
		treinamento.POST("/sessoes", handler.CreateSessao)
		treinamento.GET("/programas", handler.GetProgramas)
		treinamento.POST("/programas", handler.CreatePrograma)
	}
}
