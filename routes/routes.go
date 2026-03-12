package routes

import (
	"backend/controllers"
	"backend/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "server running",
		})
	})

	// WebSocket route
	router.GET("/ws", middleware.AuthMiddleware(), controllers.ServeWS)

	UserRoutes(router)
	AdminRoutes(router)

	return router
}