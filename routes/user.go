package routes

import (
	"backend/controllers"
	"backend/middleware"

	"github.com/gin-gonic/gin"
)

func UserRoutes(router *gin.Engine) {

	users := router.Group("/users")

	// Public user routes
	{
		users.POST("/register", controllers.RegisterUser)
		users.POST("/login", controllers.LoginUser)
		users.POST("/refresh", controllers.RefreshAccessToken)
	}

	// Authenticated user routes
	auth := users.Group("/")
	auth.Use(middleware.AuthMiddleware())
	{
		auth.POST("/logout", controllers.LogoutUser)

		auth.GET("/me", controllers.GetCurrentUser)
		auth.PATCH("/change-password", controllers.ChangePassword)

		// Wallet APIs
		auth.GET("/wallet", controllers.GetWallet)
		auth.GET("/wallet/transactions", controllers.GetWalletTransactions)
	}
}