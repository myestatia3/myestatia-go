package entity

import (
	"time"

	"gorm.io/gorm"
)

// PasswordReset represents a password reset token
type PasswordReset struct {
	ID        string         `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID    string         `gorm:"type:uuid;not null;index" json:"userId"`
	User      *Agent         `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"user,omitempty"`
	Token     string         `gorm:"type:varchar(255);not null;uniqueIndex" json:"-"` // Secure random token, not exposed in JSON
	ExpiresAt time.Time      `gorm:"not null" json:"expiresAt"`
	Used      bool           `gorm:"default:false" json:"used"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
}

// IsExpired checks if the token has expired
func (pr *PasswordReset) IsExpired() bool {
	return time.Now().After(pr.ExpiresAt)
}

// IsValid checks if the token is valid (not used and not expired)
func (pr *PasswordReset) IsValid() bool {
	return !pr.Used && !pr.IsExpired()
}
