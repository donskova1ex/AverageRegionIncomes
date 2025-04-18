package processors

import (
	"context"
	"errors"
	"fmt"
	"github.com/donskova1ex/AverageRegionIncomes/internal/domain"
	"github.com/donskova1ex/AverageRegionIncomes/internal/processors/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

type AverageIncomeTestSuite struct {
	suite.Suite
	ctrl       *gomock.Controller
	processor  *averageIncome
	repository *mocks.AverageIncomeRepository
	logger     *mocks.AverageIncomeLogger
	ctx        context.Context
}

func (s *AverageIncomeTestSuite) SetupSuite() {
	s.ctrl = gomock.NewController(s.T())
	s.logger = mocks.NewAverageIncomeLogger(s.ctrl)
	s.repository = mocks.NewAverageIncomeRepository(s.ctrl)
	s.processor = NewAverageIncome(s.repository, s.logger)
	s.ctx = context.Background()
}

func (s *AverageIncomeTestSuite) TestGetRegionIncomesSuccess() {
	expectedRegionIncomes := &domain.AverageRegionIncomes{
		RegionId:             1,
		RegionName:           "test1",
		Year:                 2024,
		Quarter:              1,
		AverageRegionIncomes: 1,
	}

	gomock.InOrder(
		s.repository.
			EXPECT().
			GetRegionIncomes(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(expectedRegionIncomes, nil),
	)
	actualRegionIncomes, err := s.processor.GetRegionIncomes(s.ctx,
		expectedRegionIncomes.RegionId,
		0,
		0)

	require.NoError(s.T(), err)
	require.Equal(s.T(), expectedRegionIncomes, actualRegionIncomes)
}

func (s *AverageIncomeTestSuite) TestGetRegionIncomesError() {
	dbError := errors.New("db error")
	expectedRegionIncomes := &domain.AverageRegionIncomes{
		RegionId:   1,
		RegionName: "test1",
		Year:       2024,
		Quarter:    1,
	}
	expectedError := fmt.Errorf("it is impossible to get a region incomes, err: %s", dbError.Error())

	gomock.InOrder(
		s.repository.
			EXPECT().
			GetRegionIncomes(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, dbError),
		s.logger.
			EXPECT().
			Error(gomock.Any(), gomock.Any()),
	)

	actualRegionIncomes, err := s.processor.GetRegionIncomes(s.ctx,
		expectedRegionIncomes.RegionId,
		0,
		0)
	require.Nil(s.T(), actualRegionIncomes)
	require.EqualError(s.T(), err, expectedError.Error())
}

func TestAverageIncomeTestSuite(t *testing.T) {
	suite.Run(t, new(AverageIncomeTestSuite))
}
