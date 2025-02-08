package middleware

import (
	"context"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

const requestIDKey = "requestID"

var logger = zerolog.New(os.Stdout).With().Timestamp().Logger()

// RequestIDMiddleware attaches a unique request ID to each request.
func RequestIDMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if the request already has an ID; otherwise, generate one.
			reqID := r.Header.Get("X-Request-ID")
			if reqID == "" {
				reqID = uuid.New().String()
			}

			// Set the request ID in the response header.
			w.Header().Set("X-Request-ID", reqID)

			ctx := r.Context()
			ctx = context.WithValue(ctx, requestIDKey, reqID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
