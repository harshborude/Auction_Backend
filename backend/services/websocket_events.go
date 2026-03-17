package services

import (
	"log"
	"time"
)

// safeBroadcast prevents API handlers or workers from blocking if the WebSocket hub is overloaded.
// If the broadcast channel is full, the message will be dropped instead of blocking the caller.
func safeBroadcast(msg Message) {
	if AuctionHub == nil {
		return
	}

	select {
	case AuctionHub.Broadcast <- msg:
		// Optional debug log
		log.Printf("Broadcasted %s event for auction %d", msg.Type, msg.AuctionID)

	default:
		// Drop the message instead of blocking
		log.Printf("Warning: WebSocket hub is busy, dropped %s event for auction %d", msg.Type, msg.AuctionID)
	}
}

// BroadcastBidUpdate sends a BID_UPDATE event when a new bid is placed.
// This updates the current price for everyone watching the auction.
func BroadcastBidUpdate(auctionID uint, amount int64, bidderID uint) {
	msg := Message{
		Type:      "BID_UPDATE",
		AuctionID: auctionID,
		Amount:    amount,
		BidderID:  bidderID,
	}

	safeBroadcast(msg)
}

// BroadcastAuctionEnd sends an AUCTION_END event when an auction finishes.
// All connected clients watching the auction will receive the final result.
func BroadcastAuctionEnd(auctionID uint, winnerID *uint, finalPrice int64) {
	msg := Message{
		Type:      "AUCTION_END",
		AuctionID: auctionID,
		Amount:    finalPrice,
	}

	// Attach winner ID only if there was a winning bidder
	if winnerID != nil {
		msg.BidderID = *winnerID
	}

	safeBroadcast(msg)
}

// BroadcastAuctionStart notifies clients that a scheduled auction has become ACTIVE.
func BroadcastAuctionStart(auctionID uint) {
	msg := Message{
		Type:      "AUCTION_STARTED",
		AuctionID: auctionID,
	}

	safeBroadcast(msg)
}

// BroadcastOutbid notifies that a user has been outbid.
// NOTE: BidderID here represents the user who was outbid (not the new highest bidder).
// The frontend can check if the logged-in user matches this ID to show a notification.
func BroadcastOutbid(auctionID uint, outbidUserID uint, newBidAmount int64) {
	msg := Message{
		Type:      "OUTBID",
		AuctionID: auctionID,
		Amount:    newBidAmount,
		BidderID:  outbidUserID,
	}

	safeBroadcast(msg)
}

func BroadcastAuctionExtended(auctionID uint, newEndTime time.Time) {
	msg := Message{
		Type:      "AUCTION_EXTENDED",
		AuctionID: auctionID,
		EndTime:   newEndTime,
	}

	safeBroadcast(msg)
}
