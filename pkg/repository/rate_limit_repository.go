package repository

import (
	"context"
	"strconv"
	"time"

	rd "github.com/redis/go-redis/v9"
)

// IRateLimiterRepository defines the interface for a sliding window rate limiter.
type IRateLimiterRepository interface {
	// AllowRequest returns true if the request is allowed under the rate limit,
	// or false if the limit has been exceeded.
	AllowRequest(ctx context.Context, key string, window time.Duration, maxRequests int) (bool, error)
}

type rateLimiterRepository struct {
	client *rd.Client
}

// NewRateLimiterRepository creates a new instance of the rate limiter repository.
func NewRateLimiterRepository(client *rd.Client) IRateLimiterRepository {
	return &rateLimiterRepository{
		client: client,
	}
}

// AllowRequest implements the sliding window rate limiter.
// It uses a Redis sorted set to store timestamps (in milliseconds) for each request.
func (r *rateLimiterRepository) AllowRequest(ctx context.Context, key string, window time.Duration, maxRequests int) (bool, error) {
	now := time.Now().UnixNano() / int64(time.Millisecond) // current time in ms
	windowMillis := int64(window / time.Millisecond)
	cutoff := now - windowMillis

	// Remove entries older than the current sliding window.
	if err := r.client.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(cutoff, 10)).Err(); err != nil {
		return false, err
	}

	// Count the number of requests in the sliding window.
	count, err := r.client.ZCard(ctx, key).Result()
	if err != nil {
		return false, err
	}

	if int(count) >= maxRequests {
		return false, nil
	}

	// Add the current request timestamp.
	if err := r.client.ZAdd(ctx, key, rd.Z{
		Score:  float64(now),
		Member: now,
	}).Err(); err != nil {
		return false, err
	}

	// Set an expiration for cleanup.
	r.client.Expire(ctx, key, window)

	return true, nil
}
