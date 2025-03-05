package repositories

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/donskova1ex/AverageRegionIncomes/internal/domain"
	"github.com/jmoiron/sqlx"
)

func (r *Repository) CreateRegionIncomes(ctx context.Context, exRegionIncomes []*domain.ExcelRegionIncome) error {
	var txCommited bool

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}

	defer func() {
		if !txCommited {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				r.logger.Error("error rolling back transaction", slog.String("err", rollbackErr.Error()))
			}
		}
	}()

	regionsMap, err := r.getRegionsMap(ctx, tx)
	if err != nil {
		return fmt.Errorf("error fetching regions map: %w", err)
	}

	regionIncomes := make([]*domain.RegionIncomes, 0, len(exRegionIncomes))
	for _, region := range exRegionIncomes {
		trimmedRegionName := strings.ReplaceAll(region.Region, " ", "")
		if regionID, ok := regionsMap[trimmedRegionName]; ok {
			regionIncomes = append(regionIncomes, &domain.RegionIncomes{
				RegionId: regionID,
				Value:    region.AverageRegionIncomes,
				Year:     region.Year,
				Quarter:  region.Quarter,
			})
		} else {
			r.logger.Warn("region not found in regions map", slog.String("region", region.Region))
		}
	}

	query := `
        INSERT INTO region_incomes (RegionId, Year, Quarter, Value) 
        VALUES (:RegionId, :Year, :Quarter, :Value)
        ON CONFLICT (RegionId, Year, Quarter, Value) DO NOTHING`

	result, err := tx.NamedExec(query, regionIncomes)
	if err != nil {
		r.logger.Error("error executing query", slog.String("err", err.Error()))
		return fmt.Errorf("error executing query: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	r.logger.Info("rows inserted/updated", slog.Int("count", int(rowsAffected)))

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}
	txCommited = true

	return nil
}

func (r *Repository) getRegionsMap(ctx context.Context, tx *sqlx.Tx) (map[string]int32, error) {
	query := `SELECT regionid, regionname FROM regions`

	rows, err := tx.QueryxContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %w", err)
	}
	defer rows.Close()

	regionsMap := make(map[string]int32)

	for rows.Next() {
		var regionID int32
		var regionName string

		if err := rows.Scan(&regionID, &regionName); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		trimmedRegionName := strings.ReplaceAll(regionName, " ", "")
		regionsMap[trimmedRegionName] = regionID
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading rows: %w", err)
	}

	return regionsMap, nil
}
