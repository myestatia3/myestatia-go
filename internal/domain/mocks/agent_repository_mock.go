package mocks

import (
	"context"

	"github.com/myestatia/myestatia-go/internal/domain/entity"
	"github.com/stretchr/testify/mock"
)

type AgentRepositoryMock struct {
	mock.Mock
}

func (m *AgentRepositoryMock) Create(agent *entity.Agent) error {
	args := m.Called(agent)
	return args.Error(0)
}

func (m *AgentRepositoryMock) FindByID(id string) (*entity.Agent, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Agent), args.Error(1)
}

func (m *AgentRepositoryMock) FindAll() ([]entity.Agent, error) {
	args := m.Called()
	return args.Get(0).([]entity.Agent), args.Error(1)
}

func (m *AgentRepositoryMock) UpdatePartial(ctx context.Context, id string, fields map[string]interface{}) error {
	args := m.Called(ctx, id, fields)
	return args.Error(0)
}

func (m *AgentRepositoryMock) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *AgentRepositoryMock) FindByEmail(ctx context.Context, email string) (*entity.Agent, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Agent), args.Error(1)
}
