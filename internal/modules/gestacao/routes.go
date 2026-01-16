package gestacao

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {
	gestacoes := rg.Group("/gestacoes")
	gestacoes.Use(authMiddleware)
	{
		gestacoes.POST("/:gestacao_id/ultrassonografias", handler.CriarUltrassonografia)
		gestacoes.POST("/:gestacao_id/parto", handler.RegistrarParto)
	}

	equinos := rg.Group("/equinos")
	equinos.Use(authMiddleware)
	{
		equinos.POST("/:equinoid/performance-materna", handler.RegistrarPerformanceMaterna)
	}
}
