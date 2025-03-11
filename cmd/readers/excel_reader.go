package main

import (
	"context"
	"github.com/donskova1ex/AverageRegionIncomes/internal"
	"github.com/donskova1ex/AverageRegionIncomes/internal/processors"
	"github.com/jmoiron/sqlx"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/donskova1ex/AverageRegionIncomes/internal/repositories"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logJSONHandler := slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.New(logJSONHandler)
	slog.SetDefault(logger)
	logger.Info(
		"Server started",
	)

	cfg, err := repositories.DefaultParserConfig("/app/.env.dev")
	if err != nil {
		logger.Error("failed to load configuration", slog.String("err", err.Error()))
		os.Exit(1)
	}
	logger.Info("Configuration loaded")

	db, err := repositories.NewPostgresDB(ctx, cfg.PGDSN)
	if err != nil {
		logger.Error("error connecting to database", slog.String("err", err.Error()))
		return
	}
	defer func(db *sqlx.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("error closing db", slog.String("err", err.Error()))
		}
		logger.Info("db closed")
	}(db)

	repository := repositories.NewRepository(db, logger)

	gracefulCloser := internal.NewGracefulCloser()
	gracefulCloser.Add(func() error {
		logger.Info("closing db connection")
		if err := db.Close(); err != nil {
			logger.Error("error closing db connection", slog.String("err", err.Error()))
			return err
		}
		logger.Info("db connection closed")
		return nil
	})

	signalCtx, signalCancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer signalCancel()

	processExcelFile(ctx, repository, logger, cfg)

	ticker := time.NewTicker(cfg.ParsingInterval)
	defer ticker.Stop()

	logger.Info("Parser is running in the background")

	go func() {
		<-signalCtx.Done()
		logger.Info("interrupt signal received")
		cancel()

		gracefulCloser.Run(signalCtx, logger)
		logger.Info("graceful shutdown complete")
	}()

	for {
		select {
		case <-ticker.C:
			logger.Info("Parser started")
			processExcelFile(ctx, repository, logger, cfg)
		case <-ctx.Done():
			logger.Info("shutting down parser")
			return
		}
	}

}

func processExcelFile(
	ctx context.Context,
	repository *repositories.Repository,
	logger *slog.Logger,
	readerCfg *repositories.ParserConfig,
) {

	reader := repositories.NewExcelReader(logger, readerCfg.MaxRetries, readerCfg.RetryDelay)

	incomes, err := reader.ReadFile(readerCfg.FilePath)
	if err != nil {
		logger.Error(
			"failed to read file",
			slog.String("err", err.Error()),
			slog.String("filepath", readerCfg.FilePath),
		)
		os.Exit(1)
	}

	logger.Info("successfully read file",
		"filepath", readerCfg.FilePath,
		"records", len(incomes))

	eReaderProcessor := processors.NewExcelReader(repository, logger)
	if err := eReaderProcessor.ExcelReaderRepository.CreateRegionIncomes(ctx, incomes); err != nil {
		logger.Error("failed to create region incomes", slog.String("err", err.Error()))
	}

	logger.Info("successfully saved records to database", "strings read", len(incomes))
}
