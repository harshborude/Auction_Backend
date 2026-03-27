package services

import (
	"backend/cache"
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
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	log.Println("Auction background worker started...")

	for range ticker.C {
		processScheduledAuctions()
		processExpiredAuctions()
	}
}

func processExpiredAuctions() {
	var auctions []models.Auction

	// Optimistic read: Find auctions that need to be closed (fetch IDs only for efficiency)
	err := db.DB.
		Model(&models.Auction{}).
		Select("id").
		Where("status = ? AND end_time <= ?", "ACTIVE", time.Now()).
		Order("end_time ASC").
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

func processScheduledAuctions() {
	var auctions []models.Auction

	err := db.DB.
		Model(&models.Auction{}).
		Select("id").
		Where("status = ? AND start_time <= ?", "SCHEDULED", time.Now()).
		Order("start_time ASC").
		Limit(50).
		Find(&auctions).Error

	if err != nil {
		log.Printf("Worker error fetching scheduled auctions: %v\n", err)
		return
	}

	for _, auction := range auctions {
		err := ActivateAuction(db.DB, auction.ID)
		if err != nil {
			log.Printf("Failed to activate auction %d: %v\n", auction.ID, err)
		} else {
			log.Printf("Activated auction %d\n", auction.ID)
		}
	}
}

// FinalizeAuction handles the transactional settlement of an ended auction
func FinalizeAuction(db *gorm.DB, auctionID uint) error {
	var winnerID *uint
	var finalPrice int64
	var wasActive bool

	err := db.Transaction(func(tx *gorm.DB) error {
		var auction models.Auction

		// Step 1: Lock auction row
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&auction, auctionID).Error; err != nil {
			return err
		}

		// Step 2: Prevent double-closing
		if auction.Status != "ACTIVE" {
			return nil
		}

		if time.Now().Before(auction.EndTime) {
			return nil
		}

		wasActive = true

		// Step 3: Settle winner
		if auction.CurrentHighestBidderID != nil {
			// Pointer safety: create a local copy of the ID
			tempID := *auction.CurrentHighestBidderID
			winnerID = &tempID

			finalPrice = auction.CurrentHighestBid

			if err := DeductReservedCredits(tx, *winnerID, finalPrice); err != nil {
				return fmt.Errorf("failed to deduct winning credits: %w", err)
			}

			if err := tx.Create(&models.CreditTransaction{
				UserID:    *winnerID,
				Amount:    finalPrice,
				Type:      "AUCTION_WIN",
				Reference: fmt.Sprintf("auction_%d_won", auction.ID),
			}).Error; err != nil {
				return fmt.Errorf("failed to log win transaction: %w", err)
			}
		}

		// Step 4: Close auction using Update instead of Save
		if err := tx.Model(&auction).Update("status", "ENDED").Error; err != nil {
			return fmt.Errorf("failed to update auction status: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	if wasActive {
		// Remove from cache so no further bids can be validated against stale state
		cache.DeleteAuctionState(auctionID)
		BroadcastAuctionEnd(auctionID, winnerID, finalPrice)
	}

	return nil
}

func ActivateAuction(db *gorm.DB, auctionID uint) error {
	var wasScheduled bool

	err := db.Transaction(func(tx *gorm.DB) error {
		var auction models.Auction

		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&auction, auctionID).Error; err != nil {
			return err
		}

		if auction.Status != "SCHEDULED" {
			return nil
		}

		if time.Now().Before(auction.StartTime) {
			return nil
		}

		wasScheduled = true

		if err := tx.Model(&auction).Update("status", "ACTIVE").Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	if wasScheduled {
		// Update status in cache so bids can immediately be validated
		cache.UpdateAuctionStatus(auctionID, "ACTIVE")
		BroadcastAuctionStart(auctionID)
	}

	return nil
}
