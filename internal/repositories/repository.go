package repositories

import (
	"context"
	"log/slog"

	"github.com/donskova1ex/AverageRegionIncomes/internal/domain"
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db     *sqlx.DB
	logger *slog.Logger
}

func NewRepository(db *sqlx.DB, logger *slog.Logger) *Repository {
	return &Repository{db: db, logger: logger}
}

func (r *Repository) GetRegionIncomes(ctx context.Context, regionid int32, year int32, quarter int32) (*domain.RegionIncomes, error) {

	return nil, nil
}
