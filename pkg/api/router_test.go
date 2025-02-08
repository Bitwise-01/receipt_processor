package api

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"receipt_processor/pkg/middleware"
	"receipt_processor/pkg/service"
)

// fakeService is a fake implementation of service.IReceiptService for testing.
type fakeService struct{}

func (f *fakeService) ProcessReceipt(ctx context.Context, receipt service.ReceiptDTO) (string, error) {
	return "fake-id", nil
}

func (f *fakeService) GetPoints(ctx context.Context, receiptID string) (int, error) {
	return 42, nil
}

// dummyMiddleware is a simple middleware that adds an "X-Dummy: dummy" header to the response.
func dummyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Dummy", "dummy")
		next.ServeHTTP(w, r)
	})
}

func TestRouterEndpoints(t *testing.T) {
	// Create a fake service and a middleware chain with dummyMiddleware.
	fs := &fakeService{}
	mws := []middleware.Middleware{dummyMiddleware}
	router := NewRouter(fs, mws)

	t.Run("POST /receipts/process", func(t *testing.T) {
		// A valid JSON payload for processing a receipt.
		payload := `{
			"retailer": "TestRetailer",
			"purchaseDate": "2022-01-01",
			"purchaseTime": "12:00",
			"total": "10.00",
			"items": [{"shortDescription": "Item A", "price": "5.00"}]
		}`
		req := httptest.NewRequest(http.MethodPost, "/receipts/process", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		// Send the request to our router.
		router.ServeHTTP(rec, req)
		res := rec.Result()
		defer res.Body.Close()

		// Verify status code.
		if res.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", res.StatusCode)
		}

		// Verify dummy middleware header.
		if res.Header.Get("X-Dummy") != "dummy" {
			t.Errorf("Expected X-Dummy header to be 'dummy', got '%s'", res.Header.Get("X-Dummy"))
		}

		// Verify the response body contains the expected receipt ID.
		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}
		bodyStr := string(bodyBytes)
		if !strings.Contains(bodyStr, `"id":"fake-id"`) {
			t.Errorf("Expected response to contain '\"id\":\"fake-id\"', got %s", bodyStr)
		}
	})

	t.Run("GET /receipts/{id}/points", func(t *testing.T) {
		// Send a GET request to the get points endpoint.
		req := httptest.NewRequest(http.MethodGet, "/receipts/test-id/points", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		res := rec.Result()
		defer res.Body.Close()

		// Verify status code.
		if res.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", res.StatusCode)
		}

		// Verify dummy middleware header.
		if res.Header.Get("X-Dummy") != "dummy" {
			t.Errorf("Expected X-Dummy header to be 'dummy', got '%s'", res.Header.Get("X-Dummy"))
		}

		// Verify the response body contains the expected points.
		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}
		bodyStr := string(bodyBytes)
		if !strings.Contains(bodyStr, `"points":42`) {
			t.Errorf("Expected response to contain '\"points\":42', got %s", bodyStr)
		}
	})
}
