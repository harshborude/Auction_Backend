package routes

import (
	"backend/controllers"
	"backend/middleware"

	"github.com/gin-gonic/gin"
)

func UserRoutes(router *gin.Engine) {

	users := router.Group("/users")

	{
		users.POST("/register", controllers.RegisterUser)
		users.POST("/login", controllers.LoginUser)
		users.POST("/refresh", controllers.RefreshAccessToken)
		users.POST("/logout", middleware.AuthMiddleware(), controllers.LogoutUser)
	}

	auth := router.Group("/users")
	auth.Use(middleware.AuthMiddleware())

	{
		auth.GET("/me", controllers.GetCurrentUser)
		auth.PATCH("/change-password", controllers.ChangePassword)
	}
}