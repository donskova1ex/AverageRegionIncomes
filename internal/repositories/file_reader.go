package repositories

import (
	"fmt"
	"github.com/donskova1ex/AverageRegionIncomes/internal/domain"
	"github.com/donskova1ex/AverageRegionIncomes/tools"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/xuri/excelize/v2"
)

type ExcelReader struct {
	logger     *slog.Logger
	maxRetries int
	retryDelay time.Duration
}

func NewExcelReader(logger *slog.Logger, maxRetries int, retryDelay time.Duration) *ExcelReader {
	return &ExcelReader{
		logger:     logger,
		maxRetries: maxRetries,
		retryDelay: retryDelay,
	}
}

func (r *ExcelReader) openFileWithRetry(filePath string) (*excelize.File, error) {
	var file *excelize.File
	var err error
	for attempt := 1; attempt <= r.maxRetries; attempt++ {
		file, err = excelize.OpenFile(filePath)
		if err == nil {
			return file, nil
		}
		r.logger.Error(
			"err",
			"failed to get file, retrying...",
			"attemp", fmt.Sprintf(": #%d of #%d", attempt, r.maxRetries),
			err.Error(),
		)
		time.Sleep(r.retryDelay)
	}
	return nil, err
}

func (r *ExcelReader) getRowsWithRetry(file *excelize.File) ([][]string, error) {
	var rows [][]string
	var err error

	for attempt := 1; attempt <= r.maxRetries; attempt++ {
		rows, err = tools.FormattingFileRows(file)
		if err == nil {
			return rows, nil
		}
		r.logger.Error(
			"err",
			"failed to get file, retrying...",
			"attemp", fmt.Sprintf(": #%d of #%d", attempt, r.maxRetries),
			err.Error(),
		)
		time.Sleep(r.retryDelay)
	}
	return nil, err
}

func (r *ExcelReader) processRows(rows [][]string) ([]*domain.ExcelRegionIncome, error) {
	var allRegionIncomes []*domain.ExcelRegionIncome
	var mu sync.Mutex
	var wg sync.WaitGroup

	wg.Add(len(rows) - 1)
	for i, row := range rows[1:] {
		go func(i int, row []string) {
			defer wg.Done()
			region := row[0]
			regionIncomes, err := r.convertRowToIncomes(rows[0][1:], row[1:], region)
			if err != nil {
				r.logger.Error("failed to convert row",
					"row", i+1,
					"error", err)
				return
			}
			mu.Lock()
			allRegionIncomes = append(allRegionIncomes, regionIncomes...)
			mu.Unlock()
		}(i, row)
	}
	wg.Wait()

	return allRegionIncomes, nil
}

func (r *ExcelReader) convertRowToIncomes(dataParts []string, valueParts []string, region string) ([]*domain.ExcelRegionIncome, error) {
	var regionIncomes []*domain.ExcelRegionIncome

	for index, value := range dataParts {
		parts := strings.Split(value, ".")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid date format: %s", value)
		}

		year, err := strconv.ParseInt(parts[0], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("failed to parse year: %w", err)
		}

		quarter, err := strconv.ParseInt(parts[1], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("failed to parse quarter: %w", err)
		}

		if index >= len(valueParts) {
			return nil, fmt.Errorf("missing value for date %s", value)
		}

		var strIncome string
		switch {
		case strings.Contains(valueParts[index], ","):
			strIncome = strings.ReplaceAll(valueParts[index], ",", "")
		case valueParts[index] == "":
			strIncome = "0"
		default:
			strIncome = valueParts[index]
		}

		income, err := strconv.ParseFloat(strIncome, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse income: %w, [%s]", err, region)
		}

		regionIncomes = append(regionIncomes, &domain.ExcelRegionIncome{
			Region:               region,
			Year:                 int32(year),
			Quarter:              int32(quarter),
			AverageRegionIncomes: float32(income),
		})
	}

	return regionIncomes, nil
}

func (r *ExcelReader) ReadFile(filepath string) ([]*domain.ExcelRegionIncome, error) {
	file, err := r.openFileWithRetry(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	sheets := file.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("no sheets found in file")
	}

	rows, err := r.getRowsWithRetry(file)
	if err != nil {
		return nil, fmt.Errorf("failed to get rows: %w", err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("file contains insufficient data")
	}

	return r.processRows(rows)
}
