package eventos

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {
	eventos := rg.Group("/eventos")
	eventos.Use(authMiddleware)
	{
		eventos.GET("", handler.ListAll)
		eventos.POST("", handler.Create)
		eventos.GET("/:id", handler.GetByID)
		eventos.PUT("/:id", handler.Update)
		eventos.DELETE("/:id", handler.Delete)
	}

	equinos := rg.Group("/equinos")
	equinos.Use(authMiddleware)
	{
		equinos.GET("/:equinoid/eventos", handler.ListByEquino)
	}
}
