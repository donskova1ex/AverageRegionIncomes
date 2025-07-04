// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/donskova1ex/AverageRegionIncomes/internal/processors (interfaces: AverageIncomeRedisRepository)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	domain "github.com/donskova1ex/AverageRegionIncomes/internal/domain"
	gomock "github.com/golang/mock/gomock"
)

// AverageIncomeRedisRepository is a mock of AverageIncomeRedisRepository interface.
type AverageIncomeRedisRepository struct {
	ctrl     *gomock.Controller
	recorder *AverageIncomeRedisRepositoryMockRecorder
}

// AverageIncomeRedisRepositoryMockRecorder is the mock recorder for AverageIncomeRedisRepository.
type AverageIncomeRedisRepositoryMockRecorder struct {
	mock *AverageIncomeRedisRepository
}

// NewAverageIncomeRedisRepository creates a new mock instance.
func NewAverageIncomeRedisRepository(ctrl *gomock.Controller) *AverageIncomeRedisRepository {
	mock := &AverageIncomeRedisRepository{ctrl: ctrl}
	mock.recorder = &AverageIncomeRedisRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *AverageIncomeRedisRepository) EXPECT() *AverageIncomeRedisRepositoryMockRecorder {
	return m.recorder
}

// GetCachedRegionIncomes mocks base method.
func (m *AverageIncomeRedisRepository) GetCachedRegionIncomes(arg0 context.Context, arg1, arg2, arg3 int32) (*domain.AverageRegionIncomes, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCachedRegionIncomes", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(*domain.AverageRegionIncomes)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCachedRegionIncomes indicates an expected call of GetCachedRegionIncomes.
func (mr *AverageIncomeRedisRepositoryMockRecorder) GetCachedRegionIncomes(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCachedRegionIncomes", reflect.TypeOf((*AverageIncomeRedisRepository)(nil).GetCachedRegionIncomes), arg0, arg1, arg2, arg3)
}

// SetCachedRegionIncomes mocks base method.
func (m *AverageIncomeRedisRepository) SetCachedRegionIncomes(arg0 context.Context, arg1 *domain.AverageRegionIncomes, arg2, arg3, arg4 int32) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetCachedRegionIncomes", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetCachedRegionIncomes indicates an expected call of SetCachedRegionIncomes.
func (mr *AverageIncomeRedisRepositoryMockRecorder) SetCachedRegionIncomes(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetCachedRegionIncomes", reflect.TypeOf((*AverageIncomeRedisRepository)(nil).SetCachedRegionIncomes), arg0, arg1, arg2, arg3, arg4)
}
