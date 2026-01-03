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

func TestCreateLead_Handler(t *testing.T) {
	// GIVEN
	mockRepo := new(mocks.LeadRepositoryMock)
	svc := service.NewLeadService(mockRepo)
	h := handler.NewLeadHandler(svc)

	leadReq := entity.Lead{
		Name:  "Test Lead",
		Email: "lead@handler.com",
		Phone: "999999",
	}
	body, _ := json.Marshal(leadReq)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/leads", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	mockRepo.On("FindByEmail", mock.Anything, "lead@handler.com").Return(nil, nil)
	mockRepo.On("Create", mock.Anything).Return(nil)

	// WHEN
	h.CreateLead(rr, req)

	// THEN
	assert.Equal(t, http.StatusCreated, rr.Code)

	var response entity.Lead
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Equal(t, "Test Lead", response.Name)

	mockRepo.AssertExpectations(t)
}

func TestGetLeadByID_Handler_Success(t *testing.T) {
	// GIVEN
	mockRepo := new(mocks.LeadRepositoryMock)
	svc := service.NewLeadService(mockRepo)
	h := handler.NewLeadHandler(svc)

	leadID := "L1"
	lead := &entity.Lead{ID: leadID, Name: "Lead One"}

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/leads/"+leadID, nil)
	rr := httptest.NewRecorder()

	// LeadRepository.FindByID takes 1 argument
	mockRepo.On("FindByID", leadID).Return(lead, nil)

	// WHEN
	h.GetLeadByID(rr, req)

	// THEN
	assert.Equal(t, http.StatusOK, rr.Code)

	var response entity.Lead
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Equal(t, leadID, response.ID)
}
