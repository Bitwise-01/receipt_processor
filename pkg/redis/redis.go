package redis

import (
	"context"
	"encoding/json"
	"time"

	rd "github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

// RedisClient wraps the go-redis client.
type RedisClient struct {
	Rdb *rd.Client
}

// New initializes and returns a new RedisClient instance.
func New() (*RedisClient, error) {
	rdb := rd.NewClient(&rd.Options{
		Addr:         viper.GetString("redis.addr"),
		Username:     viper.GetString("redis.username"),
		Password:     viper.GetString("redis.password"),
		DB:           viper.GetInt("redis.db"),
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
	})

	// Test the connection.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		return nil, err
	}

	return &RedisClient{Rdb: rdb}, nil
}

// Set serializes a value to JSON and stores it in Redis.
func (rc *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return rc.Rdb.Set(ctx, key, data, expiration).Err()
}

// Get retrieves and deserializes the value stored at the given key.
func (rc *RedisClient) Get(ctx context.Context, key string, dest interface{}) error {
	result, err := rc.Rdb.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(result), dest)
}

// GetInt retrieves an integer value from Redis.
func (rc *RedisClient) GetInt(ctx context.Context, key string) (int, error) {
	count, err := rc.Rdb.Get(ctx, key).Int()
	if err != nil {
		if err == rd.Nil {
			return 0, nil
		}
		return 0, err
	}
	return count, nil
}

// Del deletes a key from Redis.
func (rc *RedisClient) Del(ctx context.Context, key string) error {
	return rc.Rdb.Del(ctx, key).Err()
}
