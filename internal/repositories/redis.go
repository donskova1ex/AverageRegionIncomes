package repositories

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	Username string
}
type RedisOptions struct {
	options *redis.Options
	ttl     time.Duration
}

type RedisDBData struct {
	DB  *redis.Client
	TTL time.Duration
}

// newRedisOptions initializes Redis connection options from environment variables.
// It loads configuration from .env file and validates required environment variables:
// - REDIS_HOST: Redis server host
// - REDIS_PORT: Redis server port
// - REDIS_USER_PASSWORD: Redis user password
// - REDIS_USER: Redis username
// - REDIS_TTL_ACTUAL_DATA: Time-to-live duration for cached data
//
// Returns:
//   - *RedisOptions: Redis connection options and TTL settings
//   - error: error if environment variables are missing or invalid
func newRedisOptions() (*RedisOptions, error) {
	err := godotenv.Load("/app/config/.env.dev")
	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	host := os.Getenv("REDIS_HOST")
	if host == "" {
		return nil, fmt.Errorf("REDIS_HOST environment variable not set")
	}

	port := os.Getenv("REDIS_PORT")
	if port == "" {
		return nil, fmt.Errorf("REDIS_PORT environment variable not set")
	}

	password := os.Getenv("REDIS_USER_PASSWORD")
	if password == "" {
		return nil, fmt.Errorf("REDIS_PASSWORD environment variable not set")
	}

	username := os.Getenv("REDIS_USER")
	if username == "" {
		return nil, fmt.Errorf("REDIS_USERNAME environment variable not set")
	}

	actualDataTtlStr := os.Getenv("REDIS_TTL_ACTUAL_DATA")
	if actualDataTtlStr == "" {
		return nil, fmt.Errorf("REDIS_TTL_ACTUAL_DATA environment variable not set")
	}

	actualDataTtl, err := time.ParseDuration(actualDataTtlStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing REDIS_TTL_ACTUAL_DATA: %w", err)
	}

	opts := &redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		Username: username,
	}

	return &RedisOptions{options: opts, ttl: actualDataTtl}, nil
}

// NewRedisDB creates a new Redis database connection with the specified configuration.
// It initializes the connection using environment variables and performs a connection test.
// The function returns a RedisDBData structure containing the Redis client and TTL settings,
// or an error if the connection fails or configuration is invalid.
//
// Parameters:
//   - ctx: context.Context for managing the connection lifecycle
//
// Returns:
//   - *RedisDBData: pointer to the Redis database structure containing the client and TTL
//   - error: error if connection or configuration fails
func NewRedisDB(ctx context.Context) (*RedisDBData, error) {
	opts, err := newRedisOptions()
	if err != nil {
		return nil, fmt.Errorf("error creating redis options: %w", err)
	}

	db := redis.NewClient(opts.options)
	if err := db.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("error connecting to redis: %w", err)
	}

	return &RedisDBData{DB: db, TTL: opts.ttl}, nil
}
