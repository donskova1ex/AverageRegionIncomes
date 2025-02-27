package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/donskova1ex/AverageRegionIncomes/internal/repositories"
)

type RegionIncomes struct {
	Region               string
	Year                 int32
	Quarter              int32
	AverageRegionIncomes float32
}

// TODO: периодический скрипт по копированию файла в контейнер перед открытием
func main() {
	//ctx, cancel := context.WithCancel(context.Background())
	//defer cancel()

	logJSONHandler := slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.New(logJSONHandler)
	slog.SetDefault(logger)
	logger.Info(
		"Server started",
	)

	cfg := repositories.DefaultParserConfig()

	excelReader := repositories.NewExcelReader(logger, cfg.MaxRetries, cfg.RetryDelay)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(cfg.ParsingInterval)
	defer ticker.Stop()

	logger.Info("File parsing started")

	for {
		select {
		case <-ticker.C:
			processExcelFile(logger, excelReader, cfg.FilePath)
		case <-stop:
			logger.Info("shutting down server")
			return
		}
	}

}

func processExcelFile(logger *slog.Logger, reader *repositories.ExcelReader, filepath string) {
	incomes, err := reader.ReadFile(filepath)
	if err != nil {
		logger.Error(
			"failed to read file",
			slog.String("err", err.Error()),
			slog.String("filepath", filepath),
		)
		os.Exit(1)
	}

	logger.Info("successfully read file",
		"filepath", filepath,
		"records", len(incomes))

	for _, regionIncome := range incomes {
		fmt.Println(regionIncome.Region, regionIncome.Year, regionIncome.Quarter, regionIncome.AverageRegionIncomes)
	}
	logger.Info("successfully saved records to database", "count", len(incomes))
}

// func copyFilesToContainer(containerName string, mainDir string, containerDir string) error {
// 	cmd := exec.Command("docker", "cp", mainDir, fmt.Sprintf("%s:%s", containerName, containerDir))

// 	_, err := cmd.CombinedOutput()
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }
