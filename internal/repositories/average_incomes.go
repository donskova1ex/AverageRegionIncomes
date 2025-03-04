package repositories

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/donskova1ex/AverageRegionIncomes/internal/domain"
)

func (r *Repository) CreateRegionIncomes(ctx context.Context, regionIncomes []*domain.RegionIncomes) error {
	var txCommited bool

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error start transaction: %w", err)
	}

	defer func() {
		if !txCommited {
			if err := tx.Rollback(); err != nil {
				r.logger.Error("error rollback transaction", slog.String("err", err.Error()))
			}
		}
	}()

	return nil
}
