package processors

import (
	"context"
	"fmt"
	"github.com/donskova1ex/AverageRegionIncomes/internal/domain"
	"log/slog"
)

//go:generate mockgen -destination=./mocks/average_income_repository.go -package=mocks -mock_names=AverageIncomeRepository=AverageIncomeRepository . AverageIncomeRepository
type AverageIncomeRepository interface {
	GetRegionIncomes(ctx context.Context, regionId int32, year int32, quarter int32) (*domain.AverageRegionIncomes, error)
}

//go:generate mockgen -destination=./mocks/average_income_logger.go -package=mocks -mock_names=AverageIncomeLogger=AverageIncomeLogger . AverageIncomeLogger
type AverageIncomeLogger interface {
	Error(msg string, args ...any)
	Info(msg string, args ...any)
}

type averageIncome struct {
	averageIncomeRepository AverageIncomeRepository
	log                     AverageIncomeLogger
}

func NewAverageIncome(averageIncomeRepository AverageIncomeRepository, log AverageIncomeLogger) *averageIncome {
	return &averageIncome{averageIncomeRepository, log}
}

func (a *averageIncome) GetRegionIncomes(ctx context.Context, regionId int32, year int32, quarter int32) (*domain.AverageRegionIncomes, error) {
	regionIncomes, err := a.averageIncomeRepository.GetRegionIncomes(ctx, regionId, year, quarter)
	if err != nil {
		a.log.Error("it is impossible to get a region incomes", slog.String("err", err.Error()))
		return nil, fmt.Errorf("it is impossible to get a region incomes, err: %s", err.Error())
	}
	return regionIncomes, nil
}
