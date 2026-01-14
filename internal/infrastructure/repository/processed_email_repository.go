package repository

import (
	"context"
	"time"

	"github.com/myestatia/myestatia-go/internal/domain/entity"
	"gorm.io/gorm"
)

type ProcessedEmailRepository interface {
	Exists(ctx context.Context, messageID string, companyID string) (bool, error)
	Save(ctx context.Context, processedEmail *entity.ProcessedEmail) error
	CleanupOldEmails(ctx context.Context) error
}

type processedEmailRepository struct {
	db *gorm.DB
}

func NewProcessedEmailRepository(db *gorm.DB) ProcessedEmailRepository {
	return &processedEmailRepository{db: db}
}

func (r *processedEmailRepository) Exists(ctx context.Context, messageID string, companyID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.ProcessedEmail{}).
		Where("message_id = ? AND company_id = ?", messageID, companyID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *processedEmailRepository) Save(ctx context.Context, processedEmail *entity.ProcessedEmail) error {
	return r.db.WithContext(ctx).Create(processedEmail).Error
}

func (r *processedEmailRepository) CleanupOldEmails(ctx context.Context) error {
	// Delete processed emails older than 30 days
	// This prevents the table from growing indefinitely
	cutoff := time.Now().AddDate(0, 0, -30)
	return r.db.WithContext(ctx).
		Where("processed_at < ?", cutoff).
		Delete(&entity.ProcessedEmail{}).Error
}
