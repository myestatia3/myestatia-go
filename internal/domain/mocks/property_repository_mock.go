package mocks

import (
	"context"

	"github.com/myestatia/myestatia-go/internal/domain/entity"
	"github.com/stretchr/testify/mock"
)

type PropertyRepositoryMock struct {
	mock.Mock
}

func (m *PropertyRepositoryMock) Create(property *entity.Property) error {
	args := m.Called(property)
	return args.Error(0)
}

func (m *PropertyRepositoryMock) FindByID(id string) (*entity.Property, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Property), args.Error(1)
}

func (m *PropertyRepositoryMock) FindAll() ([]entity.Property, error) {
	args := m.Called()
	return args.Get(0).([]entity.Property), args.Error(1)
}

func (m *PropertyRepositoryMock) Update(property *entity.Property) error {
	args := m.Called(property)
	return args.Error(0)
}

func (m *PropertyRepositoryMock) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *PropertyRepositoryMock) FindByReference(ctx context.Context, ref string) (*entity.Property, error) {
	args := m.Called(ctx, ref)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Property), args.Error(1)
}

func (m *PropertyRepositoryMock) FindAllByCompanyID(ctx context.Context, companyID string) ([]entity.Property, error) {
	args := m.Called(ctx, companyID)
	return args.Get(0).([]entity.Property), args.Error(1)
}

func (m *PropertyRepositoryMock) Search(ctx context.Context, filter entity.PropertyFilter) ([]entity.Property, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]entity.Property), args.Error(1)
}
