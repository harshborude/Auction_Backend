package controllers

import (
	"backend/db"
	"backend/models"
	"backend/services"
	"backend/utils"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type CreateAuctionInput struct {
	Title         string `json:"title" validate:"required,min=3"`
	Description   string `json:"description"`
	ImageURL      string `json:"image_url"`
	StartingPrice int64  `json:"starting_price" validate:"required,gt=0"`
	BidIncrement  int64  `json:"bid_increment" validate:"required,gt=0"`
	StartTime     string `json:"start_time" validate:"required"`
	EndTime       string `json:"end_time" validate:"required"`
}

func PromoteUser(c *gin.Context) {
	// Redundant role check removed; handled by middleware

	userID := c.Param("user_id")

	var user models.User

	if err := db.DB.First(&user, userID).Error; err != nil {
		log.Printf("error occurred during fetching user: %v", err)

		c.JSON(http.StatusNotFound, gin.H{
			"error": "user not found",
		})
		return
	}

	user.Role = "ADMIN"

	if err := db.DB.Save(&user).Error; err != nil {
		log.Printf("error occurred during promoting user: %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to promote user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "user promoted to admin",
	})
}

func GetUsers(c *gin.Context) {
	// Redundant role check removed; handled by middleware

	var users []models.User

	if err := db.DB.Preload("Wallet").Find(&users).Error; err != nil {
		log.Printf("error occurred during fetching users: %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch users",
		})
		return
	}

	c.JSON(http.StatusOK, users)
}

func AssignCredits(c *gin.Context) {
	// Parse string param to uint safely
	userIDParam := c.Param("user_id")
	userIDUint64, err := strconv.ParseUint(userIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user id",
		})
		return
	}
	userID := uint(userIDUint64)

	var input struct {
		Amount int64 `json:"amount" validate:"required,gt=0"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request format",
		})
		return
	}

	if err := utils.Validate.Struct(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": utils.FormatValidationErrors(err),
		})
		return
	}

	tx := db.DB.Begin()

	var wallet models.Wallet

	if err := tx.Where("user_id = ?", userID).First(&wallet).Error; err != nil {
		tx.Rollback()

		c.JSON(http.StatusNotFound, gin.H{
			"error": "wallet not found",
		})
		return
	}

	// Used Balance because that matches the models.Wallet schema we designed
	wallet.Balance += input.Amount

	if err := tx.Save(&wallet).Error; err != nil {
		tx.Rollback()

		log.Printf("error updating wallet balance: %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to update wallet balance",
		})
		return
	}

	transaction := models.CreditTransaction{
		UserID: wallet.UserID,
		Amount: input.Amount,
		Type:   "ADMIN_ASSIGN",
	}

	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()

		log.Printf("error recording credit transaction: %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to record transaction",
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("transaction commit failed: %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "transaction failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "credits assigned successfully",
	})
}

func CreateAuction(c *gin.Context) {
	adminIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	adminID := adminIDValue.(uint)

	var input CreateAuctionInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request format",
		})
		return
	}

	if err := utils.Validate.Struct(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": utils.FormatValidationErrors(err),
		})
		return
	}

	startTime, err := time.Parse(time.RFC3339, input.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid start_time format (RFC3339 required)",
		})
		return
	}

	endTime, err := time.Parse(time.RFC3339, input.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid end_time format (RFC3339 required)",
		})
		return
	}

	if endTime.Sub(startTime) < time.Minute {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "auction must run for at least 1 minute",
		})
		return
	}

	if input.BidIncrement > input.StartingPrice {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "bid_increment cannot be greater than starting_price",
		})
		return
	}

	// Dynamic Status Logic
	status := "SCHEDULED"
	// Including a tiny buffer to allow "immediate" starts without failing
	if startTime.Before(time.Now().Add(5 * time.Second)) {
		status = "ACTIVE"
	}

	auction := models.Auction{
		Title:             input.Title,
		Description:       input.Description,
		ImageURL:          input.ImageURL,
		StartingPrice:     input.StartingPrice,
		BidIncrement:      input.BidIncrement,
		CurrentHighestBid: input.StartingPrice,
		BidCount:          0,
		Status:            status,
		StartTime:         startTime,
		EndTime:           endTime,
		CreatedBy:         adminID,
	}

	if err := db.DB.Create(&auction).Error; err != nil {
		log.Printf("error occurred during auction creation: %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create auction",
		})
		return
	}

	c.JSON(http.StatusCreated, auction)
}

// GetAdminAuctions returns all auctions regardless of status (admin only)
func GetAdminAuctions(c *gin.Context) {
	var auctions []models.Auction
	if err := db.DB.Order("created_at DESC").Find(&auctions).Error; err != nil {
		log.Printf("error fetching admin auctions: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch auctions"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"auctions": auctions})
}

// BanUser sets a user's IsActive to false, preventing login
func BanUser(c *gin.Context) {
	userIDParam := c.Param("user_id")
	var user models.User
	if err := db.DB.First(&user, userIDParam).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	user.IsActive = false
	if err := db.DB.Save(&user).Error; err != nil {
		log.Printf("error banning user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to ban user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user banned successfully"})
}

// ForceCloseAuction manually ends an auction and settles the credits
func ForceCloseAuction(c *gin.Context) {
	auctionIDParam := c.Param("id")

	auctionIDUint64, err := strconv.ParseUint(auctionIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid auction id"})
		return
	}

	auctionID := uint(auctionIDUint64)

	if err := services.AdminForceCloseAuction(db.DB, auctionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "auction force-closed successfully",
	})
}

// CancelAuction manually cancels an auction and refunds users
func CancelAuction(c *gin.Context) {
	auctionIDParam := c.Param("id")

	auctionIDUint64, err := strconv.ParseUint(auctionIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid auction id"})
		return
	}

	auctionID := uint(auctionIDUint64)

	if err := services.CancelAuction(db.DB, auctionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "auction cancelled and credits refunded",
	})
}
