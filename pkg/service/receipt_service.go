package service

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"receipt_processor/pkg/repository"

	"github.com/cespare/xxhash/v2"
	"github.com/google/uuid"
)

// ReceiptDTO represents the structure of a receipt as received from the API.
type ReceiptDTO struct {
	Retailer     string    `json:"retailer"`
	PurchaseDate string    `json:"purchaseDate"` // Format: YYYY-MM-DD
	PurchaseTime string    `json:"purchaseTime"` // Format: HH:MM (24-hour)
	Total        string    `json:"total"`        // E.g. "35.35"
	Items        []ItemDTO `json:"items"`
}

// ItemDTO represents an individual item within a receipt.
type ItemDTO struct {
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"` // E.g. "12.25"
}

// IReceiptService defines the methods available in the service layer.
type IReceiptService interface {
	// ProcessReceipt validates and processes a receipt.
	// It generates a receipt ID, calculates the points, and saves the receipt.
	ProcessReceipt(ctx context.Context, receipt ReceiptDTO) (string, error)
	// GetPoints retrieves the points awarded for a given receipt ID.
	GetPoints(ctx context.Context, receiptID string) (int, error)
}

// receiptService is the concrete implementation of IReceiptService.
type receiptService struct {
	receiptRepo repository.IReceiptRepository
}

// NewReceiptService creates a new instance of the receipt service.
func NewReceiptService(receiptRepo repository.IReceiptRepository) IReceiptService {
	return &receiptService{
		receiptRepo: receiptRepo,
	}
}

// ProcessReceipt handles receipt processing: it calculates points, checks for duplicates,
// saves the receipt (if not a duplicate), and returns the generated or existing receipt ID.
func (s *receiptService) ProcessReceipt(ctx context.Context, receipt ReceiptDTO) (string, error) {
	// Compute a hash for the receipt to detect duplicates.
	hash := computeReceiptHash(receipt)

	// Check for a duplicate receipt using the hash.
	existing, err := s.receiptRepo.FindByHash(ctx, hash)
	if err == nil {
		// Duplicate found: return the existing receipt's ID.
		return existing.ID, nil
	}

	// Generate a new unique receipt ID.
	receiptID := uuid.New().String()

	// Calculate points based on the receipt's data using the defined rules.
	points, err := calculatePoints(receipt)
	if err != nil {
		return "", err
	}

	// Convert the ReceiptDTO to the repository's model, including the computed hash.
	model := repository.ReceiptModel{
		ID:           receiptID,
		Retailer:     receipt.Retailer,
		PurchaseDate: receipt.PurchaseDate,
		PurchaseTime: receipt.PurchaseTime,
		Total:        receipt.Total,
		Hash:         hash,
		Items:        convertItems(receipt.Items),
		Points:       points,
	}

	// Save the receipt using the repository.
	if err := s.receiptRepo.Save(ctx, model); err != nil {
		return "", err
	}

	return receiptID, nil
}

// GetPoints retrieves the points associated with a receipt by its ID.
func (s *receiptService) GetPoints(ctx context.Context, receiptID string) (int, error) {
	model, err := s.receiptRepo.GetByID(ctx, receiptID)
	if err != nil {
		return 0, err
	}
	return model.Points, nil
}

// computeReceiptHash computes a hash for the receipt based on its content.
// It concatenates key fields and uses xxhash to generate a hash string.
func computeReceiptHash(receipt ReceiptDTO) string {
	var sb strings.Builder
	sb.WriteString(receipt.Retailer)
	sb.WriteString(receipt.PurchaseDate)
	sb.WriteString(receipt.PurchaseTime)
	sb.WriteString(receipt.Total)
	for _, item := range receipt.Items {
		// Use the trimmed description.
		sb.WriteString(strings.TrimSpace(item.ShortDescription))
		sb.WriteString(item.Price)
	}
	// Compute the hash using xxhash.
	hashValue := xxhash.Sum64String(sb.String())
	return fmt.Sprintf("%x", hashValue)
}

// calculatePoints computes the total points for a receipt based on the following rules:
//  1. One point for every alphanumeric character in the retailer name.
//  2. 50 points if the total is a round dollar amount with no cents.
//  3. 25 points if the total is a multiple of 0.25.
//  4. 5 points for every two items on the receipt.
//  5. For each item, if the trimmed description length is a multiple of 3,
//     add ceil(price * 0.2) to the points.
//  6. 5 points if the total is greater than 10.00.
//  7. 6 points if the day in the purchase date is odd.
//  8. 10 points if the time of purchase is after 2:00pm and before 4:00pm.
func calculatePoints(receipt ReceiptDTO) (int, error) {
	var points int

	// Rule 1: One point for every alphanumeric character in the retailer name.
	for _, ch := range receipt.Retailer {
		if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') {
			points++
		}
	}

	// Parse the total amount.
	total, err := strconv.ParseFloat(receipt.Total, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid total amount: %v", err)
	}

	// Rule 2: 50 points if the total is a round dollar amount (no cents).
	if total == float64(int(total)) {
		points += 50
	}

	// Rule 3: 25 points if the total is a multiple of 0.25.
	if math.Mod(total, 0.25) == 0 {
		points += 25
	}

	// Rule 4: 5 points for every two items on the receipt.
	itemCount := len(receipt.Items)
	points += (itemCount / 2) * 5

	// Rule 5: For each item, if the trimmed description length is a multiple of 3,
	// add ceil(price * 0.2) to the points.
	for _, item := range receipt.Items {
		trimmed := strings.TrimSpace(item.ShortDescription)
		if len(trimmed)%3 == 0 {
			price, err := strconv.ParseFloat(item.Price, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid item price: %v", err)
			}
			extra := math.Ceil(price * 0.2)
			points += int(extra)
		}
	}

	// Rule 6: 5 points if the total is greater than 10.00.
	if total > 10.00 {
		points += 5
	}

	// Rule 7: 6 points if the day in the purchase date is odd.
	parts := strings.Split(receipt.PurchaseDate, "-")
	if len(parts) == 3 {
		day, err := strconv.Atoi(parts[2])
		if err == nil && day%2 == 1 {
			points += 6
		}
	}

	// Rule 8: 10 points if the time of purchase is after 2:00pm and before 4:00pm.
	timeParts := strings.Split(receipt.PurchaseTime, ":")
	if len(timeParts) == 2 {
		hour, err := strconv.Atoi(timeParts[0])
		if err == nil && hour >= 14 && hour < 16 {
			points += 10
		}
	}

	return points, nil
}

// convertItems transforms a slice of ItemDTO into a slice of repository.ItemModel.
func convertItems(items []ItemDTO) []repository.ItemModel {
	var models []repository.ItemModel
	for _, item := range items {
		models = append(models, repository.ItemModel{
			ShortDescription: item.ShortDescription,
			Price:            item.Price,
		})
	}
	return models
}
