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

type ExcelReaderTestSuite struct {
	suite.Suite
	ctrl       *gomock.Controller
	processor  *excelReader
	repository *mocks.ExcelReaderRepository
	logger     *mocks.ExcelReaderLogger
	ctx        context.Context
}

func (s *ExcelReaderTestSuite) SetupSuite() {
	s.ctrl = gomock.NewController(s.T())
	s.logger = mocks.NewExcelReaderLogger(s.ctrl)
	s.repository = mocks.NewExcelReaderRepository(s.ctrl)
	s.processor = NewExcelReader(s.repository, s.logger)
	s.ctx = context.Background()
}

func (s *ExcelReaderTestSuite) TestCreateRegionIncomeError() {
	dbError := errors.New("db error")
	expectedError := fmt.Errorf("error creating region incomes: %w", dbError)

	exRegionIncomes := []*domain.ExcelRegionIncome{
		{
			Region:               "test1",
			Year:                 1990,
			Quarter:              1,
			AverageRegionIncomes: 10,
		},
	}

	gomock.InOrder(
		s.repository.
			EXPECT().
			CreateRegionIncomes(gomock.Any(), gomock.Any()).
			Return(dbError),
		s.logger.
			EXPECT().
			Error(gomock.Any(), gomock.Any()),
	)
	err := s.processor.CreateRegionIncomes(s.ctx, exRegionIncomes)
	require.EqualError(s.T(), err, expectedError.Error())
}

func TestExcelReaderTestSuite(t *testing.T) {
	suite.Run(t, new(ExcelReaderTestSuite))
}
