package models

import "time"

type User struct {
	ID uint `gorm:"primaryKey"`

	Username string `gorm:"uniqueIndex;not null"`
	Email    string `gorm:"uniqueIndex;not null"`

	PasswordHash string `gorm:"not null" json:"-"`

	Role string `gorm:"default:USER"`

	IsActive bool `gorm:"default:true"`

	RefreshToken string `gorm:"type:text" json:"-"`

	Rating               float64 `gorm:"default:0"`
	TotalAuctionsWon     uint    `gorm:"default:0"`
	TotalAuctionsCreated uint    `gorm:"default:0"`

	Wallet Wallet `gorm:"foreignKey:UserID"`

	CreatedAt time.Time
	UpdatedAt time.Time
}