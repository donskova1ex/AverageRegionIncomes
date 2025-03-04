package processors

import (
	"context"
	"github.com/donskova1ex/AverageRegionIncomes/internal/domain"
)

type ExcelReaderRepository interface {
	CreateRegionIncomes(ctx context.Context, exRegionIncomes []*domain.ExcelRegionIncome) error
}

type ExcelReaderLogger interface {
	Error(msg string, args ...any)
	Info(msg string, args ...any)
}

type ExcelReader struct {
	ExcelReaderRepository ExcelReaderRepository
	Logger                ExcelReaderLogger
}

func NewExcelReader(repository ExcelReaderRepository, log ExcelReaderLogger) *ExcelReader {
	return &ExcelReader{
		ExcelReaderRepository: repository,
		Logger:                log,
	}
}
