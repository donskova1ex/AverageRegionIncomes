package repositories

import (
	"log/slog"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db     *sqlx.DB
	logger *slog.Logger
}

func NewRepository(db *sqlx.DB, logger *slog.Logger) *Repository {
	return &Repository{db: db, logger: logger}
}
