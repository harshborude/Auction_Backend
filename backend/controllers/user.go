package controllers

import (
	"backend/db"
	"backend/models"
	"backend/utils"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"strings"
	"strconv"
)

type RegisterInput struct {
	Username string `json:"username" validate:"required,min=3,max=20,alphanum"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type ChangePasswordInput struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=72"`
}

type RefreshTokenInput struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

func RegisterUser(c *gin.Context) {

	var input RegisterInput


	// Step 1: Parse request body
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Printf("error occurred during request parsing: %v", err)

		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request format",
		})
		return
	}

	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	// Step 2: Validate input
	if err := utils.Validate.Struct(&input); err != nil {

	log.Printf("validation error: %v", err)

	c.JSON(http.StatusBadRequest, gin.H{
		"errors": utils.FormatValidationErrors(err),
	})

	return
}

	// Step 3: Check if user already exists
	var existingUser models.User
	err := db.DB.
		Where("email = ? OR username = ?", input.Email, input.Username).
		First(&existingUser).Error

	if err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": "username or email already exists",
		})
		return
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("error occurred during user lookup: %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "error occurred during user lookup",
		})
		return
	}

	// Step 4: Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("error occurred during password hashing: %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "error occurred during password hashing",
		})
		return
	}

	// Step 5: Create user and wallet in a single transaction
	tx := db.DB.Begin()

	user := models.User{
		Username:     input.Username,
		Email:        input.Email,
		PasswordHash: string(hashedPassword),
	}

	// Use tx.Create instead of db.DB.Create
	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback() // Abort transaction on error
		log.Printf("error occurred during user creation in database: %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "error occurred during user creation",
		})
		return
	}

	// Create user wallet using the new user's ID
	wallet := models.Wallet{
		UserID: user.ID,
	}

	// Use tx.Create instead of db.DB.Create
	if err := tx.Create(&wallet).Error; err != nil {
		tx.Rollback() // Abort transaction and undo user creation on error
		log.Printf("error occurred during wallet creation: %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "error occurred during wallet creation",
		})
		return
	}

	// Commit the transaction to save both user and wallet permanently
	tx.Commit()

	c.JSON(http.StatusCreated, gin.H{
		"message": "user created successfully",
	})
	
}

func LoginUser(c *gin.Context) {

	var input LoginInput


	// Step 1: Parse request body
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Printf("error occurred during login request parsing: %v", err)

		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request format",
		})
		return
	}
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	// Step 2: Validate input
	if err := utils.Validate.Struct(&input); err != nil {

	log.Printf("validation error: %v", err)

	c.JSON(http.StatusBadRequest, gin.H{
		"errors": utils.FormatValidationErrors(err),
	})

	return
}

	var user models.User

	// Step 3: Find user
	if err := db.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid email or password",
			})
			return
		}

		log.Printf("error occurred during fetching user for login: %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "error occurred during login",
		})
		return
	}

	// Step 4: Compare password
	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(input.Password),
	); err != nil {

		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid email or password",
		})
		return
	}

	// Step 5: Generate JWT
	accessToken, err := utils.GenerateAccessToken(user.ID, user.Role)
if err != nil {
    log.Printf("error occurred during access token generation: %v", err)

    c.JSON(http.StatusInternalServerError, gin.H{
        "error": "error occurred during token generation",
    })
    return
}

refreshToken, err := utils.GenerateRefreshToken(user.ID)
if err != nil {
    log.Printf("error occurred during refresh token generation: %v", err)

    c.JSON(http.StatusInternalServerError, gin.H{
        "error": "error occurred during token generation",
    })
    return
}

user.RefreshToken = refreshToken

if err := db.DB.Save(&user).Error; err != nil {
    log.Printf("error occurred during refresh token storage: %v", err)

    c.JSON(http.StatusInternalServerError, gin.H{
        "error": "error occurred during login",
    })
    return
}

c.JSON(http.StatusOK, gin.H{
    "access_token":  accessToken,
    "refresh_token": refreshToken,
})
}


func GetCurrentUser(c *gin.Context) {

	userID, exists := c.Get("user_id")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	var user models.User

	if err := db.DB.Preload("Wallet").First(&user, userID).Error; err != nil {

		log.Printf("error occurred during fetching current user: %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "error occurred during fetching user",
		})
		return
	}

	c.JSON(http.StatusOK, user)
}


func ChangePassword(c *gin.Context) {

	userID, exists := c.Get("user_id")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	var input ChangePasswordInput

	if err := c.ShouldBindJSON(&input); err != nil {
		log.Printf("error occurred during password change parsing: %v", err)

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

	var user models.User

	if err := db.DB.First(&user, userID).Error; err != nil {

		log.Printf("error occurred during user lookup: %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "error occurred during user lookup",
		})
		return
	}

	// verify old password
	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(input.OldPassword),
	); err != nil {

		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "old password incorrect",
		})
		return
	}

	// hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(input.NewPassword),
		bcrypt.DefaultCost,
	)

	if err != nil {

		log.Printf("error occurred during password hashing: %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "error occurred during password hashing",
		})
		return
	}

	user.PasswordHash = string(hashedPassword)

	if err := db.DB.Save(&user).Error; err != nil {

		log.Printf("error occurred during password update: %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "error occurred during password update",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "password updated successfully",
	})
}

func RefreshAccessToken(c *gin.Context) {

	var input RefreshTokenInput

	// parse request
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Printf("error occurred during refresh request parsing: %v", err)

		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request format",
		})
		return
	}

	// validate input
	if err := utils.Validate.Struct(&input); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"errors": utils.FormatValidationErrors(err),
		})

		return
	}

	// validate refresh token
	claims, err := utils.ValidateRefreshToken(input.RefreshToken)
	if err != nil {

		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid refresh token",
		})

		return
	}

	userID64, err := strconv.ParseUint(claims.Subject, 10, 64)
if err != nil {
	c.JSON(http.StatusUnauthorized, gin.H{
		"error": "invalid refresh token",
	})
	return
}

userID := uint(userID64)

	var user models.User

	// find user
	if err := db.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid refresh token",
		})
		return
	}

	// check account status
	if !user.IsActive {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "account disabled",
		})
		return
	}

	// verify stored refresh token
	if user.RefreshToken != input.RefreshToken {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid refresh token",
		})
		return
	}
	// generate new access token
	accessToken, err := utils.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		log.Printf("error occurred during access token generation: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "error occurred during token generation",
		})
		return
	}

	// rotate refresh token — invalidate the old one immediately
	newRefreshToken, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		log.Printf("error occurred during refresh token rotation: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "error occurred during token generation",
		})
		return
	}

	user.RefreshToken = newRefreshToken
	if err := db.DB.Save(&user).Error; err != nil {
		log.Printf("error occurred during refresh token storage: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "error occurred during token rotation",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": newRefreshToken,
	})
}

func LogoutUser(c *gin.Context) {

	userIDRaw, exists := c.Get("user_id")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	userID := userIDRaw.(uint)

	var user models.User

	if err := db.DB.First(&user, userID).Error; err != nil {

		log.Printf("error occurred during logout user lookup: %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "error occurred during logout",
		})

		return
	}

	// revoke refresh token
	user.RefreshToken = ""

	if err := db.DB.Save(&user).Error; err != nil {

		log.Printf("error occurred during logout token revocation: %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "error occurred during logout",
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "logged out successfully",
	})
}