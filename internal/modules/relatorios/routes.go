package relatorios

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {
	reports := rg.Group("/reports")
	reports.Use(authMiddleware)
	{
		reports.GET("/dashboard", handler.GetDashboardStats)
	}
}
