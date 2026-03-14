package routes

import (
	"backend/controllers"
	"backend/middleware"

	"github.com/gin-gonic/gin"
)

func AdminRoutes(router *gin.Engine) {

	admin := router.Group("/admin")

	admin.Use(
		middleware.AuthMiddleware(),
		middleware.RoleRequired("ADMIN"),
	)

	{
		admin.PATCH("/promote/:user_id", controllers.PromoteUser)
		admin.GET("/users", controllers.GetUsers)
		admin.PATCH("/users/:user_id/credits", controllers.AssignCredits)
		admin.POST("/auctions", controllers.CreateAuction)
		admin.POST("/auctions/:id/end", controllers.ForceCloseAuction)
		admin.POST("/auctions/:id/cancel", controllers.CancelAuction)
	}
}
