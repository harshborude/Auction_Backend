package controllers

import (
	"backend/services"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// WARNING: Allow all origins for development/hackathon.
	// In production, restrict this to allowed domains.
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func ServeWS(c *gin.Context) {

	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user context"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}

	client := &services.Client{
		Hub:      services.AuctionHub,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Auctions: make(map[uint]bool),
		UserID:   userID,
	}

	log.Printf("WebSocket connected: user %d", userID)

	client.Hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}