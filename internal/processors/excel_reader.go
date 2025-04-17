package processors

import (
	"context"
	"github.com/donskova1ex/AverageRegionIncomes/internal/domain"
)

// TODO:Tests
//
//go:generate mockgen -destination=./mocks/excel_reader_repository.go -package=mocks -mock_names=ExcelReaderRepository=ExcelReaderRepository . ExcelReaderRepository
type ExcelReaderRepository interface {
	CreateRegionIncomes(ctx context.Context, exRegionIncomes []*domain.ExcelRegionIncome) error
}

//go:generate mockgen -destination=./mocks/excel_reader_logger.go -package=mocks -mock_names=ExcelReaderLogger=ExcelReaderLogger . ExcelReaderLogger
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
