package processors

import (
	"context"
	"errors"
	"fmt"
	"github.com/donskova1ex/AverageRegionIncomes/internal/domain"
	"github.com/redis/go-redis/v9"
	"log/slog"
)

//go:generate mockgen -destination=./mocks/average_income_db_repository.go -package=mocks -mock_names=AverageIncomeDBRepository=AverageIncomeDBRepository . AverageIncomeDBRepository
type AverageIncomeDBRepository interface {
	GetRegionIncomes(ctx context.Context, regionId int32, year int32, quarter int32) (*domain.AverageRegionIncomes, error)
}

//go:generate mockgen -destination=./mocks/average_income_redis_repository.go -package=mocks -mock_names=AverageIncomeRedisRepository=AverageIncomeRedisRepository . AverageIncomeRedisRepository
type AverageIncomeRedisRepository interface {
	GetCachedRegionIncomes(ctx context.Context, regionId int32, year int32, quarter int32) (*domain.AverageRegionIncomes, error)
	SetCachedRegionIncomes(ctx context.Context, averageRegionIncomes *domain.AverageRegionIncomes, regionId int32, year int32, quarter int32) error
}

//go:generate mockgen -destination=./mocks/average_income_logger.go -package=mocks -mock_names=AverageIncomeLogger=AverageIncomeLogger . AverageIncomeLogger
type AverageIncomeLogger interface {
	Error(msg string, args ...any)
	Info(msg string, args ...any)
}

type averageIncome struct {
	averageIncomeRepository      AverageIncomeDBRepository
	averageIncomeRedisRepository AverageIncomeRedisRepository
	logger                       AverageIncomeLogger
}

func NewAverageIncome(averageIncomeRepository AverageIncomeDBRepository, averageIncomeRedisRepository AverageIncomeRedisRepository, log AverageIncomeLogger) *averageIncome {
	return &averageIncome{averageIncomeRepository, averageIncomeRedisRepository, log}
}

func (a *averageIncome) GetRegionIncomes(ctx context.Context, regionId int32, year int32, quarter int32) (*domain.AverageRegionIncomes, error) {
	cachedRegionIncomes, err := a.averageIncomeRedisRepository.GetCachedRegionIncomes(ctx, regionId, year, quarter)
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("it is impossible to get a cached region incomes: %w", err)
	}

	if cachedRegionIncomes == nil {
		regionIncomes, err := a.averageIncomeRepository.GetRegionIncomes(ctx, regionId, year, quarter)
		if err != nil {
			a.logger.Error("it is impossible to get a region incomes", slog.String("err", err.Error()))
			return nil, fmt.Errorf("it is impossible to get a region incomes, err: %s", err.Error())
		}
		err = a.averageIncomeRedisRepository.SetCachedRegionIncomes(ctx, regionIncomes, regionId, year, quarter)
		if err != nil {
			a.logger.Error("it is impossible to set cached region incomes", slog.String("err", err.Error()))
			return nil, fmt.Errorf("it is impossible to set cached region incomes: %w", err)
		}
		return regionIncomes, nil
	}
	return cachedRegionIncomes, nil

}
