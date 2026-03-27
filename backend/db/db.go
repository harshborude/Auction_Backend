package db

import (
	"backend/models"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"golang.org/x/crypto/bcrypt"

	"fmt"
	"os"
)

var DB *gorm.DB

func seedAdmin() {

	var admin models.User

	result := DB.Where("email = ?", "admin@auction.com").First(&admin)

	if result.Error == gorm.ErrRecordNotFound {

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("error occurred during admin password hashing: %v", err)
			return
		}

		adminUser := models.User{
			Username:     "admin",
			Email:        "admin@auction.com",
			PasswordHash: string(hashedPassword),
			Role:         "ADMIN",
		}

		if err := DB.Create(&adminUser).Error; err != nil {
			log.Printf("error occurred during admin creation: %v", err)
			return
		}

		wallet := models.Wallet{
			UserID: adminUser.ID,
		}

		if err := DB.Create(&wallet).Error; err != nil {
			log.Printf("error occurred during wallet creation: %v", err)
			return
		}

		log.Println("Default admin created")
	}
}

func ConnectDatabase() {

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	sslmode := os.Getenv("DB_SSLMODE")
	if sslmode == "" {
		sslmode = "require"
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		host, user, password, dbname, port, sslmode,
	)

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		log.Fatal("Failed to connect to database")
	}

	DB = database

	DB.AutoMigrate(
		&models.User{},
		&models.Wallet{},
		&models.CreditTransaction{},
		&models.Auction{},
		&models.Bid{},
	)

	seedAdmin()

	log.Println("Database connected successfully")
}
