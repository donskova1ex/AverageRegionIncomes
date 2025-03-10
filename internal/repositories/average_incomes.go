package repositories

import (
	"context"
	"database/sql"
	"errors"
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
		return fmt.Errorf("error filling regions map: %w", err)
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
			r.logger.Warn(
				"region not found in regions map",
				slog.String("region", fmt.Sprintf("[%s]", region.Region)),
			)
		}
	}

	query := `
        INSERT INTO region_incomes (RegionId, Year, Quarter, Value) 
        VALUES (:RegionId, :Year, :Quarter, :Value)
        ON CONFLICT (RegionId, Year, Quarter, Value) DO NOTHING`

	//TODO: изоляция транзакций
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

func (r *Repository) GetRegionIncomes(ctx context.Context, regionId int32, year int32, quarter int32) (*domain.AverageRegionIncomes, error) {

	if year == 0 && quarter == 0 {
		averageRegionIncomes, err := r.getIncomesByRegionID(ctx, regionId)
		if err != nil {
			return nil, err
		}
		return averageRegionIncomes, nil
	}

	if quarter == 0 {
		averageRegionIncomes, err := r.getIncomesByRegionIDAndYear(ctx, regionId, year)
		if err != nil {
			return nil, err
		}
		return averageRegionIncomes, nil
	}

	averageRegionIncomes, err := r.getRegionIncomesByAllParameters(ctx, regionId, year, quarter)
	if err != nil {
		return nil, err
	}

	return averageRegionIncomes, nil
}
func (r *Repository) getIncomesByRegionID(ctx context.Context, regionId int32) (*domain.AverageRegionIncomes, error) {
	averageRegionIncomes := &domain.AverageRegionIncomes{}
	query := `SELECT 
					r.regionname AS RegionName, 
					ri.regionid AS RegionId, 
					EXTRACT(YEAR FROM CURRENT_DATE) AS Year,
					FLOOR((EXTRACT(MONTH FROM CURRENT_DATE) - 1) / 3) + 1 AS Quarter, 
					AVG(ri.value) AS AverageRegionIncomes 
				FROM (
				    SELECT regionid, value 
						FROM region_incomes 
						WHERE regionid = $1 
						ORDER BY year DESC, quarter DESC 
						LIMIT 4) AS ri 
				JOIN regions r ON ri.regionid = r.regionid 
				GROUP BY r.regionname, ri.regionid`

	err := r.db.GetContext(ctx, averageRegionIncomes, query, regionId)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("region not found with region_id [%d]: %w", regionId, err)
	}
	if err != nil {
		return nil, fmt.Errorf("err getting incomes by region_id [%d]: %w", regionId, err)
	}

	return averageRegionIncomes, nil
}

func (r *Repository) getIncomesByRegionIDAndYear(ctx context.Context, regionId int32, year int32) (*domain.AverageRegionIncomes, error) {
	averageRegionIncomes := &domain.AverageRegionIncomes{}
	query := `SELECT 
					r.regionname AS RegionName,
					ri.regionid AS RegionId,
					EXTRACT(YEAR FROM CURRENT_DATE) AS Year,
					FLOOR((EXTRACT(MONTH FROM CURRENT_DATE) - 1) / 3) + 1 AS Quarter,
					AVG(ri.value) AS AverageRegionIncomes
				FROM (
					SELECT regionid, year, quarter, value
						FROM region_incomes
						WHERE regionid = $1
						  AND year = $2
					UNION ALL
					SELECT regionid, year, quarter, value
						FROM region_incomes
						WHERE regionid = $1
						  AND year = $2 - 1
					ORDER BY year DESC, quarter DESC
					LIMIT 4
				) AS ri
				JOIN regions r ON ri.regionid = r.regionid
				GROUP BY r.regionname, ri.regionid`

	err := r.db.GetContext(ctx, averageRegionIncomes, query, regionId, year)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("region not found with region_id [%d], year [%d]: %w", regionId, year, err)
	}
	if err != nil {
		return nil, fmt.Errorf("err getting incomes by region_id [%d], year [%d]: %w", regionId, year, err)
	}

	return averageRegionIncomes, nil
}

func (r *Repository) getRegionIncomesByAllParameters(ctx context.Context, regionId int32, year int32, quarter int32) (*domain.AverageRegionIncomes, error) {
	averageRegionIncomes := &domain.AverageRegionIncomes{}

	query := `SELECT 
					incomes.RegionId,
					r.regionname AS RegionName,
					$3 AS Quarter,
					$2 AS Year,
					AVG(incomes.value) AS AverageRegionIncomes
				FROM (
					SELECT 
						ri.RegionId, 
						ri.year, 
						ri.quarter, 
						ri.value
					FROM 
						region_incomes ri
					WHERE 
						ri.RegionId = $1 
						AND ri.Year <= $2 
						AND NOT (ri.Year = $2 AND ri.Quarter >= $3) 
					ORDER BY 
						ri.Year DESC, 
						ri.Quarter DESC 
					LIMIT 4
				) AS incomes
				JOIN 
					regions r 
				ON 
					incomes.RegionId = r.RegionId
				GROUP BY 
					incomes.RegionId, 
					r.regionname`

	err := r.db.GetContext(ctx, averageRegionIncomes, query, regionId, year, quarter)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("region not found with region_id [%d], year [%d], quarter [%d]: %w", regionId, year, quarter, err)
	}
	if err != nil {
		return nil, fmt.Errorf("err getting incomes by region_id [%d], year [%d], quarter [%d]: %w", regionId, year, quarter, err)
	}

	return averageRegionIncomes, nil
}
