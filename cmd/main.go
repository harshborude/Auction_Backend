package main

import (
	"backend/db"
	"backend/routes"
	"backend/services"
	"backend/utils"
	"log"

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

	go services.StartAuctionWorker()
	log.Println("Auction worker started")

	services.InitHub()
	go services.AuctionHub.Run()
	log.Println("WebSocket hub started")

	router := routes.SetupRouter()

	log.Println("Server running on port 8080")

	if err := router.Run(":8080"); err != nil {
		log.Fatal("server failed:", err)
	}
}