package entity

type PropertySubtype struct {
	ID     string       `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name   string       `gorm:"uniqueIndex;not null"`
	Type   PropertyType `gorm:"type:varchar(20);not null"`
	Active bool         `gorm:"default:true"`
}
