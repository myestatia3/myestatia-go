package email

import "time"

// ParsedEmail represents a parsed email message (common interface for IMAP and Gmail API)
type ParsedEmail struct {
	MessageID string
	From      string
	Subject   string
	Body      string
	Date      time.Time
}
