package mocks

import (
	"context"

	"github.com/myestatia/myestatia-go/internal/domain/entity"
	"github.com/stretchr/testify/mock"
)

type LeadRepositoryMock struct {
	mock.Mock
}

func (m *LeadRepositoryMock) Create(lead *entity.Lead) error {
	args := m.Called(lead)
	return args.Error(0)
}

func (m *LeadRepositoryMock) FindByID(id string) (*entity.Lead, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Lead), args.Error(1)
}

func (m *LeadRepositoryMock) FindAll() ([]entity.Lead, error) {
	args := m.Called()
	return args.Get(0).([]entity.Lead), args.Error(1)
}

func (m *LeadRepositoryMock) Update(lead *entity.Lead) error {
	args := m.Called(lead)
	return args.Error(0)
}

func (m *LeadRepositoryMock) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *LeadRepositoryMock) FindByEmail(ctx context.Context, email string) (*entity.Lead, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Lead), args.Error(1)
}

func (m *LeadRepositoryMock) FindByCompanyId(ctx context.Context, companyId string) ([]entity.Lead, error) {
	args := m.Called(ctx, companyId)
	return args.Get(0).([]entity.Lead), args.Error(1)
}

func (m *LeadRepositoryMock) FindByPropertyId(ctx context.Context, propertyId string) ([]entity.Lead, error) {
	args := m.Called(ctx, propertyId)
	return args.Get(0).([]entity.Lead), args.Error(1)
}
