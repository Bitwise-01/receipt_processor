package repository

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	rd "github.com/redis/go-redis/v9"
)

func TestAllowRequest(t *testing.T) {
	// Create a Redis client and a redismock instance.
	client, mock := redismock.NewClientMock()
	repo := NewRateLimiterRepository(client)
	ctx := context.Background()

	// Use a fixed key representing a client's IP.
	key := "rate_limit:127.0.0.1"
	window := time.Minute
	maxRequests := 2

	// Set a cutoff value for the current sliding window.
	nowMillis := time.Now().UnixNano() / int64(time.Millisecond)
	windowMillis := int64(window / time.Millisecond)
	cutoff := nowMillis - windowMillis
	cutoffStr := strconv.FormatInt(cutoff, 10)

	// ---- First Request: No previous requests.
	mock.ExpectZRemRangeByScore(key, "0", cutoffStr).SetVal(0)
	mock.ExpectZCard(key).SetVal(0)
	// For testing, we don't care about the exact timestamp value passed into ZAdd.
	mock.ExpectZAdd(key, rd.Z{
		Score:  float64(nowMillis),
		Member: nowMillis,
	}).SetVal(1)
	mock.ExpectExpire(key, window).SetVal(true)

	allowed, err := repo.AllowRequest(ctx, key, window, maxRequests)
	if err != nil {
		t.Fatalf("unexpected error on first request: %v", err)
	}
	if !allowed {
		t.Errorf("expected first request to be allowed, but it was denied")
	}

	// ---- Second Request: Count should be 1.
	nowMillis = time.Now().UnixNano() / int64(time.Millisecond)
	mock.ExpectZRemRangeByScore(key, "0", cutoffStr).SetVal(0)
	mock.ExpectZCard(key).SetVal(1)
	mock.ExpectZAdd(key, rd.Z{
		Score:  float64(nowMillis),
		Member: nowMillis,
	}).SetVal(1)
	mock.ExpectExpire(key, window).SetVal(true)

	allowed, err = repo.AllowRequest(ctx, key, window, maxRequests)
	if err != nil {
		t.Fatalf("unexpected error on second request: %v", err)
	}
	if !allowed {
		t.Errorf("expected second request to be allowed, but it was denied")
	}

	// ---- Third Request: Count is now 2 so should be denied.
	mock.ExpectZRemRangeByScore(key, "0", cutoffStr).SetVal(0)
	mock.ExpectZCard(key).SetVal(2)

	allowed, err = repo.AllowRequest(ctx, key, window, maxRequests)
	if err != nil {
		t.Fatalf("unexpected error on third request: %v", err)
	}
	if allowed {
		t.Errorf("expected third request to be denied, but it was allowed")
	}

	// Verify that all expectations were met.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}
