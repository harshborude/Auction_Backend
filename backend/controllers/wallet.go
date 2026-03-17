package controllers

import (
	"backend/db"
	"backend/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetWallet fetches the current user's wallet balances
func GetWallet(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user context"})
		return
	}

	var wallet models.Wallet

	// Fetch the wallet for the logged-in user
	if err := db.DB.Where("user_id = ?", userID).First(&wallet).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "wallet not found",
		})
		return
	}

	c.JSON(http.StatusOK, wallet)
}

// GetWalletTransactions fetches the user's credit history with pagination
func GetWalletTransactions(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user context"})
		return
	}

	// Pagination setup
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	var transactions []models.CreditTransaction

	// Fetch transactions ordered by newest first
	if err := db.DB.
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch transaction history",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"page":         page,
		"limit":        limit,
		"transactions": transactions,
	})
}