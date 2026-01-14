package entity

import "time"

type EmailSource string

const (
	EmailSourceFotocasa  EmailSource = "Fotocasa"
	EmailSourceIdealista EmailSource = "Idealista"
)

// ParsedLead represents the intermediate structure after parsing an email
type ParsedLead struct {
	Name              string
	Email             string
	Phone             string
	Message           string
	PropertyReference string      // e.g., "R4786633"
	Source            EmailSource // Fotocasa or Idealista
	ContactDate       time.Time
}
