package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/donskova1ex/AverageRegionIncomes/internal/config"

	"github.com/donskova1ex/AverageRegionIncomes/internal"
	"github.com/donskova1ex/AverageRegionIncomes/internal/processors"
	"github.com/jmoiron/sqlx"

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

	cfg, err := config.DefaultParserConfig("/app/config/.env.dev")
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

	repository := repositories.NewSQLRepository(db, logger)

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

	logger.Info("First initialization started")
	firstInitialization(logger, cfg, ctx, repository)
	logger.Info("First initialization finished successfully")

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
			logger.Info("Download file started")
			downloadFile(cfg, logger)
			logger.Info("Download file finished")
			logger.Info("Parser started")
			processExcelFile(ctx, repository, logger, cfg)
			logger.Info("Parser finished")
		case <-ctx.Done():
			logger.Info("shutting down parser")
			return
		}
	}

}

func processExcelFile(
	ctx context.Context,
	repository *repositories.SQLRepository,
	logger *slog.Logger,
	readerCfg *config.ParserConfig,
) {

	reader := repositories.NewExcelReader(logger, readerCfg.MaxRetries, readerCfg.RetryDelay)

	fPath := filePathConstructor(readerCfg.ContainerDir, readerCfg.DefaultFileName)

	incomes, err := reader.ReadFile(fPath)
	if err != nil {
		logger.Error(
			"failed to read file",
			slog.String("err", err.Error()),
			slog.String("filepath", fPath),
		)
		os.Exit(1)
	}

	logger.Info("Successfully read file",
		"filepath", fPath,
		"records", len(incomes))

	eReaderProcessor := processors.NewExcelReader(repository, logger)
	if err := eReaderProcessor.ExcelReaderRepository.CreateRegionIncomes(ctx, incomes); err != nil {
		logger.Error("failed to create region incomes", slog.String("err", err.Error()))
	}

	logger.Info("Successfully saved records to database", "strings read", len(incomes))
}

func filePathConstructor(filePath, fileName string) string {
	fileExtension := filepath.Ext(fileName)
	name := fileName[0 : len(fileName)-len(fileExtension)]
	datedFileName := fmt.Sprintf("%s%s_%s%s",
		filePath,
		name,
		time.Now().Format("2006-01-02"), fileExtension)

	return datedFileName
}

func downloadFile(cfg *config.ParserConfig, logger *slog.Logger) {
	flag.Parse()

	jar, err := cookiejar.New(nil)
	if err != nil {
		logger.Error(
			"failed to create cookie jar",
			slog.String("err", err.Error()),
		)
	}
	logger.Info("Successfully created cookie jar")

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Jar:       jar,
		Transport: transport,
	}

	sslURL, err := url.Parse(cfg.SslCookieURL)
	if err != nil {
		logger.Error(
			"failed to parse ssl cookie url",
			slog.String("err", err.Error()),
		)

	}
	logger.Info("Successfully parsed ssl cookie url")

	resp, err := client.Get(sslURL.String())
	if err != nil {
		logger.Error(
			"error establishing session",
			slog.String("err", err.Error()),
		)
	}
	logger.Info("Successfully established session")

	if err := resp.Body.Close(); err != nil {
		logger.Error(
			"failed to close getting cookie response body",
			slog.String("err", err.Error()),
		)
	}
	logger.Info("Successfully closed response body")

	cookies := jar.Cookies(sslURL)
	logger.Info(fmt.Sprintf("Recived [%d] cookies from [%s]", len(cookies), sslURL.String()))

	fileURL := fmt.Sprintf("%s%s", cfg.FileStorageURL, cfg.DefaultFileName)
	resp, err = client.Get(fileURL)
	if err != nil {
		logger.Error("failed to download file", slog.String("err", err.Error()))
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Error(
				"failed to close download response body",
				slog.String("err", err.Error()),
			)
		}
	}(resp.Body)

	fileExt := filepath.Ext(cfg.DefaultFileName)
	fileName := (cfg.DefaultFileName)[:len(cfg.DefaultFileName)-len(fileExt)]
	datedFileName := fmt.Sprintf(
		"%s%s_%s%s",
		cfg.ContainerDir,
		fileName,
		time.Now().Format("2006-01-02"),
		fileExt,
	)

	file, err := os.Create(datedFileName)
	if err != nil {
		logger.Error("failed to create file", slog.String("err", err.Error()))
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			logger.Error(
				"failed to close file",
				slog.String("err", err.Error()),
			)
		}
	}(file)
	logger.Info("Successfully created file")

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		logger.Error(
			"failed writing file",
			slog.String("err", err.Error()),
		)
	}
	logger.Info("Successfully wrote file")
}

func firstInitialization(logger *slog.Logger, cfg *config.ParserConfig, ctx context.Context, repository *repositories.SQLRepository) {
	logger.Info("Starting download file")
	downloadFile(cfg, logger)
	logger.Info("Downloading complete")
	logger.Info("File parsing started")
	processExcelFile(ctx, repository, logger, cfg)
	logger.Info("File parsing finished")
}
