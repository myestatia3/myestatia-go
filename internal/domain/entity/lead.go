package entity

import (
	"time"

	"gorm.io/gorm"
)

type LeadStatus string

const (
	LeadStatusNew       LeadStatus = "new"
	LeadStatusQualified LeadStatus = "qualified"
	LeadStatusContacted LeadStatus = "contacted"
	LeadStatusClosed    LeadStatus = "closed"
)

type Lead struct {
	ID              string     `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name            string     `gorm:"not null"`
	Email           string     `gorm:"uniqueIndex;not null"`
	Phone           string     `gorm:"not null"`
	Status          LeadStatus `gorm:"type:varchar(20);default:'new'"`
	PropertyID      string     `gorm:"not null;type:uuid;index" json:"property_id"`
	Property        *Property  `gorm:"foreignKey:PropertyID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"property,omitempty"`
	CompanyID       string     `gorm:"not null;type:uuid;index" json:"company_id"`
	Company         *Company   `gorm:"foreignKey:CompanyID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"company,omitempty"`
	AssignedAgentID *string
	LastInteraction *time.Time

	Messages  []Message `gorm:"foreignKey:LeadID"`
	Summaries []Summary `gorm:"foreignKey:LeadID"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
