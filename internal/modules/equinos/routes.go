package equinos

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {
	equinos := rg.Group("/equinos")
	equinos.Use(authMiddleware)
	{
		equinos.GET("", handler.ListEquinos)
		equinos.POST("", handler.CreateEquino)
		equinos.GET("/:equinoid", handler.GetEquino)
		equinos.PUT("/:equinoid", handler.UpdateEquino)
		equinos.DELETE("/:equinoid", handler.DeleteEquino)
		equinos.POST("/:equinoid/transferir", handler.TransferOwnership)
	}
}
