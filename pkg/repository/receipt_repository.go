package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

// ReceiptModel represents the receipt stored in the database.
type ReceiptModel struct {
	ID           string `gorm:"primaryKey;type:varchar(36)"`
	Retailer     string
	PurchaseDate string
	PurchaseTime string
	Total        string
	Points       int
	Hash         string      `gorm:"uniqueIndex;not null"`
	Items        []ItemModel `gorm:"foreignKey:ReceiptID"`
}

// ItemModel represents an individual item within a receipt.
type ItemModel struct {
	ID               uint   `gorm:"primaryKey;autoIncrement"`
	ReceiptID        string `gorm:"index;type:varchar(36)"`
	ShortDescription string
	Price            string
}

// IReceiptRepository defines the interface for interacting with receipt persistence.
type IReceiptRepository interface {
	Save(ctx context.Context, receipt ReceiptModel) error
	GetByID(ctx context.Context, id string) (ReceiptModel, error)
	FindByHash(ctx context.Context, hash string) (ReceiptModel, error)
}

// receiptRepository is a concrete implementation of IReceiptRepository using GORM.
type receiptRepository struct {
	db *gorm.DB
}

// NewReceiptRepository creates a new instance of the receipt repository.
// It performs auto-migration to ensure the schema is up to date.
func NewReceiptRepository(db *gorm.DB) IReceiptRepository {
	// AutoMigrate ReceiptModel and ItemModel.
	db.AutoMigrate(&ReceiptModel{}, &ItemModel{})
	return &receiptRepository{
		db: db,
	}
}

// Save stores a receipt and its items in the database.
func (r *receiptRepository) Save(ctx context.Context, receipt ReceiptModel) error {
	result := r.db.WithContext(ctx).Create(&receipt)
	return result.Error
}

// GetByID retrieves a receipt by its ID, preloading associated items.
func (r *receiptRepository) GetByID(ctx context.Context, id string) (ReceiptModel, error) {
	var receipt ReceiptModel
	result := r.db.WithContext(ctx).Preload("Items").First(&receipt, "id = ?", id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return receipt, result.Error
	}
	return receipt, result.Error
}

// FindByHash retrieves a receipt by its computed hash. If not found, returns an error.
func (r *receiptRepository) FindByHash(ctx context.Context, hash string) (ReceiptModel, error) {
	var receipt ReceiptModel
	result := r.db.WithContext(ctx).Where("hash = ?", hash).Preload("Items").First(&receipt)
	return receipt, result.Error
}
