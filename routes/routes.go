package routes

import "github.com/gin-gonic/gin"

func SetupRouter() *gin.Engine {

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "server running",
		})
	})

	UserRoutes(router)
	AdminRoutes(router)
	return router
}