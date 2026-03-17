package services

import (
	"encoding/json"
	"log"

	"time"

	"github.com/gorilla/websocket"
)

// Message defines the JSON protocol between client and server
type Message struct {
	Type      string    `json:"type"` // e.g., JOIN_AUCTION, LEAVE_AUCTION, BID_UPDATE, AUCTION_END
	AuctionID uint      `json:"auction_id"`
	Amount    int64     `json:"amount,omitempty"`
	BidderID  uint      `json:"bidder_id,omitempty"`
	Message   string    `json:"message,omitempty"`
	EndTime   time.Time `json:"end_time,omitempty"` // Added to tell frontend the new clock time
}

// Client wraps the WebSocket connection and state
type Client struct {
	Hub      *Hub
	Conn     *websocket.Conn
	Send     chan []byte
	Auctions map[uint]bool // Tracks which auctions this client has joined
	UserID   uint          // Useful for targeted notifications later
}

// RoomAction is used to safely join/leave auction rooms via channels
type RoomAction struct {
	Client    *Client
	AuctionID uint
}

// Hub maintains the set of active clients and broadcasts messages to the rooms
type Hub struct {
	Clients      map[*Client]bool
	AuctionRooms map[uint]map[*Client]bool // Map of AuctionID -> Connected Clients

	Broadcast  chan Message
	Register   chan *Client
	Unregister chan *Client
	JoinRoom   chan RoomAction
	LeaveRoom  chan RoomAction
}

// Global Hub instance
var AuctionHub *Hub

// InitHub initializes the global Hub
func InitHub() {
	AuctionHub = &Hub{
		Clients:      make(map[*Client]bool),
		AuctionRooms: make(map[uint]map[*Client]bool),
		Broadcast:    make(chan Message),
		Register:     make(chan *Client),
		Unregister:   make(chan *Client),
		JoinRoom:     make(chan RoomAction),
		LeaveRoom:    make(chan RoomAction),
	}
}

// Run listens on channels and manages state (Concurrent-safe)
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = true

		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				// 1. Remove from master client list
				delete(h.Clients, client)
				close(client.Send)

				// 2. Remove client from all auction rooms they were in
				for auctionID := range client.Auctions {
					if rooms, exists := h.AuctionRooms[auctionID]; exists {
						delete(rooms, client)
						// Clean up empty rooms to save memory
						if len(rooms) == 0 {
							delete(h.AuctionRooms, auctionID)
						}
					}
				}

				// 3. Clear the client's map to prevent reuse issues
				client.Auctions = make(map[uint]bool)
			}

		case action := <-h.JoinRoom:
			// Prevent duplicate joins
			if !action.Client.Auctions[action.AuctionID] {
				// Create room if it doesn't exist
				if h.AuctionRooms[action.AuctionID] == nil {
					h.AuctionRooms[action.AuctionID] = make(map[*Client]bool)
				}
				// Add client to room and track it in the client's state
				h.AuctionRooms[action.AuctionID][action.Client] = true
				action.Client.Auctions[action.AuctionID] = true
				log.Printf("Client joined auction %d", action.AuctionID)
			}

		case action := <-h.LeaveRoom:
			// Remove client from specific room
			if rooms, ok := h.AuctionRooms[action.AuctionID]; ok {
				delete(rooms, action.Client)
				if len(rooms) == 0 {
					delete(h.AuctionRooms, action.AuctionID)
				}
			}
			delete(action.Client.Auctions, action.AuctionID)
			log.Printf("Client left auction %d", action.AuctionID)

		case message := <-h.Broadcast:
			// Handle JSON Marshal error properly before trying to send
			msgBytes, err := json.Marshal(message)
			if err != nil {
				log.Printf("Broadcast marshal error: %v", err)
				continue
			}

			// Broadcast only to clients in the specific auction room
			if roomClients, ok := h.AuctionRooms[message.AuctionID]; ok {
				for client := range roomClients {
					select {
					case client.Send <- msgBytes:
					default:
						// If the client's send buffer is full/blocked, disconnect them
						close(client.Send)
						delete(h.Clients, client)
						delete(roomClients, client)
						// Remove from client's internal map for total consistency
						delete(client.Auctions, message.AuctionID)
					}
				}
			}
		}
	}
}
