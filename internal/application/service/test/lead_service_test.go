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

func TestCreateLead(t *testing.T) {
	// GIVEN
	mockRepo := new(mocks.LeadRepositoryMock)
	svc := service.NewLeadService(mockRepo)
	ctx := context.TODO()

	lead := &entity.Lead{
		Email: "lead@test.com",
		Name:  "Test Lead",
	}

	// Mocking behavior
	mockRepo.On("FindByEmail", ctx, "lead@test.com").Return(nil, nil)
	mockRepo.On("Create", mock.MatchedBy(func(l *entity.Lead) bool {
		return l.Email == "lead@test.com" && l.ID != ""
	})).Return(nil)

	// WHEN
	createdLead, created, err := svc.Create(ctx, lead)

	// THEN
	assert.NoError(t, err)
	assert.True(t, created)
	assert.NotNil(t, createdLead)
	assert.NotEmpty(t, createdLead.ID)
	assert.Equal(t, "lead@test.com", createdLead.Email)

	mockRepo.AssertExpectations(t)
}

func TestCreateLead_AlreadyExists(t *testing.T) {
	// GIVEN
	mockRepo := new(mocks.LeadRepositoryMock)
	svc := service.NewLeadService(mockRepo)
	ctx := context.TODO()

	lead := &entity.Lead{Email: "existing@lead.com"}
	existingLead := &entity.Lead{ID: "L1", Email: "existing@lead.com"}

	mockRepo.On("FindByEmail", ctx, "existing@lead.com").Return(existingLead, nil)

	// WHEN
	result, created, err := svc.Create(ctx, lead)

	// THEN
	assert.NoError(t, err)
	assert.False(t, created)
	assert.Equal(t, existingLead, result)
	mockRepo.AssertNotCalled(t, "Create")
}
