package repository

import (
	"github.com/myestatia/myestatia-go/internal/domain/entity"
	"gorm.io/gorm"
)

type MessageRepository interface {
	Create(message *entity.Message) error
	FindByLeadID(leadID string) ([]entity.Message, error)
}

type messageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(message *entity.Message) error {
	return r.db.Create(message).Error
}

func (r *messageRepository) FindByLeadID(leadID string) ([]entity.Message, error) {
	var messages []entity.Message
	if err := r.db.Where("lead_id = ?", leadID).Order("timestamp asc").Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}
