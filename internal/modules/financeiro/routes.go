package financeiro

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {
	financeiro := rg.Group("/financeiro")
	financeiro.Use(authMiddleware)
	{
		financeiro.GET("/stats", handler.GetStats)
		financeiro.GET("/monthly", handler.GetMonthlyData)
		financeiro.GET("/breakdown", handler.GetExpenseBreakdown)
		financeiro.GET("/transactions", handler.ListTransactions)
		financeiro.POST("/transactions", handler.CreateTransaction)
	}
}
