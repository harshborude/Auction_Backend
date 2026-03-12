package models

import "time"

type CreditTransaction struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null;index"`
	Amount    int64     `gorm:"not null"`
	Type      string    `gorm:"not null"`
	Reference string
	CreatedAt time.Time
}