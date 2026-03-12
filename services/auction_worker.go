package services

import (
	"backend/db"
	"backend/models"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// StartAuctionWorker begins the background polling process
func StartAuctionWorker() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	log.Println("Auction background worker started...")

	for {
		select {
		case <-ticker.C:
			processExpiredAuctions()
		}
	}
}

func processExpiredAuctions() {
	var auctions []models.Auction

	// Optimistic read: Find auctions that need to be closed (fetch IDs only for efficiency)
	err := db.DB.Model(&models.Auction{}).
		Select("id").
		Where("status = ? AND end_time <= ?", "ACTIVE", time.Now()).
		Limit(50).
		Find(&auctions).Error

	if err != nil {
		log.Printf("Worker error fetching expired auctions: %v\n", err)
		return
	}

	for _, auction := range auctions {
		err := FinalizeAuction(db.DB, auction.ID)
		if err != nil {
			log.Printf("Failed to finalize auction %d: %v\n", auction.ID, err)
		} else {
			log.Printf("Successfully finalized auction %d\n", auction.ID)
		}
	}
}

// FinalizeAuction handles the transactional settlement of an ended auction
func FinalizeAuction(db *gorm.DB, auctionID uint) error {

	return db.Transaction(func(tx *gorm.DB) error {

		var auction models.Auction

		// Step 1: Lock auction row
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&auction, auctionID).Error; err != nil {
			return err
		}

		// Step 2: Double-check status and time to prevent race conditions
		if auction.Status != "ACTIVE" {
			return nil
		}

		if time.Now().Before(auction.EndTime) {
			return nil
		}

		// Step 3: Settle with the winner (if one exists)
		if auction.CurrentHighestBidderID != nil {
			winnerID := *auction.CurrentHighestBidderID
			winningAmount := auction.CurrentHighestBid

			// Deduct the reserved credits permanently
			if err := DeductReservedCredits(tx, winnerID, winningAmount); err != nil {
				return fmt.Errorf("failed to deduct winning credits: %w", err)
			}

			// Record the win transaction for the audit trail
			if err := tx.Create(&models.CreditTransaction{
				UserID:    winnerID,
				Amount:    winningAmount,
				Type:      "AUCTION_WIN", // Or TxAuctionWin if constants are shared in services
				Reference: fmt.Sprintf("auction_%d_won", auction.ID),
			}).Error; err != nil {
				return fmt.Errorf("failed to log win transaction: %w", err)
			}
		}

		// Step 4: Mark auction as ENDED
		auction.Status = "ENDED"

		if err := tx.Save(&auction).Error; err != nil {
			return fmt.Errorf("failed to update auction status: %w", err)
		}

		return nil
	})
}