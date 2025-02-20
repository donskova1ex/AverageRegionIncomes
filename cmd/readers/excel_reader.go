package main

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/xuri/excelize/v2"
)

type RegionIncomes struct {
	Region               string
	Year                 int32
	Quarter              int32
	AverageRegionIncomes float32
}

func main() {

	log.Printf("Server started")
	logJSONHandler := slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.New(logJSONHandler)
	slog.SetDefault(logger)
	filepath := `/mnt/network_share/AverageIncomes.xlsx`

	timoutInterval := time.Second * 5
	ticker := time.NewTicker(timoutInterval)
	defer ticker.Stop()

	for range ticker.C {
		readingDbFile(logger, filepath)
		logger.Info(
			"file is reading from Excel",
			slog.String("file", filepath),
		)
	}

}

func readingDbFile(logger *slog.Logger, filepath string) {
	const maxRetries = 3

	file, err := excelize.OpenFile(filepath)
	if err != nil {
		logger.Error(
			"err",
			"failed to open file",
			err.Error(),
		)
		os.Exit(1)
	}

	defer func(file *excelize.File) {
		if file != nil {
			err := file.Close()
			if err != nil {
				logger.Error(
					"err",
					"failed to close file",
					err.Error(),
				)
			}
		}
	}(file)

	sheets := file.GetSheetList()
	if len(sheets) == 0 {
		logger.Error(
			"err",
			"failed to get file sheets",
			errors.New("no file sheets found"),
		)
		os.Exit(1)
	}

	var rows [][]string
	for attempt := 1; attempt <= maxRetries; attempt++ {
		if rows, err = file.GetRows(sheets[0]); err == nil {
			break
		}
		logger.Error(
			"err",
			"failed to get rows, retrying...",
			"attemp", fmt.Sprintf(": #%d of #%d", attempt, maxRetries),
			err.Error(),
		)
		time.Sleep(time.Second)
	}
	if err != nil {
		logger.Error(
			"err",
			"failed to get table rows",
			err.Error(),
		)
		os.Exit(1)
	}

	var allRegionIncomes []*RegionIncomes

	var mu = &sync.Mutex{}
	var wg = &sync.WaitGroup{}
	wg.Add(len(rows[1:]))
	for i, row := range rows[1:] {
		go func(i int, row []string) {
			defer wg.Done()
			region := row[0]
			regionIncomes, err := convertingStringsToStruct(rows[0][1:], rows[i+1][1:], region)
			if err != nil {
				logger.Error(
					"err",
					"failed to convert strings to domain struct",
					err.Error(),
				)
				return
			}
			mu.Lock()
			allRegionIncomes = append(allRegionIncomes, regionIncomes...)
			mu.Unlock()
		}(i, row)
	}
	wg.Wait()

	for _, regionIncome := range allRegionIncomes {
		fmt.Println(regionIncome.Region, regionIncome.Year, regionIncome.Quarter, regionIncome.AverageRegionIncomes)
	}
}

func convertingStringsToStruct(dataParts []string, valueParts []string, region string) ([]*RegionIncomes, error) {
	var regionIncomes []*RegionIncomes

	for index, value := range dataParts {
		parts := strings.Split(value, ".")

		year, err := strconv.ParseInt(parts[0], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("failed to convert year to int: %w", err)
		}

		quarter, err := strconv.ParseInt(parts[1], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("failed to convert quarter to int: %w", err)
		}

		income, err := strconv.ParseFloat(valueParts[index], 32)
		if err != nil {
			return nil, fmt.Errorf("failed to convert income to float: %w", err)
		}

		regionIncomes = append(regionIncomes, &RegionIncomes{
			Region:               region,
			Year:                 int32(year),
			Quarter:              int32(quarter),
			AverageRegionIncomes: float32(income),
		})
	}

	return regionIncomes, nil
}
