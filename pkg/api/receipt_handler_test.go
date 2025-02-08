package api

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"receipt_processor/pkg/service"
)

// fakeReceiptService is a fake implementation of service.IReceiptService for testing.
type fakeReceiptService struct{}

// ProcessReceipt returns "test-id" unless the retailer is "error", in which case it returns an error.
func (f *fakeReceiptService) ProcessReceipt(ctx context.Context, receipt service.ReceiptDTO) (string, error) {
	if receipt.Retailer == "error" {
		return "", errors.New("processing error")
	}
	return "test-id", nil
}

// GetPoints returns 42 points unless the receiptID is "error-id", in which case it returns an error.
func (f *fakeReceiptService) GetPoints(ctx context.Context, receiptID string) (int, error) {
	if receiptID == "error-id" {
		return 0, errors.New("receipt not found")
	}
	return 42, nil
}

func TestProcessReceiptHandler(t *testing.T) {
	// Create a fake service and a Router that uses it.
	fakeService := &fakeReceiptService{}
	router := &Router{receiptService: fakeService}

	// Define table test cases for the ProcessReceiptHandler.
	testCases := []struct {
		name                      string
		method                    string
		url                       string
		body                      string
		expectedStatus            int
		expectedResponseSubstring string
	}{
		{
			name:           "Valid Request",
			method:         http.MethodPost,
			url:            "/receipts/process",
			body:           `{"retailer": "Target", "purchaseDate": "2022-01-01", "purchaseTime": "13:01", "total": "35.35", "items": [{"shortDescription": "Item A", "price": "10.00"}]}`,
			expectedStatus: http.StatusOK,
			// We expect our fake service to return "test-id".
			expectedResponseSubstring: `"id":"test-id"`,
		},
		{
			name:           "Invalid Method",
			method:         http.MethodGet,
			url:            "/receipts/process",
			body:           ``,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Invalid JSON",
			method:         http.MethodPost,
			url:            "/receipts/process",
			body:           `invalid-json`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Service Error",
			method:         http.MethodPost,
			url:            "/receipts/process",
			body:           `{"retailer": "error", "purchaseDate": "2022-01-01", "purchaseTime": "13:01", "total": "35.35", "items": [{"shortDescription": "Item A", "price": "10.00"}]}`,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.url, bytes.NewBufferString(tc.body))
			w := httptest.NewRecorder()
			// Call the ProcessReceiptHandler directly.
			router.ProcessReceiptHandler(w, req)
			resp := w.Result()
			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, resp.StatusCode)
			}
			// Read the response body.
			responseData, _ := io.ReadAll(resp.Body)
			bodyStr := string(responseData)
			if tc.expectedResponseSubstring != "" && !strings.Contains(bodyStr, tc.expectedResponseSubstring) {
				t.Errorf("expected response to contain %q, got %q", tc.expectedResponseSubstring, bodyStr)
			}
		})
	}
}

func TestGetPointsHandler(t *testing.T) {
	// Create a fake service and a Router that uses it.
	fakeService := &fakeReceiptService{}
	router := &Router{receiptService: fakeService}

	// Define table test cases for the GetPointsHandler.
	testCases := []struct {
		name                      string
		method                    string
		url                       string
		expectedStatus            int
		expectedResponseSubstring string
	}{
		{
			name:                      "Valid Get",
			method:                    http.MethodGet,
			url:                       "/receipts/test-id/points",
			expectedStatus:            http.StatusOK,
			expectedResponseSubstring: `"points":42`,
		},
		{
			name:           "Invalid Method",
			method:         http.MethodPost,
			url:            "/receipts/test-id/points",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Invalid URL",
			method:         http.MethodGet,
			url:            "/invalid/test-id/points",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:                      "Service Error",
			method:                    http.MethodGet,
			url:                       "/receipts/error-id/points",
			expectedStatus:            http.StatusNotFound,
			expectedResponseSubstring: "", // We don't expect a valid JSON body in this error case.
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.url, nil)
			w := httptest.NewRecorder()
			// Call the GetPointsHandler directly.
			router.GetPointsHandler(w, req)
			resp := w.Result()
			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, resp.StatusCode)
			}
			responseData, _ := io.ReadAll(resp.Body)
			bodyStr := string(responseData)
			if tc.expectedResponseSubstring != "" && !strings.Contains(bodyStr, tc.expectedResponseSubstring) {
				t.Errorf("expected response to contain %q, got %q", tc.expectedResponseSubstring, bodyStr)
			}
		})
	}
}
