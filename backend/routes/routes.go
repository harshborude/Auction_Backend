package routes

import (
	"backend/controllers"
	"backend/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12,
	}))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "server running",
		})
	})

	// WebSocket route
	router.GET("/ws", middleware.AuthMiddleware(), controllers.ServeWS)

	UserRoutes(router)
	AdminRoutes(router)
	AuctionRoutes(router)

	return router
}
