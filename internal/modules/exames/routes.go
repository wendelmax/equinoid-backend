package exames

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {
	exames := rg.Group("/exames-laboratoriais")
	exames.Use(authMiddleware)
	{
		exames.GET("", handler.List)
		exames.POST("", handler.Create)
		exames.GET("/:id", handler.GetByID)
		exames.PUT("/:id", handler.Update)
		exames.DELETE("/:id", handler.Delete)
		
		exames.PUT("/:id/receber-amostra", handler.ReceberAmostra)
		exames.PUT("/:id/iniciar-analise", handler.IniciarAnalise)
		exames.PUT("/:id/concluir", handler.ConcluirExame)
	}
}
