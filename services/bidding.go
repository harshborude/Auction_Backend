package services

import (
	"backend/models"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	TxBidReserve = "BID_RESERVE"
	TxBidRelease = "BID_RELEASE"
	TxAuctionWin = "AUCTION_WIN"
)

// PlaceBid handles the entire transactional bidding process
func PlaceBid(db *gorm.DB, auctionID uint, userID uint, bidAmount int64) (*models.Bid, error) {
	var placedBid *models.Bid

	err := db.Transaction(func(tx *gorm.DB) error {
		
		// 1. Lock Auction Row
		var auction models.Auction
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", auctionID).First(&auction).Error; err != nil {
			return fmt.Errorf("auction not found: %w", err)
		}

		// 2. Validate Auction Status & Timer
		if auction.Status != "ACTIVE" {
			return errors.New("auction is closed")
		}
		
		now := time.Now()
		if now.Before(auction.StartTime) || now.After(auction.EndTime) {
			return errors.New("auction is not currently open for bidding")
		}

		if auction.CreatedBy == userID {
			return errors.New("sellers cannot bid on their own auctions")
		}

		// 3. Validate Bid Amount
		var minRequiredBid int64
		if auction.BidCount == 0 {
			minRequiredBid = auction.StartingPrice
		} else {
			minRequiredBid = auction.CurrentHighestBid + auction.BidIncrement
		}

		if bidAmount < minRequiredBid {
			return fmt.Errorf("bid amount too low. minimum required bid is %d", minRequiredBid)
		}

		// 4. Handle Credit Reservations & Refunds
		isSelfOutbid := auction.CurrentHighestBidderID != nil && *auction.CurrentHighestBidderID == userID

		if isSelfOutbid && bidAmount <= auction.CurrentHighestBid {
			return errors.New("new bid must be higher than your current bid")
		}

		if isSelfOutbid {
			// Reserve only the difference
			difference := bidAmount - auction.CurrentHighestBid
			
			if _, err := ReserveCredits(tx, userID, difference); err != nil {
				return fmt.Errorf("insufficient credits to increase bid: %w", err)
			}

			if err := tx.Create(&models.CreditTransaction{
				UserID:    userID,
				Amount:    difference,
				Type:      TxBidReserve,
				Reference: fmt.Sprintf("auction_%d_increase", auction.ID),
			}).Error; err != nil {
				return fmt.Errorf("failed to record credit transaction: %w", err)
			}

		} else {
			// Reserve full amount for the new bidder
			if _, err := ReserveCredits(tx, userID, bidAmount); err != nil {
				return fmt.Errorf("insufficient credits: %w", err)
			}

			if err := tx.Create(&models.CreditTransaction{
				UserID:    userID,
				Amount:    bidAmount,
				Type:      TxBidReserve,
				Reference: fmt.Sprintf("auction_%d", auction.ID),
			}).Error; err != nil {
				return fmt.Errorf("failed to record credit transaction: %w", err)
			}

			// Release Previous Bidder
			if auction.CurrentHighestBidderID != nil {
				prevBidderID := *auction.CurrentHighestBidderID
				prevBidAmount := auction.CurrentHighestBid

				if err := ReleaseCredits(tx, prevBidderID, prevBidAmount); err != nil {
					return fmt.Errorf("failed to refund previous bidder: %w", err)
				}

				if err := tx.Create(&models.CreditTransaction{
					UserID:    prevBidderID,
					Amount:    prevBidAmount,
					Type:      TxBidRelease,
					Reference: fmt.Sprintf("auction_%d_refund", auction.ID),
				}).Error; err != nil {
					return fmt.Errorf("failed to record refund transaction: %w", err)
				}
			}
		}

		// 5. Insert Bid Record
		bid := models.Bid{
			AuctionID: auction.ID,
			UserID:    userID,
			Amount:    bidAmount,
		}
		if err := tx.Create(&bid).Error; err != nil {
			return fmt.Errorf("failed to record bid: %w", err)
		}
		
		placedBid = &bid 

		// 6. Update Auction
		auction.BidCount++

		if err := tx.Model(&auction).Updates(map[string]interface{}{
			"current_highest_bid":       bidAmount,
			"current_highest_bidder_id": userID,
			"bid_count":                 auction.BidCount,
		}).Error; err != nil {
			return fmt.Errorf("failed to update auction state: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return placedBid, nil
}