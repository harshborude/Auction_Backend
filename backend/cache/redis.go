package cache

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

// Lua script: delete the lock only if we are still the owner.
// This makes the check-and-delete atomic — no race between checking and deleting.
var releaseLockScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
    return redis.call("DEL", KEYS[1])
else
    return 0
end
`)

func Connect() {
	var opts *redis.Options

	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	opts = &redis.Options{
		Addr:     addr,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	}

	if os.Getenv("REDIS_TLS") == "true" {
		opts.TLSConfig = &tls.Config{}
	}

	RDB = redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := RDB.Ping(ctx).Err(); err != nil {
		log.Fatalf("Redis connection failed: %v", err)
	}

	log.Println("Redis connected successfully")
}

// ─── Distributed Lock ────────────────────────────────────────────────────────
//
// We use the standard SET key value NX EX pattern:
//   NX  = only set if key does not exist (atomic "grab if free")
//   EX  = auto-expire after TTL seconds (safety net if the server crashes
//         mid-bid and never releases the lock)
//
// The value is a random token unique to this acquisition. The release script
// checks the token before deleting, so one caller can never release another
// caller's lock.

func lockKey(auctionID uint) string {
	return fmt.Sprintf("auction:%d:lock", auctionID)
}

// AcquireLock tries to grab the distributed bid lock for auctionID.
// Returns (token, true) on success, ("", false) if the lock is already held.
func AcquireLock(auctionID uint, ttl time.Duration) (string, bool) {
	b := make([]byte, 16)
	rand.Read(b)
	token := hex.EncodeToString(b)

	ok, err := RDB.SetNX(context.Background(), lockKey(auctionID), token, ttl).Result()
	if err != nil || !ok {
		return "", false
	}
	return token, true
}

// ReleaseLock releases the lock only if we are still the owner.
// Safe to call even if the lock already expired.
func ReleaseLock(auctionID uint, token string) {
	releaseLockScript.Run(
		context.Background(), RDB,
		[]string{lockKey(auctionID)},
		token,
	)
}

// ─── Auction State Cache ─────────────────────────────────────────────────────
//
// We cache the hot fields needed for bid validation so each bid doesn't require
// a PostgreSQL round-trip just to check "is the auction active and what is the
// current highest bid?". The cache is stored as a Redis Hash so individual
// fields can be updated cheaply (e.g. after a new bid, we only HSET the
// bid-related fields rather than rewriting the whole entry).

// AuctionState holds every field required to validate and process a bid.
type AuctionState struct {
	CurrentHighestBid      int64
	CurrentHighestBidderID *uint
	BidCount               int64
	ExtensionCount         int
	Status                 string
	StartingPrice          int64
	BidIncrement           int64
	EndTime                time.Time
	CreatedBy              uint
}

func stateKey(auctionID uint) string {
	return fmt.Sprintf("auction:%d:state", auctionID)
}

// GetAuctionState retrieves cached auction state.
// Returns (state, true, nil) on hit, (nil, false, nil) on miss.
func GetAuctionState(auctionID uint) (*AuctionState, bool, error) {
	vals, err := RDB.HGetAll(context.Background(), stateKey(auctionID)).Result()
	if err != nil {
		return nil, false, err
	}
	if len(vals) == 0 {
		return nil, false, nil // cache miss
	}

	s := &AuctionState{}

	if v := vals["current_highest_bid"]; v != "" {
		s.CurrentHighestBid, _ = strconv.ParseInt(v, 10, 64)
	}
	if v := vals["current_highest_bidder_id"]; v != "" {
		id, _ := strconv.ParseUint(v, 10, 64)
		uid := uint(id)
		s.CurrentHighestBidderID = &uid
	}
	if v := vals["bid_count"]; v != "" {
		s.BidCount, _ = strconv.ParseInt(v, 10, 64)
	}
	if v := vals["extension_count"]; v != "" {
		n, _ := strconv.Atoi(v)
		s.ExtensionCount = n
	}
	s.Status = vals["status"]
	if v := vals["starting_price"]; v != "" {
		s.StartingPrice, _ = strconv.ParseInt(v, 10, 64)
	}
	if v := vals["bid_increment"]; v != "" {
		s.BidIncrement, _ = strconv.ParseInt(v, 10, 64)
	}
	if v := vals["end_time"]; v != "" {
		s.EndTime, _ = time.Parse(time.RFC3339Nano, v)
	}
	if v := vals["created_by"]; v != "" {
		id, _ := strconv.ParseUint(v, 10, 64)
		s.CreatedBy = uint(id)
	}

	return s, true, nil
}

// SetAuctionState writes the full auction state to the cache with a TTL.
func SetAuctionState(auctionID uint, s *AuctionState, ttl time.Duration) error {
	ctx := context.Background()
	key := stateKey(auctionID)

	bidderID := ""
	if s.CurrentHighestBidderID != nil {
		bidderID = strconv.FormatUint(uint64(*s.CurrentHighestBidderID), 10)
	}

	fields := map[string]interface{}{
		"current_highest_bid":       s.CurrentHighestBid,
		"current_highest_bidder_id": bidderID,
		"bid_count":                 s.BidCount,
		"extension_count":           s.ExtensionCount,
		"status":                    s.Status,
		"starting_price":            s.StartingPrice,
		"bid_increment":             s.BidIncrement,
		"end_time":                  s.EndTime.Format(time.RFC3339Nano),
		"created_by":                s.CreatedBy,
	}

	if err := RDB.HSet(ctx, key, fields).Err(); err != nil {
		return err
	}
	RDB.Expire(ctx, key, ttl)
	return nil
}

// UpdateAuctionBid patches only the bid-related fields in the cache.
// Called after a successful bid so we don't rewrite the entire hash.
func UpdateAuctionBid(auctionID uint, bid int64, bidderID uint, bidCount int64, extCount int, endTime time.Time) {
	RDB.HSet(context.Background(), stateKey(auctionID), map[string]interface{}{
		"current_highest_bid":       bid,
		"current_highest_bidder_id": strconv.FormatUint(uint64(bidderID), 10),
		"bid_count":                 bidCount,
		"extension_count":           extCount,
		"end_time":                  endTime.Format(time.RFC3339Nano),
	})
}

// UpdateAuctionStatus patches only the status field (used when worker activates/ends an auction).
func UpdateAuctionStatus(auctionID uint, status string) {
	RDB.HSet(context.Background(), stateKey(auctionID), "status", status)
}

// DeleteAuctionState removes the auction from the cache entirely.
// Called when an auction is finalized, cancelled, or force-closed.
func DeleteAuctionState(auctionID uint) {
	RDB.Del(context.Background(), stateKey(auctionID))
}
