package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type ParserConfig struct {
	FilePath        string
	ParsingInterval time.Duration
	MaxRetries      int
	RetryDelay      time.Duration
	PGDSN           string
	ContainerName   string
	MainDir         string
	ContainerDir    string
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

	defaultFilePath := os.Getenv("DEFAULT_FILE_PATH")
	if defaultFilePath == "" {
		return nil, fmt.Errorf("DEFAULT_FILE_PATH environment variable not set")
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
		FilePath:        defaultFilePath,
		ParsingInterval: parsedInterval,
		MaxRetries:      maxRetries,
		RetryDelay:      time.Second,
		PGDSN:           pgDSN,
	}, nil
}

func NewParserConfig(filepath string, parsingInterval time.Duration, maxRetries int, retryDelay time.Duration) *ParserConfig {
	return &ParserConfig{
		FilePath:        filepath,
		ParsingInterval: parsingInterval,
		MaxRetries:      maxRetries,
		RetryDelay:      retryDelay,
	}
}
