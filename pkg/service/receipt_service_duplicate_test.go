package service

import (
	"context"
	"testing"

	"receipt_processor/pkg/database"
	"receipt_processor/pkg/repository"
)

func TestProcessReceiptDuplicatePrevention(t *testing.T) {
	// Set up an in-memory SQLite database.
	db, err := database.New("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("failed to create in-memory db: %v", err)
	}
	repo := repository.NewReceiptRepository(db)
	svc := NewReceiptService(repo)
	ctx := context.Background()

	// Define a base receipt.
	baseReceipt := ReceiptDTO{
		Retailer:     "Target",
		PurchaseDate: "2022-01-01",
		PurchaseTime: "13:01",
		Total:        "35.35",
		Items: []ItemDTO{
			{ShortDescription: "Mountain Dew 12PK", Price: "6.49"},
			{ShortDescription: "Emils Cheese Pizza", Price: "12.25"},
		},
	}

	// Table-driven test cases.
	testCases := []struct {
		name       string
		receipt1   ReceiptDTO
		receipt2   ReceiptDTO
		expectSame bool // true if we expect duplicate receipts to return the same id.
	}{
		{
			name:       "Identical receipts yield same id",
			receipt1:   baseReceipt,
			receipt2:   baseReceipt,
			expectSame: true,
		},
		{
			name:     "Different receipts yield different id",
			receipt1: baseReceipt,
			receipt2: func() ReceiptDTO {
				// Modify the receipt slightly to get a different hash.
				r := baseReceipt
				r.Retailer = "TargetX" // Change retailer name.
				return r
			}(),
			expectSame: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Process the first receipt.
			id1, err := svc.ProcessReceipt(ctx, tc.receipt1)
			if err != nil {
				t.Fatalf("failed to process first receipt: %v", err)
			}
			// Process the second receipt.
			id2, err := svc.ProcessReceipt(ctx, tc.receipt2)
			if err != nil {
				t.Fatalf("failed to process second receipt: %v", err)
			}

			if tc.expectSame && id1 != id2 {
				t.Errorf("expected same id for duplicate receipts, got %s and %s", id1, id2)
			}
			if !tc.expectSame && id1 == id2 {
				t.Errorf("expected different ids for distinct receipts, but both returned %s", id1)
			}
		})
	}
}
