package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/donskova1ex/AverageRegionIncomes/internal/domain"
	"github.com/jmoiron/sqlx"
	"log/slog"
	"strings"
)

func (r *SQLRepository) CreateRegionIncomes(ctx context.Context, exRegionIncomes []*domain.ExcelRegionIncome) error {
	var txCommited bool
	//TODO: serializable выставить на всю таблицу, а все остальное ReadCommited, на уровне БД
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
	query := `SELECT 
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
				GROUP BY r.region_name, ri.region_id`

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
	return nil
}

func createCachedKey(regionId int32, year int32, quarter int32) string {
	return fmt.Sprintf("region_incomes_%d_%d_%d", regionId, year, quarter)
}
