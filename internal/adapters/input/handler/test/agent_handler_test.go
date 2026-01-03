package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/myestatia/myestatia-go/internal/adapters/input/handler"
	"github.com/myestatia/myestatia-go/internal/application/service"
	"github.com/myestatia/myestatia-go/internal/domain/entity"
	"github.com/myestatia/myestatia-go/internal/domain/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateAgent_Handler(t *testing.T) {
	// GIVEN
	mockAgentRepo := new(mocks.AgentRepositoryMock)
	mockPropRepo := new(mocks.PropertyRepositoryMock)
	svc := service.NewAgentService(mockAgentRepo, mockPropRepo)
	h := handler.NewAgentHandler(svc)

	agentReq := entity.Agent{
		Name:  "Test Handler Agent",
		Email: "agent@handler.com",
		Phone: "123456",
	}
	body, _ := json.Marshal(agentReq)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/agents", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	mockAgentRepo.On("FindByEmail", mock.Anything, "agent@handler.com").Return(nil, nil)
	mockAgentRepo.On("Create", mock.Anything).Return(nil)

	// WHEN
	h.CreateAgent(rr, req)

	// THEN
	assert.Equal(t, http.StatusCreated, rr.Code)

	var response entity.Agent
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Equal(t, "Test Handler Agent", response.Name)

	mockAgentRepo.AssertExpectations(t)
}

func TestGetAgentByID_Handler_Success(t *testing.T) {
	// GIVEN
	mockAgentRepo := new(mocks.AgentRepositoryMock)
	mockPropRepo := new(mocks.PropertyRepositoryMock)
	svc := service.NewAgentService(mockAgentRepo, mockPropRepo)
	h := handler.NewAgentHandler(svc)

	agentID := "a1"
	agent := &entity.Agent{ID: agentID, Name: "Agent One"}

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/agents/"+agentID, nil)
	rr := httptest.NewRecorder()

	// AgentRepository.FindByID only takes 1 argument: id
	mockAgentRepo.On("FindByID", agentID).Return(agent, nil)

	// WHEN
	h.GetAgentByID(rr, req)

	// THEN
	assert.Equal(t, http.StatusOK, rr.Code)

	var response entity.Agent
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Equal(t, agentID, response.ID)
}
