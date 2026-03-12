package controllers

import (
	"backend/db"
	"backend/models"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetAuctions(c *gin.Context) {

	// Optional pagination parameters (safe defaults)
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 20
	}

	// Prevent large queries
	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	var auctions []models.Auction

	if err := db.DB.
		Where("status = ?", "ACTIVE").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&auctions).Error; err != nil {

		log.Printf("error occurred during fetching auctions: %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch auctions",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"page":     page,
		"limit":    limit,
		"auctions": auctions,
	})
}

func GetAuctionByID(c *gin.Context) {

	id := c.Param("id")

	var auction models.Auction

	if err := db.DB.First(&auction, id).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "auction not found",
			})
			return
		}

		log.Printf("error occurred during fetching auction: %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch auction",
		})
		return
	}

	c.JSON(http.StatusOK, auction)
}