package services

import (
	"backend/models"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// AdminForceCloseAuction ends an ACTIVE auction immediately, bypassing the end-time check.
// Unlike FinalizeAuction (used by the background worker), this ignores whether the
// auction's EndTime has passed — it forces settlement right now.
func AdminForceCloseAuction(db *gorm.DB, auctionID uint) error {
	var winnerID *uint
	var finalPrice int64
	var wasActive bool

	err := db.Transaction(func(tx *gorm.DB) error {
		var auction models.Auction

		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&auction, auctionID).Error; err != nil {
			return fmt.Errorf("auction not found: %w", err)
		}

		if auction.Status != "ACTIVE" {
			return fmt.Errorf("can only force-close ACTIVE auctions, current status: %s", auction.Status)
		}

		wasActive = true

		if auction.CurrentHighestBidderID != nil {
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
				Reference: fmt.Sprintf("auction_%d_force_ended", auction.ID),
			}).Error; err != nil {
				return fmt.Errorf("failed to log win transaction: %w", err)
			}
		}

		if err := tx.Model(&auction).Update("status", "ENDED").Error; err != nil {
			return fmt.Errorf("failed to update auction status: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	if wasActive {
		BroadcastAuctionEnd(auctionID, winnerID, finalPrice)
	}

	return nil
}

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