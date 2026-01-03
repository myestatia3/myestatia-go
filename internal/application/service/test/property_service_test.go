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

func TestCreateProperty(t *testing.T) {
	// GIVEN
	mockRepo := new(mocks.PropertyRepositoryMock)
	svc := service.NewPropertyService(mockRepo)
	ctx := context.TODO()

	prop := &entity.Property{
		Reference: "REF123",
		CompanyID: "C1",
		Title:     "Luxury Villa",
	}

	// Mocking behavior
	mockRepo.On("FindByReference", ctx, "REF123").Return(nil, nil)
	mockRepo.On("Create", mock.MatchedBy(func(p *entity.Property) bool {
		return p.Reference == "REF123" && p.ID != ""
	})).Return(nil)

	// WHEN
	createdProp, created, err := svc.CreateProperty(ctx, prop)

	// THEN
	assert.NoError(t, err)
	assert.True(t, created)
	assert.NotNil(t, createdProp)
	assert.Equal(t, "REF123", createdProp.Reference)

	mockRepo.AssertExpectations(t)
}

func TestGetPropertyByID(t *testing.T) {
	// GIVEN
	mockRepo := new(mocks.PropertyRepositoryMock)
	svc := service.NewPropertyService(mockRepo)
	ctx := context.TODO()
	prop := &entity.Property{ID: "P1", Title: "Test Prop"}

	mockRepo.On("FindByID", "P1").Return(prop, nil)

	// WHEN
	result, err := svc.GetPropertyByID(ctx, "P1")

	// THEN
	assert.NoError(t, err)
	assert.Equal(t, prop, result)
}
