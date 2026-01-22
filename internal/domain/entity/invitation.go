package entity

import (
	"time"

	"gorm.io/gorm"
)

type Invitation struct {
	ID        string         `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Token     string         `gorm:"uniqueIndex;not null" json:"token"`
	Email     string         `gorm:"not null" json:"email"`
	CompanyID string         `gorm:"type:uuid;not null;index" json:"company_id"`
	Company   *Company       `gorm:"foreignKey:CompanyID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"company,omitempty"`
	Used      bool           `gorm:"default:false" json:"used"`
	UsedAt    *time.Time     `json:"used_at,omitempty"`
	ExpiresAt time.Time      `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (i *Invitation) IsExpired() bool {
	return time.Now().After(i.ExpiresAt)
}

func (i *Invitation) IsValid() bool {
	return !i.Used && !i.IsExpired()
}
