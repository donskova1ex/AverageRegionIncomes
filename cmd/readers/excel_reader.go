package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"strconv"
	"strings"

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

	filepath := "C:/Users/Admin/Desktop/testingAverageIncome.xlsx"
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
		err := file.Close()
		if err != nil {
			logger.Error(
				"err",
				"failed to close file",
				err.Error(),
			)
		}
	}(file)

	sheets := file.GetSheetList()
	rows, err := file.GetRows(sheets[0])

	if err != nil {
		logger.Error(
			"err",
			"failed to get table rows",
			err.Error(),
		)
		//TODO:Retry reading rows
	}

	for i, row := range rows[1:] {
		region := row[0]
		regionIncomes, err := convertingStringsToStruct(rows[0][1:], rows[i+1][1:], region)
		if err != nil {
			logger.Error(
				"err",
				"failed to converting strings to domain struct",
				err.Error(),
			)
		}

		for _, regionIncome := range regionIncomes {
			fmt.Println(regionIncome.Region, regionIncome.Year, regionIncome.Quarter, regionIncome.AverageRegionIncomes)
		}
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
