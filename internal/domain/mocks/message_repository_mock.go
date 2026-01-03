package mocks

import (
	"github.com/myestatia/myestatia-go/internal/domain/entity"
	"github.com/stretchr/testify/mock"
)

type MessageRepositoryMock struct {
	mock.Mock
}

func (m *MessageRepositoryMock) Create(msg *entity.Message) error {
	args := m.Called(msg)
	return args.Error(0)
}

func (m *MessageRepositoryMock) FindByLeadID(leadID string) ([]entity.Message, error) {
	args := m.Called(leadID)
	return args.Get(0).([]entity.Message), args.Error(1)
}
