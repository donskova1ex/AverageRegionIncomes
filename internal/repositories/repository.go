package repositories

import (
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type SQLRepository struct {
	db     *sqlx.DB
	logger *slog.Logger
}

type RedisRepository struct {
	db     *redis.Client
	ttl    time.Duration
	logger *slog.Logger
}

func NewSQLRepository(db *sqlx.DB, logger *slog.Logger) *SQLRepository {
	return &SQLRepository{db: db, logger: logger}
}

func NewRedisRepository(db *redis.Client, ttl time.Duration, logger *slog.Logger) *RedisRepository {
	return &RedisRepository{db: db, ttl: ttl, logger: logger}
}
