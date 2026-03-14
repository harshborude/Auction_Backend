package routes

import (
	"backend/controllers"
	// "backend/middleware"

	"github.com/gin-gonic/gin"
)

func AuctionRoutes(router *gin.Engine) {

    auctions := router.Group("/auctions")

    {
        auctions.GET("", controllers.GetAuctions)
        auctions.GET("/:id", controllers.GetAuctionByID)
        auctions.GET("/:id/bids", controllers.GetBidHistory)
    }
}

