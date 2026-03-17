package services

import (
	"backend/models"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func ReserveCredits(tx *gorm.DB, userID uint, amount int64) (*models.Wallet, error) {

	if amount <= 0 {
		return nil, errors.New("invalid amount")
	}

	var wallet models.Wallet

	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ?", userID).
		First(&wallet).Error; err != nil {
		return nil, err
	}

	available := wallet.Balance - wallet.ReservedBalance

	if available < amount {
		return nil, errors.New("insufficient credits")
	}

	wallet.ReservedBalance += amount

	if err := tx.Model(&wallet).
		Update("reserved_balance", wallet.ReservedBalance).Error; err != nil {
		return nil, err
	}

	return &wallet, nil
}

func ReleaseCredits(tx *gorm.DB, userID uint, amount int64) error {

	if amount <= 0 {
		return errors.New("invalid amount")
	}

	var wallet models.Wallet

	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ?", userID).
		First(&wallet).Error; err != nil {
		return err
	}

	// Fail fast instead of silently masking a potential bug
	if wallet.ReservedBalance < amount {
		return errors.New("invalid release amount: exceeds reserved balance")
	}

	wallet.ReservedBalance -= amount

	return tx.Save(&wallet).Error
}

func DeductReservedCredits(tx *gorm.DB, userID uint, amount int64) error {

	if amount <= 0 {
		return errors.New("invalid amount")
	}

	var wallet models.Wallet

	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ?", userID).
		First(&wallet).Error; err != nil {
		return err
	}

	if wallet.Balance < amount || wallet.ReservedBalance < amount {
		return errors.New("insufficient balance to deduct")
	}

	wallet.Balance -= amount
	wallet.ReservedBalance -= amount

	return tx.Save(&wallet).Error
}

func AddCredits(tx *gorm.DB, userID uint, amount int64) error {

	// Prevent accidental deduction or pointless zero transactions
	if amount <= 0 {
		return errors.New("invalid credit amount")
	}

	var wallet models.Wallet

	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ?", userID).
		First(&wallet).Error; err != nil {
		return err
	}

	wallet.Balance += amount

	return tx.Save(&wallet).Error
}