package port

import "github.com/myestatia/myestatia-go/internal/domain/entity"

// EmailParser defines the interface for parsing email content
// Different implementations handle different portal formats (Strategy Pattern)
type EmailParser interface {
	CanParse(subject, from string) bool

	Parse(subject, body string) (*entity.ParsedLead, error)
}
