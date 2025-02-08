package repository

import (
	"context"
	"testing"

	"receipt_processor/pkg/database"
)

func TestReceiptRepository_SaveAndGet(t *testing.T) {
	// Set up an in-memory SQLite database.
	db, err := database.New("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("failed to create in-memory db: %v", err)
	}

	repo := NewReceiptRepository(db)
	ctx := context.Background()

	// Define table-driven test cases.
	testCases := []struct {
		name    string
		receipt ReceiptModel
	}{
		{
			name: "Valid receipt 1",
			receipt: ReceiptModel{
				ID:           "id1",
				Retailer:     "Target",
				PurchaseDate: "2022-01-01",
				PurchaseTime: "13:01",
				Total:        "35.35",
				Points:       33, // assumed calculated points for test purposes
				Hash:         "hash1",
				Items: []ItemModel{
					{ShortDescription: "Item A", Price: "10.00"},
					{ShortDescription: "Item B", Price: "25.35"},
				},
			},
		},
		{
			name: "Valid receipt 2",
			receipt: ReceiptModel{
				ID:           "id2",
				Retailer:     "M&M Corner Market",
				PurchaseDate: "2022-03-20",
				PurchaseTime: "14:33",
				Total:        "9.00",
				Points:       109,
				Hash:         "hash2",
				Items: []ItemModel{
					{ShortDescription: "Gatorade", Price: "2.25"},
					{ShortDescription: "Gatorade", Price: "2.25"},
					{ShortDescription: "Gatorade", Price: "2.25"},
					{ShortDescription: "Gatorade", Price: "2.25"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Save the receipt.
			if err := repo.Save(ctx, tc.receipt); err != nil {
				t.Fatalf("failed to save receipt: %v", err)
			}

			// Retrieve the receipt by ID.
			saved, err := repo.GetByID(ctx, tc.receipt.ID)
			if err != nil {
				t.Fatalf("failed to get receipt by ID: %v", err)
			}
			if saved.ID != tc.receipt.ID {
				t.Errorf("expected ID %s, got %s", tc.receipt.ID, saved.ID)
			}
			if saved.Hash != tc.receipt.Hash {
				t.Errorf("expected Hash %s, got %s", tc.receipt.Hash, saved.Hash)
			}
			if len(saved.Items) != len(tc.receipt.Items) {
				t.Errorf("expected %d items, got %d", len(tc.receipt.Items), len(saved.Items))
			}

			// Retrieve the receipt by Hash.
			found, err := repo.FindByHash(ctx, tc.receipt.Hash)
			if err != nil {
				t.Fatalf("failed to find receipt by hash: %v", err)
			}
			if found.ID != tc.receipt.ID {
				t.Errorf("expected found ID %s, got %s", tc.receipt.ID, found.ID)
			}
		})
	}
}

func TestReceiptRepository_Duplicate(t *testing.T) {
	// Set up an in-memory SQLite database.
	db, err := database.New("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("failed to create in-memory db: %v", err)
	}

	repo := NewReceiptRepository(db)
	ctx := context.Background()

	// Prepare a receipt.
	receipt := ReceiptModel{
		ID:           "dup1",
		Retailer:     "Target",
		PurchaseDate: "2022-01-01",
		PurchaseTime: "13:01",
		Total:        "35.35",
		Points:       33,
		Hash:         "dup-hash",
		Items: []ItemModel{
			{ShortDescription: "Item A", Price: "10.00"},
		},
	}

	// First insertion should succeed.
	if err := repo.Save(ctx, receipt); err != nil {
		t.Fatalf("failed to save receipt first time: %v", err)
	}

	// Second insertion with the same hash should fail.
	duplicate := receipt
	duplicate.ID = "dup2" // Different ID but same hash.
	err = repo.Save(ctx, duplicate)
	if err == nil {
		t.Fatalf("expected error when saving duplicate receipt, got nil")
	}
}
