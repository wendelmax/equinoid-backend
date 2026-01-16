package users

import (
	"github.com/equinoid/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {
	users := rg.Group("/users")
	users.Use(authMiddleware)
	{
		users.GET("/me", handler.GetProfile)
		users.PUT("/me", handler.UpdateProfile)
		users.DELETE("/me", handler.DeleteAccount)
		users.POST("/me/change-password", handler.ChangePassword)
		users.GET("/check-email", handler.CheckEmailAvailability)
		users.GET("", handler.ListUsers)
		
		adminProtected := users.Group("")
		adminProtected.Use(middleware.RequireAdminMiddleware())
		{
			adminProtected.POST("", handler.CreateUser)
			adminProtected.GET("/:id", handler.GetUserByID)
			adminProtected.PUT("/:id", handler.UpdateUserByID)
			adminProtected.DELETE("/:id", handler.DeleteUserByID)
			adminProtected.PUT("/:id/activate", handler.ActivateUser)
			adminProtected.PUT("/:id/deactivate", handler.DeactivateUser)
		}
	}
}
