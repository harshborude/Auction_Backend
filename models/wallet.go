package models

import "time"

type Wallet struct {
	ID              uint      `gorm:"primaryKey"`
	UserID          uint      `gorm:"uniqueIndex;not null"`
	Balance         int64     `gorm:"not null;default:0"`
	ReservedBalance int64     `gorm:"not null;default:0"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}