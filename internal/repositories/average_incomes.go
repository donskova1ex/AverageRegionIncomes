package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"log/slog"
	"strings"

	"github.com/donskova1ex/AverageRegionIncomes/internal/domain"
	"github.com/jmoiron/sqlx"
)

func (r *SQLRepository) CreateRegionIncomes(ctx context.Context, exRegionIncomes []*domain.ExcelRegionIncome) error {
	const maxRetries = 5
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		err := r.createRegionIncomesWithTx(ctx, exRegionIncomes)
		if err == nil {
			r.logger.Info("Re")
			return nil
		}

		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "40001" {
			r.logger.Warn("serialization failure, retrying transaction",
				slog.Int("attempt", attempt+1),
				slog.String("err", pqErr.Error()))
			lastErr = fmt.Errorf("serialization error (retry %d): %w", attempt+1, err)
			continue
		}

		r.logger.Error("non-retryable error on attempt",
			slog.Int("attempt", attempt+1),
			slog.String("err", err.Error()))
		return fmt.Errorf("non-retryable error on attempt %d: %w", attempt+1, err)
	}

	return fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

func (r *SQLRepository) getRegionsMap(ctx context.Context, tx *sqlx.Tx) (map[string]int32, error) {
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

func (r *SQLRepository) createRegionIncomesWithTx(ctx context.Context, exRegionIncomes []*domain.ExcelRegionIncome) error {
	var txCommited bool

	serializableIsolation := &sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	}
	tx, err := r.db.BeginTxx(ctx, serializableIsolation)
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
		}
	}

	query := `
        INSERT INTO region_incomes (region_id, year, quarter, value) 
        VALUES (:region_id, :year, :quarter, :value)
        ON CONFLICT (region_id, year, quarter, value) DO NOTHING`

	result, err := tx.NamedExec(query, regionIncomes)
	if err != nil {
		r.logger.Error("error executing query", slog.String("err", err.Error()))
		return fmt.Errorf("error executing query: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	r.logger.Info("rows inserted", slog.Int("count", int(rowsAffected)))

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}
	txCommited = true

	return nil
}

func (r *SQLRepository) GetRegionIncomes(ctx context.Context, regionId int32, year int32, quarter int32) (*domain.AverageRegionIncomes, error) {
	var txCommited bool

	readOnlyTx := &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
		ReadOnly:  true,
	}

	tx, err := r.db.BeginTxx(ctx, readOnlyTx)
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}

	defer func() {
		if !txCommited {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				r.logger.Error("error rolling back transaction", slog.String("err", rollbackErr.Error()))
			}
		}
	}()

	var result *domain.AverageRegionIncomes
	var queryErr error

	if year == 0 && quarter == 0 {
		result, queryErr = r.getIncomesByRegionID(ctx, tx, regionId)
	} else if quarter == 0 {
		result, queryErr = r.getIncomesByRegionIDAndYear(ctx, tx, regionId, year)
	} else {
		result, queryErr = r.getRegionIncomesByAllParameters(ctx, tx, regionId, year, quarter)
	}

	if queryErr != nil {
		return nil, queryErr
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}
	txCommited = true
	return result, nil
}

func (r *SQLRepository) getIncomesByRegionID(ctx context.Context, tx *sqlx.Tx, regionId int32) (*domain.AverageRegionIncomes, error) {
	averageRegionIncomes := &domain.AverageRegionIncomes{}
	/*query := `SELECT
		r.region_name AS Region_Name,
		ri.region_id AS Region_Id,
		EXTRACT(YEAR FROM CURRENT_DATE) AS Year,
		FLOOR((EXTRACT(MONTH FROM CURRENT_DATE) - 1) / 3) + 1 AS Quarter,
		AVG(ri.value) AS Average_Region_Incomes
	FROM (
	    SELECT region_id, value
			FROM region_incomes
			WHERE region_id = $1
			ORDER BY year DESC, quarter DESC
			LIMIT 4) AS ri
	JOIN regions r ON ri.region_id = r.region_id
	GROUP BY r.region_name, ri.region_id`*/

	query := `SELECT
					R.REGION_NAME AS REGION_NAME,
					RI.REGION_ID AS REGION_ID,
					EXTRACT(YEAR FROM CURRENT_DATE) AS YEAR,
					FLOOR((EXTRACT(MONTH FROM CURRENT_DATE) - 1) / 3) + 1 AS QUARTER,
					AVG(RI.VALUE) AS AVERAGE_REGION_INCOMES
				FROM
					(
					SELECT REGION_ID, VALUE
					FROM
						(SELECT DISTINCT ON (QUARTER, YEAR)
							REGION_ID,
							VALUE,
							LOADED_AT,
							QUARTER,
							YEAR
						FROM
							REGION_INCOMES
						WHERE
							REGION_ID = $1
						ORDER BY
							QUARTER DESC,
							YEAR DESC,
							LOADED_AT DESC
				) AS LATEST_QUARTERS
					ORDER BY
						YEAR DESC,
						QUARTER DESC
					LIMIT 4) AS RI
				JOIN REGIONS R ON
					RI.REGION_ID = R.REGION_ID
				GROUP BY
					R.REGION_NAME,
					RI.REGION_ID`

	err := tx.GetContext(ctx, averageRegionIncomes, query, regionId)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("region not found with region_id [%d]: %w", regionId, err)
	}
	if err != nil {
		return nil, fmt.Errorf("err getting incomes by region_id [%d]: %w", regionId, err)
	}

	return averageRegionIncomes, nil
}

