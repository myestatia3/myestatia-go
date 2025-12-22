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
	ID              string     `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Name            string     `gorm:"not null" json:"name"`
	Email           string     `gorm:"uniqueIndex;not null" json:"email"`
	Phone           string     `gorm:"not null" json:"phone"`
	Status          LeadStatus `gorm:"type:varchar(20);default:'new'" json:"status"`
	PropertyID      *string    `gorm:"type:uuid;index" json:"propertyId"`
	Property        *Property  `gorm:"foreignKey:PropertyID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"property,omitempty"`
	CompanyID       string     `gorm:"not null;type:uuid;index" json:"companyId"`
	Company         *Company   `gorm:"foreignKey:CompanyID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"company,omitempty"`
	AssignedAgentID *string    `json:"assignedAgentId"`
	LastInteraction *time.Time `json:"lastInteraction"`

	// New fields for Frontend integration
	Language                 string  `gorm:"type:varchar(10);default:'es'" json:"language"`
	Source                   string  `gorm:"type:varchar(50)" json:"source"`
	Budget                   float64 `json:"budget"`
	Zone                     string  `json:"zone"`
	PropertyType             string  `json:"propertyType"`
	Channel                  string  `json:"channel"`
	SuggestedPropertiesCount int     `json:"suggestedPropertiesCount"`

	Messages  []Message `gorm:"foreignKey:LeadID" json:"messages,omitempty"`
	Summaries []Summary `gorm:"foreignKey:LeadID" json:"summaries,omitempty"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
}
