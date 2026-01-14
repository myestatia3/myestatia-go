package entity

import (
	"time"

	"gorm.io/gorm"
)

// CompanyEmailConfig stores email configuration for a company's email inbox
// Supports both IMAP password auth and OAuth2 (Google)
type CompanyEmailConfig struct {
	ID        string   `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CompanyID string   `gorm:"not null;type:uuid;uniqueIndex;index" json:"companyId"`
	Company   *Company `gorm:"foreignKey:CompanyID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"company,omitempty"`

	// Auth method: "password" or "oauth2"
	AuthMethod string `gorm:"type:varchar(20);default:'password'" json:"authMethod"`

	// IMAP fields (for password auth)
	IMAPHost     string `gorm:"type:varchar(255)" json:"imapHost"`
	IMAPPort     int    `gorm:"default:993" json:"imapPort"`
	IMAPUsername string `gorm:"type:varchar(255)" json:"imapUsername"`
	IMAPPassword string `gorm:"type:text" json:"-"` // Encrypted, never returned in JSON

	// OAuth2 fields (for oauth2 auth)
	OAuth2Provider string     `gorm:"type:varchar(50)" json:"oauth2Provider"` // "google"
	AccessToken    string     `gorm:"type:text" json:"-"`                     // Encrypted
	RefreshToken   string     `gorm:"type:text" json:"-"`                     // Encrypted
	TokenExpiry    *time.Time `json:"tokenExpiry"`

	// Common fields
	InboxFolder      string         `gorm:"type:varchar(100);default:'INBOX'" json:"inboxFolder"`
	PollIntervalSecs int            `gorm:"default:300" json:"pollIntervalSecs"` // 5 minutes default
	IsEnabled        bool           `gorm:"default:true;index" json:"isEnabled"`
	LastSyncAt       *time.Time     `json:"lastSyncAt"`
	CreatedAt        time.Time      `json:"createdAt"`
	UpdatedAt        time.Time      `json:"updatedAt"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
}
