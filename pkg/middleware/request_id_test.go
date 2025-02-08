package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestIDMiddleware(t *testing.T) {
	// Define a dummy handler that simply returns 200 OK.
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Table-driven test cases.
	testCases := []struct {
		name               string
		requestHeaderValue string // X-Request-ID header in request.
		expectSame         bool   // If true, expect the same header value in the response.
	}{
		{
			name:               "No Request ID Provided",
			requestHeaderValue: "",
			expectSame:         false,
		},
		{
			name:               "Request ID Provided",
			requestHeaderValue: "test-request-id",
			expectSame:         true,
		},
	}

	// Create the middleware.
	mw := RequestIDMiddleware()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new HTTP request.
			req := httptest.NewRequest("GET", "http://example.com", nil)
			if tc.requestHeaderValue != "" {
				req.Header.Set("X-Request-ID", tc.requestHeaderValue)
			}

			// Create a ResponseRecorder to capture the response.
			rr := httptest.NewRecorder()

			// Wrap the dummy handler with the middleware.
			handler := mw(dummyHandler)
			handler.ServeHTTP(rr, req)

			// Retrieve the X-Request-ID from the response header.
			respID := rr.Header().Get("X-Request-ID")
			if respID == "" {
				t.Errorf("expected X-Request-ID header to be set in the response, but it was empty")
			}

			// If a header was provided in the request, the response should contain the same value.
			if tc.expectSame && respID != tc.requestHeaderValue {
				t.Errorf("expected X-Request-ID header to be %q, got %q", tc.requestHeaderValue, respID)
			}

			// If no header was provided, ensure that a non-empty header was generated.
			if !tc.expectSame && tc.requestHeaderValue == "" && respID == "" {
				t.Errorf("expected middleware to generate a new request ID, but got empty")
			}
		})
	}
}
