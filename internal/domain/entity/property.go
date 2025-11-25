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
	ID                 string   `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Reference          string   `gorm:"uniqueIndex;not null"`
	CompanyID          string   `gorm:"not null;type:uuid;index" json:"company_id"`
	Company            *Company `gorm:"foreignKey:CompanyID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"company,omitempty"`
	CreatedByAgentID   *string
	Origin             PropertyOrigin `gorm:"type:varchar(20);not null"`
	Status             string         `gorm:"type:varchar(20)"`
	Title              string
	Description        string
	Type               PropertyType `gorm:"type:varchar(20)"`
	SubtypeID          *string
	Country            string
	Province           string
	City               string
	Address            string
	Lat                float64
	Lon                float64
	AreaM2             float64
	Rooms              int
	Bathrooms          int
	Price              float64
	Currency           string
	Floor              *int
	Features           datatypes.JSON
	Photos             datatypes.JSON
	SharedWithNetwork  bool
	PublishedOnPortals datatypes.JSON
	Metadata           datatypes.JSON

	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	PublishedAt *time.Time     `json:"published_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}
