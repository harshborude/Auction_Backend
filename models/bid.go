package models

import "time"

type Bid struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null;index"`

	Amount    int64     `gorm:"not null"`

	// Relationships
	Auction Auction `gorm:"foreignKey:AuctionID"`
	// User    User    `gorm:"foreignKey:UserID"`
	// User User `gorm:"foreignKey:UserID" json:"user"`
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`

	AuctionID uint `gorm:"not null;index:idx_auction_bids"`
	CreatedAt time.Time `gorm:"autoCreateTime;index:idx_auction_bids"`
}