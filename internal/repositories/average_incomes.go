package repositories

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/donskova1ex/AverageRegionIncomes/internal/domain"
	"github.com/jmoiron/sqlx"
)

func (r *Repository) CreateRegionIncomes(ctx context.Context, exRegionIncomes []*domain.ExcelRegionIncome) error {
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

	regionsMap, err := r.getRegionsMap(ctx, tx)
	regionIncomesSlice := make([]*domain.RegionIncomes, len(exRegionIncomes))

	for _, region := range exRegionIncomes {
		regionIncomesSlice = append(regionIncomesSlice, &domain.RegionIncomes{
			RegionId:             regionsMap[region.Region],
			AverageRegionIncomes: region.AverageRegionIncomes,
			Year:                 region.Year,
			Quarter:              region.Quarter,
		})
	}

	query := `INSERT INTO region_incomes (RegionId, Year, Quarter, Value) 
	VALUES (:RegionId, :AverageRegionIncomes, :Year, :Quarter)
	ON CONFLICT (RegionId, Year, Quarter, Value) DO NOTHING`

	rows, err := tx.QueryxContext(ctx, query)
	if err != nil {
		return fmt.Errorf("error execute query: %w", err)
	}
	if rows.Err() != nil {
		return fmt.Errorf("error inserting rows: %w", rows.Err())
	}

	return nil
}

func (r *Repository) getRegionsMap(ctx context.Context, tx *sqlx.Tx) (map[string]int32, error) {
	query := `SELECT regionid, regionname FROM regions`

	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error execute query: %w", err)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("error reading rows: %w", err)
	}
	regionsMap := make(map[string]int32)
	var regionID int32
	var regionName string

	for rows.Next() {
		if err := rows.Scan(&regionID, &regionName); err != nil {
			return nil, fmt.Errorf("error scan row: %w", err)
		}
		regionsMap[regionName] = regionID
	}
	return regionsMap, nil
}
