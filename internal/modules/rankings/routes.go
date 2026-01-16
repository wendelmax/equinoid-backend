package rankings

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {
	rankings := rg.Group("/rankings")
	rankings.Use(authMiddleware)
	{
		rankings.GET("/:tipo", handler.GetRankingGeral)
	}

	equinos := rg.Group("/equinos")
	equinos.Use(authMiddleware)
	{
		equinos.GET("/:equinoid/rankings", handler.GetRankingsEquino)
	}
}
