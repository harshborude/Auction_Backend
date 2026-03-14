package services

import (
	"backend/models"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CancelAuction stops an auction and refunds the highest bidder if one exists
func CancelAuction(db *gorm.DB, auctionID uint) error {

	var refundedUserID *uint
	var refundAmount int64
	var wasCancelled bool

	err := db.Transaction(func(tx *gorm.DB) error {

		var auction models.Auction

		// Lock auction row
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&auction, auctionID).Error; err != nil {
			return fmt.Errorf("auction not found: %w", err)
		}

		if auction.Status != "ACTIVE" && auction.Status != "SCHEDULED" {
			return fmt.Errorf("cannot cancel auction with status: %s", auction.Status)
		}

		// Refund highest bidder
		if auction.CurrentHighestBidderID != nil {

			refundedUserID = auction.CurrentHighestBidderID
			refundAmount = auction.CurrentHighestBid

			if err := ReleaseCredits(tx, *refundedUserID, refundAmount); err != nil {
				return fmt.Errorf("failed to refund bidder during cancellation: %w", err)
			}

			if err := tx.Create(&models.CreditTransaction{
				UserID:    *refundedUserID,
				Amount:    refundAmount,
				Type:      TxBidRelease,
				Reference: fmt.Sprintf("auction_%d_cancelled", auction.ID),
			}).Error; err != nil {
				return fmt.Errorf("failed to log refund transaction: %w", err)
			}
		}

		auction.Status = "CANCELLED"

		if err := tx.Save(&auction).Error; err != nil {
			return fmt.Errorf("failed to update auction status: %w", err)
		}

		wasCancelled = true

		return nil
	})

	if err != nil {
		return err
	}

	// Broadcast event AFTER commit
	if wasCancelled {
		BroadcastAuctionEnd(auctionID, refundedUserID, 0)
	}

	return nil
}