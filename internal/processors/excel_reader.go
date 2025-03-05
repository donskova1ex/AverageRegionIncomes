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

type excelReader struct {
	ExcelReaderRepository ExcelReaderRepository
	Logger                ExcelReaderLogger
}

func NewExcelReader(repository ExcelReaderRepository, log ExcelReaderLogger) *excelReader {
	return &excelReader{
		ExcelReaderRepository: repository,
		Logger:                log,
	}
}

func (er *excelReader) CreateRegionIncomes(ctx context.Context, exRegionIncomes []*domain.ExcelRegionIncome) error {
	return er.ExcelReaderRepository.CreateRegionIncomes(ctx, exRegionIncomes)
}
