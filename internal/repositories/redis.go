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

	password := os.Getenv("REDIS_PASSWORD")
	if password == "" {
		return nil, fmt.Errorf("REDIS_PASSWORD environment variable not set")
	}

	username := os.Getenv("REDIS_USERNAME")
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
