package services

func BroadcastBidUpdate(auctionID uint, amount int64, bidderID uint) {
	if AuctionHub == nil {
		return
	}

	msg := Message{
		Type:      "BID_UPDATE",
		AuctionID: auctionID,
		Amount:    amount,
		BidderID:  bidderID,
	}

	AuctionHub.Broadcast <- msg
}