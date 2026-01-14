package email

import (
	"os"
	"strconv"
)

// Config holds IMAP email configuration
// Currently loaded from environment variables
// Architecture note: Designed to be migrated to database-driven per-agent configuration
type Config struct {
	IMAPHost         string
	IMAPPort         int
	Username         string
	Password         string
	InboxFolder      string
	PollIntervalSecs int
	DefaultCompanyID string // Mocked for now: ecf4ed64-06b5-4129-af4e-72718751e087
}

// LoadConfig loads email configuration from environment variables
func LoadConfig() Config {
	port := 993
	if portStr := os.Getenv("EMAIL_IMAP_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	pollInterval := 300 // default: 5 minutes
	if intervalStr := os.Getenv("EMAIL_POLL_INTERVAL_SECONDS"); intervalStr != "" {
		if i, err := strconv.Atoi(intervalStr); err == nil {
			pollInterval = i
		}
	}

	inbox := os.Getenv("EMAIL_INBOX_FOLDER")
	if inbox == "" {
		inbox = "INBOX"
	}

	return Config{
		IMAPHost:         os.Getenv("EMAIL_IMAP_HOST"),
		IMAPPort:         port,
		Username:         os.Getenv("EMAIL_IMAP_USERNAME"),
		Password:         os.Getenv("EMAIL_IMAP_PASSWORD"),
		InboxFolder:      inbox,
		PollIntervalSecs: pollInterval,
		DefaultCompanyID: os.Getenv("EMAIL_DEFAULT_COMPANY_ID"),
	}
}

// IsValid checks if the configuration has required fields
func (c Config) IsValid() bool {
	return c.IMAPHost != "" && c.Username != "" && c.Password != "" && c.DefaultCompanyID != ""
}
