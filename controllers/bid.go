package controllers

import (
	"backend/db"
	"backend/models"
	"backend/services"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PlaceBidRequest struct {
	Amount int64 `json:"amount" binding:"required"`
}

func PlaceBid(c *gin.Context) {

	var req PlaceBidRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	// Improvement 2: Early validation
	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "bid amount must be greater than zero",
		})
		return
	}

	auctionIDParam := c.Param("id")
	auctionIDUint64, err := strconv.ParseUint(auctionIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid auction id",
		})
		return
	}

	auctionID := uint(auctionIDUint64)

	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	// Improvement 1: Safe type assertion
	userID, ok := userIDValue.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid user context",
		})
		return
	}

	bid, err := services.PlaceBid(db.DB, auctionID, userID, req.Amount)
	if err != nil {
		// Improvement 3: Basic dynamic HTTP status mapping
		status := http.StatusBadRequest
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		}

		c.JSON(status, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Used 201 Created since a new bid record was inserted
	c.JSON(http.StatusCreated, gin.H{
		"message": "bid placed successfully",
		"bid":     bid,
	})
}

// GetBidHistory fetches all bids for a specific auction
func GetBidHistory(c *gin.Context) {

	auctionIDParam := c.Param("id")

	auctionIDUint64, err := strconv.ParseUint(auctionIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid auction id",
		})
		return
	}

	auctionID := uint(auctionIDUint64)

	var bids []models.Bid

	if err := db.DB.
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, username")
		}).
		Where("auction_id = ?", auctionID).
		Order("created_at desc").
		Find(&bids).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch bid history",
		})
		return
	}

	c.JSON(http.StatusOK, bids)
}
