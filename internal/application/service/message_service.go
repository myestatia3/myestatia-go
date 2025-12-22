package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/myestatia/myestatia-go/internal/domain/entity"
	"github.com/myestatia/myestatia-go/internal/infrastructure/repository"
)

type MessageService struct {
	Repo repository.MessageRepository
}

func NewMessageService(repo repository.MessageRepository) *MessageService {
	return &MessageService{Repo: repo}
}

func (s *MessageService) GetMessagesByLeadID(ctx context.Context, leadID string) ([]entity.Message, error) {
	return s.Repo.FindByLeadID(leadID)
}

func (s *MessageService) CreateMessage(ctx context.Context, leadID string, senderType string, content string) (*entity.Message, error) {
	if leadID == "" {
		return nil, errors.New("leadID is required")
	}
	if content == "" {
		return nil, errors.New("content is required")
	}

	msg := &entity.Message{
		ID:         uuid.New().String(),
		LeadID:     leadID,
		SenderType: entity.SenderType(senderType),
		Content:    content,
		Timestamp:  time.Now(),
		CreatedAt:  time.Now(),
	}

	if err := s.Repo.Create(msg); err != nil {
		return nil, err
	}

	return msg, nil
}
