package mocks

import (
	"context"

	"github.com/myestatia/myestatia-go/internal/domain/entity"
	"github.com/stretchr/testify/mock"
)

type CompanyRepositoryMock struct {
	mock.Mock
}

func (m *CompanyRepositoryMock) Create(company *entity.Company) error {
	args := m.Called(company)
	return args.Error(0)
}

func (m *CompanyRepositoryMock) FindByID(id string) (*entity.Company, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Company), args.Error(1)
}

func (m *CompanyRepositoryMock) FindAll() ([]entity.Company, error) {
	args := m.Called()
	return args.Get(0).([]entity.Company), args.Error(1)
}

func (m *CompanyRepositoryMock) UpdatePartial(ctx context.Context, id string, fields map[string]interface{}) error {
	args := m.Called(ctx, id, fields)
	return args.Error(0)
}

func (m *CompanyRepositoryMock) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *CompanyRepositoryMock) FindByName(ctx context.Context, name string) (*entity.Company, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Company), args.Error(1)
}
