package repositories

//TODO: перенести в отдельную папочку config
import (
	"fmt"
	"os"
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

	return &ParserConfig{
		FilePath:        defaultFilePath,
		ParsingInterval: 5 * time.Hour,
		MaxRetries:      3,
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
