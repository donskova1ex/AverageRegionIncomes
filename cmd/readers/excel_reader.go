package main

import (
	"context"
	"fmt"
	"github.com/donskova1ex/AverageRegionIncomes/internal/processors"
	"github.com/joho/godotenv"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/donskova1ex/AverageRegionIncomes/internal/repositories"
)

// TODO: периодический скрипт по копированию файла в контейнер перед открытием
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logJSONHandler := slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.New(logJSONHandler)
	slog.SetDefault(logger)
	logger.Info(
		"Server started",
	)

	err := godotenv.Load(".env.dev")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	pgDSN := os.Getenv("POSTGRES_DSN")
	if pgDSN == "" {
		logger.Error("empty POSTGRES_DSN")
		os.Exit(1)
	}

	db, err := repositories.NewPostgresDB(ctx, pgDSN)
	if err != nil {
		logger.Error("error connecting to database", slog.String("err", err.Error()))
		return
	}
	defer db.Close()

	repository := repositories.NewRepository(db, logger)
	excelReaderProcessor := processors.NewExcelReader(repository, logger)
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
			processExcelFile(ctx, excelReaderProcessor, logger, excelReader, cfg.FilePath)
		case <-stop:
			logger.Info("shutting down server")
			return
		}
	}

}

func processExcelFile(
	ctx context.Context,
	p *processors.ExcelReader,
	logger *slog.Logger,
	reader *repositories.ExcelReader,
	filepath string,
) {
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

	if err := p.ExcelReaderRepository.CreateRegionIncomes(ctx, incomes); err != nil {
		logger.Error("failed to create region incomes", slog.String("err", err.Error()))
	}

	logger.Info("successfully saved records to database", "count", len(incomes))
}

func copyFilesToContainer(containerName string, mainDir string, containerDir string) error {

	checkCmd := exec.Command("docker", "ps", "-a", "--filter", fmt.Sprintf("name=%s", containerName), "--format", "{{.Names}}")

	out, err := checkCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check check container existence. Error: %w", err)
	}

	if len(out) == 0 || string(out) != containerName+"\n" {
		return fmt.Errorf("container %s does not exists", containerName)
	}

	files, err := filepath.Glob(filepath.Join(mainDir, "*"))
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("directory %s is empty", mainDir)
	}

	for _, file := range files {
		copyCmd := exec.Command("docker", "cp", file, fmt.Sprintf("%s:%s", containerName, containerDir))
		output, err := copyCmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to copy [%s]: [%s]. Error: %w", file, string(output), err)
		}
	}

	return nil
}