func (r *SQLRepository) getIncomesByRegionIDAndYear(ctx context.Context, tx *sqlx.Tx, regionId int32, year int32) (*domain.AverageRegionIncomes, error) {
	averageRegionIncomes := &domain.AverageRegionIncomes{}
	query := `SELECT 
					r.region_name AS Region_Name,
					ri.region_id AS Region_Id,
					EXTRACT(YEAR FROM CURRENT_DATE) AS Year,
					FLOOR((EXTRACT(MONTH FROM CURRENT_DATE) - 1) / 3) + 1 AS Quarter,
					AVG(ri.value) AS Average_Region_Incomes
				FROM (
					SELECT region_id, value
					FROM (
						SELECT DISTINCT ON (year, quarter)
							region_id,
							year,
							quarter,
							value,
							loaded_at
						FROM region_incomes
						WHERE region_id = $1
						  AND (
							  year = $2
							  OR year = $2 - 1
						  )
						ORDER BY year DESC, quarter DESC, loaded_at DESC
					) AS latest_quarters
					ORDER BY year DESC, quarter DESC
					LIMIT 4
				) AS ri
				JOIN regions r ON ri.region_id = r.region_id
				GROUP BY r.region_name, ri.region_id`

	err := tx.GetContext(ctx, averageRegionIncomes, query, regionId, year)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("region not found with region_id [%d], year [%d]: %w", regionId, year, err)
	}
	if err != nil {
		return nil, fmt.Errorf("err getting incomes by region_id [%d], year [%d]: %w", regionId, year, err)
	}

	return averageRegionIncomes, nil
}

func (r *SQLRepository) getRegionIncomesByAllParameters(ctx context.Context, tx *sqlx.Tx, regionId int32, year int32, quarter int32) (*domain.AverageRegionIncomes, error) {
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

	err := tx.GetContext(ctx, averageRegionIncomes, query, regionId, year, quarter)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("region not found with region_id [%d], year [%d], quarter [%d]: %w", regionId, year, quarter, err)
	}
	if err != nil {
		return nil, fmt.Errorf("err getting incomes by region_id [%d], year [%d], quarter [%d]: %w", regionId, year, quarter, err)
	}

	return averageRegionIncomes, nil
}

func (r *RedisRepository) GetCachedRegionIncomes(ctx context.Context, regionId int32, year int32, quarter int32) (*domain.AverageRegionIncomes, error) {
	averageRegionIncomes := &domain.AverageRegionIncomes{}

	redisKey := createCachedKey(regionId, year, quarter)

	var averageRegionIncomesJSON string

	err := r.db.Get(ctx, redisKey).Scan(&averageRegionIncomesJSON)
	if err != nil {
		return nil, fmt.Errorf("error getting cached region incomes: %w", err)
	}

	err = json.Unmarshal([]byte(averageRegionIncomesJSON), averageRegionIncomes)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling cached region incomes: %w", err)
	}
	r.logger.Info("get cached region incomes", slog.String("region_id", fmt.Sprintf("%d", regionId)), slog.String("year", fmt.Sprintf("%d", year)), slog.String("quarter", fmt.Sprintf("%d", quarter)))
	return averageRegionIncomes, nil
}

func (r *RedisRepository) SetCachedRegionIncomes(
	ctx context.Context,
	averageRegionIncomes *domain.AverageRegionIncomes,
	regionId int32,
	year int32,
	quarter int32) error {

	redisKey := createCachedKey(regionId, year, quarter)

	averageRegionIncomesJSON, err := json.Marshal(averageRegionIncomes)
	if err != nil {
		return fmt.Errorf("error marshalling cached region incomes: %w", err)
	}

	err = r.db.Set(ctx, redisKey, averageRegionIncomesJSON, r.ttl).Err()
	if err != nil {
		return fmt.Errorf("error setting cached region incomes: %w", err)
	}
	r.logger.Info("set cached region incomes", slog.String("region_id", fmt.Sprintf("%d", regionId)), slog.String("year", fmt.Sprintf("%d", year)), slog.String("quarter", fmt.Sprintf("%d", quarter)))
	return nil
}

func createCachedKey(regionId int32, year int32, quarter int32) string {
	return fmt.Sprintf("region_incomes_%d_%d_%d", regionId, year, quarter)
}
