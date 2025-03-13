package main

import (
	"context"
	"testing"
	"time"

	"github.com/donskova1ex/AverageRegionIncomes/internal/config"
	"github.com/donskova1ex/AverageRegionIncomes/internal/repositories"
	"github.com/donskova1ex/AverageRegionIncomes/internal/processors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateRegionIncomes(ctx context.Context, incomes []repositories.Income) error {
	args := m.Called(ctx, incomes)
	return args.Error(0)
}

type MockExcelReaderRepository struct {
	mock.Mock
}

func (m *MockExcelReaderRepository) CreateRegionIncomes(ctx context.Context, incomes []repositories.Income) error {
	args := m.Called(ctx, incomes)
	return args.Error(0)
}

func TestProcessExcelFile(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	readerCfg := &config.ParserConfig{
		FilePath:     "/path/to/file",
		MaxRetries:   3,
		RetryDelay:   time.Second,
		ParsingInterval: time.Minute,
	}

	mockRepository := new(MockRepository)
	mockExcelReaderRepository := new(MockExcelReaderRepository)

	mockRepository.On("CreateRegionIncomes", ctx, mock.Anything).Return(nil)
	mockExcelReaderRepository.On("CreateRegionIncomes", ctx, mock.Anything).Return(nil)

	repository := &repositories.Repository{
		ExcelReaderRepository: mockExcelReaderRepository,
	}

	processExcelFile(ctx, repository, logger, readerCfg)

	mockRepository.AssertExpectations(t)
	mockExcelReaderRepository.AssertExpectations(t)
}

