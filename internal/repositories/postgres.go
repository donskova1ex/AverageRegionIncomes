package repositories

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	SSLMode  string
}

func NewPostgresDB(ctx context.Context, pgDSN string) (*sqlx.DB, error) {
	db, err := sqlx.ConnectContext(ctx, "postgres", pgDSN)

	if err != nil {
		return nil, fmt.Errorf("failed to open connection to postgres: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	return db, nil
}
