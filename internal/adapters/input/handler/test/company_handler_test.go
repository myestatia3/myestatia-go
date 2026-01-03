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

func TestCreateCompany_Handler(t *testing.T) {
	// GIVEN
	mockRepo := new(mocks.CompanyRepositoryMock)
	svc := service.NewCompanyService(mockRepo)
	h := handler.NewCompanyHandler(svc)

	companyReq := entity.Company{
		Name:   "Handler Corp",
		City:   "Madrid",
		Email1: "handler@corp.com",
	}
	body, _ := json.Marshal(companyReq)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/companies", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	// Mocking service dependency (the repo)
	mockRepo.On("FindByName", mock.Anything, "Handler Corp").Return(nil, nil)
	mockRepo.On("Create", mock.Anything).Return(nil)

	// WHEN
	h.CreateCompany(rr, req)

	// THEN
	assert.Equal(t, http.StatusCreated, rr.Code)

	var response entity.Company
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Handler Corp", response.Name)

	mockRepo.AssertExpectations(t)
}

func TestGetCompanyByID_Handler_NotFound(t *testing.T) {
	// GIVEN
	mockRepo := new(mocks.CompanyRepositoryMock)
	svc := service.NewCompanyService(mockRepo)
	h := handler.NewCompanyHandler(svc)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/companies/non-existent", nil)
	rr := httptest.NewRecorder()

	mockRepo.On("FindByID", "non-existent").Return(nil, nil)

	// WHEN
	h.GetCompanyByID(rr, req)

	// THEN
	assert.Equal(t, http.StatusNotFound, rr.Code)
}
