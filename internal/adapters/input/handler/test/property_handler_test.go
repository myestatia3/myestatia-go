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

func TestCreateProperty_Handler(t *testing.T) {
	// GIVEN
	mockRepo := new(mocks.PropertyRepositoryMock)
	svc := service.NewPropertyService(mockRepo)
	h := handler.NewPropertyHandler(svc)

	propReq := entity.Property{
		Reference: "REF-HTTP",
		CompanyID: "C1",
		Title:     "Handler Property",
	}
	body, _ := json.Marshal(propReq)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/properties", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	mockRepo.On("FindByReference", mock.Anything, "REF-HTTP").Return(nil, nil)
	mockRepo.On("Create", mock.Anything).Return(nil)

	// WHEN
	h.CreateProperty(rr, req)

	// THEN
	assert.Equal(t, http.StatusCreated, rr.Code)

	// The handler returns a map {"created": true, "property": {...}}
	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)

	propData := response["property"].(map[string]interface{})
	assert.Equal(t, "REF-HTTP", propData["reference"])
	assert.True(t, response["created"].(bool))
}

func TestGetPropertyByID_Handler_Success(t *testing.T) {
	// GIVEN
	mockRepo := new(mocks.PropertyRepositoryMock)
	svc := service.NewPropertyService(mockRepo)
	h := handler.NewPropertyHandler(svc)

	propID := "P1"
	prop := &entity.Property{ID: propID, Reference: "REF1"}

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/properties/"+propID, nil)
	rr := httptest.NewRecorder()

	mockRepo.On("FindByID", propID).Return(prop, nil)

	// WHEN
	h.GetPropertyByID(rr, req)

	// THEN
	assert.Equal(t, http.StatusOK, rr.Code)

	var response entity.Property
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Equal(t, propID, response.ID)
}
