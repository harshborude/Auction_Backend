package services

import (
	"backend/cache"
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

	antiSnipeWindow        = 30 * time.Second
	antiSnipeExtension     = 30 * time.Second
	maxAntiSnipeExtensions = 10

	// How long we hold the Redis bid lock.
	// Must comfortably exceed the worst-case DB transaction time.
	bidLockTTL = 10 * time.Second
)

func PlaceBid(db *gorm.DB, auctionID uint, userID uint, bidAmount int64) (*models.Bid, error) {

	// ── Step 1: Acquire the distributed Redis lock ──────────────────────────
	//
	// Only one goroutine (across all server instances) can hold this lock at a
	// time. Concurrent bid attempts for the same auction wait here instead of
	// stacking up inside PostgreSQL's row-lock queue.

	token, acquired := cache.AcquireLock(auctionID, bidLockTTL)
	if !acquired {
		return nil, errors.New("auction is busy, please try again in a moment")
	}
	defer cache.ReleaseLock(auctionID, token)

	// ── Step 2: Load auction state ──────────────────────────────────────────
	//
	// Attempt to read from the Redis cache first. If the cache is cold (first
	// bid after a restart, or after the cache was invalidated), fall back to
	// PostgreSQL and populate the cache for subsequent bids.

	state, hit, err := cache.GetAuctionState(auctionID)
	if err != nil {
		return nil, fmt.Errorf("cache read error: %w", err)
	}

	if !hit {
		var auction models.Auction
		if err := db.First(&auction, auctionID).Error; err != nil {
			return nil, fmt.Errorf("auction not found: %w", err)
		}
		state = modelToState(&auction)
		// Cache until 1 hour past the auction end time
		cache.SetAuctionState(auctionID, state, time.Until(auction.EndTime)+time.Hour)
	}

	// ── Step 3: Fast validation against cached state ────────────────────────
	//
	// All of these checks operate entirely in memory — no DB round-trip needed.
	// Invalid bids are rejected here before we ever touch PostgreSQL.

	now := time.Now()

	if state.Status != "ACTIVE" {
		return nil, errors.New("auction is closed")
	}
	if now.After(state.EndTime) {
		return nil, errors.New("auction is not currently open for bidding")
	}
	if state.CreatedBy == userID {
		return nil, errors.New("sellers cannot bid on their own auctions")
	}

	isSelfOutbid := state.CurrentHighestBidderID != nil && *state.CurrentHighestBidderID == userID

	var minRequiredBid int64
	if state.BidCount == 0 {
		minRequiredBid = state.StartingPrice
	} else {
		minRequiredBid = state.CurrentHighestBid + state.BidIncrement
	}

	if isSelfOutbid && bidAmount <= state.CurrentHighestBid {
		return nil, errors.New("new bid must be higher than your current bid")
	}
	if !isSelfOutbid && bidAmount < minRequiredBid {
		return nil, fmt.Errorf("bid amount too low, minimum required bid is %d", minRequiredBid)
	}

	// ── Step 4: Commit to PostgreSQL ────────────────────────────────────────
	//
	// Financial operations (reserve/release credits, insert bid, update auction)
	// run inside a DB transaction.
	//
	// We keep SELECT FOR UPDATE on the auction row as a safety net against the
	// background worker or admin operations that don't acquire the Redis lock.
	// The Redis lock already prevents concurrent bids, so this lock is
	// essentially free here — it never contends with another bid — but it does
	// protect against the rare race where the worker finalizes the auction at
	// the exact same millisecond as an incoming bid.

	var placedBid *models.Bid
	var outbidUserID *uint
	var newEndTime time.Time
	extended := false

	err = db.Transaction(func(tx *gorm.DB) error {
		var auction models.Auction
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&auction, auctionID).Error; err != nil {
			return fmt.Errorf("auction not found: %w", err)
		}

		// Re-check status inside the transaction — the worker may have closed
		// the auction between our cache read and now.
		if auction.Status != "ACTIVE" {
			return errors.New("auction is closed")
		}
		if now.After(auction.EndTime) {
			return errors.New("auction is not currently open for bidding")
		}

		if isSelfOutbid {
			// Raising own bid: only reserve the difference
			diff := bidAmount - auction.CurrentHighestBid
			if _, err := ReserveCredits(tx, userID, diff); err != nil {
				return fmt.Errorf("insufficient credits to increase bid: %w", err)
			}
			if err := tx.Create(&models.CreditTransaction{
				UserID:    userID,
				Amount:    diff,
				Type:      TxBidReserve,
				Reference: fmt.Sprintf("auction_%d_increase", auction.ID),
			}).Error; err != nil {
				return err
			}
		} else {
			// New highest bidder: reserve full amount, refund previous bidder
			if _, err := ReserveCredits(tx, userID, bidAmount); err != nil {
				return fmt.Errorf("insufficient credits: %w", err)
			}
			if err := tx.Create(&models.CreditTransaction{
				UserID:    userID,
				Amount:    bidAmount,
				Type:      TxBidReserve,
				Reference: fmt.Sprintf("auction_%d", auction.ID),
			}).Error; err != nil {
				return err
			}

			if auction.CurrentHighestBidderID != nil {
				prevID := *auction.CurrentHighestBidderID
				outbidUserID = &prevID
				if err := ReleaseCredits(tx, prevID, auction.CurrentHighestBid); err != nil {
					return fmt.Errorf("failed to refund previous bidder: %w", err)
				}
				if err := tx.Create(&models.CreditTransaction{
					UserID:    prevID,
					Amount:    auction.CurrentHighestBid,
					Type:      TxBidRelease,
					Reference: fmt.Sprintf("auction_%d_refund", auction.ID),
				}).Error; err != nil {
					return err
				}
			}
		}

		bid := models.Bid{AuctionID: auction.ID, UserID: userID, Amount: bidAmount}
		if err := tx.Create(&bid).Error; err != nil {
			return fmt.Errorf("failed to record bid: %w", err)
		}
		placedBid = &bid
		auction.BidCount++

		updates := map[string]interface{}{
			"current_highest_bid":       bidAmount,
			"current_highest_bidder_id": userID,
			"bid_count":                 auction.BidCount,
		}

		// Anti-sniping: extend the auction if bid is within the final window
		newEndTime = auction.EndTime
		if auction.EndTime.Sub(now) <= antiSnipeWindow && auction.ExtensionCount < maxAntiSnipeExtensions {
			newEndTime = auction.EndTime.Add(antiSnipeExtension)
			updates["end_time"] = newEndTime
			updates["extension_count"] = auction.ExtensionCount + 1
			extended = true
		}

		return tx.Model(&auction).Updates(updates).Error
	})

	if err != nil {
		return nil, err
	}

	// ── Step 5: Update the cache (while we still hold the Redis lock) ────────
	//
	// Patch only the bid-related fields. The Redis lock ensures no other bid
	// can read a stale state between the DB commit above and this cache update.

	newExtCount := state.ExtensionCount
	if extended {
		newExtCount++
	}
	cache.UpdateAuctionBid(auctionID, bidAmount, userID, state.BidCount+1, newExtCount, newEndTime)

	// ── Step 6: Broadcast WebSocket events ──────────────────────────────────

	BroadcastBidUpdate(auctionID, bidAmount, userID)

	if outbidUserID != nil && *outbidUserID != userID {
		BroadcastOutbid(auctionID, *outbidUserID, bidAmount)
	}
	if extended {
		BroadcastAuctionExtended(auctionID, newEndTime)
	}

	return placedBid, nil
}

// modelToState converts a models.Auction into the lightweight cache struct.
func modelToState(a *models.Auction) *cache.AuctionState {
	return &cache.AuctionState{
		CurrentHighestBid:      a.CurrentHighestBid,
		CurrentHighestBidderID: a.CurrentHighestBidderID,
		BidCount:               a.BidCount,
		ExtensionCount:         a.ExtensionCount,
		Status:                 a.Status,
		StartingPrice:          a.StartingPrice,
		BidIncrement:           a.BidIncrement,
		EndTime:                a.EndTime,
		CreatedBy:              a.CreatedBy,
	}
}
