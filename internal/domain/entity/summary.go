package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Summary: conversation summary for a lead
type Summary struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	LeadID      uuid.UUID      `gorm:"type:uuid;uniqueIndex" json:"lead_id"`
	LastUpdated time.Time      `json:"last_updated"`
	SummaryText string         `gorm:"type:text" json:"summaryText"`
	NextAction  *string        `json:"next_action,omitempty"`
	Lead        Lead           `gorm:"foreignKey:LeadID"`
	Metadata    datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}
