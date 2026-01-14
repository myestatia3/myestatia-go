package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProcessedEmail represents an email that has already been processed by the worker.
// This prevents re-processing the same email multiple times without modifying the source inbox state.
type ProcessedEmail struct {
	ID          string    `gorm:"type:uuid;primary_key;" json:"id"`
	MessageID   string    `gorm:"type:varchar(255);index;not null" json:"messageId"` // Unique ID from email provider
	CompanyID   string    `gorm:"type:uuid;index;not null" json:"companyId"`
	Company     Company   `gorm:"foreignKey:CompanyID" json:"-"`
	ProcessedAt time.Time `json:"processedAt"`
}

func (pe *ProcessedEmail) BeforeCreate(tx *gorm.DB) (err error) {
	if pe.ID == "" {
		pe.ID = uuid.New().String()
	}
	if pe.ProcessedAt.IsZero() {
		pe.ProcessedAt = time.Now()
	}
	return
}
