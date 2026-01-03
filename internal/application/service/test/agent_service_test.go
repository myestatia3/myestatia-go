package test

import (
	"context"
	"testing"

	"github.com/myestatia/myestatia-go/internal/application/service"
	"github.com/myestatia/myestatia-go/internal/domain/entity"
	"github.com/myestatia/myestatia-go/internal/domain/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateAgent(t *testing.T) {
	// GIVEN
	mockAgentRepo := new(mocks.AgentRepositoryMock)
	mockPropRepo := new(mocks.PropertyRepositoryMock)
	svc := service.NewAgentService(mockAgentRepo, mockPropRepo)
	ctx := context.TODO()

	agent := &entity.Agent{
		Email: "agent@test.com",
		Name:  "Test Agent",
	}

	// Mocking behavior
	// 1. FindByEmail returns nil (agent does not exist)
	mockAgentRepo.On("FindByEmail", ctx, "agent@test.com").Return(nil, nil)
	// 2. Create is called and returns no error
	mockAgentRepo.On("Create", mock.MatchedBy(func(a *entity.Agent) bool {
		return a.Email == "agent@test.com" && a.ID != ""
	})).Return(nil)

	// WHEN
	createdAgent, created, err := svc.Create(ctx, agent)

	// THEN
	assert.NoError(t, err)
	assert.True(t, created)
	assert.NotNil(t, createdAgent)
	assert.NotEmpty(t, createdAgent.ID)
	assert.Equal(t, "agent@test.com", createdAgent.Email)

	mockAgentRepo.AssertExpectations(t)
}

func TestCreateAgent_AlreadyExists(t *testing.T) {
	// GIVEN
	mockAgentRepo := new(mocks.AgentRepositoryMock)
	mockPropRepo := new(mocks.PropertyRepositoryMock)
	svc := service.NewAgentService(mockAgentRepo, mockPropRepo)
	ctx := context.TODO()

	agent := &entity.Agent{Email: "existing@test.com"}
	existingAgent := &entity.Agent{ID: "1", Email: "existing@test.com"}

	mockAgentRepo.On("FindByEmail", ctx, "existing@test.com").Return(existingAgent, nil)

	// WHEN
	result, created, err := svc.Create(ctx, agent)

	// THEN
	assert.NoError(t, err)
	assert.False(t, created)
	assert.Equal(t, existingAgent, result)
	mockAgentRepo.AssertNotCalled(t, "Create")
}

func TestAssociateProperties(t *testing.T) {
	// GIVEN
	mockAgentRepo := new(mocks.AgentRepositoryMock)
	mockPropRepo := new(mocks.PropertyRepositoryMock)
	svc := service.NewAgentService(mockAgentRepo, mockPropRepo)

	prop1 := &entity.Property{ID: "p1"}
	prop2 := &entity.Property{ID: "p2"}
	agent := &entity.Agent{
		Properties: []*entity.Property{
			{ID: "p1"},
			{ID: "p2"},
			{ID: "p3"}, // This one doesn't exist
		},
	}

	mockPropRepo.On("FindByID", "p1").Return(prop1, nil)
	mockPropRepo.On("FindByID", "p2").Return(prop2, nil)
	mockPropRepo.On("FindByID", "p3").Return(nil, nil)

	// WHEN
	err := svc.AssociateProperties(agent)

	// THEN
	assert.NoError(t, err)
	assert.Len(t, agent.Properties, 2)
	assert.Equal(t, "p1", agent.Properties[0].ID)
	assert.Equal(t, "p2", agent.Properties[1].ID)
}
