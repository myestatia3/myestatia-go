package entity

import (
	"time"

	"gorm.io/gorm"
)

type Company struct {
	ID             string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name           string `gorm:"not null" json:"name"`
	Address        string `json:"address"`
	PostalCode     string `json:"postal_code"`
	City           string `gorm:"not null" json:"city"`
	Province       string `json:"province"`
	Country        string `json:"country"`
	OfficeLocation string `json:"office_location"`
	ContactPerson  string `json:"contact_person"`
	Email1         string `gorm:"not null" json:"email1"`
	Email2         string `json:"email2"`
	Phone1         string `json:"phone1"`
	Phone2         string `json:"phone2"`
	Website        string `json:"website"`
	PageLink       string `json:"page_link"`
	WebDeveloper   string `json:"web_developer"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
