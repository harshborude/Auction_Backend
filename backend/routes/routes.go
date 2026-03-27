package routes

import (
	"backend/controllers"
	"backend/middleware"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {

	router := gin.Default()

	allowedOrigins := []string{"http://localhost:5173"}
	if raw := os.Getenv("ALLOWED_ORIGINS"); raw != "" {
		allowedOrigins = strings.Split(raw, ",")
	}

	router.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
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

	// Static file serving for local dev uploads
	router.Static("/uploads", "./uploads")

	// Image upload (local dev fallback — production uses Cloudinary)
	router.POST("/upload", middleware.AuthMiddleware(), controllers.UploadImage)

	// WebSocket route
	router.GET("/ws", middleware.AuthMiddleware(), controllers.ServeWS)

	UserRoutes(router)
	AdminRoutes(router)
	AuctionRoutes(router)

	return router
}
