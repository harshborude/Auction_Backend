package models

import "time"

type Auction struct {
	ID uint `gorm:"primaryKey"`

	Title       string `gorm:"not null"`
	Description string `gorm:"type:text"`
	ImageURL    string

	StartingPrice int64 `gorm:"not null;check:starting_price >= 0"`
	BidIncrement  int64 `gorm:"not null;check:bid_increment > 0"`

	CurrentHighestBid      int64
	CurrentHighestBidderID *uint `gorm:"index"`
	BidCount               int64 `gorm:"default:0"`
	ExtensionCount         int   `gorm:"default:0"`

	Status string `gorm:"type:varchar(20);default:'ACTIVE';index:idx_status_endtime;index:idx_status_starttime"`

	StartTime time.Time `gorm:"index:idx_status_starttime"`
	EndTime   time.Time `gorm:"index:idx_status_endtime"`

	CreatedBy uint `gorm:"index"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
