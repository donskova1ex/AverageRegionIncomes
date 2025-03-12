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
        INSERT INTO region_incomes (region_id, year, quarter, value) 
        VALUES (:region_id, :year, :quarter, :value)
        ON CONFLICT (region_id, year, quarter, value) DO NOTHING`

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
	query := `SELECT region_id, region_name FROM regions`

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
					r.region_name AS RegionName, 
					ri.region_id AS RegionId, 
					EXTRACT(YEAR FROM CURRENT_DATE) AS Year,
					FLOOR((EXTRACT(MONTH FROM CURRENT_DATE) - 1) / 3) + 1 AS Quarter, 
					AVG(ri.value) AS AverageRegionIncomes 
				FROM (
				    SELECT region_id, value 
						FROM region_incomes 
						WHERE region_id = $1 
						ORDER BY year DESC, quarter DESC 
						LIMIT 4) AS ri 
				JOIN regions r ON ri.region_id = r.region_id 
				GROUP BY r.region_name, ri.region_id`

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
					r.region_name AS RegionName,
					ri.region_id AS RegionId,
					EXTRACT(YEAR FROM CURRENT_DATE) AS Year,
					FLOOR((EXTRACT(MONTH FROM CURRENT_DATE) - 1) / 3) + 1 AS Quarter,
					AVG(ri.value) AS AverageRegionIncomes
				FROM (
					SELECT region_id, year, quarter, value
						FROM region_incomes
						WHERE region_id = $1
						  AND year = $2
					UNION ALL
					SELECT region_id, year, quarter, value
						FROM region_incomes
						WHERE region_id = $1
						  AND year = $2 - 1
					ORDER BY year DESC, quarter DESC
					LIMIT 4
				) AS ri
				JOIN regions r ON ri.region_id = r.region_id
				GROUP BY r.region_name, ri.region_id`

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
					incomes.region_id,
					r.region_name AS region_name,
					$3 AS quarter,
					$2 AS year,
					AVG(incomes.value) AS average_region_incomes
				FROM (
					SELECT 
						ri.region_id, 
						ri.year, 
						ri.quarter, 
						ri.value
					FROM 
						region_incomes ri
					WHERE 
						ri.region_id = $1 
						AND ri.year <= $2 
						AND NOT (ri.year = $2 AND ri.quarter >= $3) 
					ORDER BY 
						ri.year DESC, 
						ri.quarter DESC 
					LIMIT 4
				) AS incomes
				JOIN 
					regions r 
				ON 
					incomes.region_id = r.region_id
				GROUP BY 
					incomes.region_id, 
					r.region_name`

	err := r.db.GetContext(ctx, averageRegionIncomes, query, regionId, year, quarter)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("region not found with region_id [%d], year [%d], quarter [%d]: %w", regionId, year, quarter, err)
	}
	if err != nil {
		return nil, fmt.Errorf("err getting incomes by region_id [%d], year [%d], quarter [%d]: %w", regionId, year, quarter, err)
	}

	return averageRegionIncomes, nil
}
