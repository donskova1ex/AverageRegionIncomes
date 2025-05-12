package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type ParserConfig struct {
	DefaultFileName string
	ParsingInterval time.Duration
	MaxRetries      int
	RetryDelay      time.Duration
	PGDSN           string
	ContainerName   string
	MainDir         string
	ContainerDir    string
	SslCookieURL    string
	FileStorageURL  string
}

func DefaultParserConfig(envPath string) (*ParserConfig, error) {
	err := godotenv.Load(envPath)
	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	pgDSN := os.Getenv("POSTGRES_DSN")
	if pgDSN == "" {
		return nil, fmt.Errorf("POSTGRES_DSN environment variable not set")
	}

	containerName := os.Getenv("READER_NAME")
	if containerName == "" {
		return nil, fmt.Errorf("READER_NAME environment variable not set")
	}

	mainReaderDir := os.Getenv("READER_MAIN_DIR")
	if mainReaderDir == "" {
		return nil, fmt.Errorf("READER_MAIN_DIR environment variable not set")
	}

	readerContainerDir := os.Getenv("READER_CONTAINER_DIR")
	if readerContainerDir == "" {
		return nil, fmt.Errorf("READER_CONTAINER_DIR environment variable not set")
	}

	defaultFileName := os.Getenv("DEFAULT_FILE_NAME")
	if defaultFileName == "" {
		return nil, fmt.Errorf("DEFAULT_FILE_Name environment variable not set")
	}

	parsingInterval := os.Getenv("PARSING_INTERVAL")
	var parsedInterval time.Duration
	if parsingInterval == "" {
		parsedInterval = 5 * time.Hour
	}
	parsedInterval, err = time.ParseDuration(parsingInterval)
	if err != nil {
		return nil, fmt.Errorf("error parsing PARSING_INTERVAL: %w", err)
	}

	sslCookieURL := os.Getenv("SSL_COOKIE_URL")
	if sslCookieURL == "" {
		return nil, fmt.Errorf("SSL_COOKIE_URL environment variable not set")
	}

	fileUrlStorage := os.Getenv("FILE_URL_STORAGE")

	maxRetriesStr := os.Getenv("MAX_RETRIES")
	var maxRetries int
	if maxRetriesStr == "" {
		maxRetries = 3
	}
	maxRetries, err = strconv.Atoi(maxRetriesStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing MAX_RETRIES: %w", err)
	}

	return &ParserConfig{
		DefaultFileName: defaultFileName,
		ParsingInterval: parsedInterval,
		MaxRetries:      maxRetries,
		RetryDelay:      time.Second,
		PGDSN:           pgDSN,
		ContainerDir:    readerContainerDir,
		SslCookieURL:    sslCookieURL,
		FileStorageURL:  fileUrlStorage,
	}, nil
}

func NewParserConfig(fileName string, parsingInterval time.Duration, maxRetries int, retryDelay time.Duration) *ParserConfig {
	return &ParserConfig{
		DefaultFileName: fileName,
		ParsingInterval: parsingInterval,
		MaxRetries:      maxRetries,
		RetryDelay:      retryDelay,
	}
}
