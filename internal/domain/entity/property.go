package entity

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type PropertyOrigin string
type PropertyType string

const (
	OriginManual     PropertyOrigin = "MANUAL"
	OriginAIEmail    PropertyOrigin = "AI_EMAIL"
	OriginAIWhatsapp PropertyOrigin = "AI_WHATSAPP"
	OriginAISpeech   PropertyOrigin = "AI_SPEECH"
	OriginImport     PropertyOrigin = "PORTAL_IMPORT"

	TypeApartment  PropertyType = "APARTMENT"
	TypeHouse      PropertyType = "HOUSE"
	TypeLand       PropertyType = "LAND"
	TypeCommercial PropertyType = "COMMERCIAL"
	TypeOther      PropertyType = "OTHER"
)

type PropertyFilter struct {
	MinRooms   *int
	MaxRooms   *int
	MinPrice   *float64
	MaxPrice   *float64
	HasParking *bool
	MinAreaM2  *int
	MaxAreaM2  *int
	Province   *string
	Address    *string

	Limit  int
	Offset int
}

type Property struct {
	ID               string         `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Reference        string         `gorm:"uniqueIndex;not null" json:"reference"`
	CompanyID        string         `gorm:"not null;type:uuid;index" json:"companyId"`
	Company          *Company       `gorm:"foreignKey:CompanyID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"company,omitempty"`
	CreatedByAgentID *string        `json:"createdByAgentId"`
	Origin           PropertyOrigin `gorm:"type:varchar(20);not null" json:"origin"`
	Status           string         `gorm:"type:varchar(20)" json:"status"`
	Title            string         `json:"title"`
	Description      string         `json:"description"`
	Type             PropertyType   `gorm:"type:varchar(20)" json:"type"`
	SubtypeID        *string        `json:"subtypeId"`
	Country          string         `json:"country"`
	Province         string         `json:"province"`
	City             string         `json:"city"`
	Address          string         `json:"address"`
	Lat              float64        `json:"lat"`
	Lon              float64        `json:"lon"`
	AreaM2           float64        `json:"area"`
	Rooms            int            `json:"rooms"`
	Bathrooms        int            `json:"bathrooms"`
	Price            float64        `json:"price"`
	Currency         string         `json:"currency"`
	Floor            *int           `json:"floor"`

	// Frontend Integration Fields
	Image                string `json:"image"`
	Zone                 string `json:"zone"`
	IsNew                bool   `json:"isNew"`
	CompatibleLeadsCount int    `json:"compatibleLeadsCount"`

	Features           datatypes.JSON `json:"features"`
	Photos             datatypes.JSON `json:"photos"`
	SharedWithNetwork  bool           `json:"sharedWithNetwork"`
	PublishedOnPortals datatypes.JSON `json:"publishedOnPortals"`
	Metadata           datatypes.JSON `json:"metadata"`

	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	PublishedAt *time.Time     `json:"publishedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
}
