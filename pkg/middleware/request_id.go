package middleware

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// RequestIDMiddleware attaches a unique request ID to each incoming request,
// sets it as a header, and injects a logger (with that request ID) into the context.
func RequestIDMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check for an existing request ID in the header.
			reqID := r.Header.Get("X-Request-ID")
			if reqID == "" {
				// Generate a new request ID if one isn't provided.
				reqID = uuid.New().String()
			}
			// Set the request ID in the response header.
			w.Header().Set("X-Request-ID", reqID)
			// Create a logger that includes the request_id field.
			logger := log.With().Str("request_id", reqID).Logger()
			// Attach the logger to the request context.
			ctx := logger.WithContext(r.Context())
			// Pass the request with the updated context to the next handler.
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
