package service

import (
	"testing"
)

func TestCalculatePoints(t *testing.T) {
	testCases := []struct {
		name           string
		receipt        ReceiptDTO
		expectedPoints int
		expectError    bool
	}{
		{
			name: "Example 1",
			receipt: ReceiptDTO{
				Retailer:     "Target",
				PurchaseDate: "2022-01-01",
				PurchaseTime: "13:01",
				Total:        "35.35",
				Items: []ItemDTO{
					{ShortDescription: "Mountain Dew 12PK", Price: "6.49"},
					{ShortDescription: "Emils Cheese Pizza", Price: "12.25"},
					{ShortDescription: "Knorr Creamy Chicken", Price: "1.26"},
					{ShortDescription: "Doritos Nacho Cheese", Price: "3.35"},
					{ShortDescription: "   Klarbrunn 12-PK 12 FL OZ  ", Price: "12.00"},
				},
			},
			// Calculation breakdown (including the LLM rule):
			// Rule 1: "Target" -> 6 points.
			// Rule 4: 5 items => floor(5/2) * 5 = 2*5 = 10 points.
			// Rule 5: "Emils Cheese Pizza" qualifies -> ceil(12.25*0.2)=3 points,
			//         "Klarbrunn 12-PK 12 FL OZ" qualifies -> ceil(12.00*0.2)=3 points;
			//         Total from items = 3+3 = 6 points.
			// Rule 6: Total (35.35) > 10.00 -> +5 points.
			// Rule 7: Purchase day "01" (odd) -> +6 points.
			// Total = 6 + 10 + 6 + 5 + 6 = 33 points.
			expectedPoints: 33,
			expectError:    false,
		},
		{
			name: "Example 2",
			receipt: ReceiptDTO{
				Retailer:     "M&M Corner Market",
				PurchaseDate: "2022-03-20",
				PurchaseTime: "14:33",
				Total:        "9.00",
				Items: []ItemDTO{
					{ShortDescription: "Gatorade", Price: "2.25"},
					{ShortDescription: "Gatorade", Price: "2.25"},
					{ShortDescription: "Gatorade", Price: "2.25"},
					{ShortDescription: "Gatorade", Price: "2.25"},
				},
			},
			// Calculation breakdown:
			// Rule 1: "M&M Corner Market" -> 14 alphanumeric characters.
			// Rule 2: Total is round (9.00) -> +50 points.
			// Rule 3: 9.00 is a multiple of 0.25 -> +25 points.
			// Rule 4: 4 items => floor(4/2) * 5 = 2*5 = 10 points.
			// Rule 5: "Gatorade" length = 8 (not a multiple of 3) -> 0 points.
			// Rule 6: Total is not >10.00 -> 0 points.
			// Rule 7: Purchase day "20" (even) -> 0 points.
			// Rule 8: Purchase time "14:33" is between 14:00 and 16:00 -> +10 points.
			// Total = 14 + 50 + 25 + 10 + 10 = 109 points.
			expectedPoints: 109,
			expectError:    false,
		},
		{
			name: "Invalid Total",
			receipt: ReceiptDTO{
				Retailer:     "Invalid",
				PurchaseDate: "2022-01-01",
				PurchaseTime: "13:01",
				Total:        "abc", // Invalid total
				Items: []ItemDTO{
					{ShortDescription: "Item", Price: "1.00"},
				},
			},
			expectedPoints: 0,
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			points, err := calculatePoints(tc.receipt)
			if tc.expectError {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("did not expect error but got: %v", err)
				}
				if points != tc.expectedPoints {
					t.Errorf("expected %d points, got %d", tc.expectedPoints, points)
				}
			}
		})
	}
}
