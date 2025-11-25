package entity

import (
	"time"

	"gorm.io/gorm"
)

type SenderType string

const (
	SenderLead  SenderType = "lead"
	SenderAgent SenderType = "agent"
	SenderAI    SenderType = "ai"
)

type Message struct {
	ID         string     `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	LeadID     string     `gorm:"index;not null"`
	SenderType SenderType `gorm:"type:varchar(10);not null"`
	Content    string     `gorm:"type:text;not null"`
	Timestamp  time.Time  `gorm:"not null"`

	Lead Lead `gorm:"foreignKey:LeadID"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
