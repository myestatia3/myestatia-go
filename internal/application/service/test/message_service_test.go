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

func TestCreateMessage(t *testing.T) {
	// GIVEN
	mockRepo := new(mocks.MessageRepositoryMock)
	svc := service.NewMessageService(mockRepo)
	ctx := context.TODO()

	leadID := "L1"
	content := "Hello, I am interested"
	senderType := "user"

	mockRepo.On("Create", mock.MatchedBy(func(m *entity.Message) bool {
		return m.LeadID == leadID && m.Content == content
	})).Return(nil)

	// WHEN
	msg, err := svc.CreateMessage(ctx, leadID, senderType, content)

	// THEN
	assert.NoError(t, err)
	assert.NotNil(t, msg)
	assert.Equal(t, content, msg.Content)
	assert.Equal(t, entity.SenderType(senderType), msg.SenderType)

	mockRepo.AssertExpectations(t)
}

func TestGetMessagesByLeadID(t *testing.T) {
	// GIVEN
	mockRepo := new(mocks.MessageRepositoryMock)
	svc := service.NewMessageService(mockRepo)
	ctx := context.TODO()
	leadID := "L1"
	messages := []entity.Message{{ID: "M1", Content: "Hi"}}

	mockRepo.On("FindByLeadID", leadID).Return(messages, nil)

	// WHEN
	result, err := svc.GetMessagesByLeadID(ctx, leadID)

	// THEN
	assert.NoError(t, err)
	assert.Equal(t, messages, result)
}
