package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"receipt_processor/pkg/service"

	"github.com/rs/zerolog/log"
)

// ProcessReceiptHandler handles POST /receipts/process.
// It reads and validates the incoming JSON, delegates processing to the service layer,
// and returns a JSON response with the generated receipt ID.
func (r *Router) ProcessReceiptHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the request body.
	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Ctx(req.Context()).Error().Err(err).Msg("Failed to read request body")
		http.Error(w, "The receipt is invalid.", http.StatusBadRequest)
		return
	}

	// Unmarshal the JSON into a ReceiptDTO.
	var receipt service.ReceiptDTO
	if err := json.Unmarshal(body, &receipt); err != nil {
		log.Ctx(req.Context()).Error().Err(err).Msg("Invalid JSON in request body")
		http.Error(w, "Invalid JSON in request body", http.StatusBadRequest)
		return
	}

	// Validate the receipt fields using regex rules.
	if err := validateReceipt(receipt); err != nil {
		log.Ctx(req.Context()).Error().Err(err).Msg("Validation failed")
		http.Error(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Delegate to the service layer to process the receipt.
	id, err := r.receiptService.ProcessReceipt(req.Context(), receipt)
	if err != nil {
		log.Ctx(req.Context()).Error().Err(err).Msg("Failed to process receipt")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Return the generated receipt ID in JSON format.
	response := map[string]string{"id": id}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Ctx(req.Context()).Error().Err(err).Msg("Failed to write response")
	}
}

// GetPointsHandler handles GET /receipts/{id}/points.
// It extracts the receipt ID from the URL, validates it, and returns the points awarded.
func (r *Router) GetPointsHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Expecting URL format: /receipts/{id}/points.
	parts := strings.Split(strings.Trim(req.URL.Path, "/"), "/")
	if len(parts) != 3 || parts[0] != "receipts" || parts[2] != "points" {
		http.NotFound(w, req)
		return
	}
	receiptID := parts[1]

	// Validate receiptID using the regex pattern: "^\S+$".
	idRegex := regexp.MustCompile(`^\S+$`)
	if !idRegex.MatchString(receiptID) {
		http.Error(w, "No receipt found for that ID.", http.StatusBadRequest)
		return
	}

	// Retrieve the points via the service layer.
	points, err := r.receiptService.GetPoints(req.Context(), receiptID)
	if err != nil {
		log.Ctx(req.Context()).Error().Err(err).Msg("Failed to get points")
		http.Error(w, "No receipt found for that ID.", http.StatusNotFound)
		return
	}

	// Return the points in JSON format.
	response := map[string]int{"points": points}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Ctx(req.Context()).Error().Err(err).Msg("Failed to write response")
	}
}

// validateReceipt checks the receipt fields against the regex patterns from the OpenAPI spec.
func validateReceipt(receipt service.ReceiptDTO) error {
	// Validate "retailer": pattern "^[\w\s\-\&]+$".
	retailerRegex := regexp.MustCompile(`^[\w\s\-\&]+$`)
	if !retailerRegex.MatchString(receipt.Retailer) {
		return fmt.Errorf("invalid retailer format")
	}
	// Validate "purchaseDate": basic pattern for YYYY-MM-DD.
	dateRegex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	if !dateRegex.MatchString(receipt.PurchaseDate) {
		return fmt.Errorf("invalid purchaseDate format")
	}
	// Validate "purchaseTime": expecting HH:MM in 24-hour format.
	timeRegex := regexp.MustCompile(`^\d{2}:\d{2}$`)
	if !timeRegex.MatchString(receipt.PurchaseTime) {
		return fmt.Errorf("invalid purchaseTime format")
	}
	// Validate "total": pattern "^\d+\.\d{2}$".
	totalRegex := regexp.MustCompile(`^\d+\.\d{2}$`)
	if !totalRegex.MatchString(receipt.Total) {
		return fmt.Errorf("invalid total format")
	}
	// Ensure there is at least one item.
	if len(receipt.Items) == 0 {
		return fmt.Errorf("at least one item is required")
	}
	// Validate each item.
	itemDescRegex := regexp.MustCompile(`^[\w\s\-]+$`)
	priceRegex := regexp.MustCompile(`^\d+\.\d{2}$`)
	for _, item := range receipt.Items {
		if !itemDescRegex.MatchString(item.ShortDescription) {
			return fmt.Errorf("invalid item shortDescription format")
		}
		if !priceRegex.MatchString(item.Price) {
			return fmt.Errorf("invalid item price format")
		}
	}
	return nil
}
