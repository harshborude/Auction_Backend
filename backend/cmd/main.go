package main

import (
	"backend/cache"
	"backend/db"
	"backend/routes"
	"backend/services"
	"backend/utils"
	"log"

	// "github.com/gin-contrib/cors"
	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	utils.InitJWT()
	log.Println("JWT initialized")

	db.ConnectDatabase()
	log.Println("Database connected")

	cache.Connect()
	log.Println("Redis connected")

	go services.StartAuctionWorker()
	log.Println("Auction worker started")

	services.InitHub()
	go services.AuctionHub.Run()
	log.Println("WebSocket hub started")

	router := routes.SetupRouter()

	// router.Use(cors.New(cors.Config{
	// 	AllowOrigins:     []string{"http://localhost:5173"},
	// 	AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	// 	AllowHeaders:     []string{"Content-Type", "Authorization"},
	// 	ExposeHeaders:    []string{"Content-Length"},
	// 	AllowCredentials: true,
	// 	MaxAge:           12,
	// }))

	log.Println("Server running on port 8080")

	if err := router.Run(":8080"); err != nil {
		log.Fatal("server failed:", err)
	}
}
