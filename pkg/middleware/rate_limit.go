package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"receipt_processor/pkg/repository"
)

const (
	// maxRequests is the maximum number of requests allowed in the window period.
	maxRequests = 5 // Keep it low to test the effects
	// windowPeriod is the duration of the sliding window.
	windowPeriod = time.Minute
)

// RateLimitMiddleware enforces rate limiting using a sliding window algorithm with Redis.
func RateLimitMiddleware(rateLimiter repository.IRateLimiterRepository) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			host, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				host = strings.Split(r.RemoteAddr, ":")[0]
			}
			// Use the remote IP address as a unique key.
			key := fmt.Sprintf("rate_limit:%s", host)

			allowed, err := rateLimiter.AllowRequest(r.Context(), key, windowPeriod, maxRequests)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if !allowed {
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
