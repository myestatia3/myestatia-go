package test

import (
	"context"
	"errors"
	"testing"

	"github.com/myestatia/myestatia-go/internal/application/service"
	"github.com/myestatia/myestatia-go/internal/domain/entity"
	"github.com/myestatia/myestatia-go/internal/domain/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateCompany(t *testing.T) {
	// GIVEN
	mockRepo := new(mocks.CompanyRepositoryMock)
	svc := service.NewCompanyService(mockRepo)
	ctx := context.TODO()

	company := &entity.Company{
		Name: "Test Company",
	}

	// Mocking behavior
	mockRepo.On("FindByName", ctx, "Test Company").Return(nil, nil)
	mockRepo.On("Create", mock.MatchedBy(func(c *entity.Company) bool {
		return c.Name == "Test Company" && c.ID != ""
	})).Return(nil)

	// WHEN
	createdCompany, created, err := svc.Create(ctx, company)

	// THEN
	assert.NoError(t, err)
	assert.True(t, created)
	assert.NotNil(t, createdCompany)
	assert.NotEmpty(t, createdCompany.ID)
	assert.Equal(t, "Test Company", createdCompany.Name)

	mockRepo.AssertExpectations(t)
}

func TestCreateCompany_AlreadyExists(t *testing.T) {
	// GIVEN
	mockRepo := new(mocks.CompanyRepositoryMock)
	svc := service.NewCompanyService(mockRepo)
	ctx := context.TODO()

	company := &entity.Company{Name: "Existing Company"}
	existingCompany := &entity.Company{ID: "1", Name: "Existing Company"}

	mockRepo.On("FindByName", ctx, "Existing Company").Return(existingCompany, nil)

	// WHEN
	result, created, err := svc.Create(ctx, company)

	// THEN
	assert.NoError(t, err)
	assert.False(t, created)
	assert.Equal(t, existingCompany, result)
	mockRepo.AssertNotCalled(t, "Create")
}

func TestCreateCompany_Error(t *testing.T) {
	// GIVEN
	mockRepo := new(mocks.CompanyRepositoryMock)
	svc := service.NewCompanyService(mockRepo)
	ctx := context.TODO()

	company := &entity.Company{Name: "Error Company"}

	mockRepo.On("FindByName", ctx, "Error Company").Return(nil, errors.New("db error"))

	// WHEN
	_, _, err := svc.Create(ctx, company)

	// THEN
	assert.Error(t, err)
	assert.Equal(t, "db error", err.Error())
}
