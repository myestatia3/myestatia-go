package entity

import (
	"time"

	"gorm.io/gorm"
)

type Agent struct {
	ID    string `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Name  string `gorm:"not null" json:"name"`
	Email string `gorm:"not null" json:"email"`
	Phone string `json:"phone"`

	CompanyID string   `gorm:"type:uuid;not null;index" json:"company_id"`
	Company   *Company `gorm:"foreignKey:CompanyID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"company,omitempty"`

	Properties []*Property `gorm:"many2many:agent_properties;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"properties,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
