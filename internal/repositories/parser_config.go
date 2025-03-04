package repositories

import (
	"time"
)

type ParserConfig struct {
	FilePath        string
	ParsingInterval time.Duration
	MaxRetries      int
	RetryDelay      time.Duration
}

func DefaultParserConfig() *ParserConfig {
	return &ParserConfig{
		FilePath:        "/db-files/AverageIncomes.xlsx",
		ParsingInterval: 5 * time.Hour,
		MaxRetries:      3,
		RetryDelay:      time.Second,
	}
}
func NewParserConfig(filepath string, parsingInterval time.Duration, maxRetries int, retryDelay time.Duration) *ParserConfig {
	return &ParserConfig{
		FilePath:        filepath,
		ParsingInterval: parsingInterval,
		MaxRetries:      maxRetries,
		RetryDelay:      retryDelay,
	}
}
